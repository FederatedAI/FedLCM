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

package service

import (
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"gorm.io/gorm"
)

// UserApp provides user management service
type UserApp struct {
	UserRepo repo.UserRepository
}

// PublicUser represents a user info viewable to the public
type PublicUser struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
	UUID string `json:"uuid"`
	valueobject.UserPermissionInfo
}

// LoginInfo represents fields related with login
type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// PwdChangeInfo represents fields related with login
type PwdChangeInfo struct {
	CurPassword string `json:"cur_password"`
	NewPassword string `json:"new_password"`
}

// GetUsers returns all the available users as a list of PublicUser
func (app *UserApp) GetUsers() ([]PublicUser, error) {
	repoUsers, err := app.UserRepo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	userList := repoUsers.([]entity.User)
	users := make([]PublicUser, len(userList))
	for index, repoUser := range userList {
		users[index] = PublicUser{
			Name:               repoUser.Name,
			ID:                 repoUser.ID,
			UUID:               repoUser.UUID,
			UserPermissionInfo: repoUser.PermissionInfo,
		}
	}
	return users, nil
}

// UpdateUserPermission changes a user's valueobject.UserPermissionInfo
func (app *UserApp) UpdateUserPermission(publicUser *PublicUser) error {
	user := &entity.User{
		Model: gorm.Model{
			ID: publicUser.ID,
		},
		Repo: app.UserRepo,
	}
	if err := user.LoadById(); err != nil {
		return err
	}
	return user.UpdatePermissionInfo(publicUser.UserPermissionInfo)
}

// Login validates the loginInfo and returns a publicUser object on success
func (app *UserApp) Login(info *LoginInfo) (*PublicUser, error) {
	loginService := &service.UserService{
		Repo: app.UserRepo,
	}
	user, err := loginService.LoginService(info.Username, info.Password)
	if err != nil {
		return nil, err
	}
	publicUser := PublicUser{
		Name:               user.Name,
		ID:                 user.ID,
		UUID:               user.UUID,
		UserPermissionInfo: user.PermissionInfo,
	}
	return &publicUser, nil
}

// CheckAccess validates if the user can access site portal
func (app *UserApp) CheckAccess(publicUser *PublicUser) error {
	user := &entity.User{
		Model: gorm.Model{
			ID: publicUser.ID,
		},
		Repo: app.UserRepo,
	}
	if err := user.LoadById(); err != nil {
		return err
	}
	return user.CheckSitePortalAccess()
}

// UpdateUserPassword changes a user's password
func (app *UserApp) UpdateUserPassword(userId int, info *PwdChangeInfo) error {
	user := &entity.User{
		Model: gorm.Model{
			ID: uint(userId),
		},
		Repo: app.UserRepo,
	}
	if err := user.LoadById(); err != nil {
		return err
	}
	return user.UpdatePwdInfo(info.CurPassword, info.NewPassword)
}
