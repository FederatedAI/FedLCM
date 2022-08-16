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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Project is a container for federated machine learning jobs
type Project struct {
	gorm.Model
	UUID string `gorm:"type:varchar(36);index;unique"`
	// Name is the name of the project
	Name string `gorm:"type:varchar(255);not null"`
	// Description contains more text about the project
	Description string `json:"description" gorm:"type:text"`
	// AutoApprovalEnabled is whether new jobs will be automatically approved
	AutoApprovalEnabled bool `json:"auto_approval_enabled"`
	// Type is the project type
	Type ProjectType
	// Status is the status of the project
	Status ProjectStatus
	// Creating/Managing site info
	valueobject.ProjectCreatorInfo
	// The repo for persistence
	Repo repo.ProjectRepository `json:"-" gorm:"-"`
}

// ProjectType is the project type
type ProjectType uint8

const (
	ProjectTypeLocal ProjectType = iota + 1
	// ProjectTypeFederatedLocal means the project is locally created and is tracked in the FML manager
	ProjectTypeFederatedLocal
	ProjectTypeRemote
)

// ProjectStatus is the status of a project
type ProjectStatus uint8

const (
	ProjectStatusManaged ProjectStatus = iota + 1
	ProjectStatusPending
	ProjectStatusJoined
	ProjectStatusRejected
	ProjectStatusLeft
	ProjectStatusClosed
	ProjectStatusDismissed
)

// Create creates the project
func (p *Project) Create() error {
	if p.Name == "" {
		return errors.New("empty project name")
	}
	if p.ProjectCreatorInfo.ManagingSiteName == "" || p.ProjectCreatorInfo.ManagingSitePartyID == 0 {
		return errors.New("site info not configured")
	}
	switch p.Type {
	case ProjectTypeLocal:
		p.UUID = uuid.NewV4().String()
		if err := p.Repo.CheckNameConflict(p.Name); err != nil {
			return err
		}
		p.Status = ProjectStatusManaged
	case ProjectTypeRemote:
		if p.UUID == "" {
			return errors.New("missing project uuid")
		}
		p.Status = ProjectStatusPending
	default:
		return errors.Errorf("invalid project type: %d", p.Type)
	}
	log.Info().Str("name", p.Name).Interface("type", p.Type).Msg("create project")
	return p.Repo.Create(p)
}
