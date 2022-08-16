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

package repo

// InfraProviderRepository is the interface to handle infra provider's persistence related actions
type InfraProviderRepository interface {
	// Create takes a *entity.InfraProviderBase's derived struct and creates a provider info record in the repository
	Create(interface{}) error
	// List returns provider info list, currently []entity.InfraProviderKubernetes
	List() (interface{}, error)
	// DeleteByUUID delete provider info with the specified uuid
	DeleteByUUID(string) error
	// GetByUUID returns an *entity.InfraProviderBase or its derived struct of the specified provider
	GetByUUID(string) (interface{}, error)
	// UpdateByUUID takes an *entity.InfraProviderBase or its derived struct and updates the infra provider config by uuid
	UpdateByUUID(interface{}) error
	// GetByAddress returns an *entity.InfraProviderBase or its derived struct who contains the specified address
	GetByAddress(string) (interface{}, error)
	// ProviderExists checks if a provider already exists in the  database
	ProviderExists(interface{}) error
}
