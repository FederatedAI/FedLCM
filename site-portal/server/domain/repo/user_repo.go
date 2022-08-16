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

import "github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"

// UserRepository holds methods to access user repos
type UserRepository interface {
	// GetAllUsers returns all the saved users
	GetAllUsers() (interface{}, error)
	// CreateUser create a new user
	CreateUser(user interface{}) error
	// UpdatePermissionInfoById updates a users permission
	UpdatePermissionInfoById(id uint, info valueobject.UserPermissionInfo) error
	// LoadById loads info of a user from the repo
	LoadById(user interface{}) error
	// LoadByName loads info of a user from the repo
	LoadByName(user interface{}) error
	//UpdatePasswordById updates a users password
	UpdatePasswordById(id uint, newPassword string) error
}
