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

package gorm

import (
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/pkg/errors"
)

// ParticipantFATERepo implements the repo.ParticipantFATERepository interface
type ParticipantFATERepo struct{}

var _ repo.ParticipantFATERepository = (*ParticipantFATERepo)(nil)

// ErrParticipantExist means new participant cannot be created due to name conflicts
var ErrParticipantExist = errors.New("participant already exists")

func (r *ParticipantFATERepo) Create(instance interface{}) error {
	var count int64
	participant := instance.(*entity.ParticipantFATE)
	db.Model(&entity.ParticipantFATE{}).Where("name = ? AND federation_uuid = ?", participant.Name, participant.FederationUUID).Count(&count)
	if count > 0 {
		return ErrParticipantExist
	}

	if err := db.Create(participant).Error; err != nil {
		return err
	}
	return nil
}

func (r *ParticipantFATERepo) List() (interface{}, error) {
	var participantList []entity.ParticipantFATE
	if err := db.Find(&participantList).Error; err != nil {
		return nil, err
	}
	return participantList, nil
}

func (r *ParticipantFATERepo) DeleteByUUID(uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&entity.ParticipantFATE{}).Error
}

func (r *ParticipantFATERepo) GetByUUID(uuid string) (interface{}, error) {
	participant := &entity.ParticipantFATE{}
	if err := db.Where("uuid = ?", uuid).First(participant).Error; err != nil {
		return nil, err
	}
	return participant, nil
}

func (r *ParticipantFATERepo) ListByFederationUUID(federationUUID string) (interface{}, error) {
	var participants []entity.ParticipantFATE
	err := db.Where("federation_uuid = ?", federationUUID).Find(&participants).Error
	if err != nil {
		return 0, err
	}
	return participants, nil
}

func (r *ParticipantFATERepo) ListByEndpointUUID(endpointUUID string) (interface{}, error) {
	var participants []entity.ParticipantFATE
	err := db.Where("endpoint_uuid = ?", endpointUUID).Find(&participants).Error
	if err != nil {
		return 0, err
	}
	return participants, nil
}

func (r *ParticipantFATERepo) UpdateStatusByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantFATE)
	return db.Model(&entity.ParticipantFATE{}).Where("uuid = ?", participant.UUID).
		Update("status", participant.Status).Error
}

func (r *ParticipantFATERepo) UpdateDeploymentYAMLByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantFATE)
	return db.Model(&entity.ParticipantFATE{}).Where("uuid = ?", participant.UUID).
		Update("deployment_yaml", participant.DeploymentYAML).Error
}

func (r *ParticipantFATERepo) UpdateInfoByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantFATE)
	return db.Where("uuid = ?", participant.UUID).
		Select("cluster_uuid", "status", "access_info", "cert_config", "extra_attribute", "ingress_info", "job_uuid", "deployment_yaml", "chart_uuid").
		Updates(participant).Error
}

func (r *ParticipantFATERepo) GetExchangeByFederationUUID(uuid string) (interface{}, error) {
	participant := &entity.ParticipantFATE{}
	if err := db.Where("federation_uuid = ? AND type = ?", uuid, entity.ParticipantFATETypeExchange).First(participant).Error; err != nil {
		return nil, err
	}
	return participant, nil
}

func (r *ParticipantFATERepo) IsExchangeCreatedByFederationUUID(uuid string) (bool, error) {
	var count int64
	if err := db.Model(&entity.ParticipantFATE{}).Where("federation_uuid = ? AND type = ?", uuid, entity.ParticipantFATETypeExchange).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ParticipantFATERepo) IsConflictedByFederationUUIDAndPartyID(federationUUID string, partyID int) (bool, error) {
	var count int64
	err := db.Model(&entity.ParticipantFATE{}).Where("federation_uuid = ? AND party_id = ?", federationUUID, partyID).Count(&count).Error
	return count > 0, err
}

// InitTable makes sure the table is created in the db
func (r *ParticipantFATERepo) InitTable() {
	if err := db.AutoMigrate(entity.ParticipantFATE{}); err != nil {
		panic(err)
	}
}
