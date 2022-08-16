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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
)

// LocalDataRepo implements repo.LocalDataRepository using gorm and PostgreSQL
type LocalDataRepo struct{}

// make sure LocalDataRepo implements the repo.LocalDataRepository interface
var _ repo.LocalDataRepository = (*LocalDataRepo)(nil)

// ErrLocalDataNameConflict means data with same name exists
var ErrLocalDataNameConflict = errors.New("data name cannot be the same with existing one")

func (r *LocalDataRepo) Create(instance interface{}) error {
	newData := instance.(*entity.LocalData)
	if err := r.CheckNameConflict(newData.Name); err != nil {
		return err
	}

	if err := db.Model(&entity.LocalData{}).Create(newData).Error; err != nil {
		return err
	}
	return nil
}

func (r *LocalDataRepo) UpdateJobInfoByUUID(instance interface{}) error {
	localData := instance.(*entity.LocalData)
	return db.Model(localData).Where("uuid = ?", localData.UUID).
		Select("job_id", "job_conf", "job_status", "job_error_msg").
		Updates(localData).Error
}

func (r *LocalDataRepo) GetAll() (interface{}, error) {
	var localDataList []entity.LocalData
	if err := db.Find(&localDataList).Error; err != nil {
		return nil, err
	}
	return localDataList, nil
}

func (r *LocalDataRepo) GetByUUID(uuid string) (interface{}, error) {
	localData := &entity.LocalData{}
	if err := db.Model(&entity.LocalData{}).Where("uuid = ?", uuid).First(&localData).Error; err != nil {
		return nil, err
	}
	return localData, nil
}

func (r *LocalDataRepo) DeleteByUUID(uuid string) error {
	return db.Model(&entity.LocalData{}).Where("uuid = ?", uuid).Delete(&entity.LocalData{}).Error
}

func (r *LocalDataRepo) UpdateIDMetaInfoByUUID(uuid string, instance interface{}) error {
	metaInfo := instance.(*valueobject.IDMetaInfo)
	return db.Model(&entity.LocalData{}).Where("uuid = ?", uuid).
		Update("id_meta_info", metaInfo).Error
}

func (r *LocalDataRepo) CheckNameConflict(name string) error {
	var count int64
	err := db.Model(&entity.LocalData{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrLocalDataNameConflict
	}
	return nil
}

// InitTable make sure the table is created in the db
func (r *LocalDataRepo) InitTable() {
	if err := db.AutoMigrate(&entity.LocalData{}); err != nil {
		panic(err)
	}
}
