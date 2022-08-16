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

// CertificateAuthorityRepo is the implementation of the repo.CertificateAuthorityRepository interface
type CertificateAuthorityRepo struct{}

var _ repo.CertificateAuthorityRepository = (*CertificateAuthorityRepo)(nil)

// ErrCertificateAuthorityExist means new CA cannot be created due to the existence of the same-name CA
var ErrCertificateAuthorityExist = errors.New("CA already exists")

func (r *CertificateAuthorityRepo) Create(instance interface{}) error {
	ca := instance.(*entity.CertificateAuthority)

	var count int64
	db.Model(&entity.CertificateAuthority{}).Count(&count)
	if count > 0 {
		return ErrCertificateAuthorityExist
	}

	if err := db.Create(ca).Error; err != nil {
		return err
	}
	return nil
}

func (r *CertificateAuthorityRepo) UpdateByUUID(instance interface{}) error {
	ca := instance.(*entity.CertificateAuthority)
	return db.Where("uuid = ?", ca.UUID).
		Select("name", "description", "type", "config_json").Updates(ca).Error
}

func (r *CertificateAuthorityRepo) GetFirst() (interface{}, error) {
	ca := &entity.CertificateAuthority{}
	if err := db.First(ca).Error; err != nil {
		return nil, err
	}
	return ca, nil
}

// InitTable makes sure the table is created in the db
func (r *CertificateAuthorityRepo) InitTable() {
	if err := db.AutoMigrate(entity.CertificateAuthority{}); err != nil {
		panic(err)
	}
}
