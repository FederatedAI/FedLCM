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

// UserRepository holds methods to access user repos
type UserRepository interface {
	// CreateUser takes an *entity.User and creates a new user record in the repo
	CreateUser(user interface{}) error
	// LoadById takes an *entity.User and loads info of a user with the specified id from the repo into it
	LoadById(user interface{}) error
	// LoadByName takes an *entity.User loads info of a user with the specified name from the repo into it
	LoadByName(user interface{}) error
	// UpdatePasswordById updates a users password
	UpdatePasswordById(id uint, newPassword string) error
}
