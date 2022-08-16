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
	"github.com/FederatedAI/FedLCM/server/domain/service"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
)

// ParticipantOpenFLListItem contains basic info of an OpenFL participant
type ParticipantOpenFLListItem struct {
	UUID              string                                   `json:"uuid"`
	Name              string                                   `json:"name"`
	Description       string                                   `json:"description"`
	CreatedAt         time.Time                                `json:"created_at"`
	Type              entity.ParticipantOpenFLType             `json:"type"`
	EndpointName      string                                   `json:"endpoint_name"`
	EndpointUUID      string                                   `json:"endpoint_uuid"`
	InfraProviderName string                                   `json:"infra_provider_name"`
	InfraProviderUUID string                                   `json:"infra_provider_uuid"`
	Namespace         string                                   `json:"namespace"`
	ClusterUUID       string                                   `json:"cluster_uuid"`
	Status            entity.ParticipantOpenFLStatus           `json:"status"`
	AccessInfo        entity.ParticipantOpenFLModulesAccessMap `json:"access_info"`
	TokenStr          string                                   `json:"token_str"`
	TokenName         string                                   `json:"token_name"`
	Labels            valueobject.Labels                       `json:"labels"`
}

// ParticipantOpenFLListInFederation contains all the participants in an OpenFL federation
type ParticipantOpenFLListInFederation struct {
	Director *ParticipantOpenFLListItem   `json:"director"`
	Envoy    []*ParticipantOpenFLListItem `json:"envoy"`
}

// OpenFLDirectorDetail contains detailed info of an OpenFL director
type OpenFLDirectorDetail struct {
	ParticipantOpenFLListItem
	ChartUUID              string                              `json:"chart_uuid"`
	DeploymentYAML         string                              `json:"deployment_yaml"`
	DirectorServerCertInfo entity.ParticipantComponentCertInfo `json:"director_server_cert_info"`
	JupyterClientCertInfo  entity.ParticipantComponentCertInfo `json:"jupyter_client_cert_info"`
}

// OpenFLEnvoyDetail contains detailed info of an OpenFL envoy
type OpenFLEnvoyDetail struct {
	ParticipantOpenFLListItem
	ChartUUID           string                              `json:"chart_uuid"`
	EnvoyClientCertInfo entity.ParticipantComponentCertInfo `json:"envoy_client_cert_info"`
}

func (app *ParticipantApp) getOpenFLDomainService() *service.ParticipantOpenFLService {
	eventService := &service.EventService{
		EventRepo: app.EventRepo,
	}
	return &service.ParticipantOpenFLService{
		ParticipantOpenFLRepo: app.ParticipantOpenFLRepo,
		TokenRepo:             app.RegistrationTokenOpenFLRepo,
		InfraRepo:             app.InfraProviderKubernetesRepo,
		ParticipantService: service.ParticipantService{
			FederationRepo: app.FederationOpenFLRepo,
			ChartRepo:      app.ChartRepo,
			CertificateService: &service.CertificateService{
				CertificateAuthorityRepo: app.CertificateAuthorityRepo,
				CertificateRepo:          app.CertificateRepo,
				CertificateBindingRepo:   app.CertificateBindingRepo,
			},
			EventService: eventService,
			EndpointService: &service.EndpointService{
				InfraProviderKubernetesRepo: app.InfraProviderKubernetesRepo,
				EndpointKubeFATERepo:        app.EndpointKubeFATERepo,
				ParticipantFATERepo:         app.ParticipantFATERepo,
				ParticipantOpenFLRepo:       app.ParticipantOpenFLRepo,
				EventService:                eventService,
			},
		},
	}
}

// GetOpenFLDirectorDeploymentYAML returns the deployment yaml for an OpenFL director
func (app *ParticipantApp) GetOpenFLDirectorDeploymentYAML(req *service.ParticipantOpenFLDirectorYAMLCreationRequest) (string, error) {
	return app.getOpenFLDomainService().GetOpenFLDirectorYAML(req)
}

// CreateOpenFLDirector creates an OpenFL director using the specified endpoint
func (app *ParticipantApp) CreateOpenFLDirector(req *service.ParticipantOpenFLDirectorCreationRequest) (string, error) {
	director, _, err := app.getOpenFLDomainService().CreateDirector(req)
	if err != nil {
		return "", err
	}
	return director.UUID, err
}

// HandleOpenFLEnvoyRegistration handles registration request from an Envoy node
func (app *ParticipantApp) HandleOpenFLEnvoyRegistration(req *service.ParticipantOpenFLEnvoyRegistrationRequest) (string, error) {
	envoy, err := app.getOpenFLDomainService().HandleRegistrationRequest(req)
	if err != nil {
		return "", err
	}
	return envoy.UUID, err
}

// RemoveOpenFLDirector removes and uninstalls an OpenFL director deployment
func (app *ParticipantApp) RemoveOpenFLDirector(uuid string, force bool) error {
	_, err := app.getOpenFLDomainService().RemoveDirector(uuid, force)
	return err
}

// RemoveOpenFLEnvoy removes and uninstalls an OpenFL envoy deployment
func (app *ParticipantApp) RemoveOpenFLEnvoy(uuid string, force bool) error {
	return app.getOpenFLDomainService().RemoveEnvoy(uuid, force)
}

// GetOpenFLParticipantList returns the current participants in an OpenFL federation
func (app *ParticipantApp) GetOpenFLParticipantList(federationUUID string) (*ParticipantOpenFLListInFederation, error) {
	var participants ParticipantOpenFLListInFederation
	instanceList, err := app.ParticipantOpenFLRepo.ListByFederationUUID(federationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list participant by federation")
	}
	domainParticipantList := instanceList.([]entity.ParticipantOpenFL)

	for _, domainParticipant := range domainParticipantList {
		item := &ParticipantOpenFLListItem{
			UUID:              domainParticipant.UUID,
			Name:              domainParticipant.Name,
			Description:       domainParticipant.Description,
			CreatedAt:         domainParticipant.CreatedAt,
			Type:              domainParticipant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      domainParticipant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			Namespace:         domainParticipant.Namespace,
			ClusterUUID:       domainParticipant.ClusterUUID,
			Status:            domainParticipant.Status,
			AccessInfo:        domainParticipant.AccessInfo,
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
		if domainParticipant.Type == entity.ParticipantOpenFLTypeDirector {
			participants.Director = item
		} else {
			item.Labels = domainParticipant.Labels
			item.TokenName = "Unknown"
			item.TokenStr = "Unknown"
			if instance, err := app.RegistrationTokenOpenFLRepo.GetByUUID(domainParticipant.TokenUUID); err == nil {
				token := instance.(*entity.RegistrationTokenOpenFL)
				item.TokenStr = token.Display()
				item.TokenName = token.Name
			}
			participants.Envoy = append(participants.Envoy, item)
		}
	}
	return &participants, nil
}

// GetOpenFLDirectorDetail returns the detailed information of a OpenFL director
func (app *ParticipantApp) GetOpenFLDirectorDetail(uuid string) (*OpenFLDirectorDetail, error) {
	participantInstance, err := app.ParticipantOpenFLRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	participant := participantInstance.(*entity.ParticipantOpenFL)
	participantDetail := &OpenFLDirectorDetail{
		ParticipantOpenFLListItem: ParticipantOpenFLListItem{
			UUID:              participant.UUID,
			Name:              participant.Name,
			Description:       participant.Description,
			CreatedAt:         participant.CreatedAt,
			Type:              participant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      participant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			Namespace:         participant.Namespace,
			ClusterUUID:       participant.ClusterUUID,
			Status:            participant.Status,
			AccessInfo:        participant.AccessInfo,
		},
		ChartUUID:              participant.ChartUUID,
		DeploymentYAML:         participant.DeploymentYAML,
		DirectorServerCertInfo: participant.CertConfig.DirectorServerCertInfo,
		JupyterClientCertInfo:  participant.CertConfig.JupyterClientCertInfo,
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

// GetOpenFLEnvoyDetail returns the detailed information of a OpenFL envoy
func (app *ParticipantApp) GetOpenFLEnvoyDetail(uuid string) (*OpenFLEnvoyDetail, error) {
	participantInstance, err := app.ParticipantOpenFLRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	participant := participantInstance.(*entity.ParticipantOpenFL)
	participantDetail := &OpenFLEnvoyDetail{
		ParticipantOpenFLListItem: ParticipantOpenFLListItem{
			UUID:              participant.UUID,
			Name:              participant.Name,
			Description:       participant.Description,
			CreatedAt:         participant.CreatedAt,
			Type:              participant.Type,
			EndpointName:      "Unknown",
			EndpointUUID:      participant.EndpointUUID,
			InfraProviderName: "Unknown",
			InfraProviderUUID: "Unknown",
			Namespace:         participant.Namespace,
			ClusterUUID:       participant.ClusterUUID,
			Status:            participant.Status,
			AccessInfo:        participant.AccessInfo,
			TokenStr:          "Unknown",
			TokenName:         "Unknown",
			Labels:            participant.Labels,
		},
		ChartUUID:           participant.ChartUUID,
		EnvoyClientCertInfo: participant.CertConfig.EnvoyClientCertInfo,
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
	if tokenInstance, err := app.RegistrationTokenOpenFLRepo.GetByUUID(participant.TokenUUID); err == nil {
		token := tokenInstance.(*entity.RegistrationTokenOpenFL)
		participantDetail.TokenStr = token.Display()
		participantDetail.TokenName = token.Name
	}
	return participantDetail, nil
}

// GetOpenFLEnvoyDetailWithTokenVerification returns the detailed information of a OpenFL envoy if the supplied token string is correct
func (app *ParticipantApp) GetOpenFLEnvoyDetailWithTokenVerification(uuid, tokenStr string) (*OpenFLEnvoyDetail, error) {
	envoy, err := app.GetOpenFLEnvoyDetail(uuid)
	if err != nil {
		return nil, err
	}
	if envoy.TokenStr != tokenStr {
		return nil, errors.New("invalid token")
	}
	return envoy, nil
}
