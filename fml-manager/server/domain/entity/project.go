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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/valueobject"
	"gorm.io/gorm"
)

// Project contains jobs, sites and their data for collaborating via FATE
type Project struct {
	gorm.Model
	UUID                string `gorm:"type:varchar(36);index;unique"`
	Name                string `gorm:"type:varchar(255);not null"`
	Description         string `json:"description" gorm:"type:text"`
	AutoApprovalEnabled bool   `json:"auto_approval_enabled"`
	Status              ProjectStatus
	*valueobject.ProjectCreatorInfo
	Repo repo.ProjectRepository `json:"-" gorm:"-"`
}

// ProjectStatus is the status of a project
type ProjectStatus uint8

const (
	ProjectStatusUnknown ProjectStatus = iota
	ProjectStatusManaged
	ProjectStatusPending
	ProjectStatusJoined
	ProjectStatusRejected
	ProjectStatusLeft
	ProjectStatusClosed
	ProjectStatusDismissed
)
