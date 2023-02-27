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
	"net/http/httputil"

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	jwt "github.com/appleboy/gin-jwt/v2"
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
	modelRepo repo.ModelRepository) *JobController {
	return &JobController{
		jobApp: &service.JobApp{
			SiteRepo:        siteRepo,
			JobRepo:         jobRepo,
			ParticipantRepo: jobParticipantRepo,
			ProjectRepo:     projectRepo,
			ProjectDataRepo: projectDataRepo,
			ModelRepo:       modelRepo,
		},
	}
}

// Route set up route mappings to job related APIs
func (controller *JobController) Route(r *gin.RouterGroup) {
	job := r.Group("job")
	// internal APIs are used by FML manager only
	internal := job.Group("internal")
	if viper.GetBool("siteportal.tls.enabled") {
		internal.Use(certAuthenticator())
	}
	{
		internal.POST("/create", controller.createRemoteJob)
		internal.POST("/:uuid/response", controller.handleJobResponse)
		internal.POST("/:uuid/status", controller.handleJobStatusUpdate)
	}

	job.Use(authMiddleware.MiddlewareFunc())
	{
		job.POST("/:uuid/approve", controller.approveJob)
		job.POST("/:uuid/reject", controller.rejectJob)
		job.POST("/:uuid/refresh", controller.refreshJob)
		job.GET("/:uuid", controller.get)
		job.DELETE("/:uuid", controller.delete)
		job.POST("/conf/create", controller.generateConf)
		job.GET("/predict/participant", controller.getPredictingJobParticipant)
		job.GET("/:uuid/data-result/download", controller.downloadDataResult)
		job.GET("/components", controller.getJobComponents)
		job.POST("/generateDslFromDag", controller.generateDslFromDag)
		job.POST("/generateConfFromDag", controller.generateConfFromDag)
	}
}

// approveJob approves the job
//	@Summary	Approve a pending job
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid}/approve [post]
func (controller *JobController) approveJob(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		return controller.jobApp.Approve(jobUUID)
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

// rejectJob rejects the job
//	@Summary	Disapprove a pending job
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid}/reject [post]
func (controller *JobController) rejectJob(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		return controller.jobApp.Reject(jobUUID)
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

// refreshJob retrieve the latest job status
//	@Summary	Refresh the latest job status
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid}/refresh [post]
func (controller *JobController) refreshJob(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		return controller.jobApp.Refresh(jobUUID)
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

// get returns detailed information of a job
//	@Summary	Get job's detailed info
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string									true	"Job UUID"
//	@Success	200		{object}	GeneralResponse{data=service.JobDetail}	"Success"
//	@Failure	401		{object}	GeneralResponse							"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}				"Internal server error"
//	@Router		/job/{uuid} [get]
func (controller *JobController) get(c *gin.Context) {
	if job, err := func() (*service.JobDetail, error) {
		jobUUID := c.Param("uuid")
		return controller.jobApp.GetJobDetail(jobUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: job,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// generateConf returns job configuration and DSL content
//	@Summary	Get a job's config template, used in json template mode
//	@Tags		Job
//	@Produce	json
//	@Param		request	body		service.JobSubmissionRequest			true	"Job requests, not all fields are required: only need to fill related ones according to job type"
//	@Success	200		{object}	GeneralResponse{data=service.JobConf}	"Success"
//	@Failure	401		{object}	GeneralResponse							"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}				"Internal server error"
//	@Router		/job/conf/create [post]
func (controller *JobController) generateConf(c *gin.Context) {
	if conf, err := func() (*service.JobConf, error) {
		request := &service.JobSubmissionRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return nil, err
		}
		claims := jwt.ExtractClaims(c)
		// the auth middleware makes sure username exists
		username := claims[nameKey].(string)
		if conf, err := controller.jobApp.GenerateConfig(username, request); err != nil {
			return nil, err
		} else {
			return conf, nil
		}
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: conf,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createRemoteJob creates a new job
//	@Summary	Create a new job that is created by other site, only called by FML manager
//	@Tags		Job
//	@Produce	json
//	@Param		request	body		service.RemoteJobCreationRequest	true	"Job info"
//	@Success	200		{object}	GeneralResponse{}					"Success"
//	@Failure	401		{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/job/internal/create [post]
func (controller *JobController) createRemoteJob(c *gin.Context) {
	if err := func() error {
		request := &service.RemoteJobCreationRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return err
		}
		return controller.jobApp.ProcessNewRemoteJob(request)
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

// handleJobResponse process job approval response
//	@Summary	Handle job response from other sites, only called by FML manager
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Param		context	body		service.JobApprovalContext	true	"Approval context, containing the sender UUID and approval status"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/internal/{uuid}/response [post]
func (controller *JobController) handleJobResponse(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		approvalContext := &service.JobApprovalContext{}
		if err := c.ShouldBindJSON(approvalContext); err != nil {
			return err
		}
		return controller.jobApp.ProcessJobResponse(jobUUID, approvalContext)
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

// handleJobStatusUpdate process job status update
//	@Summary	Handle job status updates, only called by FML manager
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string							true	"Job UUID"
//	@Param		context	body		service.JobStatusUpdateContext	true	"Job status update context, containing the latest job and participant status"
//	@Success	200		{object}	GeneralResponse{}				"Success"
//	@Failure	401		{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/job/internal/{uuid}/status [post]
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

// getPredictingJobParticipant returns participant list for creating a predicting job from a model
//	@Summary	Get allowed participants for a predicting job from a model
//	@Tags		Job
//	@Produce	json
//	@Param		modelUUID	query		string													true	"UUID of a trained model"
//	@Success	200			{object}	GeneralResponse{data=[]service.JobParticipantInfoBase}	"Success"
//	@Failure	401			{object}	GeneralResponse											"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}								"Internal server error"
//	@Router		/job/predict/participant [get]
func (controller *JobController) getPredictingJobParticipant(c *gin.Context) {
	if data, err := func() ([]service.JobParticipantInfoBase, error) {
		modelUUID := c.Query("modelUUID")
		return controller.jobApp.GeneratePredictingJobParticipants(modelUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: data,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// downloadDataResult returns predict/PSI job result data
//	@Summary	the result data of a Predicting or PSI job, XXX: currently it will return an error message due to a bug in FATE
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid}/data-result/download [get]
func (controller *JobController) downloadDataResult(c *gin.Context) {
	if req, err := func() (*http.Request, error) {
		jobUUID := c.Param("uuid")
		return controller.jobApp.GetDataResultDownloadRequest(jobUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		proxy := &httputil.ReverseProxy{
			Director: func(*http.Request) {
				// no-op as we don't need to change the request
			},
		}
		proxy.ServeHTTP(c.Writer, req)
	}
}

// delete a job
//	@Summary	Delete the job. The job will be marked as delete in this site, but still viewable in other sites
//	@Tags		Job
//	@Produce	json
//	@Param		uuid	path		string						true	"Job UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/job/{uuid} [delete]
func (controller *JobController) delete(c *gin.Context) {
	if err := func() error {
		jobUUID := c.Param("uuid")
		return controller.jobApp.DeleteJob(jobUUID)
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

// getJobComponents returns all the components for a model and their default configs
//	@Summary	Get all the components and their default configs. The returned format is json
//	@Tags		Job
//	@Produce	json
//	@Success	200	{object}	GeneralResponse{data=string}	"Success"
//	@Failure	401	{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500	{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/job/components [get]
func (controller *JobController) getJobComponents(c *gin.Context) {
	resp := &GeneralResponse{
		Code: constants.RespNoErr,
		Data: controller.jobApp.LoadJobComponents(),
	}
	c.JSON(http.StatusOK, resp)
}

// generateDslFromDag returns the DSL json file from the DAG the user draw
//	@Summary	Generate the DSL json file from the DAG the user draw, should be called by UI only
//	@Tags		Job
//	@Produce	json
//	@Param		rawJson	body		service.JobRawDagJson			true	"The raw json, the value should be a serialized json string"
//	@Success	200		{object}	GeneralResponse{data=string}	"Success"
//	@Failure	401		{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/job/generateDslFromDag [post]
func (controller *JobController) generateDslFromDag(c *gin.Context) {
	if rawJson, err := func() (string, error) {
		request := &service.JobRawDagJson{}
		if err := c.ShouldBindJSON(request); err != nil {
			return "", err
		}
		return request.RawJson, nil
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		res, err := controller.jobApp.GenerateDslFromDag(rawJson)
		if err != nil {
			resp := &GeneralResponse{
				Code:    constants.RespInternalErr,
				Message: err.Error(),
			}
			c.JSON(http.StatusInternalServerError, resp)
		} else {
			resp := &GeneralResponse{
				Code: constants.RespNoErr,
				Data: res,
			}
			c.JSON(http.StatusOK, resp)
		}
	}
}

// generateConfFromDag returns the conf json file from the DAG the user draw
//	@Summary	Generate the conf json file from the DAG the user draw, the conf file can be consumed by Fateflow
//	@Tags		Job
//	@Produce	json
//	@Param		generateJobConfRequest	body		service.GenerateJobConfRequest	true	"The request for generate the conf json file"
//	@Success	200						{object}	GeneralResponse{data=string}	"Success"
//	@Failure	401						{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500						{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/job/generateConfFromDag [post]
func (controller *JobController) generateConfFromDag(c *gin.Context) {
	if generateJobConfRequest, err := func() (*service.GenerateJobConfRequest, error) {
		request := &service.GenerateJobConfRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return nil, err
		}
		return request, nil
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		claims := jwt.ExtractClaims(c)
		username := claims[nameKey].(string)
		res, err := controller.jobApp.GenerateConfFromDag(username, generateJobConfRequest)
		if err != nil {
			resp := &GeneralResponse{
				Code:    constants.RespInternalErr,
				Message: err.Error(),
			}
			c.JSON(http.StatusInternalServerError, resp)
		} else {
			resp := &GeneralResponse{
				Code: constants.RespNoErr,
				Data: res,
			}
			c.JSON(http.StatusOK, resp)
		}
	}
}
