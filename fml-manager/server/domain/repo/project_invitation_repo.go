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

// ProjectInvitationRepository is an interface for working with the invitation repo
type ProjectInvitationRepository interface {
	// Create takes an *entity.ProjectInvitation to create the record
	Create(interface{}) error
	// UpdateStatusByUUID takes an *entity.ProjectInvitation and updates the status
	UpdateStatusByUUID(interface{}) error
	// GetByProjectUUID returns an *entity.ProjectInvitation for the specified project. it is the latest one for the project
	GetByProjectUUID(string) (interface{}, error)
	// GetByUUID returns an *entity.ProjectInvitation indexed by the specified uuid
	GetByUUID(string) (interface{}, error)
}
