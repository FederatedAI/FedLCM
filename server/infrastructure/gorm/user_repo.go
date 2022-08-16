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

package gorm

import (
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// UserRepo implements repo.UserRepository using gorm and PostgreSQL
type UserRepo struct{}

// make sure UserRepo implements the repo.UserRepository interface
var _ repo.UserRepository = (*UserRepo)(nil)

// ErrUserExist means new user cannot be created due to the existence of the same-name user
var ErrUserExist = errors.New("user already exists")

// CreateUser creates a new user
func (r *UserRepo) CreateUser(instance interface{}) error {
	// check name
	var count int64
	newUser := instance.(*entity.User)
	db.Model(&entity.User{}).Where("name = ?", newUser.Name).Count(&count)
	if count > 0 {
		return ErrUserExist
	}

	// add data
	if err := db.Create(newUser).Error; err != nil {
		return err
	}

	return nil
}

// UpdatePasswordById changes the user's hashed password
func (r *UserRepo) UpdatePasswordById(id uint, hashedPassword string) error {
	toUpdateUser := &entity.User{}
	if err := db.Where("id = ?", id).First(&toUpdateUser).Error; err != nil {
		return err
	}
	return db.Model(toUpdateUser).Update("password", hashedPassword).Error
}

// UpdateByName changes the specified user's info
func (r *UserRepo) UpdateByName(updatedUser *entity.User) error {
	return db.Model(&entity.User{}).Where("name = ?", updatedUser.Name).Updates(updatedUser).Error
}

// GetByName returns the user info indexed by the name
func (r *UserRepo) GetByName(name string) (*entity.User, error) {
	user := &entity.User{}
	if err := db.Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// LoadById loads the user info by id
func (r *UserRepo) LoadById(instance interface{}) error {
	user := instance.(*entity.User)
	return db.Where("id = ?", user.ID).First(&user).Error
}

// LoadByName loads the user info by name
func (r *UserRepo) LoadByName(instance interface{}) error {
	user := instance.(*entity.User)
	return db.Where("name = ?", user.Name).First(&user).Error
}

// InitTable makes sure the table is created in the db
func (r *UserRepo) InitTable() {
	if err := db.AutoMigrate(entity.User{}); err != nil {
		panic(err)
	}
}

// InitData inserts a default users information
func (r *UserRepo) InitData() {

	adminPassword := viper.GetString("lifecyclemanager.initial.admin.password")
	if adminPassword == "" {
		adminPassword = "admin"
	}
	hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// init 'admin' user
	admin := &entity.User{
		UUID:     uuid.NewV4().String(),
		Name:     "Admin",
		Password: string(hashedAdminPassword),
	}

	// if 'admin' exists, we keep using the original password
	if err := r.CreateUser(admin); err != nil {
		if err == ErrUserExist {
			log.Info().Msgf("user: %s exists", admin.Name)
		} else {
			panic(err)
		}
	}

}
