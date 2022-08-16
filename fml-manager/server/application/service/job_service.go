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

package service

import (
	"encoding/json"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/service"
	"github.com/pkg/errors"
)

// JobApp provides functions to handle job related events
type JobApp struct {
	SiteRepo        repo.SiteRepository
	JobRepo         repo.JobRepository
	ParticipantRepo repo.JobParticipantRepository
	ProjectRepo     repo.ProjectRepository
	ProjectDataRepo repo.ProjectDataRepository
}

// JobDataBase describes one data configuration for a job
type JobDataBase struct {
	DataUUID  string `json:"data_uuid"`
	LabelName string `json:"label_name"`
}

// JobApprovalContext contains the issuing site and the approval result
type JobApprovalContext struct {
	SiteUUID string `json:"site_uuid"`
	Approved bool   `json:"approved"`
}

// JobRemoteJobCreationRequest is the structure containing necessary info to create a job
type JobRemoteJobCreationRequest struct {
	UUID                   string                  `json:"uuid"`
	ConfJson               string                  `json:"conf_json"`
	DSLJson                string                  `json:"dsl_json"`
	Name                   string                  `json:"name"`
	Description            string                  `json:"description"`
	Type                   entity.JobType          `json:"type"`
	ProjectUUID            string                  `json:"project_uuid"`
	InitiatorData          JobDataBase             `json:"initiator_data"`
	OtherData              []JobDataBase           `json:"other_site_data"`
	ValidationEnabled      bool                    `json:"training_validation_enabled"`
	ValidationSizePercent  uint                    `json:"training_validation_percent"`
	ModelName              string                  `json:"training_model_name"`
	AlgorithmType          entity.JobAlgorithmType `json:"training_algorithm_type"`
	AlgorithmComponentName string                  `json:"algorithm_component_name"`
	EvaluateComponentName  string                  `json:"evaluate_component_name"`
	ComponentsToDeploy     []string                `json:"training_component_list_to_deploy"`
	ModelUUID              string                  `json:"predicting_model_uuid"`
	Username               string                  `json:"username"`
}

// JobStatusUpdateContext contain necessary info for updating job status, including status of the participants
type JobStatusUpdateContext struct {
	Status               entity.JobStatus                       `json:"status"`
	StatusMessage        string                                 `json:"status_message"`
	FATEJobID            string                                 `json:"fate_job_id"`
	FATEJobStatus        string                                 `json:"fate_job_status"`
	FATEModelID          string                                 `json:"fate_model_id"`
	FATEModelVersion     string                                 `json:"fate_model_version"`
	ParticipantStatusMap map[string]entity.JobParticipantStatus `json:"participant_status_map"`
}

// ProcessJobCreationRequest builds the creation requests and calls the domain service
func (app *JobApp) ProcessJobCreationRequest(request *JobRemoteJobCreationRequest) error {
	projectDataInstance, err := app.ProjectDataRepo.GetByProjectAndDataUUID(request.ProjectUUID, request.InitiatorData.DataUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get initiator data")
	}
	projectData := projectDataInstance.(*entity.ProjectData)

	siteInstance, err := app.SiteRepo.GetByUUID(projectData.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find initiating site")
	}
	site := siteInstance.(*entity.Site)

	requestJsonByte, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return err
	}
	requestJsonStr := string(requestJsonByte)

	jobService := &service.JobService{
		JobRepo:         app.JobRepo,
		ParticipantRepo: app.ParticipantRepo,
	}

	creationRequest := &service.JobCreationRequest{
		Job: &entity.Job{
			Name:          request.Name,
			Description:   request.Description,
			UUID:          request.UUID,
			ProjectUUID:   request.ProjectUUID,
			Type:          request.Type,
			AlgorithmType: request.AlgorithmType,
			AlgorithmConfig: entity.AlgorithmConfig{
				TrainingValidationEnabled:     request.ValidationEnabled,
				TrainingValidationSizePercent: request.ValidationSizePercent,
				TrainingComponentsToDeploy:    request.ComponentsToDeploy,
			},
			ModelName:             request.ModelName,
			PredictingModelUUID:   request.ModelUUID,
			InitiatingSiteUUID:    site.UUID,
			InitiatingSiteName:    site.Name,
			InitiatingSitePartyID: site.PartyID,
			InitiatingUser:        request.Username,
			Conf:                  request.ConfJson,
			DSL:                   request.DSLJson,
			RequestJson:           requestJsonStr,
			Repo:                  app.JobRepo,
		},
		Initiator: service.JobParticipantSiteInfo{
			JobParticipant: &entity.JobParticipant{
				JobUUID:            request.UUID,
				SiteUUID:           site.UUID,
				SiteName:           site.Name,
				SitePartyID:        site.PartyID,
				DataUUID:           projectData.DataUUID,
				DataName:           projectData.Name,
				DataDescription:    projectData.Description,
				DataTableName:      projectData.TableName,
				DataTableNamespace: projectData.TableNamespace,
				DataLabelName:      request.InitiatorData.LabelName,
				Repo:               app.ParticipantRepo,
			},
			JobParticipantConnectionInfo: service.JobParticipantConnectionInfo{
				ExternalHost: site.ExternalHost,
				ExternalPort: site.ExternalPort,
				HTTPS:        site.HTTPS,
				ServerName:   site.ServerName,
			},
		},
		Participants: map[string]service.JobParticipantSiteInfo{},
	}

	for _, otherData := range request.OtherData {
		projectDataInstance, err = app.ProjectDataRepo.GetByProjectAndDataUUID(request.ProjectUUID, otherData.DataUUID)
		if err != nil {
			return errors.Wrapf(err, "failed to get other site data: %s", otherData.DataUUID)
		}
		projectData = projectDataInstance.(*entity.ProjectData)

		siteInstance, err := app.SiteRepo.GetByUUID(projectData.SiteUUID)
		if err != nil {
			return errors.Wrapf(err, "failed to find other site")
		}
		otherSite := siteInstance.(*entity.Site)

		creationRequest.Participants[projectData.SiteUUID] = service.JobParticipantSiteInfo{
			JobParticipant: &entity.JobParticipant{
				JobUUID:            request.UUID,
				SiteUUID:           otherSite.UUID,
				SiteName:           otherSite.Name,
				SitePartyID:        otherSite.PartyID,
				DataUUID:           projectData.DataUUID,
				DataName:           projectData.Name,
				DataDescription:    projectData.Description,
				DataTableName:      projectData.TableName,
				DataTableNamespace: projectData.TableNamespace,
				DataLabelName:      request.InitiatorData.LabelName,
				Repo:               app.ParticipantRepo,
			},
			JobParticipantConnectionInfo: service.JobParticipantConnectionInfo{
				ExternalHost: otherSite.ExternalHost,
				ExternalPort: otherSite.ExternalPort,
				HTTPS:        otherSite.HTTPS,
				ServerName:   otherSite.ServerName,
			},
		}
	}
	return jobService.HandleNewJobCreation(creationRequest)
}

// ProcessJobApprovalResponse calls the domain service to process the job approval response
func (app *JobApp) ProcessJobApprovalResponse(jobUUID string, context *JobApprovalContext) error {
	jobInstance, err := app.JobRepo.GetByUUID(jobUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query job")
	}
	job := jobInstance.(*entity.Job)
	job.Repo = app.JobRepo

	siteInstance, err := app.SiteRepo.GetByUUID(job.InitiatingSiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find initiating site")
	}
	initiatingSite := siteInstance.(*entity.Site)

	siteInstance, err = app.SiteRepo.GetByUUID(context.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find initiating site")
	}
	approvingSite := siteInstance.(*entity.Site)

	jobParticipantInstance, err := app.ParticipantRepo.GetByJobAndSiteUUID(job.UUID, approvingSite.UUID)
	if err != nil {
		return errors.Wrap(err, "failed to get participant info")
	}
	jobParticipant := jobParticipantInstance.(*entity.JobParticipant)
	jobParticipant.Repo = app.ParticipantRepo

	response := &service.JobApprovalResponse{
		Initiator: service.JobParticipantConnectionInfo{
			ExternalHost: initiatingSite.ExternalHost,
			ExternalPort: initiatingSite.ExternalPort,
			HTTPS:        initiatingSite.HTTPS,
			ServerName:   initiatingSite.ServerName,
		},
		Participants:  map[string]service.JobParticipantConnectionInfo{},
		ApprovingSite: jobParticipant,
		Approved:      context.Approved,
		JobUUID:       jobUUID,
	}

	participantListInstance, err := app.ParticipantRepo.GetListByJobUUID(jobUUID)
	if err != nil {
		return err
	}
	participantList := participantListInstance.([]entity.JobParticipant)
	for index := range participantList {
		if participantList[index].SiteUUID != initiatingSite.UUID {
			siteInstance, err := app.SiteRepo.GetByUUID(participantList[index].SiteUUID)
			if err != nil {
				return errors.Wrapf(err, "failed to find initiating site")
			}
			site := siteInstance.(*entity.Site)
			response.Participants[site.UUID] = service.JobParticipantConnectionInfo{
				ExternalHost: site.ExternalHost,
				ExternalPort: site.ExternalPort,
				HTTPS:        site.HTTPS,
				ServerName:   site.ServerName,
			}
		}
	}
	jobService := &service.JobService{
		JobRepo:         app.JobRepo,
		ParticipantRepo: app.ParticipantRepo,
	}
	return jobService.HandleJobApprovalResponse(response)
}

// ProcessJobStatusUpdate calls the domain service to handle the job status update event
func (app *JobApp) ProcessJobStatusUpdate(jobUUID string, context *JobStatusUpdateContext) error {
	jobService := &service.JobService{
		JobRepo:         app.JobRepo,
		ParticipantRepo: app.ParticipantRepo,
	}

	jobInstance, err := app.JobRepo.GetByUUID(jobUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query job")
	}
	job := jobInstance.(*entity.Job)
	job.Repo = app.JobRepo

	updateContext := &service.JobStatusUpdateContext{
		JobUUID: jobUUID,
		NewJobStatus: &entity.Job{
			Status:           context.Status,
			StatusMessage:    context.StatusMessage,
			FATEJobID:        context.FATEJobID,
			FATEJobStatus:    context.FATEJobStatus,
			FATEModelID:      context.FATEModelID,
			FATEModelVersion: context.FATEModelVersion,
		},
		ParticipantStatusMap: map[string]service.JobParticipantStatusInfo{},
	}
	requestJsonByte, err := json.Marshal(context)
	if err != nil {
		return err
	}
	updateContext.RequestJson = string(requestJsonByte)
	for siteUUID := range context.ParticipantStatusMap {
		siteInstance, err := app.SiteRepo.GetByUUID(siteUUID)
		if err != nil {
			return errors.Wrapf(err, "failed to find initiating site")
		}
		site := siteInstance.(*entity.Site)
		updateContext.ParticipantStatusMap[siteUUID] = service.JobParticipantStatusInfo{
			JobParticipantStatus: context.ParticipantStatusMap[siteUUID],
			JobParticipantConnectionInfo: service.JobParticipantConnectionInfo{
				ExternalHost: site.ExternalHost,
				ExternalPort: site.ExternalPort,
				HTTPS:        site.HTTPS,
				ServerName:   site.ServerName,
			},
		}
	}
	return jobService.HandleJobStatusUpdate(updateContext)
}
