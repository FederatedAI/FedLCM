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
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// FederationOpenFLCreationRequest contains necessary info to create an OpenFL federation
type FederationOpenFLCreationRequest struct {
	FederationCreationRequest
	Domain                       string                             `json:"domain"`
	UseCustomizedShardDescriptor bool                               `json:"use_customized_shard_descriptor"`
	ShardDescriptorConfig        *valueobject.ShardDescriptorConfig `json:"shard_descriptor_config"`
}

// FederationOpenFLDetail contains specific info for an OpenFL federation
type FederationOpenFLDetail struct {
	FederationListItem
	Domain                       string                             `json:"domain"`
	UseCustomizedShardDescriptor bool                               `json:"use_customized_shard_descriptor"`
	ShardDescriptorConfig        *valueobject.ShardDescriptorConfig `json:"shard_descriptor_config"`
}

// RegistrationTokenOpenFLBasicInfo contains necessary info to generate token for an OpenFL federation
type RegistrationTokenOpenFLBasicInfo struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	ExpiredAt   time.Time          `json:"expired_at"`
	Limit       int                `json:"limit"`
	Used        int                `json:"used"`
	Labels      valueobject.Labels `json:"labels"`
}

type RegistrationTokenOpenFLListItem struct {
	RegistrationTokenOpenFLBasicInfo
	UUID              string `json:"uuid"`
	DisplayedTokenStr string `json:"token_str"`
}

// CreateOpenFLFederation creates an OpenFL federation entity
func (app *FederationApp) CreateOpenFLFederation(req *FederationOpenFLCreationRequest) (string, error) {
	federation := &entity.FederationOpenFL{
		Federation: entity.Federation{
			UUID:        uuid.NewV4().String(),
			Name:        req.Name,
			Description: req.Description,
			Type:        entity.FederationTypeOpenFL,
			Repo:        app.FederationOpenFLRepo,
		},
		Domain:                       req.Domain,
		UseCustomizedShardDescriptor: req.UseCustomizedShardDescriptor,
		ShardDescriptorConfig:        req.ShardDescriptorConfig,
	}
	if err := federation.Create(); err != nil {
		return "", err
	}
	return federation.UUID, nil
}

// DeleteOpenFLFederation deletes an OpenFL federation
func (app *FederationApp) DeleteOpenFLFederation(uuid string) error {
	//check participant existence
	instanceList, err := app.ParticipantOpenFLRepo.ListByFederationUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "failed to list federation participants")
	}
	participantList := instanceList.([]entity.ParticipantOpenFL)
	if len(participantList) > 0 {
		return errors.Errorf("cannot remove federation that still contains %v OpenFL participants", len(participantList))
	}
	if err := app.RegistrationTokenOpenFLRepo.DeleteByFederation(uuid); err != nil {
		return errors.Wrap(err, "failed to clean up tokens")
	}
	return app.FederationOpenFLRepo.DeleteByUUID(uuid)
}

// GetOpenFLFederation returns basic info of a specific OpenFL federation
func (app *FederationApp) GetOpenFLFederation(uuid string) (*FederationOpenFLDetail, error) {
	instance, err := app.FederationOpenFLRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	domainFederation := instance.(*entity.FederationOpenFL)
	return &FederationOpenFLDetail{
		FederationListItem: FederationListItem{
			UUID:        domainFederation.UUID,
			Name:        domainFederation.Name,
			Description: domainFederation.Description,
			Type:        domainFederation.Type,
			CreatedAt:   domainFederation.CreatedAt,
		},
		Domain:                       domainFederation.Domain,
		UseCustomizedShardDescriptor: domainFederation.UseCustomizedShardDescriptor,
		ShardDescriptorConfig:        domainFederation.ShardDescriptorConfig,
	}, nil
}

// GeneratedOpenFLToken creates a new token for an OpenFL federation
func (app *FederationApp) GeneratedOpenFLToken(req *RegistrationTokenOpenFLBasicInfo, federationUUID string) error {
	token := &entity.RegistrationTokenOpenFL{
		RegistrationToken: entity.RegistrationToken{
			Name:        req.Name,
			Description: req.Description,
			Repo:        app.RegistrationTokenOpenFLRepo,
		},
		ExpiredAt:      req.ExpiredAt,
		Limit:          req.Limit,
		Labels:         req.Labels,
		FederationUUID: federationUUID,
	}
	return token.Create()
}

// ListOpenFLToken list all registration tokens in an OpenFL federation
func (app *FederationApp) ListOpenFLToken(federationUUID string) ([]RegistrationTokenOpenFLListItem, error) {
	instanceList, err := app.RegistrationTokenOpenFLRepo.ListByFederation(federationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query token list")
	}
	domainTokenList := instanceList.([]entity.RegistrationTokenOpenFL)
	var tokenList []RegistrationTokenOpenFLListItem
	for _, token := range domainTokenList {
		count, err := app.ParticipantOpenFLRepo.CountByTokenUUID(token.UUID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to query token count")
		}
		tokenList = append(tokenList, RegistrationTokenOpenFLListItem{
			RegistrationTokenOpenFLBasicInfo: RegistrationTokenOpenFLBasicInfo{
				Name:        token.Name,
				Description: token.Description,
				ExpiredAt:   token.ExpiredAt,
				Limit:       token.Limit,
				Used:        count,
				Labels:      token.Labels,
			},
			UUID:              token.UUID,
			DisplayedTokenStr: token.Display(),
		})
	}
	return tokenList, nil
}

// DeleteOpenFLToken removes the specified token
func (app *FederationApp) DeleteOpenFLToken(uuid, federationUUID string) error {
	// TODO: validate federation UUID
	return app.RegistrationTokenOpenFLRepo.DeleteByUUID(uuid)
}
