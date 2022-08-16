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
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/service"
	"github.com/pkg/errors"
)

// ParticipantApp provide functions to manage the participants
type ParticipantApp struct {
	ParticipantFATERepo repo.ParticipantFATERepository
	FederationFATERepo  repo.FederationRepository

	FederationOpenFLRepo        repo.FederationRepository
	ParticipantOpenFLRepo       repo.ParticipantOpenFLRepository
	RegistrationTokenOpenFLRepo repo.RegistrationTokenRepository

	EndpointKubeFATERepo        repo.EndpointRepository
	InfraProviderKubernetesRepo repo.InfraProviderRepository
	ChartRepo                   repo.ChartRepository
	CertificateAuthorityRepo    repo.CertificateAuthorityRepository
	CertificateRepo             repo.CertificateRepository
	CertificateBindingRepo      repo.CertificateBindingRepository
	EventRepo                   repo.EventRepository
}

// ParticipantFATEListItem contains basic information of a FATE participant
type ParticipantFATEListItem struct {
	UUID              string                                 `json:"uuid"`
	Name              string                                 `json:"name"`
	Description       string                                 `json:"description"`
	CreatedAt         time.Time                              `json:"created_at"`
	Type              entity.ParticipantFATEType             `json:"type"`
	EndpointName      string                                 `json:"endpoint_name"`
	EndpointUUID      string                                 `json:"endpoint_uuid"`
	InfraProviderName string                                 `json:"infra_provider_name"`
	InfraProviderUUID string                                 `json:"infra_provider_uuid"`
	ChartUUID         string                                 `json:"chart_uuid"`
	Namespace         string                                 `json:"namespace"`
	PartyID           int                                    `json:"party_id"`
	ClusterUUID       string                                 `json:"cluster_uuid"`
	Status            entity.ParticipantFATEStatus           `json:"status"`
	AccessInfo        entity.ParticipantFATEModulesAccessMap `json:"access_info"`
	IsManaged         bool                                   `json:"is_managed"`
}

// ParticipantFATEListInFederation has all the participants in a FATE federation
type ParticipantFATEListInFederation struct {
	Exchange *ParticipantFATEListItem   `json:"exchange"`
	Clusters []*ParticipantFATEListItem `json:"clusters"`
}

// FATEExchangeDetail contains the detailed info of a FATE exchange
type FATEExchangeDetail struct {
	ParticipantFATEListItem
	DeploymentYAML           string                              `json:"deployment_yaml"`
	ProxyServerCertInfo      entity.ParticipantComponentCertInfo `json:"proxy_server_cert_info"`
	FMLManagerServerCertInfo entity.ParticipantComponentCertInfo `json:"fml_manager_server_cert_info"`
	FMLManagerClientCertInfo entity.ParticipantComponentCertInfo `json:"fml_manager_client_cert_info"`
}

// FATEClusterDetail contains the detailed info a FATE cluster
type FATEClusterDetail struct {
	ParticipantFATEListItem
	DeploymentYAML           string                              `json:"deployment_yaml"`
	IngressInfo              entity.ParticipantFATEIngressMap    `json:"ingress_info"`
	PulsarServerCertInfo     entity.ParticipantComponentCertInfo `json:"pulsar_server_cert_info"`
	SitePortalServerCertInfo entity.ParticipantComponentCertInfo `json:"site_portal_server_cert_info"`
	SitePortalClientCertInfo entity.ParticipantComponentCertInfo `json:"site_portal_client_cert_info"`
}

// CheckFATPartyID returns error if the current party id is taken in the specified federation
func (app *ParticipantApp) CheckFATPartyID(federationUUID string, partyID int) error {
	return app.getFATEDomainService().CheckPartyIDConflict(federationUUID, partyID)
}

// GetFATEExchangeDeploymentYAML returns the deployment yaml for a FATE exchange
func (app *ParticipantApp) GetFATEExchangeDeploymentYAML(req *service.ParticipantFATEExchangeYAMLCreationRequest) (string, error) {
	return app.getFATEDomainService().GetExchangeDeploymentYAML(req)
}

// GetFATEClusterDeploymentYAML returns the deployment yaml for a FATE cluster
func (app *ParticipantApp) GetFATEClusterDeploymentYAML(req *service.ParticipantFATEClusterYAMLCreationRequest) (string, error) {
	return app.getFATEDomainService().GetClusterDeploymentYAML(req)
}

// CreateFATEExchange creates a FATE exchange using the specified endpoint
func (app *ParticipantApp) CreateFATEExchange(req *service.ParticipantFATEExchangeCreationRequest) (string, error) {
	exchange, _, err := app.getFATEDomainService().CreateExchange(req)
	if err != nil {
		return "", err
	}
	return exchange.UUID, err
}

// CreateExternalFATEExchange creates an external FATE exchange
func (app *ParticipantApp) CreateExternalFATEExchange(req *service.ParticipantFATEExternalExchangeCreationRequest) (string, error) {
	exchange, err := app.getFATEDomainService().CreateExternalExchange(req)
	if err != nil {
		return "", err
	}
	return exchange.UUID, err
}

// CreateFATECluster creates a FATE cluster using the specified endpoint
func (app *ParticipantApp) CreateFATECluster(req *service.ParticipantFATEClusterCreationRequest) (string, error) {
	cluster, _, err := app.getFATEDomainService().CreateCluster(req)
	if err != nil {
		return "", err
	}
	return cluster.UUID, err
}

// CreateExternalFATECluster creates an external FATE cluster
func (app *ParticipantApp) CreateExternalFATECluster(req *service.ParticipantFATEExternalClusterCreationRequest) (string, error) {
	cluster, _, err := app.getFATEDomainService().CreateExternalCluster(req)
	if err != nil {
		return "", err
	}
	return cluster.UUID, err
}

// RemoveFATEExchange removes and uninstalls a FATE exchange deployment
func (app *ParticipantApp) RemoveFATEExchange(uuid string, force bool) error {
	_, err := app.getFATEDomainService().RemoveExchange(uuid, force)
	return err
}

// RemoveFATECluster removes and uninstalls a FATE cluster deployment
func (app *ParticipantApp) RemoveFATECluster(uuid string, force bool) error {
	_, err := app.getFATEDomainService().RemoveCluster(uuid, force)
	return err
}

func (app *ParticipantApp) getFATEDomainService() *service.ParticipantFATEService {
	return &service.ParticipantFATEService{
		ParticipantFATERepo: app.ParticipantFATERepo,
		ParticipantService: service.ParticipantService{
			FederationRepo: app.FederationFATERepo,
			ChartRepo:      app.ChartRepo,
			CertificateService: &service.CertificateService{
				CertificateAuthorityRepo: app.CertificateAuthorityRepo,
				CertificateRepo:          app.CertificateRepo,
				CertificateBindingRepo:   app.CertificateBindingRepo,
			},
			EventService: &service.EventService{
				EventRepo: app.EventRepo,
			},
			EndpointService: &service.EndpointService{
				InfraProviderKubernetesRepo: app.InfraProviderKubernetesRepo,
				EndpointKubeFATERepo:        app.EndpointKubeFATERepo,
				ParticipantFATERepo:         app.ParticipantFATERepo,
			},
		},
	}
}

// GetFATEParticipantList returns the current participants in a FATE federation
func (app *ParticipantApp) GetFATEParticipantList(federationUUID string) (*ParticipantFATEListInFederation, error) {
	var participants ParticipantFATEListInFederation
	instanceList, err := app.ParticipantFATERepo.ListByFederationUUID(federationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list participant by federation")
	}
	domainParticipantList := instanceList.([]entity.ParticipantFATE)

	for _, domainParticipant := range domainParticipantList {
		item := &ParticipantFATEListItem{
			UUID:              domainParticipant.UUID,
			Name:              domainParticipant.Name,
			Description:       domainParticipant.Description,
			CreatedAt:         domainParticipant.CreatedAt,
			Type:              domainParticipant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      domainParticipant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			ChartUUID:         domainParticipant.ChartUUID,
			Namespace:         domainParticipant.Namespace,
			PartyID:           domainParticipant.PartyID,
			ClusterUUID:       domainParticipant.ClusterUUID,
			Status:            domainParticipant.Status,
			AccessInfo:        domainParticipant.AccessInfo,
			IsManaged:         domainParticipant.IsManaged,
		}
		if endpointInstance, err := app.EndpointKubeFATERepo.GetByUUID(domainParticipant.EndpointUUID); err == nil {
			endpoint := endpointInstance.(*entity.EndpointKubeFATE)
			item.EndpointName = endpoint.Name
			item.InfraProviderUUID = endpoint.InfraProviderUUID
			if infraInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(endpoint.InfraProviderUUID); err == nil {
				infra := infraInstance.(*entity.InfraProviderKubernetes)
				item.InfraProviderName = infra.Name
			}
		}
		if domainParticipant.Type == entity.ParticipantFATETypeExchange {
			participants.Exchange = item
		} else {
			participants.Clusters = append(participants.Clusters, item)
		}
	}
	return &participants, nil
}

// GetFATEExchangeDetail returns detailed info of a exchange
func (app *ParticipantApp) GetFATEExchangeDetail(uuid string) (*FATEExchangeDetail, error) {
	participantInstance, err := app.ParticipantFATERepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	participant := participantInstance.(*entity.ParticipantFATE)
	participantDetail := &FATEExchangeDetail{
		ParticipantFATEListItem: ParticipantFATEListItem{
			UUID:              participant.UUID,
			Name:              participant.Name,
			Description:       participant.Description,
			CreatedAt:         participant.CreatedAt,
			Type:              participant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      participant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			ChartUUID:         participant.ChartUUID,
			Namespace:         participant.Namespace,
			PartyID:           participant.PartyID,
			ClusterUUID:       participant.ClusterUUID,
			Status:            participant.Status,
			AccessInfo:        participant.AccessInfo,
			IsManaged:         participant.IsManaged,
		},
		DeploymentYAML:           participant.DeploymentYAML,
		ProxyServerCertInfo:      participant.CertConfig.ProxyServerCertInfo,
		FMLManagerServerCertInfo: participant.CertConfig.FMLManagerServerCertInfo,
		FMLManagerClientCertInfo: participant.CertConfig.FMLManagerClientCertInfo,
	}
	if endpointInstance, err := app.EndpointKubeFATERepo.GetByUUID(participant.EndpointUUID); err == nil {
		endpoint := endpointInstance.(*entity.EndpointKubeFATE)
		participantDetail.EndpointName = endpoint.Name
		participantDetail.InfraProviderUUID = endpoint.InfraProviderUUID
		if infraInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(endpoint.InfraProviderUUID); err == nil {
			infra := infraInstance.(*entity.InfraProviderKubernetes)
			participantDetail.InfraProviderName = infra.Name
		}
	}
	return participantDetail, nil
}

// GetFATEClusterDetail returns detailed info of a exchange or cluster
func (app *ParticipantApp) GetFATEClusterDetail(uuid string) (*FATEClusterDetail, error) {
	participantInstance, err := app.ParticipantFATERepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	participant := participantInstance.(*entity.ParticipantFATE)
	participantDetail := &FATEClusterDetail{
		ParticipantFATEListItem: ParticipantFATEListItem{
			UUID:              participant.UUID,
			Name:              participant.Name,
			Description:       participant.Description,
			CreatedAt:         participant.CreatedAt,
			Type:              participant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      participant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			ChartUUID:         participant.ChartUUID,
			Namespace:         participant.Namespace,
			PartyID:           participant.PartyID,
			ClusterUUID:       participant.ClusterUUID,
			Status:            participant.Status,
			AccessInfo:        participant.AccessInfo,
			IsManaged:         participant.IsManaged,
		},
		DeploymentYAML:           participant.DeploymentYAML,
		IngressInfo:              participant.IngressInfo,
		PulsarServerCertInfo:     participant.CertConfig.PulsarServerCertInfo,
		SitePortalServerCertInfo: participant.CertConfig.SitePortalServerCertInfo,
		SitePortalClientCertInfo: participant.CertConfig.SitePortalClientCertInfo,
	}

	if endpointInstance, err := app.EndpointKubeFATERepo.GetByUUID(participant.EndpointUUID); err == nil {
		endpoint := endpointInstance.(*entity.EndpointKubeFATE)
		participantDetail.EndpointName = endpoint.Name
		participantDetail.InfraProviderUUID = endpoint.InfraProviderUUID
		if infraInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(endpoint.InfraProviderUUID); err == nil {
			infra := infraInstance.(*entity.InfraProviderKubernetes)
			participantDetail.InfraProviderName = infra.Name
		}
	}
	return participantDetail, nil
}
