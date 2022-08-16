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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"gorm.io/gorm"
	"time"
)

// ProjectData is a data association from a site to a project
type ProjectData struct {
	gorm.Model
	Name           string `gorm:"type:varchar(255)"`
	Description    string `gorm:"type:text"`
	UUID           string `gorm:"type:varchar(36)"`
	ProjectUUID    string `gorm:"type:varchar(36)"`
	DataUUID       string `gorm:"type:varchar(36)"`
	SiteUUID       string `gorm:"type:varchar(36)"`
	SiteName       string `gorm:"type:varchar(255)"`
	SitePartyID    uint
	Status         ProjectDataStatus
	TableName      string `gorm:"type:varchar(255)"`
	TableNamespace string `gorm:"type:varchar(255)"`
	CreationTime   time.Time
	UpdateTime     time.Time
	Repo           repo.ProjectDataRepository `gorm:"-"`
}

// ProjectDataStatus is the status of the association
type ProjectDataStatus uint8

const (
	ProjectDataStatusUnknown ProjectDataStatus = iota
	ProjectDataStatusDismissed
	ProjectDataStatusAssociated
)
