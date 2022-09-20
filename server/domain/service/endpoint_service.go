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
	"fmt"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/kubefate"
	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	k8sErr "k8s.io/apimachinery/pkg/api/errors"
)

// EndpointService is the domain service for the endpoint management
type EndpointService struct {
	InfraProviderKubernetesRepo repo.InfraProviderRepository
	EndpointKubeFATERepo        repo.EndpointRepository
	ParticipantFATERepo         repo.ParticipantFATERepository
	ParticipantOpenFLRepo       repo.ParticipantOpenFLRepository
	EventService                EventServiceInt
}

// EndpointScanResult records the scanning result from an infra provider
type EndpointScanResult struct {
	*entity.EndpointBase
	IsManaged    bool
	IsCompatible bool
}

var minKubeFATEVersion = func() *version.Version {
	version, _ := version.NewVersion("1.4.4")
	return version
}()

var inCompatibleEndpoint = EndpointScanResult{
	EndpointBase: &entity.EndpointBase{
		UUID:        "",
		Name:        "incompatible endpoint",
		Description: "",
		Type:        entity.EndpointTypeKubeFATE,
	},
	IsManaged:    false,
	IsCompatible: false,
}

var compatibleEndpoint = EndpointScanResult{
	EndpointBase: &entity.EndpointBase{
		UUID:        "",
		Name:        "compatible endpoint",
		Description: "",
		Type:        entity.EndpointTypeKubeFATE,
	},
	IsManaged:    false,
	IsCompatible: true,
}

// for mocking purpose
var (
	newK8sClientFn   = kubernetes.NewKubernetesClient
	newKubeFATEMgrFn = kubefate.NewManager
)

const (
	imagePullSecretsNameKubeFATE = "registrykeykubefate"
)

// CreateKubeFATEEndpoint install a kubefate endpoint or add an existing kubefate as a managed endpoint
func (s *EndpointService) CreateKubeFATEEndpoint(infraUUID, namespace, name, description, yaml string,
	install bool,
	ingressControllerServiceMode entity.EndpointKubeFATEIngressControllerServiceMode) (string, error) {
	log.Info().Msgf("creating KubeFATE endpoint with name: %s, install: %v in namespace %s on infra with uuid: %s", name, install, namespace, infraUUID)
	if name == "" {
		return "", errors.New("name cannot be empty")
	}
	if namespace != "" {
		ingressControllerServiceMode = entity.EndpointKubeFATEIngressControllerServiceModeSkip
	}
	endpointListInstance, err := s.EndpointKubeFATERepo.ListByInfraProviderUUIDAndNamespace(infraUUID, namespace)
	if err != nil {
		return "", errors.Wrapf(err, "failed to query current KubeFATE endpoint info in namespace %s of infra %s", namespace, infraUUID)
	}
	domainEndpointList := endpointListInstance.([]entity.EndpointKubeFATE)
	if len(domainEndpointList) != 0 {
		return "", errors.Errorf("current infra provider %s already contains a KubeFATE endpoint in namespace %s", infraUUID, namespace)
	}

	endpoint := &entity.EndpointKubeFATE{
		EndpointBase: entity.EndpointBase{
			UUID:              uuid.NewV4().String(),
			InfraProviderUUID: infraUUID,
			Namespace:         namespace,
			Name:              name,
			Description:       description,
			Version:           "",
			Type:              entity.EndpointTypeKubeFATE,
			Status:            entity.EndpointStatusCreating,
		},
		Config: entity.KubeFATEConfig{
			IngressAddress:    "",
			IngressRuleHost:   "",
			UsePortForwarding: ingressControllerServiceMode == entity.EndpointKubeFATEIngressControllerServiceModeModeNonexistent,
		},
		DeploymentYAML:        "",
		IngressControllerYAML: "",
	}

	ingressControllerYAML := ""
	if install {
		if !endpoint.Config.UsePortForwarding {
			log.Info().Msgf("ingress controller is needed, service mode: %v", ingressControllerServiceMode)
			ingressControllerYAML, err = s.GetIngressControllerDeploymentYAML(ingressControllerServiceMode)
			if err != nil {
				return "", err
			}
		}
	}

	if yaml == "" {
		log.Info().Msgf("generating default deployment yaml")
		yaml, err = s.GetDeploymentYAML(namespace, "admin", "admin", "kubefate.net", valueobject.KubeRegistryConfig{})
		if err != nil {
			return "", err
		}
	}

	mgr, err := s.BuildKubeFATEManager(infraUUID, namespace, yaml, ingressControllerYAML)
	if err != nil {
		return "", errors.Wrapf(err, "failed to build kubefate manager")
	}

	fillInfo := func() error {
		kfc, closer, err := mgr.BuildPFClient()
		if closer != nil {
			defer closer()
		}
		if err != nil {
			return errors.Wrapf(err, "failed to get kubefate client")
		}
		versionStr, err := kfc.CheckVersion()
		if err != nil {
			return errors.Wrapf(err, "failed to get kubefate version")
		}
		endpoint.Version = versionStr
		endpoint.Config.IngressAddress = kfc.IngressAddress()
		endpoint.Config.IngressRuleHost = kfc.IngressRuleHost()
		endpoint.Description = strings.TrimSpace(endpoint.Description + " The address is a dummy one because we will use port-forwarder to do the connection.")
		return nil
	}

	if install {
		endpoint.DeploymentYAML = yaml
		endpoint.IngressControllerYAML = ingressControllerYAML
	} else {
		// ignore ingress requirements when adding existing KubeFATE
		log.Info().Msgf("adding existing KubeFATE instead of installing a new one")
		endpoint.Config.UsePortForwarding = true
		if err := fillInfo(); err != nil {
			return "", errors.Wrapf(err, "failed to query existing kubefate info")
		}
		endpoint.Status = entity.EndpointStatusReady
	}
	if err := s.EndpointKubeFATERepo.Create(endpoint); err != nil {
		return "", errors.Wrapf(err, "failed to create endpoint")
	}
	//record event of creating endpoint
	err = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, "start creating endpoint", entity.EventLogLevelInfo)
	if err != nil {
		return endpoint.UUID, err
	}
	if install {
		go func() {
			message := fmt.Sprintf("installing kubefate endpoint (%s) on infra provider: %s", name, infraUUID)
			log.Info().Msgf(message)
			_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, message, entity.EventLogLevelInfo)
			err := func() error {
				if endpoint.IngressControllerYAML != "" {
					message = fmt.Sprintf("installing ingress controller")
					log.Info().Msgf(message)
					_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, message, entity.EventLogLevelInfo)
					if err := mgr.InstallIngressNginxController(); err != nil {
						return errors.Wrapf(err, "failed to install ingress controller")
					}
				}
				message = fmt.Sprintf("remove any old KubeFATE instance")
				log.Info().Msgf(message)
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, message, entity.EventLogLevelInfo)
				if err := mgr.Uninstall(); err != nil {
					return errors.Wrapf(err, "failed to uninstall old installation")
				}
				message = fmt.Sprintf("installing new KubeFATE instance")
				log.Info().Msgf(message)
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, message, entity.EventLogLevelInfo)
				if err := mgr.Install(!endpoint.Config.UsePortForwarding); err != nil {
					return errors.Wrapf(err, "failed to install kubefate")
				}
				if err := fillInfo(); err != nil {
					return errors.Wrapf(err, "failed to get installed kubefate info")
				}
				endpoint.Status = entity.EndpointStatusReady
				if err := s.EndpointKubeFATERepo.UpdateInfoByUUID(endpoint); err != nil {
					return errors.Wrapf(err, "failed to update endpoint info")
				}
				message = fmt.Sprintf("KubeFATE installed without error")
				log.Info().Msgf(message)
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, message, entity.EventLogLevelInfo)
				return nil
			}()
			if err != nil {
				message = "error installing kubefate endpoint"
				log.Err(err).Msgf(message)
				//record event of error in installing
				eventDesc := errors.Wrapf(err, message).Error()
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, eventDesc, entity.EventLogLevelError)
				endpoint.Status = entity.EndpointStatusUnavailable
				if err := s.EndpointKubeFATERepo.UpdateStatusByUUID(endpoint); err != nil {
					message = "failed to update endpoint status"
					log.Err(err).Msg(message)
					//record event of error in updating endpoint status
					eventDesc := errors.Wrapf(err, message).Error()
					_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, endpoint.UUID, eventDesc, entity.EventLogLevelError)
				}
			}
		}()
	}
	return endpoint.UUID, nil
}

// FindKubeFATEEndpoint returns endpoints installation status from an infra provider within particular namespace
func (s *EndpointService) FindKubeFATEEndpoint(infraUUID string, namespace string) ([]EndpointScanResult, error) {
	// 1. find from repo
	endpointInstance, err := s.EndpointKubeFATERepo.ListByInfraProviderUUIDAndNamespace(infraUUID, namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get KubeFATE endpoint in namespace %s of infra %s: ", namespace, infraUUID)
	}
	domainEndpointList := endpointInstance.([]entity.EndpointKubeFATE)
	if len(domainEndpointList) != 0 {
		var resultList []EndpointScanResult
		for _, domainEndpoint := range domainEndpointList {
			resultList = append(resultList, EndpointScanResult{
				EndpointBase: &entity.EndpointBase{
					Model:             domainEndpoint.Model,
					UUID:              domainEndpoint.UUID,
					InfraProviderUUID: infraUUID,
					Namespace:         domainEndpoint.Namespace,
					Name:              domainEndpoint.Name,
					Description:       domainEndpoint.Description,
					Type:              domainEndpoint.Type,
					Status:            domainEndpoint.Status,
				},
				IsManaged:    true,
				IsCompatible: true,
			})
		}
		return resultList, nil
	}

	// 2. scan the infra
	yaml, err := s.GetDeploymentYAML(namespace, "admin", "admin", "kubefate.net", valueobject.KubeRegistryConfig{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get deployment yaml ")
	}
	kubefateMgr, err := s.BuildKubeFATEManager(infraUUID, namespace, yaml, "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kubefate manager")
	}
	deployment, err := kubefateMgr.GetKubeFATEDeployment()
	if err != nil {
		if k8sErr.IsNotFound(err) {
			log.Info().Msgf("no kubefate endpoint found in namespace %s of infra: %s", namespace, infraUUID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "error querying kubefate installation")
	}

	// we don't require ingress to be available so we can use the client based on port-forward
	kfClient, closer, err := kubefateMgr.BuildPFClient()
	if closer != nil {
		defer closer()
	}
	if err != nil {
		return []EndpointScanResult{
			inCompatibleEndpoint,
		}, nil
	}

	versionStr, err := kfClient.CheckVersion()
	if err != nil {
		return []EndpointScanResult{
			inCompatibleEndpoint,
		}, nil
	}

	currentVersion, err := version.NewVersion(versionStr)
	if err != nil {
		return []EndpointScanResult{
			inCompatibleEndpoint,
		}, nil
	}
	if currentVersion.LessThan(minKubeFATEVersion) {
		return []EndpointScanResult{
			inCompatibleEndpoint,
		}, nil
	}

	compatibleEndpointCopy := compatibleEndpoint
	compatibleEndpointCopy.CreatedAt = deployment.CreationTimestamp.Time
	return []EndpointScanResult{
		compatibleEndpointCopy,
	}, nil
}

// TestKubeFATE checks a KubeFATE connection
func (s *EndpointService) TestKubeFATE(uuid string) error {
	endpointInstance, err := s.EndpointKubeFATERepo.GetByUUID(uuid)
	if err != nil {
		return errors.Wrapf(err, "failed to get KubeFAET endpoint instance")
	}
	endpoint := endpointInstance.(*entity.EndpointKubeFATE)

	if err := func() error {
		kubefateMgr, err := s.buildKubeFATEClientManagerFromEndpointUUID(uuid)
		if err != nil {
			return err
		}
		// for testing purpose we use PFClient
		kfClient, closer, err := kubefateMgr.BuildPFClient()
		if closer != nil {
			defer closer()
		}
		if err != nil {
			return errors.Wrapf(err, "failed to build kubefate client")
		}

		_, err = kfClient.CheckVersion()
		if err != nil {
			return errors.Wrapf(err, "failed to get kubefate str")
		}
		endpoint.Status = entity.EndpointStatusReady
		if endpoint.Config.UsePortForwarding {
			endpoint.Config.IngressRuleHost = kfClient.IngressRuleHost()
			endpoint.Config.IngressAddress = kfClient.IngressAddress()
		}
		return nil
	}(); err != nil {
		endpoint.Status = entity.EndpointStatusUnavailable
		return err
	}

	_ = s.EndpointKubeFATERepo.UpdateInfoByUUID(endpoint)
	return nil
}

// RemoveEndpoint removes the specified endpoint
func (s *EndpointService) RemoveEndpoint(uuid string, uninstall bool) error {
	log.Info().Msgf("remove endpoint with uuid %s, uninstall: %v", uuid, uninstall)
	endpointInstance, err := s.EndpointKubeFATERepo.GetByUUID(uuid)
	if err != nil {
		return errors.Wrapf(err, "failed to get KubeFAET endpoint instance")
	}
	domainEndpointKubeFATE := endpointInstance.(*entity.EndpointKubeFATE)

	instanceList, err := s.ParticipantFATERepo.ListByEndpointUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "failed to query endpoint participants")
	}
	participantList := instanceList.([]entity.ParticipantFATE)
	if len(participantList) > 0 {
		return errors.Errorf("cannot remove endpoint that still contains %v participants", len(participantList))
	}

	instanceListOpenFL, err := s.ParticipantOpenFLRepo.ListByEndpointUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "failed to query endpoint participants")
	}

	participantListOpenFL := instanceListOpenFL.([]entity.ParticipantOpenFL)
	if len(participantListOpenFL) > 0 {
		return errors.Errorf("cannot remove endpoint that still contains %v OpenFL participants", len(participantListOpenFL))
	}

	domainEndpointKubeFATE.Status = entity.EndpointStatusDeleting
	if err := s.EndpointKubeFATERepo.UpdateStatusByUUID(domainEndpointKubeFATE); err != nil {
		return errors.Wrapf(err, "failed to update status to deleting")
	}

	err = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, uuid, "start deleting endpoint", entity.EventLogLevelInfo)
	if err != nil {
		return err
	}
	go func() {
		// continue the removing even if uninstallation failed
		if uninstall {
			yaml := domainEndpointKubeFATE.DeploymentYAML
			// yaml is empty means the KubeFATE was not installed by FedLCM but directly added to the database,
			// so we don't know the actual yaml user applied.
			// Here we build a default deployment yaml and delete it for future installation.
			if yaml == "" {
				yaml, err = s.GetDeploymentYAML(domainEndpointKubeFATE.Namespace, "admin", "admin", "kubefate.net", valueobject.KubeRegistryConfig{})
			}
			kubefateMgr, err := s.BuildKubeFATEManager(domainEndpointKubeFATE.InfraProviderUUID, domainEndpointKubeFATE.Namespace, yaml, domainEndpointKubeFATE.IngressControllerYAML)
			if err != nil {
				message := "failed to get kubefate manager"
				log.Err(err).Msgf(message)
				//record error event
				eventDesc := errors.Wrapf(err, message).Error()
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, uuid, eventDesc, entity.EventLogLevelError)
			} else if err := kubefateMgr.Uninstall(); err != nil {
				message := "failed to remove kubefate installation"
				log.Err(err).Msgf(message)
				//record error event
				eventDesc := errors.Wrapf(err, message).Error()
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, uuid, eventDesc, entity.EventLogLevelError)
			}
		}
		if err := s.EndpointKubeFATERepo.DeleteByUUID(uuid); err != nil {
			message := "failed to delete endpoint"
			log.Err(err).Msgf(message)
			//record error event
			eventDesc := errors.Wrapf(err, message).Error()
			_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeEndpoint, uuid, eventDesc, entity.EventLogLevelError)
		}
	}()
	return nil
}

func (s *EndpointService) buildKubeFATEClientManagerFromEndpointUUID(uuid string) (kubefate.ClientManager, error) {
	endpointInstance, err := s.EndpointKubeFATERepo.GetByUUID(uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get KubeFAET endpoint instance")
	}
	endpoint := endpointInstance.(*entity.EndpointKubeFATE)
	if endpoint.DeploymentYAML == "" {
		endpoint.DeploymentYAML, err = s.GetDeploymentYAML(endpoint.Namespace, "admin", "admin", "kubefate.net", valueobject.KubeRegistryConfig{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get default deployment yaml")
		}
	}
	return s.BuildKubeFATEManager(endpoint.InfraProviderUUID, endpoint.Namespace, endpoint.DeploymentYAML, endpoint.IngressControllerYAML)
}

// BuildKubeFATEManager retrieve a KubeFATE manager instance from the provided endpoint uuid
func (s *EndpointService) BuildKubeFATEManager(infraUUID, namespace, yaml, ingressControllerYAML string) (kubefate.Manager, error) {
	providerInstance, err := s.InfraProviderKubernetesRepo.GetByUUID(infraUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "error query infra provider")
	}
	provider := providerInstance.(*entity.InfraProviderKubernetes)

	client, err := newK8sClientFn("", provider.Config.KubeConfigContent)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kubernetes client")
	}

	installMeta, err := kubefate.BuildInstallationMetaFromYAML(namespace, yaml, ingressControllerYAML)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kubefate meta")
	}
	kubefateMgr := newKubeFATEMgrFn(client, installMeta)
	return kubefateMgr, nil
}

// GetDeploymentYAML returns the default kubefate deployment yaml
func (s *EndpointService) GetDeploymentYAML(namespace, serviceUserName, servicePassword, hostname string, registryConfig valueobject.KubeRegistryConfig) (string, error) {
	t, err := template.New("kubefate").Parse(kubefate.GetDefaultYAML())
	if err != nil {
		return "", err
	}

	isClusterAdmin := false
	if namespace == "" {
		isClusterAdmin = true
		namespace = "kube-fate"
	}

	registrySecretData := ""
	if registryConfig.UseRegistrySecret {
		registrySecretData, err = registryConfig.RegistrySecretConfig.BuildSecretB64String()
		if err != nil {
			return "", err
		}
	}
	data := struct {
		Namespace            string
		IsClusterAdmin       bool
		ServiceUserName      string
		ServicePassword      string
		Hostname             string
		UseRegistry          bool
		Registry             string
		UseImagePullSecrets  bool
		ImagePullSecretsName string
		RegistrySecretData   string
	}{
		Namespace:            namespace,
		IsClusterAdmin:       isClusterAdmin,
		ServiceUserName:      serviceUserName,
		ServicePassword:      servicePassword,
		Hostname:             hostname,
		UseRegistry:          registryConfig.UseRegistry,
		Registry:             registryConfig.Registry,
		UseImagePullSecrets:  registryConfig.UseRegistrySecret,
		ImagePullSecretsName: imagePullSecretsNameKubeFATE,
		RegistrySecretData:   registrySecretData,
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetIngressControllerDeploymentYAML returns the default ingress controller deployment yaml
func (s *EndpointService) GetIngressControllerDeploymentYAML(mode entity.EndpointKubeFATEIngressControllerServiceMode) (string, error) {
	if mode == entity.EndpointKubeFATEIngressControllerServiceModeSkip {
		return "", nil
	}
	// TODO: specify namespace
	t, err := template.New("ingress-nginx").Parse(kubefate.GetDefaultIngressControllerYAML())
	if err != nil {
		return "", err
	}
	serviceType := "NodePort"
	if mode == entity.EndpointKubeFATEIngressControllerServiceModeLoadBalancer {
		serviceType = "LoadBalancer"
	}
	data := struct {
		ServiceType string
	}{
		ServiceType: serviceType,
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ensureEndpointExist returns try to add/install a KubeFATE into the specified cluster
func (s *EndpointService) ensureEndpointExist(infraUUID string, namespace string) (string, error) {
	endpointScanResult, err := s.FindKubeFATEEndpoint(infraUUID, namespace)
	if err != nil {
		return "", err
	}
	install := true
	if len(endpointScanResult) > 0 {
		scannedEndpoint := endpointScanResult[0]
		if scannedEndpoint.IsManaged {
			log.Info().Msgf("re-use endpoint %s(%s)", scannedEndpoint.Name, scannedEndpoint.UUID)
			return scannedEndpoint.UUID, nil
		}
		if scannedEndpoint.IsCompatible {
			install = false
		} else {
			return "", errors.New("infra contains in-compatible KubeFATE")
		}
	}
	providerInstance, err := s.InfraProviderKubernetesRepo.GetByUUID(infraUUID)
	if err != nil {
		return "", errors.Wrapf(err, "error query infra provider")
	}
	provider := providerInstance.(*entity.InfraProviderKubernetes)
	u, err := url.Parse(provider.APIHost)
	if err != nil {
		return "", err
	}
	endpointUUID, err := s.CreateKubeFATEEndpoint(infraUUID, namespace,
		fmt.Sprintf("kubefate-%s", u.Hostname()),
		fmt.Sprintf("Automatically added KubeFATE on Kubernetes %s.", u.Hostname()),
		"",
		install,
		entity.EndpointKubeFATEIngressControllerServiceModeModeNonexistent)

	var installErr error
	if err := utils.ExecuteWithTimeout(func() bool {
		log.Info().Msgf("checking kubefate endpoint status...")
		instance, err := s.EndpointKubeFATERepo.GetByUUID(endpointUUID)
		if err != nil {
			log.Err(err).Msgf("failed to query endpoint")
			return false
		}
		endpoint := instance.(*entity.EndpointKubeFATE)
		if endpoint.Status == entity.EndpointStatusCreating {
			log.Info().Msgf("endpoint status is %v, continue querying", endpoint.Status)
			return false
		}
		if endpoint.Status != entity.EndpointStatusReady {
			installErr = errors.Errorf("invalid endpoint status: %v, abort", endpoint.Status)
		}
		return true
	}, time.Minute*30, time.Second*10); err != nil {
		return "", errors.Wrapf(err, "error checking kubefate ingress deployment")
	}
	return endpointUUID, installErr
}
