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

import "gorm.io/gorm"

// ProjectParticipant is a site joining a project
type ProjectParticipant struct {
	gorm.Model
	UUID            string                   `json:"uuid" gorm:"type:varchar(36)"`
	ProjectUUID     string                   `json:"project_uuid" gorm:"type:varchar(36)"`
	SiteUUID        string                   `json:"site_uuid" gorm:"type:varchar(36)"`
	SiteName        string                   `json:"site_name" gorm:"type:varchar(255)"`
	SitePartyID     uint                     `json:"site_party_id"`
	SiteDescription string                   `json:"site_description"`
	Status          ProjectParticipantStatus `json:"status"`
}

// ProjectParticipantStatus is the status of the current participant
type ProjectParticipantStatus uint8

const (
	ProjectParticipantStatusUnknown ProjectParticipantStatus = iota
	ProjectParticipantStatusOwner
	ProjectParticipantStatusPending
	ProjectParticipantStatusJoined
	ProjectParticipantStatusRejected
	ProjectParticipantStatusLeft
	ProjectParticipantStatusDismissed
	ProjectParticipantStatusRevoked
)
