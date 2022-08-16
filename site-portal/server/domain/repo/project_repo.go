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

import "github.com/pkg/errors"

// ErrProjectNotFound is the error returned when no project is found
var ErrProjectNotFound = errors.New("project not found")

// ProjectRepository is the repo interface for project
type ProjectRepository interface {
	// Create takes an *entity.Project and create the record
	Create(interface{}) error
	// GetAll returns an []entity.Project
	GetAll() (interface{}, error)
	// GetByUUID returns an *entity.Project with the specified uuid
	GetByUUID(string) (interface{}, error)
	// DeleteByUUID deletes the project
	DeleteByUUID(string) error
	// CheckNameConflict returns an error if there are name conflicts
	CheckNameConflict(string) error
	// UpdateStatusByUUID takes an *entity.Project and update its status
	UpdateStatusByUUID(interface{}) error
	// UpdateTypeByUUID takes an *entity.Project and update its type
	UpdateTypeByUUID(interface{}) error
	// UpdateAutoApprovalStatusByUUID takes an *entity.Project and update its auto-approval status
	UpdateAutoApprovalStatusByUUID(interface{}) error
	// UpdateManagingSiteInfoBySiteUUID takes an *entity.Project as template and
	// updates site related info of all records containing the site uuid
	UpdateManagingSiteInfoBySiteUUID(interface{}) error
}
