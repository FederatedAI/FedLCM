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

package api

import (
	"net/http"

	"github.com/FederatedAI/FedLCM/fml-manager/server/application/service"
	"github.com/FederatedAI/FedLCM/fml-manager/server/constants"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// JobController handles job related APIs
type JobController struct {
	jobApp *service.JobApp
}

// NewJobController returns a controller instance to handle job API requests
func NewJobController(jobRepo repo.JobRepository,
	jobParticipantRepo repo.JobParticipantRepository,
	projectRepo repo.ProjectRepository,
	siteRepo repo.SiteRepository,
	projectDataRepo repo.ProjectDataRepository,
) *JobController {
	return &JobController{
		jobApp: &service.JobApp{
			SiteRepo:        siteRepo,
			JobRepo:         jobRepo,
			ProjectRepo:     projectRepo,
			ParticipantRepo: jobParticipantRepo,
			ProjectDataRepo: projectDataRepo,
		},
	}
}

// Route set up route mappings to job related APIs
func (controller *JobController) Route(r *gin.RouterGroup) {
	job := r.Group("job")
	if viper.GetBool("fmlmanager.tls.enabled") {
		job.Use(certAuthenticator())
	}
	{
		job.POST("/create", controller.handleJobCreation)
		job.POST("/:uuid/response", controller.handleJobResponse)
		job.POST("/:uuid/status", controller.handleJobStatusUpdate)
	}
}

// handleJobCreation process a job creation request
//	@Summary	Process job creation
//	@Tags		Job
//	@Produce	json
//	@Param		project	body		service.JobRemoteJobCreationRequest	true	"job creation request"
//	@Success	200		{object}	GeneralResponse{}					"Success"
//	@Failure	401		{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/job/create [post]
func (controller *JobController) handleJobCreation(c *gin.Context) {
	if err := func() error {
		creationRequest := &service.JobRemoteJobCreationRequest{}
		if err := c.ShouldBindJSON(creationRequest); err != nil {
			return err
		}
		return controller.jobApp.ProcessJobCreationRequest(creationRequest)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// handleJobResponse process a job approval response
//	@Summary	Process job response
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Param		project	body		service.JobApprovalContext	true	"job approval response"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid}/response [post]
func (controller *JobController) handleJobResponse(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		context := &service.JobApprovalContext{}
		if err := c.ShouldBindJSON(context); err != nil {
			return err
		}
		return controller.jobApp.ProcessJobApprovalResponse(jobUUID, context)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// handleJobStatusUpdate process a job status update request
//	@Summary	Process job status update
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string							true	"Job UUID"
//	@Param		project	body		service.JobStatusUpdateContext	true	"job status"
//	@Success	200		{object}	GeneralResponse{}				"Success"
//	@Failure	401		{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/job/{uuid}/status [post]
func (controller *JobController) handleJobStatusUpdate(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		context := &service.JobStatusUpdateContext{}
		if err := c.ShouldBindJSON(context); err != nil {
			return err
		}
		return controller.jobApp.ProcessJobStatusUpdate(jobUUID, context)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}
