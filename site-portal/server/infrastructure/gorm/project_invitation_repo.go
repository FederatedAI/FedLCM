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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
)

// ProjectInvitationRepo is the implementation of repo.ProjectInvitationRepository
type ProjectInvitationRepo struct{}

var _ repo.ProjectInvitationRepository = (*ProjectInvitationRepo)(nil)

func (r *ProjectInvitationRepo) Create(instance interface{}) error {
	newInvitation := instance.(*entity.ProjectInvitation)
	// Add records
	return db.Model(&entity.ProjectInvitation{}).Create(newInvitation).Error
}

func (r *ProjectInvitationRepo) UpdateStatusByUUID(instance interface{}) error {
	invitation := instance.(*entity.ProjectInvitation)
	return db.Model(&entity.ProjectInvitation{}).Where("uuid = ?", invitation.UUID).
		Update("status", invitation.Status).Error
}

func (r *ProjectInvitationRepo) GetByProjectUUID(uuid string) (interface{}, error) {
	invitation := &entity.ProjectInvitation{}
	if err := db.Model(&entity.ProjectInvitation{}).Where("project_uuid = ?", uuid).
		Last(invitation).Error; err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *ProjectInvitationRepo) GetByProjectAndSiteUUID(projectUUID, siteUUID string) (interface{}, error) {
	invitation := &entity.ProjectInvitation{}
	if err := db.Model(&entity.ProjectInvitation{}).
		Where("project_uuid = ? AND site_uuid = ?", projectUUID, siteUUID).
		Last(invitation).Error; err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *ProjectInvitationRepo) GetByUUID(uuid string) (interface{}, error) {
	invitation := &entity.ProjectInvitation{}
	if err := db.Model(&entity.ProjectInvitation{}).Where("uuid = ?", uuid).
		First(invitation).Error; err != nil {
		return nil, err
	}
	return invitation, nil
}

// InitTable make sure the table is created in the db
func (r *ProjectInvitationRepo) InitTable() {
	if err := db.AutoMigrate(&entity.ProjectInvitation{}); err != nil {
		panic(err)
	}
}
