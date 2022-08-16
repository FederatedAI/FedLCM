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

// ProjectDataRepo is the implementation of repo.ProjectDataRepository
type ProjectDataRepo struct{}

var _ repo.ProjectDataRepository = (*ProjectDataRepo)(nil)

func (r *ProjectDataRepo) Create(instance interface{}) error {
	newData := instance.(*entity.ProjectData)
	// Add records
	return db.Model(&entity.ProjectData{}).Create(newData).Error
}

func (r *ProjectDataRepo) GetByProjectAndDataUUID(projectUUID string, dataUUID string) (interface{}, error) {
	data := &entity.ProjectData{}
	if err := db.Model(&entity.ProjectData{}).
		Where("project_uuid = ? AND data_uuid = ?", projectUUID, dataUUID).
		First(data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrProjectDataNotFound
		}
		return nil, err
	}
	return data, nil
}

func (r *ProjectDataRepo) UpdateStatusByUUID(instance interface{}) error {
	data := instance.(*entity.ProjectData)
	return db.Model(&entity.ProjectData{}).Where("uuid = ?", data.UUID).
		Update("status", data.Status).Error
}

func (r *ProjectDataRepo) GetListByProjectUUID(projectUUID string) (interface{}, error) {
	var projectDataList []entity.ProjectData
	err := db.Where("project_uuid = ?", projectUUID).Find(&projectDataList).Error
	if err != nil {
		return 0, err
	}
	return projectDataList, nil
}

func (r *ProjectDataRepo) GetListByProjectAndSiteUUID(projectUUID string, siteUUID string) (interface{}, error) {
	var projectDataList []entity.ProjectData
	err := db.Where("project_uuid = ? AND site_uuid = ?", projectUUID, siteUUID).Find(&projectDataList).Error
	if err != nil {
		return 0, err
	}
	return projectDataList, nil
}

func (r *ProjectDataRepo) GetByDataUUID(dataUUID string) (interface{}, error) {
	projectData := &entity.ProjectData{}
	err := db.Where("data_uuid = ?", dataUUID).Last(&projectData).Error
	if err != nil {
		return 0, err
	}
	return projectData, nil
}

func (r *ProjectDataRepo) GetListByDataUUID(dataUUID string) (interface{}, error) {
	var projectDataList []entity.ProjectData
	err := db.Where("data_uuid = ?", dataUUID).Find(&projectDataList).Error
	if err != nil {
		return 0, err
	}
	return projectDataList, nil
}

func (r *ProjectDataRepo) DeleteByUUID(uuid string) error {
	return db.Unscoped().Where("uuid = ?", uuid).Delete(&entity.ProjectData{}).Error
}

func (r *ProjectDataRepo) DeleteByProjectUUID(projectUUID string) error {
	return db.Unscoped().Where("project_uuid = ?", projectUUID).Delete(&entity.ProjectData{}).Error
}

func (r *ProjectDataRepo) UpdateSiteInfoBySiteUUID(instance interface{}) error {
	site := instance.(*entity.ProjectData)
	return db.Where("site_uuid = ?", site.SiteUUID).
		Select("site_name", "site_party_id").Updates(site).Error
}

// InitTable make sure the table is created in the db
func (r *ProjectDataRepo) InitTable() {
	if err := db.AutoMigrate(&entity.ProjectData{}); err != nil {
		panic(err)
	}
}
