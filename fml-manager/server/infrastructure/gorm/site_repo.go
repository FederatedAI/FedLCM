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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
)

// SiteRepo is the implementation of the domain's repo interface
type SiteRepo struct{}

// make sure UserRepo implements the repo.UserRepository interface
var _ repo.SiteRepository = (*SiteRepo)(nil)

// GetSiteList returns all the saved sites
func (r *SiteRepo) GetSiteList() (interface{}, error) {
	var sites []entity.Site
	if err := db.Find(&sites).Error; err != nil {
		return nil, err
	}
	return sites, nil
}

// Save create a record of site in the DB
func (r *SiteRepo) Save(instance interface{}) (interface{}, error) {
	site := instance.(*entity.Site)

	var count int64
	db.Model(&entity.Site{}).Where("name = ?", site.Name).Count(&count)
	if count > 0 {
		return nil, repo.ErrSiteNameConflict
	}

	if err := db.Model(&entity.Site{}).Create(site).Error; err != nil {
		return nil, err
	}
	return site, nil
}

// ExistByUUID returns whether the site with the uuid exists
func (r *SiteRepo) ExistByUUID(uuid string) (bool, error) {
	var count int64
	err := db.Model(&entity.Site{}).Unscoped().Where("uuid = ?", uuid).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateByUUID updates the site info indexed by the uuid
func (r *SiteRepo) UpdateByUUID(instance interface{}) error {
	updatedSite := instance.(*entity.Site)
	return db.Model(&entity.Site{}).Where("uuid = ?", updatedSite.UUID).Updates(updatedSite).Error
}

// DeleteByUUID deletes records using the specified uuid
func (r *SiteRepo) DeleteByUUID(uuid string) error {
	return db.Model(&entity.Site{}).Unscoped().Where("uuid = ?", uuid).Delete(&entity.Site{}).Error
}

// GetByUUID returns an *entity.Site indexed by the uuid
func (r *SiteRepo) GetByUUID(uuid string) (interface{}, error) {
	site := &entity.Site{}
	if err := db.Model(&entity.Site{}).Where("uuid = ?", uuid).First(site).Error; err != nil {
		return nil, err
	}
	return site, nil
}

// InitTable make sure the table is created in the db
func (r *SiteRepo) InitTable() {
	if err := db.AutoMigrate(entity.Site{}); err != nil {
		panic(err)
	}
}
