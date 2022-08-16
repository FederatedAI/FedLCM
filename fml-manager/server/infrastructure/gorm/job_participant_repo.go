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
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// JobParticipantRepo implements repo.JobParticipantRepository using gorm and PostgreSQL
type JobParticipantRepo struct{}

// make sure JobParticipantRepo implements the repo.JobParticipantRepository interface
var _ repo.JobParticipantRepository = (*JobParticipantRepo)(nil)

func (r *JobParticipantRepo) Create(instance interface{}) error {
	newJobParticipant := instance.(*entity.JobParticipant)
	return db.Model(&entity.JobParticipant{}).Create(newJobParticipant).Error
}

func (r *JobParticipantRepo) UpdateStatusByUUID(instance interface{}) error {
	participant := instance.(*entity.JobParticipant)
	return db.Model(&entity.JobParticipant{}).Where("uuid = ?", participant.UUID).
		Update("status", participant.Status).Error
}

func (r *JobParticipantRepo) GetStatusByUUID(instance interface{}) interface{} {
	participant := instance.(*entity.JobParticipant)
	var status int
	row := db.Model(&entity.JobParticipant{}).Where("uuid = ?", participant.UUID).Select("status", status).Row()
	err := row.Scan(&status)
	if err != nil {
		log.Info().Msgf("Do not find any status for job-participant uuid %s, err msg %s", participant.UUID, err.Error())
		return entity.JobParticipantStatusUnknown
	}
	return status
}

func (r *JobParticipantRepo) GetByJobAndSiteUUID(jobUUID, siteUUID string) (interface{}, error) {
	participant := &entity.JobParticipant{}
	if err := db.Model(&entity.JobParticipant{}).
		Where("job_uuid = ? AND site_uuid = ?", jobUUID, siteUUID).
		First(participant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrJobParticipantNotFound
		}
		return nil, err
	}
	return participant, nil
}

func (r *JobParticipantRepo) GetListByJobUUID(jobUUID string) (interface{}, error) {
	var participantList []entity.JobParticipant
	if err := db.Where("job_uuid = ?", jobUUID).Find(&participantList).Error; err != nil {
		return nil, err
	}
	return participantList, nil
}

// InitTable make sure the table is created in the db
func (r *JobParticipantRepo) InitTable() {
	if err := db.AutoMigrate(&entity.JobParticipant{}); err != nil {
		panic(err)
	}
}
