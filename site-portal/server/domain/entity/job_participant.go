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
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// JobParticipant represents a site and its data for a job
type JobParticipant struct {
	gorm.Model
	UUID               string `gorm:"type:varchar(36)"`
	JobUUID            string `gorm:"type:varchar(36)"`
	SiteUUID           string `gorm:"type:varchar(36)"`
	SiteName           string `gorm:"type:varchar(255)"`
	SitePartyID        uint
	SiteRole           JobParticipantRole `gorm:"type:varchar(255)"`
	DataUUID           string             `gorm:"type:varchar(36)"`
	DataName           string             `gorm:"type:varchar(255)"`
	DataDescription    string             `gorm:"type:text"`
	DataTableName      string             `gorm:"type:varchar(255)"`
	DataTableNamespace string             `gorm:"type:varchar(255)"`
	DataLabelName      string             `gorm:"type:varchar(255)"`
	Status             JobParticipantStatus
	Repo               repo.JobParticipantRepository `gorm:"-"`
}

// JobParticipantStatus is the status of this participant in the job
type JobParticipantStatus uint8

const (
	JobParticipantStatusUnknown JobParticipantStatus = iota
	JobParticipantStatusInitiator
	JobParticipantStatusPending
	JobParticipantStatusApproved
	JobParticipantStatusRejected
)

func (s JobParticipantStatus) String() string {
	names := map[JobParticipantStatus]string{
		JobParticipantStatusUnknown:   "Unknown",
		JobParticipantStatusPending:   "Pending",
		JobParticipantStatusRejected:  "Rejected",
		JobParticipantStatusApproved:  "Approved",
		JobParticipantStatusInitiator: "Auto-approved as Initiator",
	}
	return names[s]
}

// JobParticipantRole is the enum of roles of a participant
type JobParticipantRole string

const (
	JobParticipantRoleGuest JobParticipantRole = "guest"
	JobParticipantRoleHost  JobParticipantRole = "host"
)

// Create initialize the participant info and create it in the repo
func (p *JobParticipant) Create() error {
	p.UUID = uuid.NewV4().String()
	if err := p.Repo.Create(p); err != nil {
		return err
	}
	return nil
}

// UpdateStatus changes the participant's status
func (p *JobParticipant) UpdateStatus(status JobParticipantStatus) error {
	p.Status = status
	return p.Repo.UpdateStatusByUUID(p)
}
