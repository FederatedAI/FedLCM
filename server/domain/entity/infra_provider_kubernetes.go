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

package entity

import (
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/utils"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// InfraProviderKubernetes is a provider that is using a kubernetes cluster
type InfraProviderKubernetes struct {
	InfraProviderBase
	Config             valueobject.KubeConfig         `json:"config" gorm:"type:text"`
	ConfigSHA256       string                         `json:"config_sha_256" gorm:"type:varchar(64)"`
	APIHost            string                         `json:"api_host" gorm:"type:varchar(256)"`
	RegistryConfigFATE valueobject.KubeRegistryConfig `json:"registry_config_fate" gorm:"type:text"`
	// TODO: add server version?
	Repo repo.InfraProviderRepository `json:"-" gorm:"-"`
}

// Validate checks if the necessary information is provided correctly
func (p *InfraProviderKubernetes) Validate() error {
	return p.Config.Validate()
}

// Create checks the config and saves the object to the repo
func (p *InfraProviderKubernetes) Create() error {
	if err := p.Validate(); err != nil {
		return err
	}
	p.UUID = uuid.NewV4().String()
	if err := p.Repo.Create(p); err != nil {
		return err
	}
	return nil
}

// Update checks the config and update the object to the repo
func (p *InfraProviderKubernetes) Update() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := p.Repo.UpdateByUUID(p); err != nil {
		return err
	}
	return nil
}

func (p *InfraProviderKubernetes) BeforeSave(tx *gorm.DB) error {
	// encrypted registry secret password
	encryptedSecret, err := utils.Encrypt(p.RegistryConfigFATE.RegistrySecretConfig.Password)
	if err != nil {
		return err
	}
	p.RegistryConfigFATE.RegistrySecretConfig.Password = encryptedSecret
	p.APIHost, _ = p.Config.APIHost()
	p.ConfigSHA256 = p.Config.SHA2565()
	return nil
}

func (p *InfraProviderKubernetes) AfterFind(tx *gorm.DB) error {
	// decrypted registry secret password
	decryptedSecret, err := utils.Decrypt(p.RegistryConfigFATE.RegistrySecretConfig.Password)
	if err != nil {
		return err
	}
	p.RegistryConfigFATE.RegistrySecretConfig.Password = decryptedSecret
	return nil
}
