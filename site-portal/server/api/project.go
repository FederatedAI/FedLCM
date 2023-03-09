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
	"strconv"

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	domainService "github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ProjectController handles project related APIs
type ProjectController struct {
	projectApp *service.ProjectApp
	jobApp     *service.JobApp
	modelApp   *service.ModelApp
}

// NewProjectController returns a controller instance to handle project API requests
func NewProjectController(projectRepo repo.ProjectRepository,
	siteRepo repo.SiteRepository,
	participantRepo repo.ProjectParticipantRepository,
	invitationRepo repo.ProjectInvitationRepository,
	projectDataRepo repo.ProjectDataRepository,
	localDataRepo repo.LocalDataRepository,
	jobRepo repo.JobRepository,
	jobParticipantRepo repo.JobParticipantRepository,
	modelRepo repo.ModelRepository) *ProjectController {

	jobApp := &service.JobApp{
		SiteRepo:        siteRepo,
		JobRepo:         jobRepo,
		ParticipantRepo: jobParticipantRepo,
		ProjectRepo:     projectRepo,
		ProjectDataRepo: projectDataRepo,
		ModelRepo:       modelRepo,
	}
	return &ProjectController{
		projectApp: &service.ProjectApp{
			ProjectRepo:        projectRepo,
			ParticipantRepo:    participantRepo,
			SiteRepo:           siteRepo,
			InvitationRepo:     invitationRepo,
			ProjectDataRepo:    projectDataRepo,
			LocalDataRepo:      localDataRepo,
			JobApp:             jobApp,
			ProjectSyncService: domainService.NewProjectSyncService(),
		},
		jobApp: jobApp,
		modelApp: &service.ModelApp{
			ModelRepo:   modelRepo,
			ProjectRepo: projectRepo,
		},
	}
}

// Route set up route mappings to project related APIs
func (controller *ProjectController) Route(r *gin.RouterGroup) {
	project := r.Group("project")
	internal := project.Group("internal")
	if viper.GetBool("siteportal.tls.enabled") {
		internal.Use(certAuthenticator())
	}
	{
		internal.POST("/:uuid/close", controller.handleProjectClosing)

		internal.POST("/invitation", controller.handleInvitation)
		internal.POST("/invitation/:uuid/accept", controller.handleInvitationAcceptance)
		internal.POST("/invitation/:uuid/reject", controller.handleInvitationRejection)
		internal.POST("/invitation/:uuid/revoke", controller.handleInvitationRevocation)

		internal.POST("/:uuid/participants", controller.createParticipants)
		internal.POST("/:uuid/participant/:siteUUID/dismiss", controller.handleParticipantDismissal)
		internal.POST("/:uuid/participant/:siteUUID/leave", controller.handleParticipantLeaving)
		internal.POST("/all/participant/:siteUUID/unregister", controller.handleParticipantUnregistration)

		internal.POST("/event/participant/update", controller.handleParticipantInfoUpdate)
		internal.POST("/event/participant/sync", controller.handleParticipantSync)
		internal.POST("/event/data/sync", controller.handleDataSync)
		internal.POST("/event/list/sync", controller.handleProjectListSync)
		internal.POST("/event/participant/unregister", controller.handleParticipantUnregistration)

		internal.POST("/:uuid/data/associate", controller.handleRemoteDataAssociation)
		internal.POST("/:uuid/data/dismiss", controller.handleRemoteDataDismissal)
	}
	project.Use(authMiddleware.MiddlewareFunc())
	{
		project.POST("", controller.create)
		project.GET("", controller.list)
		project.GET("/:uuid", controller.get)
		project.GET("/:uuid/participant", controller.listParticipants)
		project.DELETE("/:uuid/participant/:participantUUID", controller.removeParticipant)

		project.POST("/:uuid/invitation", controller.inviteParticipant)
		project.PUT("/:uuid/autoapprovalstatus", controller.toggleAutoApproval)
		project.POST("/:uuid/join", controller.joinProject)
		project.POST("/:uuid/reject", controller.rejectProject)
		project.POST("/:uuid/leave", controller.leaveProject)
		project.POST("/:uuid/close", controller.closeProject)

		project.POST("/:uuid/data", controller.addData)
		project.GET("/:uuid/data", controller.listData)
		project.GET("/:uuid/data/local", controller.listLocalAvailableData)
		project.DELETE("/:uuid/data/:dataUUID", controller.removeData)

		project.GET("/:uuid/job", controller.listJob)
		project.POST("/:uuid/job", controller.submitJob)

		project.GET("/:uuid/model", controller.listModel)
	}
}

// create Create a new local project
//	@Summary	Create a new project
//	@Tags		Project
//	@Produce	json
//	@Param		project	body		service.ProjectCreationRequest	true	"Basic project info"
//	@Success	200		{object}	GeneralResponse{}				"Success"
//	@Failure	401		{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/project [post]
func (controller *ProjectController) create(c *gin.Context) {
	if err := func() error {
		claims := jwt.ExtractClaims(c)
		// the auth middleware makes sure username exists
		username := claims[nameKey].(string)
		request := &service.ProjectCreationRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return err
		}
		if err := controller.projectApp.CreateLocalProject(request, username); err != nil {
			return err
		}
		return nil
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

// list returns all projects
//	@Summary	List all project
//	@Tags		Project
//	@Produce	json
//	@Success	200	{object}	GeneralResponse{data=service.ProjectList}	"Success"
//	@Failure	401	{object}	GeneralResponse								"Unauthorized operation"
//	@Failure	500	{object}	GeneralResponse{code=int}					"Internal server error"
//	@Router		/project [get]
func (controller *ProjectController) list(c *gin.Context) {
	if projectList, err := controller.projectApp.List(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: projectList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// listParticipants returns all participants
//	@Summary	List participants
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string												true	"project UUID"
//	@Param		all		query		bool												false	"if set to true, returns all sites, including not joined ones"
//	@Success	200		{object}	GeneralResponse{data=[]service.ProjectParticipant}	"Success"
//	@Failure	401		{object}	GeneralResponse										"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}							"Internal server error"
//	@Router		/project/{uuid}/participant [get]
func (controller *ProjectController) listParticipants(c *gin.Context) {
	queryAll, err := strconv.ParseBool(c.DefaultQuery("all", "false"))
	if err != nil {
		queryAll = false
	}
	projectUUID := c.Param("uuid")
	if participantList, err := controller.projectApp.ListParticipant(projectUUID, queryAll); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: participantList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// inviteParticipant sends project invitation to other projects
//	@Summary	Invite other site to this project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string							true	"project UUID"
//	@Param		info	body		service.ProjectParticipantBase	true	"target site information"
//	@Success	200		{object}	GeneralResponse{}				"Success"
//	@Failure	401		{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/project/{uuid}/invitation [post]
func (controller *ProjectController) inviteParticipant(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		targetSite := &service.ProjectParticipantBase{}
		if err := c.ShouldBindJSON(targetSite); err != nil {
			return err
		}
		return controller.projectApp.InviteParticipant(projectUUID, targetSite)
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

// handleInvitation process a project invitation
//	@Summary	Process project invitation, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		invitation	body		service.ProjectInvitationRequest	true	"invitation request"
//	@Success	200			{object}	GeneralResponse{}					"Success"
//	@Failure	401			{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/project/internal/invitation [post]
func (controller *ProjectController) handleInvitation(c *gin.Context) {
	if err := func() error {
		invitationRequest := &service.ProjectInvitationRequest{}
		if err := c.ShouldBindJSON(invitationRequest); err != nil {
			return err
		}
		return controller.projectApp.ProcessInvitation(invitationRequest)
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

// toggleAutoApproval changes project auto-approval status
//	@Summary	Change a project's auto-approval status
//	@Tags		Project
//	@Produce	json
//	@Param		status	body		service.ProjectAutoApprovalStatus	true	"The auto-approval status, only an 'enabled' field"
//	@Param		uuid	path		string								true	"Project UUID"
//	@Success	200		{object}	GeneralResponse						"Success"
//	@Failure	401		{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/project/{uuid}/autoapprovalstatus [put]
func (controller *ProjectController) toggleAutoApproval(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		status := &service.ProjectAutoApprovalStatus{}
		if err := c.ShouldBindJSON(status); err != nil {
			return err
		}
		return controller.projectApp.ToggleAutoApprovalStatus(projectUUID, status)
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

// get returns detailed information of a project
//	@Summary	Get project's detailed info
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string										true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{data=service.ProjectInfo}	"Success"
//	@Failure	401		{object}	GeneralResponse								"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}					"Internal server error"
//	@Router		/project/{uuid} [get]
func (controller *ProjectController) get(c *gin.Context) {
	if project, err := func() (*service.ProjectInfo, error) {
		projectUUID := c.Param("uuid")
		return controller.projectApp.GetProject(projectUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: project,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// joinProject joins a project
//	@Summary	Join a pending/invited project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/join [post]
func (controller *ProjectController) joinProject(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		return controller.projectApp.JoinOrRejectProject(projectUUID, true)
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

// rejectProject joins a project
//	@Summary	Reject to join a pending/invited project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/reject [post]
func (controller *ProjectController) rejectProject(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		return controller.projectApp.JoinOrRejectProject(projectUUID, false)
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

// leaveProject leave the specified project
//	@Summary	Leave the joined project created by other site
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/leave [post]
func (controller *ProjectController) leaveProject(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		return controller.projectApp.LeaveProject(projectUUID)
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

// closeProject close the specified project
//	@Summary	Close the managed project, only available to project managing site
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/close [post]
func (controller *ProjectController) closeProject(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		return controller.projectApp.CloseProject(projectUUID)
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

// handleInvitationAcceptance process a project invitation acceptance
//	@Summary	Process invitation acceptance response, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Invitation UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/invitation/{uuid}/accept [post]
func (controller *ProjectController) handleInvitationAcceptance(c *gin.Context) {
	if err := func() error {
		invitationUUID := c.Param("uuid")
		return controller.projectApp.ProcessInvitationResponse(invitationUUID, true)
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

// handleInvitationRejection process a project invitation rejection
//	@Summary	Process invitation rejection response, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Invitation UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/invitation/{uuid}/reject [post]
func (controller *ProjectController) handleInvitationRejection(c *gin.Context) {
	if err := func() error {
		invitationUUID := c.Param("uuid")
		return controller.projectApp.ProcessInvitationResponse(invitationUUID, false)
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

// handleInvitationRevocation process a project invitation revocation
//	@Summary	Process invitation revocation request, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Invitation UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/invitation/{uuid}/revoke [post]
func (controller *ProjectController) handleInvitationRevocation(c *gin.Context) {
	if err := func() error {
		invitationUUID := c.Param("uuid")
		return controller.projectApp.ProcessInvitationRevocation(invitationUUID)
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

// handleParticipantLeaving process a project participant removal event
//	@Summary	Process participant leaving request, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid		path		string						true	"Project UUID"
//	@Param		siteUUID	path		string						true	"Site UUID"
//	@Success	200			{object}	GeneralResponse{}			"Success"
//	@Failure	401			{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/participant/{siteUUID}/leave [post]
func (controller *ProjectController) handleParticipantLeaving(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		siteUUID := c.Param("siteUUID")
		return controller.projectApp.ProcessParticipantLeaving(projectUUID, siteUUID)
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

// handleParticipantDismissal process a project participant dismissal event
//	@Summary	Process participant dismissal event, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid		path		string						true	"Project UUID"
//	@Param		siteUUID	path		string						true	"Site UUID"
//	@Success	200			{object}	GeneralResponse{}			"Success"
//	@Failure	401			{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/participant/{siteUUID}/dismiss [post]
func (controller *ProjectController) handleParticipantDismissal(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		siteUUID := c.Param("siteUUID")
		return controller.projectApp.ProcessParticipantDismissal(projectUUID, siteUUID)
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

// createParticipants receive participants info from FML manager
//	@Summary	Create joined participants for a project, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid			path		string						true	"Project UUID"
//	@Param		participantList	body		[]entity.ProjectParticipant	true	"participants list"
//	@Success	200				{object}	GeneralResponse{}			"Success"
//	@Failure	401				{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500				{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/participants [post]
func (controller *ProjectController) createParticipants(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		var participants []entity.ProjectParticipant
		if err := c.ShouldBindJSON(&participants); err != nil {
			return err
		}
		return controller.projectApp.CreateRemoteProjectParticipants(projectUUID, participants)
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

// removeParticipant remove pending or joined participant
//	@Summary	Remove pending participant (revoke invitation) or dismiss joined participant
//	@Tags		Project
//	@Produce	json
//	@Param		uuid			path		string						true	"Project UUID"
//	@Param		participantUUID	path		string						true	"Participant UUID"
//	@Success	200				{object}	GeneralResponse{}			"Success"
//	@Failure	401				{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500				{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/participant/{participantUUID} [delete]
func (controller *ProjectController) removeParticipant(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		siteUUID := c.Param("participantUUID")
		return controller.projectApp.RemoveProjectParticipants(projectUUID, siteUUID)
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

// handleParticipantInfoUpdate process a participant info update event
//	@Summary	Process participant info update event, called by the FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		participant	body		service.ProjectParticipantBase	true	"Updated participant info"
//	@Success	200			{object}	GeneralResponse{}				"Success"
//	@Failure	401			{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/project/internal/event/participant/update [post]
func (controller *ProjectController) handleParticipantInfoUpdate(c *gin.Context) {
	if err := func() error {
		var participant service.ProjectParticipantBase
		if err := c.ShouldBindJSON(&participant); err != nil {
			return err
		}
		return controller.projectApp.ProcessParticipantInfoUpdate(&participant)
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

// listLocalAvailableData returns a list of local data that can be associated to current project
//	@Summary	Get available local data for this project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string										true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{data=[]service.ProjectData}	"Success"
//	@Failure	401		{object}	GeneralResponse								"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}					"Internal server error"
//	@Router		/project/{uuid}/data/local [get]
func (controller *ProjectController) listLocalAvailableData(c *gin.Context) {
	if data, err := func() ([]service.ProjectData, error) {
		projectUUID := c.Param("uuid")
		return controller.projectApp.ListLocalData(projectUUID)
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

// listData returns a list of associated data of current project
//	@Summary	Get associated data list for this project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid		path		string										true	"Project UUID"
//	@Param		participant	query		string										false	"participant uuid, i.e. the providing site uuid; if set, only returns the associated data of the specified participant"
//	@Success	200			{object}	GeneralResponse{data=[]service.ProjectData}	"Success"
//	@Failure	401			{object}	GeneralResponse								"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}					"Internal server error"
//	@Router		/project/{uuid}/data [get]
func (controller *ProjectController) listData(c *gin.Context) {
	if data, err := func() ([]service.ProjectData, error) {
		projectUUID := c.Param("uuid")
		participantUUID := c.DefaultQuery("participant", "")
		return controller.projectApp.ListData(projectUUID, participantUUID)
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

// addData associate local data
//	@Summary	Associate local data to current project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string									true	"project UUID"
//	@Param		request	body		service.ProjectDataAssociationRequest	true	"Local data info"
//	@Success	200		{object}	GeneralResponse{}						"Success"
//	@Failure	401		{object}	GeneralResponse							"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}				"Internal server error"
//	@Router		/project/{uuid}/data [post]
func (controller *ProjectController) addData(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		request := &service.ProjectDataAssociationRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return err
		}
		if err := controller.projectApp.CreateDataAssociation(projectUUID, request); err != nil {
			return err
		}
		return nil
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

// removeData dismiss data association
//	@Summary	Remove associated data from the current project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid		path		string						true	"Project UUID"
//	@Param		dataUUID	path		string						true	"Data UUID"
//	@Success	200			{object}	GeneralResponse{}			"Success"
//	@Failure	401			{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/{uuid}/data/{dataUUID} [delete]
func (controller *ProjectController) removeData(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		dataUUID := c.Param("dataUUID")
		if err := controller.projectApp.RemoveDataAssociation(projectUUID, dataUUID); err != nil {
			return err
		}
		return nil
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

// handleRemoteDataAssociation associate remote data
//	@Summary	Add associated remote data to current project, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"project UUID"
//	@Param		data	body		[]entity.ProjectData		true	"Remote data information"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/data/associate [post]
func (controller *ProjectController) handleRemoteDataAssociation(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		var request []entity.ProjectData
		if err := c.ShouldBindJSON(&request); err != nil {
			return err
		}
		if err := controller.projectApp.CreateRemoteProjectDataAssociation(projectUUID, request); err != nil {
			return err
		}
		return nil
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

// handleRemoteDataDismissal dismiss remote data
//	@Summary	Dismiss associated remote data from current project, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"project UUID"
//	@Param		data	body		[]string					true	"Data UUID list"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/data/dismiss [post]
func (controller *ProjectController) handleRemoteDataDismissal(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		var request []string
		if err := c.ShouldBindJSON(&request); err != nil {
			return err
		}
		if err := controller.projectApp.DismissRemoteProjectDataAssociation(projectUUID, request); err != nil {
			return err
		}
		return nil
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

// listJob returns a list of jobs in the current project
//	@Summary	Get job list for this project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string											true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{data=[]service.JobListItemBase}	"Success"
//	@Failure	401		{object}	GeneralResponse									"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}						"Internal server error"
//	@Router		/project/{uuid}/job [get]
func (controller *ProjectController) listJob(c *gin.Context) {
	if data, err := func() ([]service.JobListItemBase, error) {
		projectUUID := c.Param("uuid")
		return controller.jobApp.List(projectUUID)
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

// submitJob create new job
//	@Summary	Create a new job
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string											true	"project UUID"
//	@Param		request	body		service.JobSubmissionRequest					true	"Job requests, only fill related field according to job type"
//	@Success	200		{object}	GeneralResponse{data=service.JobListItemBase}	"Success"
//	@Failure	401		{object}	GeneralResponse									"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}						"Internal server error"
//	@Router		/project/{uuid}/job [post]
func (controller *ProjectController) submitJob(c *gin.Context) {
	if job, err := func() (*service.JobListItemBase, error) {
		projectUUID := c.Param("uuid")
		claims := jwt.ExtractClaims(c)
		// the auth middleware makes sure username exists
		username := claims[nameKey].(string)
		request := &service.JobSubmissionRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return nil, err
		}
		//log.Info().Msg(fmt.Sprint(request))

		// project status check
		if err := controller.projectApp.EnsureProjectIsOpen(projectUUID); err != nil {
			return nil, err
		}

		request.ProjectUUID = projectUUID
		job, err := controller.jobApp.SubmitJob(username, request)
		if err != nil {
			return nil, err
		}
		return job, err
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

// listModel returns a list of models in the current project
//	@Summary	Get model list for this project
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string											true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{data=[]service.ModelListItem}	"Success"
//	@Failure	401		{object}	GeneralResponse									"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}						"Internal server error"
//	@Router		/project/{uuid}/model [get]
func (controller *ProjectController) listModel(c *gin.Context) {
	if data, err := func() ([]service.ModelListItem, error) {
		projectUUID := c.Param("uuid")
		return controller.modelApp.List(projectUUID)
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

// handleProjectClosing process a project closing event
//	@Summary	Process project closing event, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		uuid	path		string						true	"Project UUID"
//	@Success	200		{object}	GeneralResponse{}			"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/{uuid}/close [post]
func (controller *ProjectController) handleProjectClosing(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		return controller.projectApp.ProcessProjectClosing(projectUUID)
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

// handleParticipantSync process a participant info sync event
//	@Summary	Process participant info sync event, to sync the participant info from fml manager
//	@Tags		Project
//	@Produce	json
//	@Param		request	body		service.ProjectResourceSyncRequest	true	"Info of the project to by synced"
//	@Success	200		{object}	GeneralResponse{}					"Success"
//	@Failure	401		{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/project/internal/event/participant/sync [post]
func (controller *ProjectController) handleParticipantSync(c *gin.Context) {
	if err := func() error {
		var event service.ProjectResourceSyncRequest
		if err := c.ShouldBindJSON(&event); err != nil {
			return err
		}
		return controller.projectApp.SyncProjectParticipant(event.ProjectUUID)
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

// handleDataSync process a data association sync event
//	@Summary	Process data sync event, to sync the data association info from fml manager
//	@Tags		Project
//	@Produce	json
//	@Param		request	body		service.ProjectResourceSyncRequest	true	"Info of the project to by synced"
//	@Success	200		{object}	GeneralResponse{}					"Success"
//	@Failure	401		{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/project/internal/event/data/sync [post]
func (controller *ProjectController) handleDataSync(c *gin.Context) {
	if err := func() error {
		var event service.ProjectResourceSyncRequest
		if err := c.ShouldBindJSON(&event); err != nil {
			return err
		}
		return controller.projectApp.SyncProjectData(event.ProjectUUID)
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

// handleProjectListSync process a project list sync event
//	@Summary	Process list sync event, to sync the projects list status from fml manager
//	@Tags		Project
//	@Produce	json
//	@Success	200	{object}	GeneralResponse{}			"Success"
//	@Failure	401	{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500	{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/event/list/sync [post]
func (controller *ProjectController) handleProjectListSync(c *gin.Context) {
	if err := controller.projectApp.SyncProject(); err != nil {
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

// handleParticipantUnregistration process a project participant unregistration event
//	@Summary	Process participant unregistration event, called by FML manager only
//	@Tags		Project
//	@Produce	json
//	@Param		siteUUID	path		string						true	"Participant Site UUID"
//	@Success	200			{object}	GeneralResponse{}			"Success"
//	@Failure	401			{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/project/internal/all/participant/{siteUUID}/unregister [post]
func (controller *ProjectController) handleParticipantUnregistration(c *gin.Context) {
	if err := func() error {
		siteUUID := c.Param("siteUUID")
		return controller.projectApp.ProcessParticipantUnregistration(siteUUID)
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
