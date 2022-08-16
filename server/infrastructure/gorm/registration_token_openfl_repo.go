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
)

// RegistrationTokenOpenFLRepo implements repo.RegistrationTokenRepository interface
type RegistrationTokenOpenFLRepo struct{}

var _ repo.RegistrationTokenRepository = (*RegistrationTokenOpenFLRepo)(nil)

// ErrTokenExist means new token cannot be created due to name conflicts
var ErrTokenExist = errors.New("token already exists")

func (r *RegistrationTokenOpenFLRepo) Create(instance interface{}) error {
	var count int64
	token := instance.(*entity.RegistrationTokenOpenFL)
	db.Model(&entity.RegistrationTokenOpenFL{}).Where("name = ? AND federation_uuid = ?", token.Name, token.FederationUUID).Count(&count)
	if count > 0 {
		return ErrTokenExist
	}

	if err := db.Create(token).Error; err != nil {
		return err
	}
	return nil
}

func (r *RegistrationTokenOpenFLRepo) ListByFederation(federationUUID string) (interface{}, error) {
	var tokens []entity.RegistrationTokenOpenFL
	err := db.Where("federation_uuid = ?", federationUUID).Find(&tokens).Error
	if err != nil {
		return 0, err
	}
	return tokens, nil
}

func (r *RegistrationTokenOpenFLRepo) DeleteByFederation(federationUUID string) error {
	return db.Unscoped().Where("federation_uuid = ?", federationUUID).Delete(&entity.RegistrationTokenOpenFL{}).Error
}

func (r *RegistrationTokenOpenFLRepo) DeleteByUUID(uuid string) error {
	return db.Unscoped().Where("uuid = ?", uuid).Delete(&entity.RegistrationTokenOpenFL{}).Error
}

func (r *RegistrationTokenOpenFLRepo) GetByUUID(uuid string) (interface{}, error) {
	token := &entity.RegistrationTokenOpenFL{}
	if err := db.Where("uuid = ?", uuid).First(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func (r *RegistrationTokenOpenFLRepo) LoadByTypeAndStr(instance interface{}) error {
	token := instance.(*entity.RegistrationTokenOpenFL)
	if err := db.Where("token_type = ? AND token_str = ?", token.TokenType, token.TokenStr).First(token).Error; err != nil {
		return err
	}
	return nil
}

// InitTable makes sure the table is created in the db
func (r *RegistrationTokenOpenFLRepo) InitTable() {
	if err := db.AutoMigrate(entity.RegistrationTokenOpenFL{}); err != nil {
		panic(err)
	}
}
