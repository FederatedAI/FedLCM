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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/service"
)

// SiteApp provide functions to manage the sites
type SiteApp struct {
	// SiteRepo is the repository for persisting site info
	SiteRepo repo.SiteRepository
}

// RegisterSite creates or updates the site info
func (app *SiteApp) RegisterSite(site *entity.Site) error {
	siteService := &service.SiteService{
		SiteRepo: app.SiteRepo,
	}
	return siteService.HandleSiteRegistration(site)
}

// GetSiteList returns all saved sites info
func (app *SiteApp) GetSiteList() ([]entity.Site, error) {
	list, err := app.SiteRepo.GetSiteList()
	return list.([]entity.Site), err
}
