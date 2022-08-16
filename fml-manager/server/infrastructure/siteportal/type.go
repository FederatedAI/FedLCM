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

package siteportal

import (
	"time"
)

// ProjectInvitationRequest is an invitation for asking a site to join a project
type ProjectInvitationRequest struct {
	UUID                       string    `json:"uuid"`
	SiteUUID                   string    `json:"site_uuid"`
	SitePartyID                uint      `json:"site_party_id"`
	ProjectUUID                string    `json:"project_uuid"`
	ProjectName                string    `json:"project_name"`
	ProjectDescription         string    `json:"project_description"`
	ProjectAutoApprovalEnabled bool      `json:"project_auto_approval_enabled"`
	ProjectManager             string    `json:"project_manager"`
	ProjectManagingSiteName    string    `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint      `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string    `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time `json:"project_creation_time"`
}

// ProjectParticipant represents a site in a project
type ProjectParticipant struct {
	UUID            string `json:"uuid"`
	ProjectUUID     string `json:"project_uuid"`
	SiteUUID        string `json:"site_uuid"`
	SiteName        string `json:"site_name"`
	SitePartyID     uint   `json:"site_party_id"`
	SiteDescription string `json:"site_description"`
	Status          uint8  `json:"status"`
}

// ProjectData represents a data association in a project
type ProjectData struct {
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ProjectUUID    string    `json:"project_uuid"`
	DataUUID       string    `json:"data_uuid"`
	SiteUUID       string    `json:"site_uuid"`
	SiteName       string    `json:"site_name"`
	SitePartyID    uint      `json:"site_party_id"`
	TableName      string    `json:"table_name"`
	TableNamespace string    `json:"table_namespace"`
	CreationTime   time.Time `json:"creation_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// ProjectParticipantUpdateEvent represents a site info update event
type ProjectParticipantUpdateEvent struct {
	UUID        string `json:"uuid"`
	PartyID     uint   `json:"party_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// JobApprovalContext is the struct containing job approval response
type JobApprovalContext struct {
	SiteUUID string `json:"site_uuid"`
	Approved bool   `json:"approved"`
}

// JobStatusUpdateContext contains info of the updated job status
type JobStatusUpdateContext struct {
	Status               uint8            `json:"status"`
	StatusMessage        string           `json:"status_message"`
	FATEJobID            string           `json:"fate_job_id"`
	FATEJobStatus        string           `json:"fate_job_status"`
	FATEModelID          string           `json:"fate_model_id"`
	FATEModelVersion     string           `json:"fate_model_version"`
	ParticipantStatusMap map[string]uint8 `json:"participant_status_map"`
}
