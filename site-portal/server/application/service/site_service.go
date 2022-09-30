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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
)

// SiteApp provide functions to manage the site
type SiteApp struct {
	// SiteRepo is the repository for persisting site info
	SiteRepo repo.SiteRepository
}

// FATEFlowConnectionInfo represent connection info to a fate flow service
type FATEFlowConnectionInfo struct {
	// Host address
	Host string `json:"host"`
	// Port is the port number
	Port uint `json:"port"`
	// Https is whether https is enabled
	Https bool `json:"https"`
}

// FMLManagerConnectionInfo contains connection settings for the fml manager
type FMLManagerConnectionInfo struct {
	// Endpoint address starting with "http" or "https"
	Endpoint string `json:"endpoint"`
	//ServerName is used by Site Portal to verify FML Manager's certificate
	ServerName string `json:"server_name"`
}

// GetSite returns the site information
func (app *SiteApp) GetSite() (*entity.Site, error) {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, err
	}
	return site, nil
}

// UpdateSite updates the site info
func (app *SiteApp) UpdateSite(updatedSiteInfo *entity.Site) error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	return site.UpdateConfigurableInfo(updatedSiteInfo)
}

// TestFATEFlowConnection tests the connection to fate flow service
func (app *SiteApp) TestFATEFlowConnection(connectionInfo *FATEFlowConnectionInfo) error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	return site.ConnectFATEFlow(connectionInfo.Host, connectionInfo.Port, connectionInfo.Https)
}

// TestKubeflowConnection tests the connection to Kubernetes and if it has KFServing installed
func (app *SiteApp) TestKubeflowConnection(connectionInfo *valueobject.KubeflowConfig) error {
	return connectionInfo.Validate()
}

// RegisterToFMLManager connects to fml manager and register the current site
func (app *SiteApp) RegisterToFMLManager(connectionInfo *FMLManagerConnectionInfo) error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	return site.RegisterToFMLManager(connectionInfo.Endpoint, connectionInfo.ServerName)
}

// UnregisterFromFMLManager connects to fml manager and unregister the current site
func (app *SiteApp) UnregisterFromFMLManager() error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	return site.UnregisterFromFMLManager()
}
