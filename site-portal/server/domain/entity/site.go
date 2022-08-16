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

package entity

import (
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/event"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fmlmanager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// Site contains all the info for the current site
type Site struct {
	gorm.Model
	UUID string `json:"uuid" gorm:"type:varchar(36);index;unique"`
	// Name is the site's name
	Name string `json:"name" gorm:"type:varchar(255);unique;not null"`
	// Description contains more text about this site
	Description string `json:"description" gorm:"type:text"`
	// PartyID is the id of this party
	PartyID uint `json:"party_id" gorm:"column:party_id"`
	// ExternalHost is the IP or hostname this site portal service is exposed
	ExternalHost string `json:"external_host" gorm:"type:varchar(255);column:external_ip"`
	// ExternalPort the port number this site portal service is exposed
	ExternalPort uint `json:"external_port" gorm:"column:external_port"`
	// HTTPS choose if site portal has HTTPS enabled
	HTTPS bool `json:"https" gorm:"column:https"`
	// FMLManagerEndpoint is of format "<http or https>://<host>:<port>"
	FMLManagerEndpoint string `json:"fml_manager_endpoint" gorm:"type:varchar(255);column:fml_manager_endpoint"`
	// FMLManagerServerName is used to verify FML Manager's certificate
	FMLManagerServerName string `json:"fml_manager_server_name" gorm:"type:varchar(255);column:fml_manager_server_name"`
	// FMLManagerConnectedAt is the last time this portal has registered to a FML manager
	FMLManagerConnectedAt time.Time `json:"fml_manager_connected_at" gorm:"column:fml_manager_connected_at"`
	// FMLManagerConnected is whether the portal is connected to FML manager
	FMLManagerConnected bool `json:"fml_manager_connected" gorm:"column:fml_manager_connected"`
	// FATEFlowHost is the host address of the FATE-flow service
	FATEFlowHost string `json:"fate_flow_host" gorm:"type:varchar(255);column:fate_flow_host"`
	// FATEFlowHTTPPort is the http port number of the FATE-flow service
	FATEFlowHTTPPort uint `json:"fate_flow_http_port" gorm:"column:fate_flow_http_port"`
	// FATEFlowGRPCPort is the grpc port number of the FATE-flow service, currently not used
	FATEFlowGRPCPort uint `json:"fate_flow_grpc_port" gorm:"column:fate_flow_grpc_port"`
	// FATEFlowConnectedAt is the last time this portal connected to the FATE flow
	FATEFlowConnectedAt time.Time `json:"fate_flow_connected_at" gorm:"column:fate_flow_connected_at"`
	// FATEFlowConnected is whether the portal has connected to FATEFlow after the address is configured
	FATEFlowConnected bool `json:"fate_flow_connected" gorm:"column:fate_flow_connected"`
	// KubeflowConfig records the Kubeflow related information for deploying horizontal model to the KFServing system
	KubeflowConfig valueobject.KubeflowConfig `json:"kubeflow_config" gorm:"type:text;column:kubeflow_config"`
	// KubeflowConnectedAt is the last time this portal has successfully connected all related Kubeflow service
	KubeflowConnectedAt time.Time `json:"kubeflow_connected_at" gorm:"column:kubeflow_connected_at"`
	// KubeflowConnected is whether this site has connected to the Kubeflow since it is configured
	KubeflowConnected bool `json:"kubeflow_connected" gorm:"column:kubeflow_connected"`
	// Repo is the repository interface
	Repo repo.SiteRepository `json:"-" gorm:"-"`
}

// Validate if the site has been properly configured
func (site *Site) Validate() error {
	if site.Name == "" || site.PartyID == 0 {
		return errors.New("Name or Party ID missing")
	}
	return nil
}

// Load site information from repository
func (site *Site) Load() error {
	return site.Repo.Load(site)
}

// UpdateConfigurableInfo changes the site information
func (site *Site) UpdateConfigurableInfo(updatedSite *Site) error {
	// load the site info first
	if err := site.Load(); err != nil {
		return err
	}

	// set the FML manager connected flag to false if key infos are changed
	if updatedSite.FMLManagerEndpoint != site.FMLManagerEndpoint ||
		updatedSite.Name != site.Name ||
		updatedSite.Description != site.Description ||
		updatedSite.PartyID != site.PartyID ||
		updatedSite.ExternalHost != site.ExternalHost ||
		updatedSite.ExternalPort != site.ExternalPort ||
		updatedSite.HTTPS != site.HTTPS ||
		updatedSite.FMLManagerServerName != site.FMLManagerServerName {
		log.Info().Msgf("site info or FML manager info changed, marking site as unregistered")
		site.FMLManagerConnected = false
	}
	// only writes the configurable info
	site.Name = updatedSite.Name
	site.Description = updatedSite.Description
	site.PartyID = updatedSite.PartyID
	site.ExternalHost = updatedSite.ExternalHost
	site.ExternalPort = updatedSite.ExternalPort
	site.FMLManagerEndpoint = updatedSite.FMLManagerEndpoint
	site.FMLManagerServerName = updatedSite.FMLManagerServerName
	site.FATEFlowHost = updatedSite.FATEFlowHost
	site.FATEFlowHTTPPort = updatedSite.FATEFlowHTTPPort
	site.FATEFlowGRPCPort = updatedSite.FATEFlowGRPCPort
	site.KubeflowConfig = updatedSite.KubeflowConfig
	site.HTTPS = updatedSite.HTTPS
	if err := site.Repo.Update(site); err != nil {
		return err
	}
	go func() {
		if err := event.NewSelfHttpExchange().PostEvent(event.ProjectParticipantUpdateEvent{
			UUID:        site.UUID,
			PartyID:     site.PartyID,
			Name:        site.Name,
			Description: site.Description,
		}); err != nil {
			log.Err(err).Msgf("failed to post site info update event")
		}
	}()
	return nil
}

// ConnectFATEFlow try to issue a test request to the FATE-flow service
func (site *Site) ConnectFATEFlow(host string, port uint, https bool) error {
	client := fateclient.NewFATEFlowClient(host, port, https)
	return client.TestConnection()
}

// RegisterToFMLManager registers this site to the FML manager service
func (site *Site) RegisterToFMLManager(endpoint string, serverName string) error {
	// load the site info first
	if err := site.Load(); err != nil {
		return err
	}
	if site.Name == "" || site.PartyID == 0 || site.ExternalHost == "" || site.ExternalPort == 0 {
		return errors.New("site info incomplete")
	}
	endpoint = strings.TrimSpace(endpoint)
	serverName = strings.TrimSpace(serverName)
	if !strings.HasPrefix(endpoint, "http") {
		return errors.New("invalid endpoint: http:// or https:// schema is needed")
	}
	sitePortalCommonName := strings.TrimSpace(viper.GetString("siteportal.tls.common.name"))
	if site.HTTPS && sitePortalCommonName == "" {
		return errors.New("missing required variable 'SITEPORTAL_TLS_COMMON_NAME' when site is using HTTPs")
	}
	endpoint = strings.TrimSuffix(endpoint, "/")
	client := fmlmanager.NewFMLManagerClient(endpoint, serverName)
	log.Info().Msgf("connecting FML manager at %s", endpoint)
	err := client.CreateSite(&fmlmanager.Site{
		UUID:         site.UUID,
		Name:         site.Name,
		Description:  site.Description,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   sitePortalCommonName,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to connect to fml-manager at %s", endpoint)
	}
	log.Info().Msgf("connected to fml manager at %s", endpoint)
	wasConnected := site.FMLManagerConnected
	site.FMLManagerConnected = true
	site.FMLManagerConnectedAt = time.Now()
	site.FMLManagerEndpoint = endpoint
	site.FMLManagerServerName = serverName
	if err := site.Repo.UpdateFMLManagerConnectionStatus(site); err != nil {
		return errors.Wrapf(err, "failed to update fml connection status")
	}
	// sync remote projects
	if wasConnected == false {
		go func() {
			log.Info().Msg("syncing projects list after re-connected to FML manager")
			exchange := event.NewSelfHttpExchange()
			if err := exchange.PostEvent(event.ProjectListSyncEvent{}); err != nil {
				log.Err(err).Msg("failed to sync projects list")
				return
			}
			log.Info().Msg("done syncing projects list after re-connected to FML manager")
		}()
	}
	return nil
}
