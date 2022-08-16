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
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// FederationApp provides application level API for federation related actions
type FederationApp struct {
	FederationFATERepo          repo.FederationRepository
	ParticipantFATERepo         repo.ParticipantFATERepository
	FederationOpenFLRepo        repo.FederationRepository
	ParticipantOpenFLRepo       repo.ParticipantOpenFLRepository
	RegistrationTokenOpenFLRepo repo.RegistrationTokenRepository
}

// FederationListItem contains basic info of a federation
type FederationListItem struct {
	UUID        string                `json:"uuid"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Type        entity.FederationType `json:"type"`
	CreatedAt   time.Time             `json:"created_at"`
}

// FederationCreationRequest contains basic info for creating a federation
type FederationCreationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// FederationFATECreationRequest contains necessary info to create a FATE federation
type FederationFATECreationRequest struct {
	FederationCreationRequest
	Domain string `json:"domain"`
}

// FederationFATEDetail contains specific info for a FATE federation
type FederationFATEDetail struct {
	FederationListItem
	Domain string `json:"domain"`
}

// List returns all saved federation
func (app *FederationApp) List() ([]FederationListItem, error) {
	instanceList, err := app.FederationFATERepo.List()
	if err != nil {
		return nil, err
	}
	domainFederationList := instanceList.([]entity.FederationFATE)
	var federationList []FederationListItem
	for _, domainFederation := range domainFederationList {
		federationList = append(federationList, FederationListItem{
			UUID:        domainFederation.UUID,
			Name:        domainFederation.Name,
			Description: domainFederation.Description,
			Type:        domainFederation.Type,
			CreatedAt:   domainFederation.CreatedAt,
		})
	}

	instanceList, err = app.FederationOpenFLRepo.List()
	if err != nil {
		return nil, err
	}
	domainFederationOpenFLList := instanceList.([]entity.FederationOpenFL)
	for _, domainFederation := range domainFederationOpenFLList {
		federationList = append(federationList, FederationListItem{
			UUID:        domainFederation.UUID,
			Name:        domainFederation.Name,
			Description: domainFederation.Description,
			Type:        domainFederation.Type,
			CreatedAt:   domainFederation.CreatedAt,
		})
	}
	return federationList, nil
}

// GetFATEFederation returns basic info of a specific FATE federation
func (app *FederationApp) GetFATEFederation(uuid string) (*FederationFATEDetail, error) {
	instance, err := app.FederationFATERepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	domainFederation := instance.(*entity.FederationFATE)
	return &FederationFATEDetail{
		FederationListItem: FederationListItem{
			UUID:        domainFederation.UUID,
			Name:        domainFederation.Name,
			Description: domainFederation.Description,
			Type:        domainFederation.Type,
			CreatedAt:   domainFederation.CreatedAt,
		},
		Domain: domainFederation.Domain,
	}, nil
}

// CreateFATEFederation creates a FATE federation
func (app *FederationApp) CreateFATEFederation(req *FederationFATECreationRequest) (string, error) {
	federation := &entity.FederationFATE{
		Federation: entity.Federation{
			UUID:        uuid.NewV4().String(),
			Name:        req.Name,
			Description: req.Description,
			Type:        entity.FederationTypeFATE,
			Repo:        app.FederationFATERepo,
		},
		Domain: req.Domain,
	}
	if err := federation.Create(); err != nil {
		return "", err
	}
	return federation.UUID, nil
}

// DeleteFATEFederation deletes a FATE federation
func (app *FederationApp) DeleteFATEFederation(uuid string) error {
	// XXX: this check should be placed in the domain level
	instanceList, err := app.ParticipantFATERepo.ListByFederationUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "failed to query federation participants")
	}
	participantList := instanceList.([]entity.ParticipantFATE)
	if len(participantList) > 0 {
		return errors.Errorf("cannot remove federation that still contains %v participants", len(participantList))
	}
	return app.FederationFATERepo.DeleteByUUID(uuid)
}
