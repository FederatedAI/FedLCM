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

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"gorm.io/gorm"
)

// ProjectData represents the data association in a project
type ProjectData struct {
	gorm.Model
	Name           string                     `json:"name" gorm:"type:varchar(255)"`
	Description    string                     `json:"description" gorm:"type:text"`
	UUID           string                     `json:"uuid" gorm:"type:varchar(36)"`
	ProjectUUID    string                     `json:"project_uuid" gorm:"type:varchar(36)"`
	DataUUID       string                     `json:"data_uuid" gorm:"type:varchar(36)"`
	SiteUUID       string                     `json:"site_uuid" gorm:"type:varchar(36)"`
	SiteName       string                     `json:"site_name" gorm:"type:varchar(255)"`
	SitePartyID    uint                       `json:"site_party_id"`
	Type           ProjectDataType            `json:"type"`
	Status         ProjectDataStatus          `json:"status"`
	TableName      string                     `json:"table_name" gorm:"type:varchar(255)"`
	TableNamespace string                     `json:"table_namespace" gorm:"type:varchar(255)"`
	CreationTime   time.Time                  `json:"creation_time"`
	UpdateTime     time.Time                  `json:"update_time"`
	Repo           repo.ProjectDataRepository `json:"-" gorm:"-"`
}

// ProjectDataType is the type of this association
type ProjectDataType uint8

const (
	ProjectDataTypeUnknown ProjectDataType = iota
	ProjectDataTypeLocal
	ProjectDataTypeRemote
)

// ProjectDataStatus is the status of this association
type ProjectDataStatus uint8

const (
	ProjectDataStatusUnknown ProjectDataStatus = iota
	ProjectDataStatusDismissed
	ProjectDataStatusAssociated
)
