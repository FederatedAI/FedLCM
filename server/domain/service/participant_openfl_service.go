// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"text/template"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/server/infrastructure/gorm/mock"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	imagePullSecretsNameOpenFL = "registrykeyopenfl"
)

// ParticipantOpenFLService is the service to manage openfl participants
type ParticipantOpenFLService struct {
	ParticipantOpenFLRepo repo.ParticipantOpenFLRepository
	TokenRepo             repo.RegistrationTokenRepository
	InfraRepo             repo.InfraProviderRepository
	ParticipantService
}

// ParticipantOpenFLDirectorYAMLCreationRequest contains necessary info to generate the deployment yaml
// content for KubeFATE to deploy OpenFL director
type ParticipantOpenFLDirectorYAMLCreationRequest struct {
	FederationUUID string                               `json:"federation_uuid" form:"federation_uuid"`
	ChartUUID      string                               `json:"chart_uuid" form:"chart_uuid"`
	Name           string                               `json:"name" form:"name"`
	Namespace      string                               `json:"namespace" form:"namespace"`
	ServiceType    entity.ParticipantDefaultServiceType `json:"service_type" form:"service_type"`
	// for generating the yaml, RegistrySecretConfig is not used in RegistryConfig
	RegistryConfig  valueobject.KubeRegistryConfig `json:"registry_config"`
	JupyterPassword string                         `json:"jupyter_password" form:"jupyter_password"`
	EnablePSP       bool                           `json:"enable_psp" form:"enable_psp"`
}

// ParticipantOpenFLDirectorCreationRequest is the director creation request
type ParticipantOpenFLDirectorCreationRequest struct {
	ParticipantOpenFLDirectorYAMLCreationRequest
	ParticipantDeploymentBaseInfo
	DirectorServerCertInfo entity.ParticipantComponentCertInfo `json:"director_server_cert_info"`
	JupyterClientCertInfo  entity.ParticipantComponentCertInfo `json:"jupyter_client_cert_info"`
}

// ParticipantOpenFLEnvoyRegistrationRequest is the registration request from an envoy
type ParticipantOpenFLEnvoyRegistrationRequest struct {
	// required
	KubeConfig string `json:"kubeconfig"`
	TokenStr   string `json:"token"`

	// optional
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	Namespace             string                         `json:"namespace"`
	ChartUUID             string                         `json:"chart_uuid"`
	Labels                valueobject.Labels             `json:"labels"`
	ConfigYAML            string                         `json:"config_yaml"`
	SkipCommonPythonFiles bool                           `json:"skip_common_python_files"`
	RegistryConfig        valueobject.KubeRegistryConfig `json:"registry_config"`
	EnablePSP             bool                           `json:"enable_psp"`

	// internal
	federation   *entity.FederationOpenFL
	caCert       *x509.Certificate
	operationLog *zerolog.Logger
	chart        *entity.Chart
}

// GetOpenFLDirectorYAML returns the exchange deployment yaml content
func (s *ParticipantOpenFLService) GetOpenFLDirectorYAML(req *ParticipantOpenFLDirectorYAMLCreationRequest) (string, error) {
	instance, err := s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to query chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeOpenFLDirector {
		return "", errors.Errorf("chart %s is not for OpenFL director deployment", chart.UUID)
	}

	t, err := template.New("openfl-director").Parse(chart.InitialYamlTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		Name                 string
		Namespace            string
		JupyterPassword      string
		ServiceType          string
		UseRegistry          bool
		Registry             string
		UseImagePullSecrets  bool
		ImagePullSecretsName string
		SampleShape          string
		TargetShape          string
		EnablePSP            bool
		Domain               string
	}{
		Name:                 toDeploymentName(req.Name),
		Namespace:            req.Namespace,
		JupyterPassword:      req.JupyterPassword,
		ServiceType:          req.ServiceType.String(),
		UseRegistry:          req.RegistryConfig.UseRegistry,
		Registry:             req.RegistryConfig.Registry,
		UseImagePullSecrets:  req.RegistryConfig.UseRegistrySecret,
		ImagePullSecretsName: imagePullSecretsNameOpenFL,
		EnablePSP:            req.EnablePSP,
	}

	federationUUID := req.FederationUUID
	instance, err = s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return "", errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationOpenFL)
	data.Domain = federation.Domain
	if federation.UseCustomizedShardDescriptor {
		var arr []string
		for _, item := range federation.ShardDescriptorConfig.SampleShape {
			arr = append(arr, `'`+item+`'`)
		}
		data.SampleShape = strings.Join(arr, ", ")
		data.SampleShape = `[` + data.SampleShape + `]`

		arr = []string{}
		for _, item := range federation.ShardDescriptorConfig.TargetShape {
			arr = append(arr, `'`+item+`'`)
		}
		data.TargetShape = strings.Join(arr, ", ")
		data.TargetShape = `[` + data.TargetShape + `]`
	} else {
		data.SampleShape = `['1']`
		data.TargetShape = `['1']`
	}

	if data.JupyterPassword != "" {
		b := ""
		for i := 0; i < 8; i++ {
			b += fmt.Sprintf("%v", rand.Intn(10))
		}
		checkSum := sha1.Sum([]byte(data.JupyterPassword + b))
		// jupyter's password mechanism - must contain a salt
		data.JupyterPassword = "sha1:" + b + ":" + hex.EncodeToString(checkSum[:])
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// CreateDirector creates an OpenFL director
func (s *ParticipantOpenFLService) CreateDirector(req *ParticipantOpenFLDirectorCreationRequest) (*entity.ParticipantOpenFL, *sync.WaitGroup, error) {
	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationOpenFL)

	if exist, err := s.ParticipantOpenFLRepo.IsDirectorCreatedByFederationUUID(federationUUID); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check director existence status")
	} else if exist {
		return nil, nil, errors.Errorf("a director is already deployed in federation %s", federationUUID)
	}

	if req.DirectorServerCertInfo.BindingMode == entity.CertBindingModeReuse ||
		req.JupyterClientCertInfo.BindingMode == entity.CertBindingModeReuse {
		return nil, nil, errors.New("cannot re-use existing certificate")
	}

	if err := s.EndpointService.TestKubeFATE(req.EndpointUUID); err != nil {
		return nil, nil, err
	}

	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeOpenFLDirector {
		return nil, nil, errors.Errorf("chart %s is not for OpenFL director deployment", chart.UUID)
	}

	var m map[string]interface{}
	err = yaml.Unmarshal([]byte(req.DeploymentYAML), &m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}

	m["name"] = toDeploymentName(req.Name)
	m["namespace"] = req.Namespace

	finalYAMLBytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get final yaml content")
	}
	req.DeploymentYAML = string(finalYAMLBytes)
	log.Debug().Msgf("openfl director deployment yaml: %s", req.DeploymentYAML)

	var caCert *x509.Certificate
	if req.DirectorServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.JupyterClientCertInfo.BindingMode == entity.CertBindingModeCreate {
		log.Info().Msg("preparing CA for issuing certificate for openfl director deployment")
		ca, err := s.CertificateService.DefaultCA()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get default CA")
		}
		caCert, err = ca.RootCert()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get CA cert")
		}
	}

	director := &entity.ParticipantOpenFL{
		Participant: entity.Participant{
			UUID:           uuid.NewV4().String(),
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.FederationUUID,
			EndpointUUID:   req.EndpointUUID,
			ChartUUID:      req.ChartUUID,
			Namespace:      req.Namespace,
			DeploymentYAML: req.DeploymentYAML,
			IsManaged:      true,
			ExtraAttribute: entity.ParticipantExtraAttribute{
				IsNewNamespace:    false,
				UseRegistrySecret: req.RegistryConfig.UseRegistrySecret,
			},
		},
		Type:   entity.ParticipantOpenFLTypeDirector,
		Status: entity.ParticipantOpenFLStatusInstallingDirector,
		CertConfig: entity.ParticipantOpenFLCertConfig{
			DirectorServerCertInfo: req.DirectorServerCertInfo,
			JupyterClientCertInfo:  req.JupyterClientCertInfo,
		},
		AccessInfo: entity.ParticipantOpenFLModulesAccessMap{},
	}
	err = s.ParticipantOpenFLRepo.Create(director)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLDirector, director.UUID, "start creating director", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "installing openfl director").Str("uuid", director.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLDirector, director.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("creating openfl director %s with UUID %s", req.Name, director.UUID)
		if err := func() error {
			endpointMgr, kfClient, clientCloser, err := s.buildKubeFATEMgrAndClient(req.EndpointUUID)
			if clientCloser != nil {
				defer clientCloser()
			}
			if err != nil {
				return err
			}
			if chart.Private {
				if err := kfClient.EnsureChartExist(chart.ChartName, chart.Version, chart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			if director.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
				return err
			}
			if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
				return errors.Wrap(err, "failed to update exchange attribute")
			}

			//Create registry secret
			if req.RegistryConfig.UseRegistrySecret {
				if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameOpenFL, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
					return errors.Wrap(err, "failed to create registry secret")
				} else {
					operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameOpenFL, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
				}
			}

			directorFQDN := req.DirectorServerCertInfo.CommonName
			if directorFQDN == "" {
				directorFQDN = fmt.Sprintf("director.%s", federation.Domain)
			}
			if req.DirectorServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.DirectorServerCertInfo.CommonName == "" {
					req.DirectorServerCertInfo.CommonName = directorFQDN
				}
				operationLog.Info().Msgf("creating certificate for the director server with CN: %s", directorFQDN)
				cert, pk, err := s.CertificateService.CreateCertificateSimple(directorFQDN, defaultCertLifetime, []string{directorFQDN, "director"})
				if err != nil {
					return errors.Wrapf(err, "failed to create director certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createDirectorSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypeOpenFLDirector, director.UUID, federationUUID, entity.FederationTypeOpenFL); err != nil {
					return err
				}
				director.CertConfig.DirectorServerCertInfo.CommonName = req.DirectorServerCertInfo.CommonName
				director.CertConfig.DirectorServerCertInfo.UUID = cert.UUID
				if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
					return errors.Wrap(err, "failed to update director cert info")
				}
			}

			jupyterCN := req.JupyterClientCertInfo.CommonName
			if jupyterCN == "" {
				jupyterCN = fmt.Sprintf("jupyter.%s", federation.Domain)
			}
			if req.JupyterClientCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.JupyterClientCertInfo.CommonName == "" {
					req.JupyterClientCertInfo.CommonName = jupyterCN
				}
				operationLog.Info().Msgf("creating certificate for the jupyter client with CN: %s", req.JupyterClientCertInfo.CommonName)
				cert, pk, err := s.CertificateService.CreateCertificateSimple(jupyterCN, defaultCertLifetime, []string{jupyterCN})
				if err != nil {
					return errors.Wrapf(err, "failed to create jupyter certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createJupyterSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypeOpenFLJupyter, director.UUID, federationUUID, entity.FederationTypeOpenFL); err != nil {
					return err
				}
				director.CertConfig.JupyterClientCertInfo.CommonName = req.JupyterClientCertInfo.CommonName
				director.CertConfig.JupyterClientCertInfo.UUID = cert.UUID
				if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
					return errors.Wrap(err, "failed to update jupyter cert info")
				}
			}

			jobUUID, err := kfClient.SubmitClusterInstallationJob(director.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster creation request")
			}
			operationLog.Info().Msgf("director cluster installing job UUID: %s", jobUUID)
			director.JobUUID = jobUUID
			if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
				return errors.Wrap(err, "failed to update director's job uuid")
			}
			clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
			if err != nil {
				return errors.Wrapf(err, "fail to get cluster uuid")
			}
			director.ClusterUUID = clusterUUID
			if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
				return errors.Wrap(err, "failed to update director cluster uuid")
			}
			//wait for job finished
			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}

			serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantOpenFLServiceNameDirector), "director")
			if err != nil {
				return errors.Wrapf(err, "fail to get director api access info")
			}
			director.AccessInfo[entity.ParticipantOpenFLServiceNameDirector] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         true,
				FQDN:        directorFQDN,
			}

			serviceType, host, port, err = getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantOpenFLServiceNameDirector), "agg")
			if err != nil {
				return errors.Wrapf(err, "fail to get director agg access info")
			}
			director.AccessInfo[entity.ParticipantOpenFLServiceNameAggregator] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         true,
				FQDN:        directorFQDN,
			}

			serviceType, host, port, err = getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantOpenFLServiceNameJupyter), "notebook")
			if err != nil {
				return errors.Wrapf(err, "fail to get jupyter access info")
			}
			director.AccessInfo[entity.ParticipantOpenFLServiceNameJupyter] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         false,
			}

			director.Status = entity.ParticipantOpenFLStatusActive
			return s.ParticipantOpenFLRepo.UpdateInfoByUUID(director)
		}(); err != nil {
			operationLog.Error().Msgf("failed to install openfl director, error: %v", err)
			director.Status = entity.ParticipantOpenFLStatusFailed
			if updateErr := s.ParticipantOpenFLRepo.UpdateStatusByUUID(director); updateErr != nil {
				operationLog.Error().Msgf("failed to update openfl director status, error: %v", updateErr)
			}
			return
		}
		operationLog.Info().Msgf("openfl director %s(%s) deployed", director.Name, director.UUID)
	}()

	return director, wg, nil
}

// RemoveDirector removes and uninstalls an OpenFL director
func (s *ParticipantOpenFLService) RemoveDirector(uuid string, force bool) (*sync.WaitGroup, error) {
	director, err := s.loadParticipant(uuid)
	if err != nil {
		return nil, err
	}
	if director.Type != entity.ParticipantOpenFLTypeDirector {
		return nil, errors.Errorf("participant %s is not an OpenFL director", director.UUID)
	}

	if !force && director.Status != entity.ParticipantOpenFLStatusActive {
		return nil, errors.Errorf("director cannot be removed when in status: %v", director.Status)
	}

	instanceList, err := s.ParticipantOpenFLRepo.ListByFederationUUID(director.FederationUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list participants in federation")
	}
	participantList := instanceList.([]entity.ParticipantOpenFL)
	if len(participantList) > 1 {
		return nil, errors.Errorf("cannot remove director as there are %v envoy(s) in this federation", len(participantList)-1)
	}

	director.Status = entity.ParticipantOpenFLStatusRemoving
	if err := s.ParticipantOpenFLRepo.UpdateStatusByUUID(director); err != nil {
		return nil, errors.Wrapf(err, "failed to update director status")
	}

	// TODO: revoke the certificate, after we have some OCSP mechanism in place
	if err := s.CertificateService.RemoveBinding(director.UUID); err != nil {
		return nil, errors.Wrapf(err, "failed to remove certificate bindings")
	}

	//record removing event
	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLDirector, director.UUID, "start removing director", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "uninstalling openfl director").Str("uuid", director.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLDirector, director.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("uninstalling OpenFL director %s with UUID %s", director.Name, director.UUID)
		err := func() error {
			endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(director.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}
			if director.JobUUID != "" {
				operationLog.Info().Msgf("try to stop KubeFATE job with UUID %s", director.JobUUID)
				err := kfClient.StopJob(director.JobUUID)
				if err != nil {
					return err
				}
			}
			if director.ClusterUUID != "" {
				operationLog.Info().Msgf("delete KubeFATE cluster with UUID %s", director.ClusterUUID)
				jobUUID, err := kfClient.SubmitClusterDeletionJob(director.ClusterUUID)
				if err != nil {
					// TODO: use helm or client-go to try to further clean things up
					return err
				}
				if jobUUID != "" {
					director.JobUUID = jobUUID
					if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(director); err != nil {
						return errors.Wrap(err, "failed to update director's job uuid")
					}
					if _, err := kfClient.WaitJob(jobUUID); err != nil {
						return err
					}
				}
			}
			// delete registry secret
			if director.ExtraAttribute.UseRegistrySecret {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(director.Namespace).
					Delete(context.TODO(), imagePullSecretsNameOpenFL, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msgf("error deleting registry secret: %v", err)
				} else {
					operationLog.Info().Msgf("deleted registry secret %s", imagePullSecretsNameOpenFL)
				}
			}

			// delete certs secrets
			if director.CertConfig.DirectorServerCertInfo.BindingMode != entity.CertBindingModeSkip {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(director.Namespace).
					Delete(context.TODO(), entity.ParticipantOpenFLSecretNameDirector, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msgf("error deleting stale director cert secret: %v", err)
				} else {
					operationLog.Info().Msgf("deleted stale director cert secret")
				}
			}
			if director.CertConfig.JupyterClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(director.Namespace).
					Delete(context.TODO(), entity.ParticipantOpenFLSecretNameJupyter, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msgf("error deleting stale %s secret: %v", entity.ParticipantOpenFLSecretNameJupyter, err)
				} else {
					operationLog.Info().Msgf("deleted stale %s secret", entity.ParticipantOpenFLSecretNameJupyter)
				}
			}
			// finally, delete the namespace
			if director.ExtraAttribute.IsNewNamespace {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Namespaces().Delete(context.TODO(), director.Namespace, v1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					return errors.Wrapf(err, "failed to delete namespace")
				}
				operationLog.Info().Msgf("namespace %s deleted", director.Namespace)
			}
			return nil
		}()
		if err != nil {
			operationLog.Error().Msgf("error uninstalling openfl director, error: %v", err)
			if !force {
				return
			}
		}
		if deleteErr := s.ParticipantOpenFLRepo.DeleteByUUID(director.UUID); deleteErr != nil {
			operationLog.Error().Msgf("error deleting director from repo, error: %v", err)
			return
		}
		operationLog.Info().Msgf("uninstalled OpenFL director %s with UUID %s", director.Name, director.UUID)
	}()
	return wg, nil
}

// HandleRegistrationRequest process a Envoy device registration request
func (s *ParticipantOpenFLService) HandleRegistrationRequest(req *ParticipantOpenFLEnvoyRegistrationRequest) (*entity.ParticipantOpenFL, error) {
	token, err := s.validateEnvoyToken(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to validate token")
	}

	instance, err := s.FederationRepo.GetByUUID(token.FederationUUID)
	if err != nil {
		return nil, err
	}
	req.federation = instance.(*entity.FederationOpenFL)

	var caCert *x509.Certificate
	ca, err := s.CertificateService.DefaultCA()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get default CA")
	}
	caCert, err = ca.RootCert()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get CA cert")
	}
	req.caCert = caCert

	if req.ChartUUID == "" {
		req.ChartUUID = "c62b27a6-bf0f-4515-840a-2554ed63aa56"
	}
	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query chart")
	}
	req.chart = instance.(*entity.Chart)
	if req.chart.Type != entity.ChartTypeOpenFLEnvoy {
		return nil, errors.Errorf("chart %s is not for OpenFL envoy deployment", req.chart.UUID)
	}

	infraProvider, err := s.configEnvoyInfra(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to prepare the envoy infra provider")
	}

	if req.Name == "" {
		infraAPIHost, err := infraProvider.Config.APIHost()
		if err != nil {
			return nil, err
		}

		u, err := url.Parse(infraAPIHost)
		if err != nil {
			return nil, err
		}
		req.Name = fmt.Sprintf("envoy-%s", toDeploymentName(u.Hostname()))
	}
	//TODO: check if same name envoy exists in the same infra

	if req.Namespace == "" {
		req.Namespace = fmt.Sprintf("%s-envoy", toDeploymentName(req.federation.Name))
	}
	K8sClient, err := kubernetes.NewKubernetesClient("", infraProvider.Config.KubeConfigContent)
	if err != nil {
		return nil, err
	}
	_, err = K8sClient.GetClientSet().CoreV1().Namespaces().Get(context.TODO(), req.Namespace, v1.GetOptions{})
	if err == nil {
		return nil, errors.Errorf("namespace %s exists. cannot override", req.Namespace)
	}

	deploymentYAML, err := s.GetOpenFLEnvoyYAML(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get deployment yaml")
	}

	envoy := &entity.ParticipantOpenFL{
		Participant: entity.Participant{
			UUID:           uuid.NewV4().String(),
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.federation.UUID,
			EndpointUUID:   "",
			ChartUUID:      req.chart.UUID,
			Namespace:      req.Namespace,
			ClusterUUID:    "",
			JobUUID:        "",
			DeploymentYAML: deploymentYAML,
			IsManaged:      true,
			ExtraAttribute: entity.ParticipantExtraAttribute{
				IsNewNamespace:    false,
				UseRegistrySecret: false,
			},
		},
		Type:      entity.ParticipantOpenFLTypeEnvoy,
		Status:    entity.ParticipantOpenFLStatusInstallingEndpoint,
		InfraUUID: infraProvider.UUID,
		TokenUUID: token.UUID,
		CertConfig: entity.ParticipantOpenFLCertConfig{
			EnvoyClientCertInfo: entity.ParticipantComponentCertInfo{
				BindingMode: entity.CertBindingModeCreate,
			},
		},
		AccessInfo: nil,
		Labels:     valueobject.Labels{},
	}

	for k, v := range token.Labels {
		envoy.Labels[k] = v
	}
	for k, v := range req.Labels {
		envoy.Labels[k] = v
	}

	if err := s.ParticipantOpenFLRepo.Create(envoy); err != nil {
		return nil, err
	}

	go func() {
		operationLog := log.Logger.With().Timestamp().Str("action", "installing envoy").Str("uuid", envoy.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLEnvoy, envoy.UUID, message, eventLvl)
			}))
		req.operationLog = &operationLog
		operationLog.Info().Msgf("creating envoy %s with UUID %s", req.Name, envoy.UUID)
		if err := func() (err error) {
			// TODO: check the namespace passed here
			endpointUUID, err := s.EndpointService.ensureEndpointExist(infraProvider.UUID, "")
			if err != nil {
				return err
			}
			envoy.Status = entity.ParticipantOpenFLStatusInstallingEnvoy
			envoy.EndpointUUID = endpointUUID
			if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
				return err
			}
			operationLog.Info().Msgf("kubefate endpoint prepared")
			return s.installEnvoyInstance(req, envoy)
		}(); err != nil {
			operationLog.Error().Msgf("failed to install openfl envoy, error: %v", err)
			envoy.Status = entity.ParticipantOpenFLStatusFailed
			if updateErr := s.ParticipantOpenFLRepo.UpdateStatusByUUID(envoy); updateErr != nil {
				operationLog.Error().Msgf("failed to update envoy status, error: %v", updateErr)
			}
		}
	}()
	return envoy, nil
}

// GetOpenFLEnvoyYAML generates the envoy deployment yaml based on the envoy registration request
func (s *ParticipantOpenFLService) GetOpenFLEnvoyYAML(req *ParticipantOpenFLEnvoyRegistrationRequest) (string, error) {
	chart := req.chart
	federation := req.federation

	instance, err := s.ParticipantOpenFLRepo.GetDirectorByFederationUUID(federation.UUID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to check director existence status")
	}
	director := instance.(*entity.ParticipantOpenFL)

	if director.Status != entity.ParticipantOpenFLStatusActive {
		return "", errors.Errorf("director %v is not in active status", director.UUID)
	}

	accessInfoMap := director.AccessInfo
	if accessInfoMap == nil {
		return "", errors.New("director access info is missing")
	}
	data := struct {
		Name                 string
		Namespace            string
		DirectorFQDN         string
		DirectorIP           string
		DirectorPort         int
		AggPort              int
		Domain               string
		UseRegistry          bool
		Registry             string
		UseImagePullSecrets  bool
		ImagePullSecretsName string
		EnvoyConfig          string
		EnablePSP            bool
	}{
		Name:                 toDeploymentName(req.Name),
		Namespace:            req.Namespace,
		Domain:               federation.Domain,
		EnvoyConfig:          mock.DefaultEnvoyConfig,
		UseRegistry:          req.RegistryConfig.UseRegistry,
		Registry:             req.RegistryConfig.Registry,
		UseImagePullSecrets:  req.RegistryConfig.UseRegistrySecret,
		ImagePullSecretsName: imagePullSecretsNameOpenFL,
		EnablePSP:            req.EnablePSP,
	}
	if directorAccess, ok := accessInfoMap[entity.ParticipantOpenFLServiceNameDirector]; !ok {
		return "", errors.New("missing director access info")
	} else {
		data.DirectorFQDN = directorAccess.FQDN
		data.DirectorIP = directorAccess.Host
		data.DirectorPort = directorAccess.Port
	}
	if aggAccess, ok := accessInfoMap[entity.ParticipantOpenFLServiceNameAggregator]; !ok {
		return "", errors.New("missing director agg access info")
	} else {
		data.AggPort = aggAccess.Port
	}

	if req.ConfigYAML != "" {
		data.EnvoyConfig = req.ConfigYAML
	} else if federation.UseCustomizedShardDescriptor {
		data.EnvoyConfig = federation.ShardDescriptorConfig.EnvoyConfigYaml
	}

	t, err := template.New("openfl-envoy").Funcs(sprig.TxtFuncMap()).Parse(chart.InitialYamlTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RemoveEnvoy removes and uninstalls an OpenFL envoy
func (s *ParticipantOpenFLService) RemoveEnvoy(uuid string, force bool) error {
	envoy, err := s.loadParticipant(uuid)
	if err != nil {
		return err
	}
	if envoy.Type != entity.ParticipantOpenFLTypeEnvoy {
		return errors.Errorf("participant %s is not an OpenFL envoy", envoy.UUID)
	}

	if !force && envoy.Status != entity.ParticipantOpenFLStatusActive {
		return errors.Errorf("director cannot be removed when in status: %v", envoy.Status)
	}

	envoy.Status = entity.ParticipantOpenFLStatusRemoving
	if err := s.ParticipantOpenFLRepo.UpdateStatusByUUID(envoy); err != nil {
		return errors.Wrapf(err, "failed to update director status")
	}

	// TODO: revoke the certificate
	if err := s.CertificateService.RemoveBinding(envoy.UUID); err != nil {
		return errors.Wrapf(err, "failed to remove certificate bindings")
	}

	// record removing event
	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLEnvoy, envoy.UUID, "start removing envoy", entity.EventLogLevelInfo)

	go func() {
		operationLog := log.Logger.With().Timestamp().Str("action", "uninstalling openfl envoy").Str("uuid", envoy.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeOpenFLEnvoy, envoy.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("uninstalling OpenFL envoy %s with UUID %s", envoy.Name, envoy.UUID)
		err := func() error {
			endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(envoy.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}
			if envoy.JobUUID != "" {
				operationLog.Info().Msgf("try to stop envoy job: %s", envoy.JobUUID)
				err := kfClient.StopJob(envoy.JobUUID)
				if err != nil {
					return err
				}
			}
			if envoy.ClusterUUID != "" {
				operationLog.Info().Msgf("try to delete envoy cluster in KubeFATE with uuid: %s", envoy.ClusterUUID)
				jobUUID, err := kfClient.SubmitClusterDeletionJob(envoy.ClusterUUID)
				if err != nil {
					// TODO: use helm or client-go to try to clean things up
					return err
				}
				if jobUUID != "" {
					envoy.JobUUID = jobUUID
					if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
						return errors.Wrap(err, "failed to update envoy's job uuid")
					}
					if _, err := kfClient.WaitJob(jobUUID); err != nil {
						return err
					}
				}
			}
			//Delete registry secret
			if envoy.ExtraAttribute.UseRegistrySecret {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(envoy.Namespace).
					Delete(context.TODO(), imagePullSecretsNameOpenFL, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msgf("error deleting registry secret: %v", err)
				} else {
					operationLog.Info().Msgf("deleted registry secret  %s", imagePullSecretsNameOpenFL)
				}
			}
			if envoy.CertConfig.EnvoyClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(envoy.Namespace).
					Delete(context.TODO(), entity.ParticipantOpenFLSecretNameEnvoy, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msgf("error deleting stale %s secret: %v", entity.ParticipantOpenFLSecretNameEnvoy, err)
				} else {
					operationLog.Info().Msgf("deleted stale %s secret", entity.ParticipantOpenFLSecretNameEnvoy)
				}
			}
			if envoy.ExtraAttribute.IsNewNamespace {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Namespaces().Delete(context.TODO(), envoy.Namespace, v1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					return errors.Wrapf(err, "failed to delete namespace")
				}
				operationLog.Info().Msgf("namespace %s deleted", envoy.Namespace)
			}
			return nil
		}()
		if err != nil {
			operationLog.Error().Msgf("error uninstalling openfl envoy: %v", err)
			if !force {
				return
			}
		}
		if deleteErr := s.ParticipantOpenFLRepo.DeleteByUUID(envoy.UUID); deleteErr != nil {
			operationLog.Info().Msgf("error deleting envoy from repo: %v", deleteErr)
			return
		}
		operationLog.Info().Msgf("uninstalled OpenFL envoy %s with UUID %s", envoy.Name, envoy.UUID)
	}()
	return nil
}

func (s *ParticipantOpenFLService) loadParticipant(uuid string) (*entity.ParticipantOpenFL, error) {
	instance, err := s.ParticipantOpenFLRepo.GetByUUID(uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query participant")
	}
	return instance.(*entity.ParticipantOpenFL), err
}

func (s *ParticipantOpenFLService) configEnvoyInfra(req *ParticipantOpenFLEnvoyRegistrationRequest) (*entity.InfraProviderKubernetes, error) {
	if req.federation == nil {
		return nil, errors.New("missing federation")
	}
	kubeconfig := valueobject.KubeConfig{
		KubeConfigContent: req.KubeConfig,
	}
	if err := kubeconfig.Validate(); err != nil {
		return nil, err
	}

	infraAPIHost, err := kubeconfig.APIHost()
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(infraAPIHost)
	if err != nil {
		return nil, err
	}

	infraProvider := &entity.InfraProviderKubernetes{
		InfraProviderBase: entity.InfraProviderBase{
			Name:        u.Hostname(),
			Description: "added during registering OpenFL envoy",
			Type:        entity.InfraProviderTypeK8s,
		},
		Config: kubeconfig,
		Repo:   s.InfraRepo,
	}

	if err := s.InfraRepo.ProviderExists(infraProvider); err == nil {
		log.Info().Msgf("creating infra provider during envoy registration, name: %s", infraProvider.Name)
		if err := infraProvider.Create(); err != nil {
			return nil, err
		}
	} else if errors.Is(err, repo.ErrProviderExist) {
		infraProviderInstance, err := s.InfraRepo.GetByConfigSHA256(infraProvider.Config.SHA2565())
		if err != nil {
			return nil, errors.Wrap(err, "failed to load infra provider")
		}
		infraProvider = infraProviderInstance.(*entity.InfraProviderKubernetes)
	} else {
		return nil, errors.Wrap(err, "failed to check provider existence")
	}
	return infraProvider, nil
}

func (s *ParticipantOpenFLService) validateEnvoyToken(req *ParticipantOpenFLEnvoyRegistrationRequest) (*entity.RegistrationTokenOpenFL, error) {
	tokenType, tokenStr, err := entity.RegistrationTokenParse(req.TokenStr)
	if err != nil {
		return nil, err
	}
	token := &entity.RegistrationTokenOpenFL{
		RegistrationToken: entity.RegistrationToken{
			TokenType: tokenType,
			TokenStr:  tokenStr,
			Repo:      s.TokenRepo,
		},
		ParticipantRepo: s.ParticipantOpenFLRepo,
	}
	if err := s.TokenRepo.LoadByTypeAndStr(token); err != nil {
		return nil, err
	}
	if err := token.Validate(); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *ParticipantOpenFLService) installEnvoyInstance(req *ParticipantOpenFLEnvoyRegistrationRequest, envoy *entity.ParticipantOpenFL) error {
	if envoy.EndpointUUID == "" {
		return errors.New("missing endpoint uuid")
	}
	if req.federation == nil {
		return errors.New("missing federation instance")
	}
	if req.caCert == nil {
		return errors.New("missing ca cert")
	}
	if req.operationLog == nil {
		return errors.New("missing logger")
	}
	if req.chart == nil {
		return errors.New("missing chart")
	}
	endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(envoy.EndpointUUID)
	if kfClientCloser != nil {
		defer kfClientCloser()
	}
	if err != nil {
		return err
	}
	if req.chart.Private {
		if err := kfClient.EnsureChartExist(req.chart.ChartName, req.chart.Version, req.chart.ArchiveContent); err != nil {
			return errors.Wrapf(err, "error uploading FedLCM private chart")
		}
	}

	if envoy.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
		return err
	}
	if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
		return errors.Wrap(err, "failed to update cluster attribute")
	}

	if !req.SkipCommonPythonFiles {
		req.operationLog.Info().Msgf("creating shard descriptor python files")
		pythonFiles := mock.DefaultEnvoyPythonConfig
		if req.federation.UseCustomizedShardDescriptor && len(req.federation.ShardDescriptorConfig.PythonFiles) > 0 {
			pythonFiles = req.federation.ShardDescriptorConfig.PythonFiles
		}
		if err := createEnvoyShardDescriptor(endpointMgr.K8sClient(), req.Namespace, pythonFiles); err != nil {
			return err
		}
	}

	if req.RegistryConfig.UseRegistrySecret {
		if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameOpenFL, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
			return errors.Wrap(err, "failed to create registry secret")
		} else {
			req.operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameOpenFL, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
		}
	}

	if envoy.CertConfig.EnvoyClientCertInfo.BindingMode == entity.CertBindingModeCreate {
		envoy.CertConfig.EnvoyClientCertInfo.CommonName = req.Name
		req.operationLog.Info().Msgf("creating certificate for the envoy client with CN: %s", req.Name)
		dnsNames := []string{req.Name}
		cert, pk, err := s.CertificateService.CreateCertificateSimple(req.Name, defaultCertLifetime, dnsNames)
		if err != nil {
			return errors.Wrapf(err, "failed to create envoy certificate")
		}
		req.operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
		err = createEnvoySecret(endpointMgr.K8sClient(), req.Namespace, req.caCert, cert, pk)
		if err != nil {
			return err
		}
		if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypeOpenFLEnvoy, envoy.UUID, req.federation.UUID, entity.FederationTypeOpenFL); err != nil {
			return err
		}
		envoy.CertConfig.EnvoyClientCertInfo.UUID = cert.UUID
		if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
			return errors.Wrap(err, "failed to update envoy cert info")
		}
		req.operationLog.Info().Msgf("certificate prepared")
	}

	jobUUID, err := kfClient.SubmitClusterInstallationJob(envoy.DeploymentYAML)
	if err != nil {
		return errors.Wrapf(err, "fail to submit envoy creation request")
	}
	envoy.JobUUID = jobUUID
	if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
		return errors.Wrap(err, "failed to update envoy's job uuid")
	}
	clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
	if err != nil {
		return errors.Wrapf(err, "fail to get cluster uuid")
	}
	envoy.ClusterUUID = clusterUUID
	if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
		return errors.Wrap(err, "failed to update cluster uuid")
	}
	req.operationLog.Info().Msgf("job submitted and running: jobUUID: %s, clusterUUID: %s", jobUUID, clusterUUID)

	job, err := kfClient.WaitJob(jobUUID)
	if err != nil {
		return err
	}
	if job.Status != modules.JobStatusSuccess {
		return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
	}
	envoy.Status = entity.ParticipantOpenFLStatusActive
	if err := s.ParticipantOpenFLRepo.UpdateInfoByUUID(envoy); err != nil {
		return errors.Wrap(err, "failed to save cluster info")
	}
	req.operationLog.Info().Msgf("envoy deployed and running, name: %s, uuid: %s", envoy.UUID, envoy.Name)
	return nil
}

var (
	createDirectorSecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate,
		directorCert *entity.Certificate, directorKey *rsa.PrivateKey) error {
		certBytes, err := directorCert.EncodePEM()
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: entity.ParticipantOpenFLSecretNameDirector,
			},
			Data: map[string][]byte{
				"priv.key": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(directorKey),
				}),
				"director.crt": certBytes,
				"root_ca.crt": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}

	createJupyterSecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate,
		jupyterCert *entity.Certificate, jupyterKey *rsa.PrivateKey) error {
		certBytes, err := jupyterCert.EncodePEM()
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: entity.ParticipantOpenFLSecretNameJupyter,
			},
			Data: map[string][]byte{
				"priv.key": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(jupyterKey),
				}),
				"notebook.crt": certBytes,
				"root_ca.crt": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}

	createEnvoySecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate,
		cert *entity.Certificate, key *rsa.PrivateKey) error {
		certBytes, err := cert.EncodePEM()
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: entity.ParticipantOpenFLSecretNameEnvoy,
			},
			Data: map[string][]byte{
				"priv.key": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(key),
				}),
				"envoy.crt": certBytes,
				"root_ca.crt": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}

	createEnvoyShardDescriptor = func(client kubernetes.Client, namespace string, pythonFiles map[string]string) error {
		cm := &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name: "envoy-python-configs",
			},
			Data: pythonFiles,
		}
		return createConfigMap(client, namespace, cm)
	}
)
