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

// ProjectRepo is the implementation of repo.ProjectRepository
type ProjectRepo struct{}

var _ repo.ProjectRepository = (*ProjectRepo)(nil)

var ErrLocalProjectNameConflict = errors.New("project name conflicts")

func (r *ProjectRepo) Create(instance interface{}) error {
	newProject := instance.(*entity.Project)
	// TODO: check name conflicts?
	// Add records
	return db.Model(&entity.Project{}).Create(newProject).Error
}

func (r *ProjectRepo) GetAll() (interface{}, error) {
	var projectList []entity.Project
	if err := db.Find(&projectList).Error; err != nil {
		return nil, err
	}
	return projectList, nil
}

func (r *ProjectRepo) GetByUUID(uuid string) (interface{}, error) {
	project := &entity.Project{}
	if err := db.Where("uuid = ?", uuid).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrProjectNotFound
		}
		return nil, err
	}
	return project, nil
}

func (r *ProjectRepo) DeleteByUUID(uuid string) error {
	return db.Unscoped().Where("uuid = ?", uuid).Delete(&entity.Project{}).Error
}

func (r *ProjectRepo) CheckNameConflict(name string) error {
	var count int64
	err := db.Model(&entity.Project{}).
		Where("name = ? AND status = ?", name, entity.ProjectStatusManaged).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrLocalProjectNameConflict
	}
	return nil
}

func (r *ProjectRepo) UpdateStatusByUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("uuid = ?", project.UUID).
		Update("status", project.Status).Error
}

func (r *ProjectRepo) UpdateTypeByUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("uuid = ?", project.UUID).
		Update("type", project.Type).Error
}

func (r *ProjectRepo) UpdateAutoApprovalStatusByUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("uuid = ?", project.UUID).
		Update("auto_approval_enabled", project.AutoApprovalEnabled).Error
}

func (r *ProjectRepo) UpdateManagingSiteInfoBySiteUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("managing_site_uuid = ?", project.ManagingSiteUUID).
		Select("managing_site_name", "managing_site_party_id").Updates(project).Error
}

// InitTable make sure the table is created in the db
func (r *ProjectRepo) InitTable() {
	if err := db.AutoMigrate(&entity.Project{}); err != nil {
		panic(err)
	}
}
