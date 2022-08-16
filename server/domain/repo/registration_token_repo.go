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

// RegistrationTokenRepository is the interface to handle federation's registration token related actions
type RegistrationTokenRepository interface {
	// Create takes an *entity.RegistrationToken's derived struct and creates a record in the repository
	Create(interface{}) error
	// ListByFederation returns token list, currently []entity.RegistrationTokenOpenFL
	ListByFederation(string) (interface{}, error)
	// DeleteByUUID delete the token with the specified uuid
	DeleteByUUID(string) error
	// GetByUUID returns an *entity.RegistrationToken's derived struct of the specified uuid
	GetByUUID(string) (interface{}, error)
	// LoadByTypeAndStr takes an *entity.RegistrationToken's derived struct and fill it with missing info based on the type and token string
	LoadByTypeAndStr(interface{}) error
	// DeleteByFederation deletes all tokens within the specified federation
	DeleteByFederation(string) error
}
