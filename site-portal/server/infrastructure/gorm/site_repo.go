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
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// SiteRepo is the implementation of the domain's repo interface
type SiteRepo struct{}

var _ repo.SiteRepository = (*SiteRepo)(nil)

// ErrSiteExist means the site info is already there
var ErrSiteExist = errors.New("site already configured")

// Load reads the site information
func (r *SiteRepo) Load(instance interface{}) error {
	site := instance.(*entity.Site)
	return db.First(site).Error
}

// GetSite returns the site information
func (r *SiteRepo) GetSite() (interface{}, error) {
	var s entity.Site
	if err := db.First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// CreateSite inserts a site info entry
func (r *SiteRepo) CreateSite(site *entity.Site) error {
	var count int64
	db.Model(&entity.Site{}).Count(&count)
	if count > 0 {
		return ErrSiteExist
	}
	if err := db.Model(&entity.Site{}).Create(site).Error; err != nil {
		return err
	}
	return nil
}

// Update updates the site info
func (r *SiteRepo) Update(instance interface{}) error {
	updatedSite := instance.(*entity.Site)
	return db.Save(updatedSite).Error
}

// UpdateFMLManagerConnectionStatus updates fml manager related information
func (r *SiteRepo) UpdateFMLManagerConnectionStatus(instance interface{}) error {
	updatedSite := instance.(*entity.Site)
	return db.Model(updatedSite).Where("id = ?", updatedSite.ID).
		Select("fml_manager_endpoint", "fml_manager_server_name", "fml_manager_connected_at", "fml_manager_connected").
		Updates(updatedSite).Error
}

// InitTable make sure the table is created in the db
func (r *SiteRepo) InitTable() {
	if err := db.AutoMigrate(entity.Site{}); err != nil {
		panic(err)
	}
}

// InitData inserts an empty site info
func (r *SiteRepo) InitData() {
	site := &entity.Site{
		UUID: uuid.NewV4().String(),
	}
	if err := r.CreateSite(site); err != nil && err != ErrSiteExist {
		panic(err)
	}
}
