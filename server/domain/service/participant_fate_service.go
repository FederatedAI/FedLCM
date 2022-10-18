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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	site_portal_client "github.com/FederatedAI/FedLCM/pkg/site-portal-client"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	imagePullSecretsNameFATE = "registrykeyfate"
)

// ParticipantFATEService is the service to manage fate participants
type ParticipantFATEService struct {
	ParticipantFATERepo repo.ParticipantFATERepository
	ParticipantService
}

// ParticipantFATEExchangeYAMLCreationRequest is the request to get the exchange deployment yaml file
type ParticipantFATEExchangeYAMLCreationRequest struct {
	ChartUUID   string                               `json:"chart_uuid"`
	Name        string                               `json:"name"`
	Namespace   string                               `json:"namespace"`
	ServiceType entity.ParticipantDefaultServiceType `json:"service_type"`
	// RegistrySecretConfig in valueobject.KubeRegistryConfig is not used for generating the yaml content
	RegistryConfig valueobject.KubeRegistryConfig `json:"registry_config"`
	EnablePSP      bool                           `json:"enable_psp"`
}

// ParticipantFATEClusterYAMLCreationRequest is the request to get the cluster deployment yaml
type ParticipantFATEClusterYAMLCreationRequest struct {
	ParticipantFATEExchangeYAMLCreationRequest
	FederationUUID    string `json:"federation_uuid"`
	PartyID           int    `json:"party_id"`
	EnablePersistence bool   `json:"enable_persistence"`
	StorageClass      string `json:"storage_class"`
}

// ParticipantFATEExternalExchangeCreationRequest is the request for creating a record of an exchange not managed by this service
type ParticipantFATEExternalExchangeCreationRequest struct {
	Name                    string                          `json:"name"`
	Description             string                          `json:"description"`
	FederationUUID          string                          `json:"federation_uuid"`
	TrafficServerAccessInfo entity.ParticipantModulesAccess `json:"traffic_server_access_info"`
	NginxAccessInfo         entity.ParticipantModulesAccess `json:"nginx_access_info"`
}

// ParticipantFATEExternalClusterCreationRequest is the request for creating a record of a FATE cluster not managed by this service
type ParticipantFATEExternalClusterCreationRequest struct {
	Name             string                          `json:"name"`
	Description      string                          `json:"description"`
	FederationUUID   string                          `json:"federation_uuid"`
	PartyID          int                             `json:"party_id"`
	PulsarAccessInfo entity.ParticipantModulesAccess `json:"pulsar_access_info"`
	NginxAccessInfo  entity.ParticipantModulesAccess `json:"nginx_access_info"`
}

type nginxRouteTableEntry struct {
	Host     string `json:"host"`
	HttpPort int    `json:"http_port"`
}

type atsRouteTableEntry struct {
	FQDN        string `json:"fqdn"`
	TunnelRoute string `json:"tunnelRoute"`
}

// ParticipantFATEExchangeCreationRequest is the exchange creation request
type ParticipantFATEExchangeCreationRequest struct {
	ParticipantFATEExchangeYAMLCreationRequest
	ParticipantDeploymentBaseInfo
	FederationUUID           string                              `json:"federation_uuid"`
	ProxyServerCertInfo      entity.ParticipantComponentCertInfo `json:"proxy_server_cert_info"`
	FMLManagerServerCertInfo entity.ParticipantComponentCertInfo `json:"fml_manager_server_cert_info"`
	FMLManagerClientCertInfo entity.ParticipantComponentCertInfo `json:"fml_manager_client_cert_info"`
}

// ParticipantFATEClusterCreationRequest is the cluster creation request
type ParticipantFATEClusterCreationRequest struct {
	ParticipantFATEClusterYAMLCreationRequest
	ParticipantDeploymentBaseInfo
	PulsarServerCertInfo     entity.ParticipantComponentCertInfo `json:"pulsar_server_cert_info"`
	SitePortalServerCertInfo entity.ParticipantComponentCertInfo `json:"site_portal_server_cert_info"`
	SitePortalClientCertInfo entity.ParticipantComponentCertInfo `json:"site_portal_client_cert_info"`
}

// CheckPartyIDConflict returns error if the party id is taken in a federation
func (s *ParticipantFATEService) CheckPartyIDConflict(federationUUID string, partyID int) error {
	conflict, err := s.ParticipantFATERepo.IsConflictedByFederationUUIDAndPartyID(federationUUID, partyID)
	if err != nil {
		return errors.Wrap(err, "failed to check party id existence")
	}
	if conflict {
		return errors.Errorf("party id %v is aleady being used", partyID)
	}
	return nil
}

// GetExchangeDeploymentYAML returns the exchange deployment yaml content
func (s *ParticipantFATEService) GetExchangeDeploymentYAML(req *ParticipantFATEExchangeYAMLCreationRequest) (string, error) {
	instance, err := s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to query chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATEExchange {
		return "", errors.Errorf("chart %s is not for FATE exchange deployment", chart.UUID)
	}

	t, err := template.New("fate-exchange").Parse(chart.InitialYamlTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		Name                 string
		Namespace            string
		ServiceType          string
		UseRegistry          bool
		Registry             string
		UseImagePullSecrets  bool
		ImagePullSecretsName string
		EnablePSP            bool
	}{
		Name:                 toDeploymentName(req.Name),
		Namespace:            req.Namespace,
		ServiceType:          req.ServiceType.String(),
		UseRegistry:          req.RegistryConfig.UseRegistry,
		Registry:             req.RegistryConfig.Registry,
		UseImagePullSecrets:  req.RegistryConfig.UseRegistrySecret,
		ImagePullSecretsName: imagePullSecretsNameFATE,
		EnablePSP:            req.EnablePSP,
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetClusterDeploymentYAML returns a cluster deployment yaml
func (s *ParticipantFATEService) GetClusterDeploymentYAML(req *ParticipantFATEClusterYAMLCreationRequest) (string, error) {
	instance, err := s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to query chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATECluster {
		return "", errors.Errorf("chart %s is not for FATE cluster deployment", chart.UUID)
	}

	federationUUID := req.FederationUUID
	instance, err = s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return "", errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	instance, err = s.ParticipantFATERepo.GetExchangeByFederationUUID(federationUUID)
	if err != nil {
		return "", errors.Wrapf(err, "failed to check exchange existence status")
	}
	exchange := instance.(*entity.ParticipantFATE)

	if exchange.Status != entity.ParticipantFATEStatusActive {
		return "", errors.Errorf("exchange %v is not in active status", exchange.UUID)
	}

	accessInfoMap := exchange.AccessInfo
	if accessInfoMap == nil {
		return "", errors.New("exchange access info is missing")
	}

	data := struct {
		Name                    string
		Namespace               string
		PartyID                 int
		ExchangeNginxHost       string
		ExchangeNginxPort       int
		ExchangeATSHost         string
		ExchangeATSPort         int
		Domain                  string
		ServiceType             string
		UseRegistry             bool
		Registry                string
		UseImagePullSecrets     bool
		ImagePullSecretsName    string
		SitePortalTLSCommonName string
		EnablePersistence       bool
		StorageClass            string
		EnablePSP               bool
	}{
		Name:                    toDeploymentName(req.Name),
		Namespace:               req.Namespace,
		PartyID:                 req.PartyID,
		Domain:                  federation.Domain,
		ServiceType:             req.ServiceType.String(),
		UseRegistry:             req.RegistryConfig.UseRegistry,
		Registry:                req.RegistryConfig.Registry,
		UseImagePullSecrets:     req.RegistryConfig.UseRegistrySecret,
		ImagePullSecretsName:    imagePullSecretsNameFATE,
		SitePortalTLSCommonName: fmt.Sprintf("site-%d.server.%s", req.PartyID, federation.Domain),
		EnablePersistence:       req.EnablePersistence,
		StorageClass:            req.StorageClass,
		EnablePSP:               req.EnablePSP,
	}
	if nginxAccess, ok := accessInfoMap[entity.ParticipantFATEServiceNameNginx]; !ok {
		return "", errors.New("missing exchange nginx access info")
	} else {
		data.ExchangeNginxHost = nginxAccess.Host
		data.ExchangeNginxPort = nginxAccess.Port
	}
	if atsAccess, ok := accessInfoMap[entity.ParticipantFATEServiceNameATS]; !ok {
		return "", errors.New("missing exchange traffic-server access info")
	} else {
		data.ExchangeATSHost = atsAccess.Host
		data.ExchangeATSPort = atsAccess.Port
	}

	t, err := template.New("fate-cluster").Parse(chart.InitialYamlTemplate)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// CreateExchange creates a FATE exchange, the returned *sync.WaitGroup can be used to wait for the completion of the async goroutine
func (s *ParticipantFATEService) CreateExchange(req *ParticipantFATEExchangeCreationRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {
	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	if exist, err := s.ParticipantFATERepo.IsExchangeCreatedByFederationUUID(federationUUID); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check exchange existence status")
	} else if exist {
		return nil, nil, errors.Errorf("an exchange is already deployed in federation %s", federationUUID)
	}

	if err := s.EndpointService.TestKubeFATE(req.EndpointUUID); err != nil {
		return nil, nil, err
	}

	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATEExchange {
		return nil, nil, errors.Errorf("chart %s is not for FATE exchange deployment", chart.UUID)
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

	if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeReuse {
		return nil, nil, errors.New("cannot re-use existing certificate")
	}

	var caCert *x509.Certificate
	if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.FMLManagerServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.FMLManagerClientCertInfo.BindingMode == entity.CertBindingModeCreate {
		ca, err := s.CertificateService.DefaultCA()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get default CA")
		}
		caCert, err = ca.RootCert()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get CA cert")
		}
	}

	exchange := &entity.ParticipantFATE{
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
		Type:       entity.ParticipantFATETypeExchange,
		PartyID:    0,
		Status:     entity.ParticipantFATEStatusInstalling,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
		CertConfig: entity.ParticipantFATECertConfig{
			ProxyServerCertInfo:      req.ProxyServerCertInfo,
			FMLManagerClientCertInfo: req.FMLManagerClientCertInfo,
			FMLManagerServerCertInfo: req.FMLManagerServerCertInfo,
		},
	}
	err = s.ParticipantFATERepo.Create(exchange)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "start creating exchange", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "installing fate exchange").Str("uuid", exchange.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("creating FATE exchange %s with UUID %s", exchange.Name, exchange.UUID)
		if err := func() error {
			endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(req.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}

			if chart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", chart.ChartName, chart.Version)
				if err := kfClient.EnsureChartExist(chart.ChartName, chart.Version, chart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			if exchange.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
				return err
			}
			if exchange.ExtraAttribute.IsNewNamespace {
				operationLog.Info().Msgf("created new namespace: %s", req.Namespace)
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange attribute")
				}
			} else {
				if clusters, err := kfClient.ListClusterByNamespace(req.Namespace); err != nil {
					return errors.Wrap(err, "failed to check existing fate or exchange installation")
				} else if len(clusters) > 0 {
					return errors.Errorf("cannot add new exchange as existing installations are running in the same namespace: %v", *clusters[0])
				}
			}

			if req.RegistryConfig.UseRegistrySecret {
				if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameFATE, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
					return errors.Wrap(err, "failed to create registry secret")
				} else {
					operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameFATE, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
				}
			}

			atsFQDN := req.ProxyServerCertInfo.CommonName
			if atsFQDN == "" {
				atsFQDN = fmt.Sprintf("proxy.%s", federation.Domain)
			}
			if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.ProxyServerCertInfo.CommonName == "" {
					req.ProxyServerCertInfo.CommonName = atsFQDN
				}
				operationLog.Info().Msgf("creating certificate for the ATS service with CN: %s", req.ProxyServerCertInfo.CommonName)
				dnsNames := []string{req.ProxyServerCertInfo.CommonName}
				cert, pk, err := s.CertificateService.CreateCertificateSimple(req.ProxyServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create ATS certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createATSSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypeATS, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.ProxyServerCertInfo.CommonName = req.ProxyServerCertInfo.CommonName
				exchange.CertConfig.ProxyServerCertInfo.UUID = cert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange cert info")
				}
			}

			fmlManagerFQDN := req.FMLManagerServerCertInfo.CommonName
			if fmlManagerFQDN == "" {
				fmlManagerFQDN = fmt.Sprintf("fmlmanager.server.%s", federation.Domain)
			}
			var serverCert, clientCert *entity.Certificate
			var serverPrivateKey, clientPrivateKey *rsa.PrivateKey
			if req.FMLManagerServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.FMLManagerServerCertInfo.CommonName == "" {
					req.FMLManagerServerCertInfo.CommonName = fmt.Sprintf("fmlmanager.server.%s", federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for FML Manager server with CN: %s", req.FMLManagerServerCertInfo.CommonName)
				dnsNames := []string{req.FMLManagerServerCertInfo.CommonName, "localhost"}
				serverCert, serverPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.FMLManagerServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create FML Manager server certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", serverCert.SerialNumber, serverCert.Subject.CommonName)
			}
			if req.FMLManagerClientCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.FMLManagerClientCertInfo.CommonName == "" {
					req.FMLManagerClientCertInfo.CommonName = fmt.Sprintf("fmlmanager.client.%s", federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for FML Manager client with CN: %s", req.FMLManagerClientCertInfo.CommonName)
				dnsNames := []string{req.FMLManagerClientCertInfo.CommonName}
				clientCert, clientPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.FMLManagerClientCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create FML Manager client certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", clientCert.SerialNumber, clientCert.Subject.CommonName)
			}
			if req.FMLManagerServerCertInfo.BindingMode != entity.CertBindingModeSkip && req.FMLManagerClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				err = createTLSSecret(endpointMgr.K8sClient(), req.Namespace, serverCert, serverPrivateKey, clientCert, clientPrivateKey, caCert, entity.ParticipantFATESecretNameFMLMgr)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(serverCert, entity.CertificateBindingServiceFMLManagerServer, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.FMLManagerServerCertInfo.CommonName = req.FMLManagerServerCertInfo.CommonName
				exchange.CertConfig.FMLManagerServerCertInfo.UUID = serverCert.UUID
				if err := s.CertificateService.CreateBinding(clientCert, entity.CertificateBindingServiceFMLManagerClient, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.FMLManagerClientCertInfo.CommonName = req.FMLManagerClientCertInfo.CommonName
				exchange.CertConfig.FMLManagerClientCertInfo.UUID = clientCert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange cert info")
				}
			}

			jobUUID, err := kfClient.SubmitClusterInstallationJob(exchange.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster creation request")
			}
			exchange.JobUUID = jobUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange's job uuid")
			}
			operationLog.Info().Msgf("kubefate job created, uuid: %s", exchange.JobUUID)
			clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
			if err != nil {
				return errors.Wrapf(err, "fail to get cluster uuid")
			}
			exchange.ClusterUUID = clusterUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange cluster uuid")
			}
			operationLog.Info().Msgf("kubefate-managed cluster created, uuid: %s", exchange.ClusterUUID)

			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}
			operationLog.Info().Msgf("kubefate job succeeded")

			serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameNginx), "http")
			if err != nil {
				return errors.Wrapf(err, "fail to get nginx access info")
			}
			exchange.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         false,
			}

			serviceType, host, port, err = getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameATS), "443")
			if err != nil {
				return errors.Wrapf(err, "fail to get ATS access info")
			}
			exchange.AccessInfo[entity.ParticipantFATEServiceNameATS] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         true,
				FQDN:        atsFQDN,
			}

			moduleList := m["modules"].([]interface{})
			for _, moduleName := range moduleList {
				if moduleName.(string) == "fmlManagerServer" {
					operationLog.Info().Msgf("fml-manager deployed, retrieving access info")
					serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameFMLMgr), "https-fml-manager-server")
					if err != nil {
						return errors.Wrapf(err, "fail to get fml manager access info")
					}
					exchange.AccessInfo[entity.ParticipantFATEServiceNameFMLMgr] = entity.ParticipantModulesAccess{
						ServiceType: serviceType,
						Host:        host,
						Port:        port,
						TLS:         true,
						FQDN:        fmlManagerFQDN,
					}
				}
			}

			exchange.Status = entity.ParticipantFATEStatusActive
			if err := s.BuildIngressInfoMap(exchange); err != nil {
				return errors.Wrapf(err, "failed to get ingress info")
			}
			return s.ParticipantFATERepo.UpdateInfoByUUID(exchange)
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrapf(err, "failed to install FATE exchange").Error())
			exchange.Status = entity.ParticipantFATEStatusFailed
			if updateErr := s.ParticipantFATERepo.UpdateStatusByUUID(exchange); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrapf(updateErr, "failed to update FATE exchange status").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE exchange %s(%s) deployed", exchange.Name, exchange.UUID)
	}()

	return exchange, wg, nil
}

// RemoveExchange removes and uninstalls a FATE exchange
func (s *ParticipantFATEService) RemoveExchange(uuid string, force bool) (*sync.WaitGroup, error) {
	exchange, err := s.loadParticipant(uuid)
	if err != nil {
		return nil, err
	}
	if exchange.Type != entity.ParticipantFATETypeExchange {
		return nil, errors.Errorf("participant %s is not a FATE exchange", exchange.UUID)
	}

	if !force && exchange.Status != entity.ParticipantFATEStatusActive {
		return nil, errors.Errorf("exchange cannot be removed when in status: %v", exchange.Status)
	}

	instanceList, err := s.ParticipantFATERepo.ListByFederationUUID(exchange.FederationUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list participants in federation")
	}
	participantList := instanceList.([]entity.ParticipantFATE)
	if len(participantList) > 1 {
		return nil, errors.Errorf("cannot remove exchange as there are %v cluster(s) in this federation", len(participantList)-1)
	}

	exchange.Status = entity.ParticipantFATEStatusRemoving
	if err := s.ParticipantFATERepo.UpdateStatusByUUID(exchange); err != nil {
		return nil, errors.Wrapf(err, "failed to update exchange status")
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "start removing exchange", entity.EventLogLevelInfo)

	// just do db deletion for unmanaged exchange
	if !exchange.IsManaged {
		if err := s.ParticipantFATERepo.DeleteByUUID(exchange.UUID); err != nil {
			log.Err(err).Msg("delete external exchange error:")
			return nil, err
		}
		_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "removed external exchange", entity.EventLogLevelInfo)
		return nil, nil
	}
	// TODO: revoke the certificate when we have some OCSP mechanism in place
	if err := s.CertificateService.RemoveBinding(exchange.UUID); err != nil {
		return nil, errors.Wrapf(err, "failed to remove certificate bindings")
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "uninstalling fate exchange").Str("uuid", exchange.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("uninstalling FATE exchange %s with UUID %s", exchange.Name, exchange.UUID)
		err := func() error {
			endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(exchange.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}
			if exchange.JobUUID != "" {
				operationLog.Info().Msgf("stopping KubeFATE job %s", exchange.JobUUID)
				err := kfClient.StopJob(exchange.JobUUID)
				if err != nil {
					return err
				}
			}
			if exchange.ClusterUUID != "" {
				operationLog.Info().Msgf("deleting KubeFATE-managed cluster %s", exchange.ClusterUUID)
				jobUUID, err := kfClient.SubmitClusterDeletionJob(exchange.ClusterUUID)
				if err != nil {
					// TODO: use helm or client-go to try to clean things up
					return err
				}
				if jobUUID != "" {
					operationLog.Info().Msgf("deleting job UUID %s", jobUUID)
					exchange.JobUUID = jobUUID
					if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
						return errors.Wrap(err, "failed to update exchange job uuid")
					}
					if job, err := kfClient.WaitJob(jobUUID); err != nil {
						return err
					} else if job.Status != modules.JobStatusSuccess {
						operationLog.Warn().Msgf("deleting job not succeeded, status: %v, job info: %v", job.Status, job)
					}
				}
			}

			if exchange.ExtraAttribute.UseRegistrySecret {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(exchange.Namespace).
					Delete(context.TODO(), imagePullSecretsNameFATE, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msg(errors.Wrap(err, "error deleting registry secret").Error())
				} else {
					operationLog.Info().Msgf("deleted registry secret %s", imagePullSecretsNameFATE)
				}
			}
			if exchange.CertConfig.ProxyServerCertInfo.BindingMode != entity.CertBindingModeSkip {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(exchange.Namespace).
					Delete(context.TODO(), entity.ParticipantFATESecretNameATS, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msg(errors.Wrapf(err, "error deleting stale cert secret: %s", entity.ParticipantFATESecretNameATS).Error())
				} else {
					operationLog.Info().Msgf("deleted stale cert secret: %s", entity.ParticipantFATESecretNameATS)
				}
			}

			if exchange.CertConfig.FMLManagerServerCertInfo.BindingMode != entity.CertBindingModeSkip && exchange.CertConfig.FMLManagerClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(exchange.Namespace).
					Delete(context.TODO(), entity.ParticipantFATESecretNameFMLMgr, v1.DeleteOptions{}); err != nil {
					operationLog.Error().Msg(errors.Wrapf(err, "error deleting stale %s secret", entity.ParticipantFATESecretNameFMLMgr).Error())
				} else {
					operationLog.Info().Msgf("deleted stale %s secret", entity.ParticipantFATESecretNameFMLMgr)
				}
			}
			if exchange.ExtraAttribute.IsNewNamespace {
				if err := endpointMgr.K8sClient().GetClientSet().CoreV1().Namespaces().Delete(context.TODO(), exchange.Namespace, v1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
					return errors.Wrapf(err, "failed to delete namespace")
				}
				operationLog.Info().Msgf("namespace %s deleted", exchange.Namespace)
			}
			return nil
		}()
		if err != nil {
			operationLog.Error().Msg(errors.Wrap(err, "error uninstalling exchange").Error())
			if !force {
				return
			}
		}
		if deleteErr := s.ParticipantFATERepo.DeleteByUUID(exchange.UUID); deleteErr != nil {
			operationLog.Error().Msgf(errors.Wrap(err, "error deleting exchange from repo").Error())
			return
		}
		operationLog.Info().Msgf("uninstalled FATE exchange %s with UUID %s", exchange.Name, exchange.UUID)
	}()
	return wg, nil
}

// CreateExternalExchange creates an external FATE exchange with the access info provided by user
func (s *ParticipantFATEService) CreateExternalExchange(req *ParticipantFATEExternalExchangeCreationRequest) (*entity.ParticipantFATE, error) {
	federationUUID := req.FederationUUID
	if exist, err := s.ParticipantFATERepo.IsExchangeCreatedByFederationUUID(federationUUID); err != nil {
		return nil, errors.Wrapf(err, "failed to check exchange existence status")
	} else if exist {
		return nil, errors.Errorf("an exchange is already existed in federation %s", federationUUID)
	}
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	exchange := &entity.ParticipantFATE{
		Participant: entity.Participant{
			UUID:           uuid.NewV4().String(),
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.FederationUUID,
			EndpointUUID:   "Unknown",
			ChartUUID:      "Unknown",
			Namespace:      "Unknown",
			ClusterUUID:    "Unknown",
			DeploymentYAML: "Unknown",
			IsManaged:      false,
		},
		Type:       entity.ParticipantFATETypeExchange,
		PartyID:    0,
		Status:     entity.ParticipantFATEStatusActive,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
	}
	exchange.AccessInfo[entity.ParticipantFATEServiceNameATS] = entity.ParticipantModulesAccess{
		ServiceType: corev1.ServiceType(entity.ParticipantDefaultServiceTypeNodePort.String()),
		Host:        req.TrafficServerAccessInfo.Host,
		Port:        req.TrafficServerAccessInfo.Port,
		TLS:         true,
		FQDN:        "proxy." + federation.Domain,
	}
	exchange.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
		ServiceType: corev1.ServiceType(entity.ParticipantDefaultServiceTypeNodePort.String()),
		Host:        req.NginxAccessInfo.Host,
		Port:        req.NginxAccessInfo.Port,
		TLS:         false,
		FQDN:        "",
	}

	if err := s.ParticipantFATERepo.Create(exchange); err != nil {
		return nil, err
	}
	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "created an external exchange", entity.EventLogLevelInfo)
	return exchange, nil
}

// CreateCluster creates a FATE cluster with exchange's access info, and will update exchange's route table
func (s *ParticipantFATEService) CreateCluster(req *ParticipantFATEClusterCreationRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {
	if err := s.CheckPartyIDConflict(req.FederationUUID, req.PartyID); err != nil {
		return nil, nil, err
	}
	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	instance, err = s.ParticipantFATERepo.GetExchangeByFederationUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check exchange existence status")
	}
	exchange := instance.(*entity.ParticipantFATE)

	if exchange.Status != entity.ParticipantFATEStatusActive {
		return nil, nil, errors.Errorf("exchange %v is not in active status", exchange.UUID)
	}
	if exchange.IsManaged {
		if err := s.EndpointService.TestKubeFATE(exchange.EndpointUUID); err != nil {
			return nil, nil, err
		}
	}

	if err := s.EndpointService.TestKubeFATE(req.EndpointUUID); err != nil {
		return nil, nil, err
	}

	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATECluster {
		return nil, nil, errors.Errorf("chart %s is not for FATE cluster deployment", chart.UUID)
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

	pulsarDomain, err := getPulsarDomainFromYAML(req.DeploymentYAML)
	if err != nil {
		return nil, nil, err
	}
	if pulsarDomain == "" {
		log.Warn().Msgf("using federation domain")
		pulsarDomain = federation.Domain
	}
	pulsarFQDN := fmt.Sprintf("%d.%s", req.PartyID, pulsarDomain)

	var caCert *x509.Certificate
	if req.PulsarServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.SitePortalClientCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.SitePortalServerCertInfo.BindingMode == entity.CertBindingModeCreate {
		ca, err := s.CertificateService.DefaultCA()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get default CA")
		}
		caCert, err = ca.RootCert()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get CA cert")
		}
	}

	cluster := &entity.ParticipantFATE{
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
		Type:       entity.ParticipantFATETypeCluster,
		PartyID:    req.PartyID,
		Status:     entity.ParticipantFATEStatusInstalling,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
		CertConfig: entity.ParticipantFATECertConfig{
			PulsarServerCertInfo:     req.PulsarServerCertInfo,
			SitePortalServerCertInfo: req.SitePortalServerCertInfo,
			SitePortalClientCertInfo: req.SitePortalClientCertInfo,
		},
	}
	err = s.ParticipantFATERepo.Create(cluster)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "start creating cluster", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "installing fate cluster").Str("uuid", cluster.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("creating FATE cluster %s with UUID %s", cluster.Name, cluster.UUID)
		if err := func() error {
			endpointMgr, kfClient, closer, err := s.buildKubeFATEMgrAndClient(req.EndpointUUID)
			if closer != nil {
				defer closer()
			}
			if err != nil {
				return err
			}
			if chart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", chart.ChartName, chart.Version)
				if err := kfClient.EnsureChartExist(chart.ChartName, chart.Version, chart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			if cluster.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
				return err
			}
			if cluster.ExtraAttribute.IsNewNamespace {
				operationLog.Info().Msgf("created new namespace: %s for this cluster", req.Namespace)
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster attribute")
				}
			} else {
				if clusters, err := kfClient.ListClusterByNamespace(req.Namespace); err != nil {
					return errors.Wrap(err, "failed to check existing fate or exchange installation")
				} else if len(clusters) > 0 {
					return errors.Errorf("cannot add new fate cluster as existing installations are running in the same namespace: %v", *clusters[0])
				}
			}

			if req.RegistryConfig.UseRegistrySecret {
				if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameFATE, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
					return errors.Wrap(err, "failed to create registry secret")
				} else {
					operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameFATE, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
				}
			}

			if req.PulsarServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				req.PulsarServerCertInfo.CommonName = fmt.Sprintf("%v.%s", req.PartyID, federation.Domain)
				operationLog.Info().Msgf("creating certificate for the pulsar service with CN: %s", req.PulsarServerCertInfo.CommonName)
				dnsNames := []string{req.PulsarServerCertInfo.CommonName}
				cert, pk, err := s.CertificateService.CreateCertificateSimple(req.PulsarServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create pulsar certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createPulsarSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypePulsarServer, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.PulsarServerCertInfo.CommonName = req.PulsarServerCertInfo.CommonName
				cluster.CertConfig.PulsarServerCertInfo.UUID = cert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster cert info")
				}
			}

			sitePortalFQDN := fmt.Sprintf("site-%v.server.%s", req.PartyID, federation.Domain)
			var serverCert, clientCert *entity.Certificate
			var serverPrivateKey, clientPrivateKey *rsa.PrivateKey
			if req.SitePortalServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.SitePortalServerCertInfo.CommonName == "" {
					req.SitePortalServerCertInfo.CommonName = fmt.Sprintf("site-%v.server.%s", req.PartyID, federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for Site Portal server with CN: %s", req.SitePortalServerCertInfo.CommonName)
				// we need localhost in the dnsNames because the site portal will call itself during some workflows
				dnsNames := []string{req.SitePortalServerCertInfo.CommonName, "localhost"}
				serverCert, serverPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.SitePortalServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create Site Portal server certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", serverCert.SerialNumber, serverCert.Subject.CommonName)
			}
			if req.SitePortalClientCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.SitePortalClientCertInfo.CommonName == "" {
					req.SitePortalClientCertInfo.CommonName = fmt.Sprintf("site-%v.client.%s", req.PartyID, federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for Site Portal client with CN: %s", req.SitePortalClientCertInfo.CommonName)
				dnsNames := []string{req.SitePortalClientCertInfo.CommonName}
				clientCert, clientPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.SitePortalClientCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create Site Portal client certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", clientCert.SerialNumber, clientCert.Subject.CommonName)
			}
			if req.SitePortalServerCertInfo.BindingMode != entity.CertBindingModeSkip && req.SitePortalClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				err = createTLSSecret(endpointMgr.K8sClient(), req.Namespace, serverCert, serverPrivateKey, clientCert, clientPrivateKey, caCert, entity.ParticipantFATESecretNamePortal)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(serverCert, entity.CertificateBindingServiceSitePortalServer, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.SitePortalServerCertInfo.CommonName = req.SitePortalServerCertInfo.CommonName
				cluster.CertConfig.SitePortalServerCertInfo.UUID = serverCert.UUID
				if err := s.CertificateService.CreateBinding(clientCert, entity.CertificateBindingServiceSitePortalClient, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.SitePortalClientCertInfo.CommonName = req.SitePortalClientCertInfo.CommonName
				cluster.CertConfig.SitePortalClientCertInfo.UUID = clientCert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster cert info")
				}
				operationLog.Info().Msgf("created cert secret and bindings for site-portal service")
			}

			// TODO: ensure there is no same-name cluster exists because we may create cluster-level resources using the cluster name
			jobUUID, err := kfClient.SubmitClusterInstallationJob(cluster.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster creation request")
			}
			cluster.JobUUID = jobUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to update cluster's job uuid")
			}
			operationLog.Info().Msgf("kubefate job created, uuid: %s", cluster.JobUUID)
			clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
			if err != nil {
				return errors.Wrapf(err, "fail to get cluster uuid")
			}
			cluster.ClusterUUID = clusterUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to update cluster uuid")
			}
			operationLog.Info().Msgf("kubefate-managed cluster created, uuid: %s", cluster.ClusterUUID)
			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}

			serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameNginx), "http")
			if err != nil {
				return errors.Wrapf(err, "fail to get nginx access info")
			}
			cluster.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         false,
			}
			operationLog.Info().Msgf("kubefate job succeeded")

			// the pulsar-public-tls service is always of type LoadBalancer, we try to use nodePort if no LoadBalancer IP is available
			serviceType, host, port, err = getServiceAccessWithFallback(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNamePulsar), "tls-port", true)
			if err != nil {
				return errors.Wrapf(err, "fail to get pulsar access info")
			}
			cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         true,
				FQDN:        pulsarFQDN,
			}

			moduleList := m["modules"].([]interface{})
			for _, moduleName := range moduleList {
				if moduleName.(string) == "frontend" {
					operationLog.Info().Msgf("site-portal deployed, retrieving access info")
					serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNamePortal), "https-frontend")
					if err != nil {
						return errors.Wrapf(err, "fail to get site portal access info")
					}
					cluster.AccessInfo[entity.ParticipantFATEServiceNamePortal] = entity.ParticipantModulesAccess{
						ServiceType: serviceType,
						Host:        host,
						Port:        port,
						TLS:         true,
						FQDN:        sitePortalFQDN,
					}
					if fmlManagerInfo, ok := exchange.AccessInfo[entity.ParticipantFATEServiceNameFMLMgr]; ok {
						go func() {
							operationLog.Info().Msgf("automatically configure site portal and its connection with fml manager")
							password, err := cluster.GetSitePortalAdminPassword()
							if err != nil {
								operationLog.Err(err).Msg("failed to get site portal admin password")
								return
							}
							sitePortalClient, err := site_portal_client.NewClient(site_portal_client.Site{
								Username:             "Admin",
								Password:             password,
								Name:                 cluster.Name,
								Description:          cluster.Description,
								PartyID:              uint(cluster.PartyID),
								ExternalHost:         host,
								ExternalPort:         uint(port),
								HTTPS:                true,
								FMLManagerEndpoint:   fmt.Sprintf("https://%s:%d", fmlManagerInfo.Host, fmlManagerInfo.Port),
								FMLManagerServerName: fmlManagerInfo.FQDN,
								FATEFlowHost:         "fateflow",
								FATEFlowHTTPPort:     9380,
							})
							if err != nil {
								operationLog.Err(err).Msg("failed to get site portal client instance")
								return
							}
							if err := sitePortalClient.ConfigAndConnectSite(); err != nil {
								operationLog.Err(err).Msg("failed to configure site portal")
								return
							}
							operationLog.Info().Msg("configured site portal and it is connected to fml manager")
						}()
					} else {
						operationLog.Warn().Msgf("fml manager is not installed in the exchange, skipping configuration of site portal")
					}
				}
			}

			cluster.Status = entity.ParticipantFATEStatusActive
			if err := s.BuildIngressInfoMap(cluster); err != nil {
				return errors.Wrapf(err, "failed to get ingress info")
			}
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to save cluster info")
			}
			if exchange.IsManaged {
				operationLog.Info().Msg("rebuilding exchange route table")
				if err := s.rebuildRouteTable(exchange); err != nil {
					operationLog.Error().Msg(errors.Wrap(err, "error rebuilding route table while creating cluster").Error())
				}
			}
			return nil
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrap(err, "failed to install FATE cluster").Error())
			cluster.Status = entity.ParticipantFATEStatusFailed
			if updateErr := s.ParticipantFATERepo.UpdateStatusByUUID(cluster); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrap(err, "failed to update FATE cluster status").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE cluster %s(%s) deployed", cluster.Name, cluster.UUID)
	}()
	return cluster, wg, nil
}

// RemoveCluster uninstall the cluster as well as remove it from the exchange's route table
func (s *ParticipantFATEService) RemoveCluster(uuid string, force bool) (*sync.WaitGroup, error) {
	cluster, err := s.loadParticipant(uuid)
	if err != nil {
		return nil, err
	}
	if cluster.Type != entity.ParticipantFATETypeCluster {
		return nil, errors.Errorf("participant %s is not a FATE cluster", cluster.UUID)
	}

	if !force && cluster.Status != entity.ParticipantFATEStatusActive {
		return nil, errors.Errorf("cluster cannot be removed when in status: %v", cluster.Status)
	}

	instance, err := s.ParticipantFATERepo.GetExchangeByFederationUUID(cluster.FederationUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to check exchange existence status")
	}
	exchange := instance.(*entity.ParticipantFATE)

	if exchange.IsManaged {
		if err := s.EndpointService.TestKubeFATE(exchange.EndpointUUID); err != nil {
			if !force {
				return nil, err
			}
			_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "failed to test exchange endpoint connection, continue as the force flag is set", entity.EventLogLevelError)
		}
	}
	if cluster.IsManaged {
		if err := s.EndpointService.TestKubeFATE(cluster.EndpointUUID); err != nil {
			if !force {
				return nil, err
			}
			_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "failed to test cluster endpoint connection, continue as the force flag is set", entity.EventLogLevelError)
		}
	}

	cluster.Status = entity.ParticipantFATEStatusRemoving
	if err := s.ParticipantFATERepo.UpdateStatusByUUID(cluster); err != nil {
		return nil, errors.Wrapf(err, "failed to update cluster status")
	}

	if cluster.IsManaged {
		if err := s.CertificateService.RemoveBinding(cluster.UUID); err != nil {
			return nil, errors.Wrapf(err, "failed to remove certificate bindings")
		}
	}

	err = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "start removing cluster", entity.EventLogLevelInfo)
	if err != nil {
		return nil, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "uninstalling fate cluster").Str("uuid", cluster.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("uninstalling FATE cluster %s with UUID %s", cluster.Name, cluster.UUID)
		err := func() error {
			if exchange.IsManaged {
				operationLog.Info().Msgf("updating exchange route table")
				if err := func() error {
					exchangeEndpointMgr, exchangeKFClient, exchangeKFClientCloser, err := s.buildKubeFATEMgrAndClient(exchange.EndpointUUID)
					if exchangeKFClientCloser != nil {
						defer exchangeKFClientCloser()
					}
					if err != nil {
						return errors.Wrap(err, "cannot get exchange endpoint manager")
					}

					// this process is faster than the rebuildRouteTable function
					var m map[string]interface{}
					_ = yaml.Unmarshal([]byte(exchange.DeploymentYAML), &m)

					if m["nginx"].(map[string]interface{})["route_table"] != nil {
						routeTable := m["nginx"].(map[string]interface{})["route_table"].(map[string]interface{})
						delete(routeTable, fmt.Sprintf("%v", cluster.PartyID))
					}

					sniTable := m["trafficServer"].(map[string]interface{})["route_table"].(map[string]interface{})["sni"]
					if sniTable != nil {
						var newTable []interface{}

						sniTableList := sniTable.([]interface{})
						for _, item := range sniTableList {
							itemBytes, _ := yaml.Marshal(item)
							var entry atsRouteTableEntry
							_ = yaml.Unmarshal(itemBytes, &entry)
							if strings.HasPrefix(entry.FQDN, fmt.Sprintf("%v.", cluster.PartyID)) {
								continue
							}
							newTable = append(newTable, item)
						}
						m["trafficServer"].(map[string]interface{})["route_table"].(map[string]interface{})["sni"] = newTable
					}

					updatedYaml, _ := yaml.Marshal(m)

					var originalMap map[string]interface{}
					_ = yaml.Unmarshal([]byte(exchange.DeploymentYAML), &originalMap)
					originalYaml, _ := yaml.Marshal(originalMap)
					if bytes.Equal(updatedYaml, originalYaml) {
						operationLog.Info().Msg("exchange yaml not changed")
						return nil
					}

					exchange.DeploymentYAML = string(updatedYaml)
					if err := s.ParticipantFATERepo.UpdateDeploymentYAMLByUUID(exchange); err != nil {
						return errors.Wrapf(err, "failed to update exchange info")
					}

					jobUUID, err := exchangeKFClient.SubmitClusterUpdateJob(exchange.DeploymentYAML)
					if err != nil {
						return errors.Wrapf(err, "failed to submit exchange update job")
					}
					exchange.JobUUID = jobUUID
					if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
						return errors.Wrap(err, "failed to update exchange's job uuid")
					}
					if job, err := exchangeKFClient.WaitJob(jobUUID); err != nil {
						return errors.Wrapf(err, "failed to query exchange update job status")
					} else if job.Status != modules.JobStatusSuccess {
						operationLog.Warn().Msgf("updating job not succeeded, status: %v, job info: %v", job.Status, job)
					}
					operationLog.Info().Msgf("restarting exchange ATS")
					return deletePodWithPrefix(exchangeEndpointMgr.K8sClient(), exchange.Namespace, "traffic-server")
				}(); err != nil {
					operationLog.Error().Msg(errors.Wrap(err, "error updating exchange route table").Error())
					if err := s.rebuildRouteTable(exchange); err != nil {
						operationLog.Error().Msg(errors.Wrap(err, "error rebuilding route table while deleting cluster").Error())
					} else {
						operationLog.Info().Msg("successfully rebuild route table")
					}
				}
			}
			if cluster.IsManaged {
				clusterEndpointMgr, clusterKFClient, clusterKFClientCloser, err := s.buildKubeFATEMgrAndClient(cluster.EndpointUUID)
				if clusterKFClientCloser != nil {
					defer clusterKFClientCloser()
				}
				if err != nil {
					return errors.Wrap(err, "cannot get cluster endpoint manager")
				}
				if cluster.JobUUID != "" {
					operationLog.Info().Msgf("stopping KubeFATE job %s", cluster.JobUUID)
					err := clusterKFClient.StopJob(cluster.JobUUID)
					if err != nil {
						return err
					}
				}

				if sitePortalSvcAccess, ok := cluster.AccessInfo[entity.ParticipantFATEServiceNamePortal]; ok {
					operationLog.Info().Msg("unregistering site portal from fml manager")
					if err := func() error {
						password, err := cluster.GetSitePortalAdminPassword()
						if err != nil {
							return err
						}
						sitePortalClient, err := site_portal_client.NewClient(site_portal_client.Site{
							Username:         "Admin",
							Password:         password,
							Name:             cluster.Name,
							Description:      cluster.Description,
							PartyID:          uint(cluster.PartyID),
							ExternalHost:     sitePortalSvcAccess.Host,
							ExternalPort:     uint(sitePortalSvcAccess.Port),
							HTTPS:            true,
							FATEFlowHost:     "fateflow",
							FATEFlowHTTPPort: 9380,
						})
						if err != nil {
							return err
						}
						return sitePortalClient.UnregisterFromFMLManager()
					}(); err != nil {
						operationLog.Info().Err(err).Msgf("cannot unregister site portal, continue")
					}
				}
				if cluster.ClusterUUID != "" {
					operationLog.Info().Msgf("deleting KubeFATE-managed cluster %s", cluster.ClusterUUID)
					jobUUID, err := clusterKFClient.SubmitClusterDeletionJob(cluster.ClusterUUID)
					if err != nil {
						// TODO: use helm or client-go to try to clean things up
						return err
					}
					if jobUUID != "" {
						operationLog.Info().Msgf("deleting job UUID %s", jobUUID)
						cluster.JobUUID = jobUUID
						if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
							return errors.Wrap(err, "failed to update cluster's job uuid")
						}
						if job, err := clusterKFClient.WaitJob(jobUUID); err != nil {
							return err
						} else if job.Status != modules.JobStatusSuccess {
							operationLog.Warn().Msgf("deleting job not succeeded, status: %v, job info: %v", job.Status, job)
						}
					}
				}

				if cluster.ExtraAttribute.UseRegistrySecret {
					if err := clusterEndpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(cluster.Namespace).
						Delete(context.TODO(), imagePullSecretsNameFATE, v1.DeleteOptions{}); err != nil {
						operationLog.Error().Msgf(errors.Wrap(err, "error deleting registry secret").Error())
					} else {
						operationLog.Info().Msgf("deleted registry secret %s", imagePullSecretsNameFATE)
					}
				}

				if cluster.CertConfig.PulsarServerCertInfo.BindingMode != entity.CertBindingModeSkip {
					if err := clusterEndpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(cluster.Namespace).
						Delete(context.TODO(), entity.ParticipantFATESecretNamePulsar, v1.DeleteOptions{}); err != nil {
						operationLog.Error().Msg(errors.Wrap(err, "error deleting stale cert secret").Error())
					} else {
						operationLog.Info().Msgf("deleted stale cert secret")
					}
				}

				if cluster.CertConfig.SitePortalServerCertInfo.BindingMode != entity.CertBindingModeSkip && cluster.CertConfig.SitePortalClientCertInfo.BindingMode != entity.CertBindingModeSkip {
					if err := clusterEndpointMgr.K8sClient().GetClientSet().CoreV1().Secrets(cluster.Namespace).
						Delete(context.TODO(), entity.ParticipantFATESecretNamePortal, v1.DeleteOptions{}); err != nil {
						operationLog.Error().Msgf(errors.Wrapf(err, "error deleting stale %s secret", entity.ParticipantFATESecretNamePortal).Error())
					} else {
						operationLog.Info().Msgf("deleted stale %s secret", entity.ParticipantFATESecretNamePortal)
					}
				}
				if cluster.ExtraAttribute.IsNewNamespace {
					if err := clusterEndpointMgr.K8sClient().GetClientSet().CoreV1().Namespaces().Delete(context.TODO(), cluster.Namespace, v1.DeleteOptions{}); err != nil && !apierr.IsNotFound(err) {
						return errors.Wrapf(err, "failed to delete namespace")
					}
					operationLog.Info().Msgf("namespace %s deleted", cluster.Namespace)
				}
			}
			return nil
		}()
		if err != nil {
			operationLog.Error().Msgf(errors.Wrapf(err, "error uninstalling cluster").Error())
			if !force {
				return
			}
		}
		if deleteErr := s.ParticipantFATERepo.DeleteByUUID(cluster.UUID); deleteErr != nil {
			operationLog.Error().Msgf(errors.Wrapf(deleteErr, "error deleting cluster from repo").Error())
			return
		}
		operationLog.Info().Msgf("uninstalled FATE cluster %s with UUID %s", cluster.Name, cluster.UUID)
	}()
	return wg, nil
}

//CreateExternalCluster creates an external FATE cluster with the access info provided by user
func (s *ParticipantFATEService) CreateExternalCluster(req *ParticipantFATEExternalClusterCreationRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {
	if err := s.CheckPartyIDConflict(req.FederationUUID, req.PartyID); err != nil {
		return nil, nil, err
	}
	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	instance, err = s.ParticipantFATERepo.GetExchangeByFederationUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check exchange existence status")
	}
	exchange := instance.(*entity.ParticipantFATE)

	if exchange.Status != entity.ParticipantFATEStatusActive {
		return nil, nil, errors.Errorf("exchange %v is not in active status", exchange.UUID)
	}
	if exchange.IsManaged {
		if err := s.EndpointService.TestKubeFATE(exchange.EndpointUUID); err != nil {
			return nil, nil, err
		}
	}
	cluster := &entity.ParticipantFATE{
		Participant: entity.Participant{
			UUID:           uuid.NewV4().String(),
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.FederationUUID,
			EndpointUUID:   "Unknown",
			ChartUUID:      "Unknown",
			Namespace:      "Unknown",
			ClusterUUID:    "Unknown",
			DeploymentYAML: "Unknown",
			IsManaged:      false,
		},
		Type:       entity.ParticipantFATETypeCluster,
		PartyID:    req.PartyID,
		Status:     entity.ParticipantFATEStatusActive,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
	}
	cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar] = entity.ParticipantModulesAccess{
		ServiceType: corev1.ServiceType(entity.ParticipantDefaultServiceTypeNodePort.String()),
		Host:        req.PulsarAccessInfo.Host,
		Port:        req.PulsarAccessInfo.Port,
		TLS:         true,
		FQDN:        fmt.Sprintf("%d.%s", req.PartyID, federation.Domain),
	}
	cluster.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
		ServiceType: corev1.ServiceType(entity.ParticipantDefaultServiceTypeNodePort.String()),
		Host:        req.NginxAccessInfo.Host,
		Port:        req.NginxAccessInfo.Port,
		TLS:         false,
		FQDN:        "",
	}
	err = s.ParticipantFATERepo.Create(cluster)
	if err != nil {
		return nil, nil, err
	}
	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "created an external cluster", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if exchange.IsManaged {
			operationLog := log.Logger.With().Timestamp().Str("action", "configuring external fate cluster").Str("uuid", cluster.UUID).Logger().
				Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
					eventLvl := entity.EventLogLevelInfo
					if level == zerolog.ErrorLevel {
						eventLvl = entity.EventLogLevelError
					}
					_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, message, eventLvl)
				}))
			if err := s.rebuildRouteTable(exchange); err != nil {
				operationLog.Error().Msg(errors.Wrap(err, "error rebuilding route table while creating cluster").Error())
				cluster.Status = entity.ParticipantFATEStatusFailed
				if err := s.ParticipantFATERepo.UpdateStatusByUUID(cluster); err != nil {
					operationLog.Error().Msg(errors.Wrap(err, "failed to update cluster status").Error())
				}
			}
		}
	}()
	return cluster, wg, nil
}

func (s *ParticipantFATEService) buildNginxRouteTable(routeTable map[string]interface{}, cluster *entity.ParticipantFATE) {
	routeTable[fmt.Sprintf("%d", cluster.PartyID)] = map[string][]nginxRouteTableEntry{
		"fateflow": {
			{
				Host:     cluster.AccessInfo[entity.ParticipantFATEServiceNameNginx].Host,
				HttpPort: cluster.AccessInfo[entity.ParticipantFATEServiceNameNginx].Port,
			},
		},
	}
}

func (s *ParticipantFATEService) buildATSRouteTable(routeTable map[string]interface{}, cluster *entity.ParticipantFATE) {
	sniTable := routeTable["sni"]
	if sniTable == nil {
		sniTable = []interface{}{}
	}
	sniTable = append(sniTable.([]interface{}), atsRouteTableEntry{
		FQDN:        cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar].FQDN,
		TunnelRoute: fmt.Sprintf("%s:%d", cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar].Host, cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar].Port),
	})
	routeTable["sni"] = sniTable
}

func (s *ParticipantFATEService) rebuildRouteTable(exchange *entity.ParticipantFATE) error {
	operationLog := log.Logger.With().Timestamp().Str("action", "rebuilding fate route table").Str("federation_uuid", exchange.FederationUUID).Logger().
		Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
			eventLvl := entity.EventLogLevelInfo
			if level == zerolog.ErrorLevel {
				eventLvl = entity.EventLogLevelError
			}
			_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, message, eventLvl)
		}))
	operationLog.Info().Msg("start rebuilding route table")
	instanceList, err := s.ParticipantFATERepo.ListByFederationUUID(exchange.FederationUUID)
	if err != nil {
		return errors.Wrap(err, "failed to list participants")
	}
	participantList := instanceList.([]entity.ParticipantFATE)

	exchangeEndpointMgr, exchangeKFClient, exchangeKFClientCloser, err := s.buildKubeFATEMgrAndClient(exchange.EndpointUUID)
	if exchangeKFClientCloser != nil {
		defer exchangeKFClientCloser()
	}
	if err != nil {
		return errors.Errorf("cannot get exchange endpoint manager")
	}

	var m map[string]interface{}
	_ = yaml.Unmarshal([]byte(exchange.DeploymentYAML), &m)

	// reset the route table
	m["nginx"].(map[string]interface{})["route_table"] = map[string]interface{}{}
	m["trafficServer"].(map[string]interface{})["route_table"].(map[string]interface{})["sni"] = []interface{}{}

	for _, participant := range participantList {
		if participant.Type == entity.ParticipantFATETypeCluster && participant.Status == entity.ParticipantFATEStatusActive {
			s.buildNginxRouteTable(m["nginx"].(map[string]interface{})["route_table"].(map[string]interface{}), &participant)
			s.buildATSRouteTable(m["trafficServer"].(map[string]interface{})["route_table"].(map[string]interface{}), &participant)
		}
	}

	updatedYaml, _ := yaml.Marshal(m)
	exchange.DeploymentYAML = string(updatedYaml)

	if err := s.ParticipantFATERepo.UpdateDeploymentYAMLByUUID(exchange); err != nil {
		return errors.Wrapf(err, "failed to update exchange info")
	}

	retry := 5
	for {
		if err := func() error {
			operationLog.Info().Msgf("perform exchange deployment update with %v retries", retry)
			jobUUID, err := exchangeKFClient.SubmitClusterUpdateJob(exchange.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "failed to submit exchange update job")
			}
			exchange.JobUUID = jobUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange's job uuid")
			}
			operationLog.Info().Msgf("KubeFATE update job uuid: %s", jobUUID)
			if job, err := exchangeKFClient.WaitJob(jobUUID); err != nil {
				return errors.Wrapf(err, "failed to query exchange update job status")
			} else if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}
			if err := deletePodWithPrefix(exchangeEndpointMgr.K8sClient(), exchange.Namespace, "traffic-server"); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			retry--
			operationLog.Error().Msgf(errors.Wrapf(err, "error updating exchange deployment, %v retries remaining", retry).Error())
			if retry == 0 {
				return err
			}
		} else {
			operationLog.Info().Msg("done updating exchange route table")
			return nil
		}
	}
}

func (s *ParticipantFATEService) loadParticipant(uuid string) (*entity.ParticipantFATE, error) {
	instance, err := s.ParticipantFATERepo.GetByUUID(uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query participant")
	}
	return instance.(*entity.ParticipantFATE), err
}

func (s *ParticipantFATEService) BuildIngressInfoMap(participant *entity.ParticipantFATE) error {
	var m map[string]interface{}
	_ = yaml.Unmarshal([]byte(participant.DeploymentYAML), &m)

	i, ok := m["ingress"]
	if !ok {
		return nil
	}
	ingressNameList := i.(map[string]interface{})

	ingressMap := entity.ParticipantFATEIngressMap{}

	mgr, KFClient, closer, err := s.buildKubeFATEMgrAndClient(participant.EndpointUUID)
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return err
	}

	for name := range ingressNameList {
		info, err := getIngressInfo(mgr.K8sClient(), name, participant.Namespace)
		if err != nil {
			return err
		}
		// if the ingress controller service is of type NodePort, then we can get the node port address from
		// this KFClient, and we should try to replace the address in the info with it
		ingressAddressFromKFClient := KFClient.IngressAddress()
		if !strings.HasPrefix(ingressAddressFromKFClient, "localhost") {
			if len(info.Addresses) == 0 ||
				(len(info.Addresses) == 1 && info.Addresses[0] != ingressAddressFromKFClient) {
				info.Addresses = []string{ingressAddressFromKFClient}
			}
		}
		ingressMap[name] = *info
	}
	participant.IngressInfo = ingressMap
	return nil
}

// For mocking purpose
var (
	createATSSecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate, serverCert *entity.Certificate, privateKey *rsa.PrivateKey) error {
		certBytes, err := serverCert.EncodePEM()
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: entity.ParticipantFATESecretNameATS,
			},
			Data: map[string][]byte{
				"proxy.key.pem": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(privateKey),
				}),
				"proxy.cert.pem": certBytes,
				"ca.cert.pem": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}

	createPulsarSecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate, serverCert *entity.Certificate, privateKey *rsa.PrivateKey) error {
		pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return err
		}

		certBytes, err := serverCert.EncodePEM()
		if err != nil {
			return err
		}

		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: entity.ParticipantFATESecretNamePulsar,
			},
			Data: map[string][]byte{
				"broker.key-pk8.pem": pem.EncodeToMemory(&pem.Block{
					Type:    "PRIVATE KEY",
					Headers: nil,
					Bytes:   pkcs8Bytes,
				}),
				"broker.cert.pem": certBytes,
				"ca.cert.pem": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}

	createTLSSecret = func(client kubernetes.Client, namespace string, serverCert *entity.Certificate, serverPrivateKey *rsa.PrivateKey, clientCert *entity.Certificate, clientPrivateKey *rsa.PrivateKey, caCert *x509.Certificate, secretName string) error {
		serverCertBytes, err := serverCert.EncodePEM()
		if err != nil {
			return err
		}
		clientCertBytes, err := clientCert.EncodePEM()
		if err != nil {
			return err
		}
		secret := &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name: secretName,
			},
			Data: map[string][]byte{
				"server.key": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(serverPrivateKey),
				}),
				"server.crt": serverCertBytes,
				"client.key": pem.EncodeToMemory(&pem.Block{
					Type:    "RSA PRIVATE KEY",
					Headers: nil,
					Bytes:   x509.MarshalPKCS1PrivateKey(clientPrivateKey),
				}),
				"client.crt": clientCertBytes,
				"ca.crt": pem.EncodeToMemory(&pem.Block{
					Type:    "CERTIFICATE",
					Headers: nil,
					Bytes:   caCert.Raw,
				}),
			},
		}
		return createSecret(client, namespace, secret)
	}
)

func getPulsarDomainFromYAML(yamlStr string) (string, error) {
	type Exchange struct {
		Domain string `json:"domain"`
	}
	type Pulsar struct {
		Exchange Exchange `json:"exchange"`
	}
	type Cluster struct {
		Pulsar Pulsar `json:"pulsar"`
	}
	var clusterDef Cluster
	if err := yaml.Unmarshal([]byte(yamlStr), &clusterDef); err != nil {
		return "", errors.Wrapf(err, "failed to extract pulsar exchange info")
	}
	return clusterDef.Pulsar.Exchange.Domain, nil
}
