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
	"time"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
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
	// HTTPS indicate whether the endpoint is over https
	HTTPS bool `json:"https"`
	// ServerName is used by fml manager to verify endpoint's certificate when HTTPS is enabled
	ServerName string `json:"server_name"`
	// LastRegisteredAt is the last time this site has tried to register to the manager
	LastRegisteredAt time.Time `json:"last_connected_at"`
	// Repo is the repository interface
	Repo repo.SiteRepository `json:"-" gorm:"-"`
}
