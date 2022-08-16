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

import (
	"github.com/pkg/errors"
)

// ErrProjectDataNotFound is an error when no data association record found
var ErrProjectDataNotFound = errors.New("this data is not in the specified project")

// ProjectDataRepository is the repo interface for persisting the project data info
type ProjectDataRepository interface {
	// Create takes an *entity.ProjectData and create the records
	Create(interface{}) error
	// GetByProjectAndDataUUID returns an *entity.ProjectData indexed by the specified project and data uuid
	GetByProjectAndDataUUID(string, string) (interface{}, error)
	// UpdateStatusByUUID takes an *entity.ProjectData and update its status
	UpdateStatusByUUID(interface{}) error
	// GetListByProjectUUID returns []entity.ProjectData that associated in the specified project
	GetListByProjectUUID(string) (interface{}, error)
	// GetListByProjectAndSiteUUID returns []entity.ProjectData that associated in the specified project by the specified site
	GetListByProjectAndSiteUUID(string, string) (interface{}, error)
	// DeleteByUUID delete the project data records by the specified uuid
	DeleteByUUID(string) error
	// DeleteByProjectUUID delete the project data records permanently by the specified project uuid
	DeleteByProjectUUID(string) error
	// UpdateSiteInfoBySiteUUID takes an *entity.ProjectData as template to update site info of the specified site uuid
	UpdateSiteInfoBySiteUUID(interface{}) error
}
