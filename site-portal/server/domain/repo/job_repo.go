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

// ErrJobNotFound is the error returned when no job is found
var ErrJobNotFound = errors.New("job not found")

// JobRepository is the interface to manage job info in the repo
type JobRepository interface {
	// Create takes an *entity.Job and creates it in the repo
	Create(interface{}) error
	// UpdateFATEJobInfoByUUID takes an *entity.Job and updates the FATE job related info
	UpdateFATEJobInfoByUUID(interface{}) error
	// UpdateFATEJobStatusByUUID takes an *entity.Job and updates the FATE job status field
	UpdateFATEJobStatusByUUID(interface{}) error
	// UpdateStatusByUUID takes an *entity.Job and updates the job status field
	UpdateStatusByUUID(interface{}) error
	// UpdateStatusMessageByUUID takes an *entity.Job and updates the job status message field
	UpdateStatusMessageByUUID(interface{}) error
	// UpdateFinishTimeByUUID takes an *entity.Job and updates the finish time
	UpdateFinishTimeByUUID(interface{}) error
	// UpdateResultInfoByUUID takes an *entity.Job and updates the result info
	UpdateResultInfoByUUID(interface{}) error
	// CheckNameConflict returns error if the same name job exists
	CheckNameConflict(string) error
	// GetAll returns []entity.Job of all not-deleted jobs
	GetAll() (interface{}, error)
	// DeleteByProjectUUID delete the job of the specified project
	DeleteByProjectUUID(string) error
	// GetListByProjectUUID returns a list of []entity.Job in the specified project
	GetListByProjectUUID(string) (interface{}, error)
	// GetByUUID returns an *entity.Job of the specified uuid
	GetByUUID(string) (interface{}, error)
}
