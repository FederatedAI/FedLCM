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
	"github.com/pkg/errors"
)

// EndpointKubeFATERepo is the implementation of the repo.EndpointRepository interface
type EndpointKubeFATERepo struct{}

var _ repo.EndpointRepository = (*EndpointKubeFATERepo)(nil)

// ErrEndpointExist means new endpoint cannot be created due to the existence of the same-name endpoint
var ErrEndpointExist = errors.New("endpoint already exists")

func (r *EndpointKubeFATERepo) Create(instance interface{}) error {
	endpoint := instance.(*entity.EndpointKubeFATE)

	var count int64
	db.Model(&entity.EndpointKubeFATE{}).Where("name = ?", endpoint.Name).Count(&count)
	if count > 0 {
		return ErrEndpointExist
	}

	if err := db.Create(endpoint).Error; err != nil {
		return err
	}
	return nil
}

func (r *EndpointKubeFATERepo) List() (interface{}, error) {
	var endpoints []entity.EndpointKubeFATE
	if err := db.Find(&endpoints).Error; err != nil {
		return nil, err
	}
	return endpoints, nil
}

func (r *EndpointKubeFATERepo) DeleteByUUID(uuid string) error {
	return db.Unscoped().Where("uuid = ?", uuid).Delete(&entity.EndpointKubeFATE{}).Error
}

func (r *EndpointKubeFATERepo) GetByUUID(uuid string) (interface{}, error) {
	endpoint := &entity.EndpointKubeFATE{}
	if err := db.Where("uuid = ?", uuid).First(endpoint).Error; err != nil {
		return nil, err
	}
	return endpoint, nil
}

func (r *EndpointKubeFATERepo) ListByInfraProviderUUID(infraUUID string) (interface{}, error) {
	var endpoints []entity.EndpointKubeFATE
	err := db.Where("infra_provider_uuid = ?", infraUUID).Find(&endpoints).Error
	if err != nil {
		return 0, err
	}
	return endpoints, nil
}

func (r *EndpointKubeFATERepo) ListByInfraProviderUUIDAndNamespace(infraUUID string, namespace string) (interface{}, error) {
	var endpoints []entity.EndpointKubeFATE
	err := db.Where("infra_provider_uuid = ? AND namespace = ?", infraUUID, namespace).Find(&endpoints).Error
	if err != nil {
		return 0, err
	}
	return endpoints, nil
}

func (r *EndpointKubeFATERepo) UpdateStatusByUUID(instance interface{}) error {
	endpoint := instance.(*entity.EndpointKubeFATE)
	return db.Model(&entity.EndpointKubeFATE{}).Where("uuid = ?", endpoint.UUID).
		Update("status", endpoint.Status).Error
}

func (r *EndpointKubeFATERepo) UpdateInfoByUUID(instance interface{}) error {
	endpoint := instance.(*entity.EndpointKubeFATE)
	return db.Where("uuid = ?", endpoint.UUID).
		Select("name", "description", "version", "config", "status").
		Updates(endpoint).Error
}

// InitTable makes sure the table is created in the db
func (r *EndpointKubeFATERepo) InitTable() {
	if err := db.AutoMigrate(entity.EndpointKubeFATE{}); err != nil {
		panic(err)
	}
}
