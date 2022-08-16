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
)

// CertificateRepo is the implementation of the repo.CertificateRepository interface
type CertificateRepo struct{}

var _ repo.CertificateRepository = (*CertificateRepo)(nil)

func (r *CertificateRepo) GetBySerialNumber(serialNumberStr string) (interface{}, error) {
	cert := &entity.Certificate{}
	if err := db.Where("serial_number_str = ?", serialNumberStr).First(cert).Error; err != nil {
		return nil, err
	}
	return cert, nil
}

func (r *CertificateRepo) Create(instance interface{}) error {
	cert := instance.(*entity.Certificate)

	if err := db.Create(cert).Error; err != nil {
		return err
	}
	return nil
}

func (r *CertificateRepo) List() (interface{}, error) {
	var certs []entity.Certificate
	if err := db.Find(&certs).Error; err != nil {
		return nil, err
	}
	return certs, nil
}

func (r *CertificateRepo) DeleteByUUID(uuid string) error {
	return db.Where("uuid = ?", uuid).Delete(&entity.Certificate{}).Error
}

func (r *CertificateRepo) GetByUUID(uuid string) (interface{}, error) {
	cert := &entity.Certificate{}
	if err := db.Where("uuid = ?", uuid).First(cert).Error; err != nil {
		return nil, err
	}
	return cert, nil
}

// InitTable makes sure the table is created in the db
func (r *CertificateRepo) InitTable() {
	if err := db.AutoMigrate(entity.Certificate{}); err != nil {
		panic(err)
	}
}
