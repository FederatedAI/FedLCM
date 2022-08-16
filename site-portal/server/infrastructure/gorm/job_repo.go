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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// JobRepo implements repo.JobRepository using gorm and PostgreSQL
type JobRepo struct{}

// make sure JobRepo implements the repo.JobRepository interface
var _ repo.JobRepository = (*JobRepo)(nil)

var ErrJobNameConflict = errors.New("job name conflicts")

func (r *JobRepo) Create(instance interface{}) error {
	newJob := instance.(*entity.Job)
	// TODO: check name conflicts?
	// Add records
	return db.Model(&entity.Job{}).Create(newJob).Error
}

func (r *JobRepo) UpdateFATEJobInfoByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Select("fate_job_id", "fate_job_status", "fate_model_id", "fate_model_version").Updates(job).Error
}

func (r *JobRepo) UpdateFATEJobStatusByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Update("fate_job_status", job.FATEJobStatus).Error
}

func (r *JobRepo) UpdateStatusByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Update("status", job.Status).Error
}

func (r *JobRepo) UpdateStatusMessageByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Update("status_message", job.StatusMessage).Error
}

func (r *JobRepo) UpdateFinishTimeByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Update("finished_at", job.FinishedAt).Error
}

func (r *JobRepo) UpdateResultInfoByUUID(instance interface{}) error {
	job := instance.(*entity.Job)
	return db.Model(&entity.Job{}).Where("uuid = ?", job.UUID).
		Update("result_json", job.ResultJson).Error
}

func (r *JobRepo) CheckNameConflict(name string) error {
	var count int64
	err := db.Model(&entity.Job{}).
		Where("name = ?", name).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrJobNameConflict
	}
	return nil
}

func (r *JobRepo) DeleteByProjectUUID(projectUUID string) error {
	return db.Unscoped().Where("project_uuid = ?", projectUUID).Delete(&entity.Job{}).Error
}

func (r *JobRepo) GetAll() (interface{}, error) {
	var jobList []entity.Job
	if err := db.Find(&jobList).Error; err != nil {
		return nil, err
	}
	return jobList, nil
}

func (r *JobRepo) GetListByProjectUUID(projectUUID string) (interface{}, error) {
	var jobList []entity.Job
	err := db.Where("project_uuid = ?", projectUUID).Find(&jobList).Error
	if err != nil {
		return 0, err
	}
	return jobList, nil
}

func (r *JobRepo) GetByUUID(uuid string) (interface{}, error) {
	job := &entity.Job{}
	if err := db.Where("uuid = ?", uuid).First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrJobNotFound
		}
		return nil, err
	}
	return job, nil
}

// InitTable make sure the table is created in the db
func (r *JobRepo) InitTable() {
	if err := db.AutoMigrate(&entity.Job{}); err != nil {
		panic(err)
	}
}
