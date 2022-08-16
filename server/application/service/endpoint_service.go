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
	"fmt"
	"time"

	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/service"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// EndpointApp provide functions to manage the endpoints
type EndpointApp struct {
	InfraProviderKubernetesRepo repo.InfraProviderRepository
	EndpointKubeFAETRepo        repo.EndpointRepository
	ParticipantFATERepo         repo.ParticipantFATERepository
	ParticipantOpenFLRepo       repo.ParticipantOpenFLRepository
	EventRepo                   repo.EventRepository
}

// EndpointListItem contains basic information of an endpoint
type EndpointListItem struct {
	UUID              string                `json:"uuid"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	Type              entity.EndpointType   `json:"type"`
	CreatedAt         time.Time             `json:"created_at"`
	InfraProviderName string                `json:"infra_provider_name"`
	InfraProviderUUID string                `json:"infra_provider_uuid"`
	KubeFATEHost      string                `json:"kubefate_host"`
	KubeFATEAddress   string                `json:"kubefate_address"`
	KubeFATEVersion   string                `json:"kubefate_version"`
	Status            entity.EndpointStatus `json:"status"`
}

// EndpointDetail contains basic information of an endpoint as well as additional information
type EndpointDetail struct {
	EndpointListItem
	KubeFATEDeploymentYAML string `json:"kubefate_deployment_yaml"`
}

// EndpointScanItem contains basic information of an endpoint that is scanned from an infra provider
type EndpointScanItem struct {
	EndpointListItem
	IsManaged    bool `json:"is_managed"`
	IsCompatible bool `json:"is_compatible"`
}

// EndpointScanRequest contains necessary request info to start an endpoint scan
type EndpointScanRequest struct {
	InfraProviderUUID string              `json:"infra_provider_uuid"`
	Type              entity.EndpointType `json:"type"`
}

// EndpointCreationRequest contains necessary request info to create an endpoint
type EndpointCreationRequest struct {
	EndpointScanRequest
	Name                         string                                              `json:"name"`
	Description                  string                                              `json:"description"`
	Install                      bool                                                `json:"install"`
	IngressControllerServiceMode entity.EndpointKubeFATEIngressControllerServiceMode `json:"ingress_controller_service_mode"`
	KubeFATEDeploymentYAML       string                                              `json:"kubefate_deployment_yaml"`
}

// GetEndpointList returns currently managed endpoints
func (app *EndpointApp) GetEndpointList() ([]EndpointListItem, error) {
	endpointListInstance, err := app.EndpointKubeFAETRepo.List()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get KubeFATE endpoint list")
	}
	domainEndpointList := endpointListInstance.([]entity.EndpointKubeFATE)

	var endpointList []EndpointListItem
	for _, domainEndpointKubeFATE := range domainEndpointList {
		endpoint := EndpointListItem{
			UUID:              domainEndpointKubeFATE.UUID,
			Name:              domainEndpointKubeFATE.Name,
			Description:       domainEndpointKubeFATE.Description,
			Type:              entity.EndpointTypeKubeFATE,
			CreatedAt:         domainEndpointKubeFATE.CreatedAt,
			InfraProviderName: "Unknown",
			InfraProviderUUID: fmt.Sprintf("Unknown (%s)", domainEndpointKubeFATE.InfraProviderUUID),
			KubeFATEHost:      domainEndpointKubeFATE.Config.IngressRuleHost,
			KubeFATEAddress:   domainEndpointKubeFATE.Config.IngressAddress,
			KubeFATEVersion:   domainEndpointKubeFATE.Version,
			Status:            domainEndpointKubeFATE.Status,
		}
		domainProviderInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(domainEndpointKubeFATE.InfraProviderUUID)
		if err != nil {
			log.Err(err).Msg("failed to query provider")
		} else {
			domainProvider := domainProviderInstance.(*entity.InfraProviderKubernetes)
			endpoint.InfraProviderUUID = domainProvider.UUID
			endpoint.InfraProviderName = domainProvider.Name
		}
		endpointList = append(endpointList, endpoint)
	}
	return endpointList, nil
}

// GetEndpointDetail returns detailed information of an endpoint
func (app *EndpointApp) GetEndpointDetail(uuid string) (*EndpointDetail, error) {
	endpointInstance, err := app.EndpointKubeFAETRepo.GetByUUID(uuid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get KubeFAET endpoint instance")
	}
	domainEndpointKubeFATE := endpointInstance.(*entity.EndpointKubeFATE)

	endpoint := &EndpointDetail{
		EndpointListItem: EndpointListItem{
			UUID:              domainEndpointKubeFATE.UUID,
			Name:              domainEndpointKubeFATE.Name,
			Description:       domainEndpointKubeFATE.Description,
			Type:              entity.EndpointTypeKubeFATE,
			CreatedAt:         domainEndpointKubeFATE.CreatedAt,
			InfraProviderName: "Unknown",
			InfraProviderUUID: fmt.Sprintf("Unknown (%s)", domainEndpointKubeFATE.InfraProviderUUID),
			KubeFATEHost:      domainEndpointKubeFATE.Config.IngressRuleHost,
			KubeFATEAddress:   domainEndpointKubeFATE.Config.IngressAddress,
			KubeFATEVersion:   domainEndpointKubeFATE.Version,
			Status:            domainEndpointKubeFATE.Status,
		},
		KubeFATEDeploymentYAML: domainEndpointKubeFATE.DeploymentYAML,
	}
	domainProviderInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(domainEndpointKubeFATE.InfraProviderUUID)
	if err != nil {
		log.Err(err).Msg("failed to query provider")
	} else {
		domainProvider := domainProviderInstance.(*entity.InfraProviderKubernetes)
		endpoint.InfraProviderUUID = domainProvider.UUID
		endpoint.InfraProviderName = domainProvider.Name
	}
	return endpoint, nil
}

// DeleteEndpoint removes the endpoint
func (app *EndpointApp) DeleteEndpoint(uuid string, uninstall bool) error {
	return app.getDomainService().RemoveEndpoint(uuid, uninstall)
}

// ScanEndpoint returns endpoints installed in an infra provider
func (app *EndpointApp) ScanEndpoint(req *EndpointScanRequest) ([]EndpointScanItem, error) {
	if req.InfraProviderUUID == "" {
		return nil, errors.New("missing providerUUID")
	}
	if req.Type == "" {
		return nil, errors.New("missing endpointType")
	}

	domainProviderInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(req.InfraProviderUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query provider")
	}
	domainProvider := domainProviderInstance.(*entity.InfraProviderKubernetes)

	scanResult, err := app.getDomainService().FindKubeFATEEndpoint(req.InfraProviderUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to scan the endpoints")
	}
	var itemList []EndpointScanItem

	for _, resultItem := range scanResult {
		itemList = append(itemList, EndpointScanItem{
			EndpointListItem: EndpointListItem{
				UUID:              resultItem.UUID,
				Name:              resultItem.Name,
				Description:       resultItem.Description,
				Type:              resultItem.Type,
				CreatedAt:         resultItem.CreatedAt,
				InfraProviderName: domainProvider.Name,
				InfraProviderUUID: domainProvider.UUID,
				// ignore kubefate related info
				KubeFATEHost:    "",
				KubeFATEAddress: "",
				Status:          resultItem.Status,
			},
			IsManaged:    resultItem.IsManaged,
			IsCompatible: resultItem.IsCompatible,
		})
	}
	return itemList, nil
}

// CheckKubeFATEConnection tests connection to a KubeFATE endpoint
func (app *EndpointApp) CheckKubeFATEConnection(uuid string) error {
	return app.getDomainService().TestKubeFATE(uuid)
}

// GetKubeFATEDeploymentYAML returns the default yaml content for deploying KubeFATE
func (app *EndpointApp) GetKubeFATEDeploymentYAML(serviceUsername, servicePassword, hostname string, registryConfig valueobject.KubeRegistryConfig) (string, error) {
	return app.getDomainService().GetDeploymentYAML(serviceUsername, servicePassword, hostname, registryConfig)
}

// CreateEndpoint add or install an endpoint
func (app *EndpointApp) CreateEndpoint(req *EndpointCreationRequest) (string, error) {
	switch req.Type {
	case entity.EndpointTypeKubeFATE:
		return app.getDomainService().CreateKubeFATEEndpoint(req.InfraProviderUUID, req.Name, req.Description, req.KubeFATEDeploymentYAML, req.Install, req.IngressControllerServiceMode)
	default:
		return "", constants.ErrNotImplemented
	}
}

func (app *EndpointApp) getDomainService() *service.EndpointService {
	return &service.EndpointService{
		InfraProviderKubernetesRepo: app.InfraProviderKubernetesRepo,
		EndpointKubeFATERepo:        app.EndpointKubeFAETRepo,
		ParticipantFATERepo:         app.ParticipantFATERepo,
		ParticipantOpenFLRepo:       app.ParticipantOpenFLRepo,
		EventService: &service.EventService{
			EventRepo: app.EventRepo,
		},
	}
}
