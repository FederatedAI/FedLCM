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

// ErrProjectParticipantNotFound is an error returned when no participant record is found
var ErrProjectParticipantNotFound = errors.New("this site is not in the current project")

// ProjectParticipantRepository is the interface to work with participant related repo
type ProjectParticipantRepository interface {
	// CountJoinedParticipantByProjectUUID returns number of participants in joined status
	CountJoinedParticipantByProjectUUID(string) (int64, error)
	// GetByProjectUUID returns a []entity.ProjectParticipant in the specified project
	GetByProjectUUID(string) (interface{}, error)
	// Create takes an *entity.ProjectParticipant cam create the records
	Create(interface{}) error
	// GetByProjectAndSiteUUID returns an *entity.ProjectParticipant from the specified project and site uuid
	GetByProjectAndSiteUUID(string, string) (interface{}, error)
	// UpdateStatusByUUID takes an *entity.ProjectParticipant and update its status
	UpdateStatusByUUID(interface{}) error
	// DeleteByProjectUUID delete the records specified by the uui
	DeleteByProjectUUID(string) error
	// UpdateParticipantInfoBySiteUUID takes an *entity.ProjectParticipant as template and
	// updates sites info of the records containing the specified site uuid
	UpdateParticipantInfoBySiteUUID(interface{}) error
}
