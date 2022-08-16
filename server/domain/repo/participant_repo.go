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

// ParticipantRepository contains common repo interface to work with federation participants
type ParticipantRepository interface {
	// Create takes an *entity.Participant's derived struct and creates a fate participant info record in the repo
	Create(interface{}) error
	// List returns []entity.Participant's derived struct list
	List() (interface{}, error)
	// DeleteByUUID deletes participant info with the specified uuid
	DeleteByUUID(string) error
	// GetByUUID returns an *entity.Participant's derived struct of the specified uuid
	GetByUUID(string) (interface{}, error)
	// ListByFederationUUID returns []entity.Participant's derived struct list that contain the specified federation uuid
	ListByFederationUUID(string) (interface{}, error)
	// ListByEndpointUUID returns []entity.Participant's derived struct list that contain the specified endpoint uuid
	ListByEndpointUUID(string) (interface{}, error)
	// UpdateStatusByUUID takes an *entity.Participant's derived struct and updates the status field
	UpdateStatusByUUID(interface{}) error
	// UpdateDeploymentYAMLByUUID takes an *entity.Participant's derived struct and updates the deployment_yaml field
	UpdateDeploymentYAMLByUUID(interface{}) error
	// UpdateInfoByUUID takes a *entity.Participant and updates the participant editable fields
	UpdateInfoByUUID(interface{}) error
}
