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

// FederationOpenFLRepo implements the repo.FederationRepository interface
type FederationOpenFLRepo struct{}

var _ repo.FederationRepository = (*FederationOpenFLRepo)(nil)

func (r *FederationOpenFLRepo) Create(instance interface{}) error {
	var count int64
	federation := instance.(*entity.FederationOpenFL)
	db.Model(&entity.FederationOpenFL{}).Where("name = ?", federation.Name).Count(&count)
	if count > 0 {
		return ErrFederationExist
	}

	if err := db.Create(federation).Error; err != nil {
		return err
	}
	return nil
}

func (r *FederationOpenFLRepo) List() (interface{}, error) {
	var federationList []entity.FederationOpenFL
	if err := db.Find(&federationList).Error; err != nil {
		return nil, err
	}
	return federationList, nil
}

func (r *FederationOpenFLRepo) DeleteByUUID(uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&entity.FederationOpenFL{}).Error
}

func (r *FederationOpenFLRepo) GetByUUID(uuid string) (interface{}, error) {
	federation := &entity.FederationOpenFL{}
	if err := db.Where("uuid = ?", uuid).First(federation).Error; err != nil {
		return nil, err
	}
	return federation, nil
}

// InitTable makes sure the table is created in the db
func (r *FederationOpenFLRepo) InitTable() {
	if err := db.AutoMigrate(entity.FederationOpenFL{}); err != nil {
		panic(err)
	}
}
