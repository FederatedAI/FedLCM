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

// ProjectRepo implements repo.ProjectRepository
type ProjectRepo struct{}

var _ repo.ProjectRepository = (*ProjectRepo)(nil)

func (r *ProjectRepo) Create(instance interface{}) error {
	newProject := instance.(*entity.Project)
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

func (r *ProjectRepo) UpdateManagingSiteInfoBySiteUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("managing_site_uuid = ?", project.ManagingSiteUUID).
		Select("managing_site_name", "managing_site_party_id").Updates(project).Error
}

func (r *ProjectRepo) UpdateStatusByUUID(instance interface{}) error {
	project := instance.(*entity.Project)
	return db.Model(&entity.Project{}).Where("uuid = ?", project.UUID).
		Update("status", project.Status).Error
}

// InitTable make sure the table is created in the db
func (r *ProjectRepo) InitTable() {
	if err := db.AutoMigrate(&entity.Project{}); err != nil {
		panic(err)
	}
}
