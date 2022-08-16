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

// LocalDataRepository holds methods to access local data repos
// The interface{} parameters in most of the functions should of type "*entity.LocalData"
type LocalDataRepository interface {
	// Create save the data records, the interface should of type "*entity.LocalData"
	Create(interface{}) error
	// UpdateJobInfoByUUID changes the data upload job status
	UpdateJobInfoByUUID(interface{}) error
	// GetAll returns all the uploaded local data
	// the returned interface{} should be of type []entity.LocalData
	GetAll() (interface{}, error)
	// LoadByUUID returns a *entity.LocalData object by providing the uuid
	GetByUUID(string) (interface{}, error)
	// DeleteByUUID deletes the data record
	DeleteByUUID(string) error
	// UpdateIDMetaInfoByUUID updates the IDMetaInfo, the passed interface{} is of type entity.IDMetaInfo
	UpdateIDMetaInfoByUUID(string, interface{}) error
	// CheckNameConflict returns an error if there are name conflicts
	CheckNameConflict(string) error
}
