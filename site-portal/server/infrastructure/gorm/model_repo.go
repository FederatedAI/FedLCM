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

// ModelRepo implements repo.ModelRepository using gorm and PostgreSQL
type ModelRepo struct{}

// make sure ModelRepo implements the repo.ModelRepository interface
var _ repo.ModelRepository = (*ModelRepo)(nil)

func (r *ModelRepo) Create(instance interface{}) error {
	newModel := instance.(*entity.Model)
	// XXX: check name conflicts?
	// Add records
	return db.Model(&entity.Model{}).Create(newModel).Error
}

func (r *ModelRepo) GetAll() (interface{}, error) {
	var modelList []entity.Model
	if err := db.Find(&modelList).Error; err != nil {
		return nil, err
	}
	return modelList, nil
}

func (r *ModelRepo) DeleteByUUID(uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&entity.Model{}).Error
}

func (r *ModelRepo) GetListByProjectUUID(projectUUID string) (interface{}, error) {
	var modelList []entity.Model
	err := db.Where("project_uuid = ?", projectUUID).Find(&modelList).Error
	if err != nil {
		return 0, err
	}
	return modelList, nil
}

func (r *ModelRepo) GetByUUID(uuid string) (interface{}, error) {
	model := &entity.Model{}
	if err := db.Where("uuid = ?", uuid).First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repo.ErrModelNotFound
		}
		return nil, err
	}
	return model, nil
}

// InitTable make sure the table is created in the db
func (r *ModelRepo) InitTable() {
	if err := db.AutoMigrate(&entity.Model{}); err != nil {
		panic(err)
	}
}
