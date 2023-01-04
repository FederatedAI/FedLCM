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

// InfraProviderKubernetesRepo is the implementation of the repo.InfraProviderRepository interface
type InfraProviderKubernetesRepo struct{}

var _ repo.InfraProviderRepository = (*InfraProviderKubernetesRepo)(nil)

// ProviderExists checks if a provider already exists in database according to provider's name and config_sha256
func (r *InfraProviderKubernetesRepo) ProviderExists(instance interface{}) error {
	var countName int64
	var countConfig int64
	provider := instance.(*entity.InfraProviderKubernetes)
	db.Model(&entity.InfraProviderKubernetes{}).Where("name = ?", provider.Name).Count(&countName)
	db.Model(&entity.InfraProviderKubernetes{}).Where("config_sha256 = ?", provider.Config.SHA2565()).Count(&countConfig)
	if countName > 0 || countConfig > 0 {
		return repo.ErrProviderExist
	}
	return nil
}

// Create creates a record in the DB
func (r *InfraProviderKubernetesRepo) Create(instance interface{}) error {

	if err := r.ProviderExists(instance); err != nil {
		return err
	}

	provider := instance.(*entity.InfraProviderKubernetes)

	if err := db.Create(provider).Error; err != nil {
		return err
	}
	return nil
}

// List returns provider list
func (r *InfraProviderKubernetesRepo) List() (interface{}, error) {
	var providers []entity.InfraProviderKubernetes
	if err := db.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// DeleteByUUID deletes records using the specified uuid
func (r *InfraProviderKubernetesRepo) DeleteByUUID(uuid string) error {
	return db.Unscoped().Where("uuid = ?", uuid).Delete(&entity.InfraProviderKubernetes{}).Error
}

// GetByUUID returns the specified provider
func (r *InfraProviderKubernetesRepo) GetByUUID(uuid string) (interface{}, error) {
	provider := &entity.InfraProviderKubernetes{}
	if err := db.Where("uuid = ?", uuid).First(provider).Error; err != nil {
		return nil, err
	}
	return provider, nil
}

func (r *InfraProviderKubernetesRepo) UpdateByUUID(instance interface{}) error {
	provider := instance.(*entity.InfraProviderKubernetes)
	return db.Where("uuid = ?", provider.UUID).
		Select("name", "description", "type", "config", "registry_config_fate", "api_host", "config_sha256").Updates(provider).Error
}

func (r *InfraProviderKubernetesRepo) GetByConfigSHA256(sha256 string) (interface{}, error) {
	provider := &entity.InfraProviderKubernetes{}
	if err := db.Where("config_sha256 = ?", sha256).First(provider).Error; err != nil {
		return nil, err
	}
	return provider, nil
}

// InitTable makes sure the table is created in the db
func (r *InfraProviderKubernetesRepo) InitTable() {
	if err := db.AutoMigrate(entity.InfraProviderKubernetes{}); err != nil {
		panic(err)
	}
}
