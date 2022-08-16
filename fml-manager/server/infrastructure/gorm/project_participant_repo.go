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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// ProjectParticipantRepo implements repo.ProjectParticipantRepository
type ProjectParticipantRepo struct{}

var _ repo.ProjectParticipantRepository = (*ProjectParticipantRepo)(nil)

func (r *ProjectParticipantRepo) GetByProjectUUID(uuid string) (interface{}, error) {
	var participantList []entity.ProjectParticipant
	err := db.Where("project_uuid = ?", uuid).Find(&participantList).Error
	if err != nil {
		return 0, err
	}
	return participantList, nil
}

func (r *ProjectParticipantRepo) GetBySiteUUID(siteUUID string) (interface{}, error) {
	var participantList []entity.ProjectParticipant
	err := db.Where("site_uuid = ?", siteUUID).Find(&participantList).Error
	if err != nil {
		return 0, err
	}
	return participantList, nil
}

func (r *ProjectParticipantRepo) GetByProjectAndSiteUUID(projectUUID, siteUUID string) (interface{}, error) {
	participant := &entity.ProjectParticipant{}
	if err := db.Model(&entity.ProjectParticipant{}).
		Where("project_uuid = ? AND site_uuid = ?", projectUUID, siteUUID).
		First(participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrProjectParticipantNotFound
		}
		return nil, err
	}
	return participant, nil
}

func (r *ProjectParticipantRepo) Create(instance interface{}) error {
	newParticipant := instance.(*entity.ProjectParticipant)
	return db.Model(&entity.ProjectParticipant{}).Create(newParticipant).Error
}

func (r *ProjectParticipantRepo) UpdateStatusByUUID(instance interface{}) error {
	participant := instance.(*entity.ProjectParticipant)
	return db.Model(&entity.ProjectParticipant{}).Where("uuid = ?", participant.UUID).
		Update("status", participant.Status).Error
}

func (r *ProjectParticipantRepo) UpdateParticipantInfoBySiteUUID(instance interface{}) error {
	participant := instance.(*entity.ProjectParticipant)
	return db.Model(&entity.ProjectParticipant{}).Where("site_uuid = ?", participant.SiteUUID).
		Select("site_name", "site_party_id", "site_description").Updates(participant).Error
}

// InitTable make sure the table is created in the db
func (r *ProjectParticipantRepo) InitTable() {
	if err := db.AutoMigrate(&entity.ProjectParticipant{}); err != nil {
		panic(err)
	}
}
