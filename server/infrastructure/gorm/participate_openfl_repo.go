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
)

// ParticipantOpenFLRepo implements repo.ParticipantOpenFLRepository interface
type ParticipantOpenFLRepo struct{}

var _ repo.ParticipantOpenFLRepository = (*ParticipantOpenFLRepo)(nil)

func (r *ParticipantOpenFLRepo) Create(instance interface{}) error {
	var count int64
	participant := instance.(*entity.ParticipantOpenFL)
	db.Model(&entity.ParticipantOpenFL{}).Where("name = ? AND federation_uuid = ?", participant.Name, participant.FederationUUID).Count(&count)
	if count > 0 {
		return ErrParticipantExist
	}

	if err := db.Create(participant).Error; err != nil {
		return err
	}
	return nil
}

func (r *ParticipantOpenFLRepo) List() (interface{}, error) {
	var participantList []entity.ParticipantOpenFL
	if err := db.Find(&participantList).Error; err != nil {
		return nil, err
	}
	return participantList, nil
}

func (r *ParticipantOpenFLRepo) DeleteByUUID(uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&entity.ParticipantOpenFL{}).Error
}

func (r *ParticipantOpenFLRepo) GetByUUID(uuid string) (interface{}, error) {
	participant := &entity.ParticipantOpenFL{}
	if err := db.Where("uuid = ?", uuid).First(participant).Error; err != nil {
		return nil, err
	}
	return participant, nil
}

func (r *ParticipantOpenFLRepo) ListByFederationUUID(federationUUID string) (interface{}, error) {
	var participants []entity.ParticipantOpenFL
	err := db.Where("federation_uuid = ?", federationUUID).Find(&participants).Error
	if err != nil {
		return 0, err
	}
	return participants, nil
}

func (r *ParticipantOpenFLRepo) ListByEndpointUUID(endpointUUID string) (interface{}, error) {
	var participants []entity.ParticipantOpenFL
	err := db.Where("endpoint_uuid = ?", endpointUUID).Find(&participants).Error
	if err != nil {
		return 0, err
	}
	return participants, nil
}

func (r *ParticipantOpenFLRepo) UpdateStatusByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantOpenFL)
	return db.Model(&entity.ParticipantOpenFL{}).Where("uuid = ?", participant.UUID).
		Update("status", participant.Status).Error
}

func (r *ParticipantOpenFLRepo) UpdateDeploymentYAMLByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantOpenFL)
	return db.Model(&entity.ParticipantOpenFL{}).Where("uuid = ?", participant.UUID).
		Update("deployment_yaml", participant.DeploymentYAML).Error
}

func (r *ParticipantOpenFLRepo) UpdateInfoByUUID(instance interface{}) error {
	participant := instance.(*entity.ParticipantOpenFL)
	return db.Where("uuid = ?", participant.UUID).
		Select("endpoint_uuid", "cluster_uuid", "status", "access_info", "cert_config", "extra_attribute", "job_uuid").
		Updates(participant).Error
}

func (r *ParticipantOpenFLRepo) IsDirectorCreatedByFederationUUID(uuid string) (bool, error) {
	var count int64
	if err := db.Model(&entity.ParticipantOpenFL{}).Where("federation_uuid = ? AND type = ?", uuid, entity.ParticipantOpenFLTypeDirector).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ParticipantOpenFLRepo) CountByTokenUUID(uuid string) (int, error) {
	var count int64
	if err := db.Unscoped().Model(&entity.ParticipantOpenFL{}).Where("token_uuid = ?", uuid).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *ParticipantOpenFLRepo) GetDirectorByFederationUUID(uuid string) (interface{}, error) {
	participant := &entity.ParticipantOpenFL{}
	if err := db.Where("federation_uuid = ? AND type = ?", uuid, entity.ParticipantOpenFLTypeDirector).First(participant).Error; err != nil {
		return nil, err
	}
	return participant, nil
}

// InitTable makes sure the table is created in the db
func (r *ParticipantOpenFLRepo) InitTable() {
	if err := db.AutoMigrate(entity.ParticipantOpenFL{}); err != nil {
		panic(err)
	}
}
