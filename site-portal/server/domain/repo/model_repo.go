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

// ErrModelNotFound is the error returned when no model is found
var ErrModelNotFound = errors.New("model not found")

// ModelRepository is the interface for managing models
type ModelRepository interface {
	// Create takes an *entity.Model and save it into the repo
	Create(interface{}) error
	// GetAll returns []entity.Model of all not-deleted models
	GetAll() (interface{}, error)
	// DeleteByUUID deletes the specified model
	DeleteByUUID(string) error
	// GetListByProjectUUID returns []entity.Model belonging to the specified project
	GetListByProjectUUID(string) (interface{}, error)
	// GetByUUID returns an *entity.Model indexed by the uuid
	GetByUUID(string) (interface{}, error)
}
