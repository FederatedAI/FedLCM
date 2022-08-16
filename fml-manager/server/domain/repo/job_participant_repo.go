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

// ErrJobParticipantNotFound is an error returned when no participant record is found
var ErrJobParticipantNotFound = errors.New("this site is not in the current job")

// JobParticipantRepository is the interface for managing job participant in the repo
type JobParticipantRepository interface {
	// Create takes an *entity.JobParticipant and save it in the repo
	Create(interface{}) error
	// UpdateStatusByUUID takes an *entity.JobParticipant and updates the status in the repo
	UpdateStatusByUUID(interface{}) error
	// GetStatusByUUID takes an *entity.JobParticipant and returns the status of the participant
	GetStatusByUUID(instance interface{}) interface{}
	// GetByJobAndSiteUUID returns an *entity.JobParticipant indexed by the job and site uuid
	GetByJobAndSiteUUID(string, string) (interface{}, error)
	// GetListByJobUUID returns an []entity.JobParticipant list in the specified job
	GetListByJobUUID(string) (interface{}, error)
}
