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

// EndpointRepository is the interface to handle endpoint's persistence related actions
type EndpointRepository interface {
	// Create takes a *entity.EndpointBase or its derived struct instance and creates an endpoint info record in the repository
	Create(interface{}) error
	// List returns []entity.EndpointBase or derived struct instances list
	List() (interface{}, error)
	// DeleteByUUID delete the endpoint record with the specified uuid
	DeleteByUUID(string) error
	// GetByUUID returns an *entity.EndpointBase or its derived struct of the specified uuid
	GetByUUID(string) (interface{}, error)
	// ListByInfraProviderUUID returns []entity.EndpointBase or derived struct instances list that contain the specified infra uuid
	ListByInfraProviderUUID(string) (interface{}, error)
	// UpdateStatusByUUID takes an *entity.EndpointBase or its derived struct and updates the status field
	UpdateStatusByUUID(interface{}) error
	// UpdateInfoByUUID takes an *entity.EndpointBase or its derived struct and updates endpoint editable fields
	UpdateInfoByUUID(interface{}) error
}
