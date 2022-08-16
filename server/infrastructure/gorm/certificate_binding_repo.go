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

// CertificateBindingRepo is the implementation of the repo.CertificateBindingRepository interface
type CertificateBindingRepo struct{}

var _ repo.CertificateBindingRepository = (*CertificateBindingRepo)(nil)

func (r *CertificateBindingRepo) ListByParticipantUUID(participantUUID string) (interface{}, error) {
	var bindings []entity.CertificateBinding
	err := db.Where("participant_uuid = ?", participantUUID).Find(&bindings).Error
	if err != nil {
		return 0, err
	}
	return bindings, nil
}

func (r *CertificateBindingRepo) Create(instance interface{}) error {
	binding := instance.(*entity.CertificateBinding)

	if err := db.Create(binding).Error; err != nil {
		return err
	}
	return nil
}

func (r *CertificateBindingRepo) ListByCertificateUUID(certificateUUID string) (interface{}, error) {
	var bindings []entity.CertificateBinding
	err := db.Where("certificate_uuid = ?", certificateUUID).Find(&bindings).Error
	if err != nil {
		return 0, err
	}
	return bindings, nil
}

func (r *CertificateBindingRepo) DeleteByParticipantUUID(participantUUID string) error {
	return db.Unscoped().Where("participant_uuid = ?", participantUUID).Delete(&entity.CertificateBinding{}).Error
}

// InitTable makes sure the table is created in the db
func (r *CertificateBindingRepo) InitTable() {
	if err := db.AutoMigrate(entity.CertificateBinding{}); err != nil {
		panic(err)
	}
}
