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

package fmlmanager

import (
	"time"
)

// CommonResponse is the structure of all FML manager response
type CommonResponse struct {
	Code    int         `json:"code" example:"0"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data" swaggertype:"object"`
}

// Site contains all the info for the current site
type Site struct {
	UUID string `json:"uuid"`
	// Name is the user's name
	Name string `json:"name"`
	// Description contains more text about this site
	Description string `json:"description"`
	// PartyID is the id of this party
	PartyID uint `json:"party_id"`
	// ExternalHost is the IP or hostname this site portal service is exposed
	ExternalHost string `json:"external_host"`
	// ExternalPort the port number this site portal service is exposed
	ExternalPort uint `json:"external_port"`
	// HTTPS choose if site portal enable HTTPS, 'true' use HTTPS, 'false'use HTTPS
	HTTPS bool `json:"https"`
	// ServerName is used by FML Manager to verify site portal's certificate when HTTPs is enabled
	ServerName string `json:"server_name"`
}

// ProjectInvitation is the invitation we send to FML manager for inviting a site
// We send targeting site uuid as well as project info and data association info
// so it can be created in the FML manager
type ProjectInvitation struct {
	UUID                       string                   `json:"uuid"`
	SiteUUID                   string                   `json:"site_uuid"`
	SitePartyID                uint                     `json:"site_party_id"`
	ProjectUUID                string                   `json:"project_uuid"`
	ProjectName                string                   `json:"project_name"`
	ProjectDescription         string                   `json:"project_description"`
	ProjectAutoApprovalEnabled bool                     `json:"project_auto_approval_enabled"`
	ProjectManager             string                   `json:"project_manager"`
	ProjectManagingSiteName    string                   `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint                     `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string                   `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time                `json:"project_creation_time"`
	AssociatedData             []ProjectDataAssociation `json:"associated_data"`
}

// ProjectDataAssociationBase contains the basic info of an association
type ProjectDataAssociationBase struct {
	DataUUID string `json:"data_uuid"`
}

// ProjectDataAssociation contains detailed info of an association
type ProjectDataAssociation struct {
	ProjectDataAssociationBase
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	SiteName       string    `json:"site_name"`
	SiteUUID       string    `json:"site_uuid"`
	SitePartyID    uint      `json:"site_party_id"`
	TableName      string    `json:"table_name"`
	TableNamespace string    `json:"table_namespace"`
	CreationTime   time.Time `json:"creation_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// ProjectParticipant is a site in a project
type ProjectParticipant struct {
	UUID            string `json:"uuid"`
	ProjectUUID     string `json:"project_uuid"`
	SiteUUID        string `json:"site_uuid"`
	SiteName        string `json:"site_name"`
	SitePartyID     uint   `json:"site_party_id"`
	SiteDescription string `json:"site_description"`
	Status          uint   `json:"status"`
}

// ProjectInfoWithStatus contains project basic information and the status inferred for certain participant
type ProjectInfoWithStatus struct {
	ProjectUUID                string    `json:"project_uuid"`
	ProjectName                string    `json:"project_name"`
	ProjectDescription         string    `json:"project_description"`
	ProjectAutoApprovalEnabled bool      `json:"project_auto_approval_enabled"`
	ProjectManager             string    `json:"project_manager"`
	ProjectManagingSiteName    string    `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint      `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string    `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time `json:"project_creation_time"`
	ProjectStatus              uint      `json:"project_status"`
}

// JobDataBase describes one data configuration for a job
type JobDataBase struct {
	DataUUID  string `json:"data_uuid"`
	LabelName string `json:"label_name"`
}

// JobRemoteJobCreationRequest is the structure containing necessary info to create a job
type JobRemoteJobCreationRequest struct {
	UUID                   string        `json:"uuid"`
	ConfJson               string        `json:"conf_json"`
	DSLJson                string        `json:"dsl_json"`
	Name                   string        `json:"name"`
	Description            string        `json:"description"`
	Type                   uint8         `json:"type"`
	ProjectUUID            string        `json:"project_uuid"`
	InitiatorData          JobDataBase   `json:"initiator_data"`
	OtherData              []JobDataBase `json:"other_site_data"`
	ValidationEnabled      bool          `json:"training_validation_enabled"`
	ValidationSizePercent  uint          `json:"training_validation_percent"`
	ModelName              string        `json:"training_model_name"`
	AlgorithmType          uint8         `json:"training_algorithm_type"`
	AlgorithmComponentName string        `json:"algorithm_component_name"`
	EvaluateComponentName  string        `json:"evaluate_component_name"`
	ComponentsToDeploy     []string      `json:"training_component_list_to_deploy"`
	ModelUUID              string        `json:"predicting_model_uuid"`
	Username               string        `json:"username"`
}

// JobApprovalContext contains the issuing site and the approval result
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
