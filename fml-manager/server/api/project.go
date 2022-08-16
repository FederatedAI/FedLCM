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
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/event"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ProjectController handles project related APIs
type ProjectController struct {
	projectApp *service.ProjectApp
}

// NewProjectController returns a controller instance to handle project API requests
func NewProjectController(projectRepo repo.ProjectRepository, siteRepo repo.SiteRepository,
	participantRepo repo.ProjectParticipantRepository,
	invitationRepo repo.ProjectInvitationRepository,
	projectDataRepo repo.ProjectDataRepository) *ProjectController {
	return &ProjectController{
		projectApp: &service.ProjectApp{
			ProjectRepo:     projectRepo,
			SiteRepo:        siteRepo,
			ParticipantRepo: participantRepo,
			InvitationRepo:  invitationRepo,
			ProjectDataRepo: projectDataRepo,
		},
	}
}

// Route set up route mappings to project related APIs
func (controller *ProjectController) Route(r *gin.RouterGroup) {
	project := r.Group("project")
	if viper.GetBool("fmlmanager.tls.enabled") {
		project.Use(certAuthenticator())
	}
	{
		project.POST("/invitation", controller.handleInvitation)
		project.POST("/invitation/:uuid/accept", controller.handleInvitationAcceptance)
		project.POST("/invitation/:uuid/reject", controller.handleInvitationRejection)
		project.POST("/invitation/:uuid/revoke", controller.handleInvitationRevocation)

		project.POST("/:uuid/participant/:siteUUID/leave", controller.handleParticipantLeaving)
		project.POST("/:uuid/participant/:siteUUID/dismiss", controller.handleParticipantDismissal)

		project.POST("/event/participant/update", controller.handleParticipantInfoUpdate)

		project.POST("/:uuid/data/associate", controller.handleDataAssociation)
		project.POST("/:uuid/data/dismiss", controller.handleDataDismissal)

		project.POST("/:uuid/close", controller.handleProjectClosing)

		project.GET("", controller.list)
		project.GET("/:uuid/participant", controller.listParticipant)
		project.GET("/:uuid/data", controller.listData)
	}
}

// handleInvitation process a project invitation
// @Summary Process project invitation
// @Tags Project
// @Produce json
// @Param project body service.ProjectInvitationRequest true "invitation request"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/invitation [post]
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

// handleInvitationAcceptance process a project invitation acceptance
// @Summary Process invitation acceptance response
// @Tags Project
// @Produce json
// @Param uuid path string true "Invitation UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/invitation/{uuid}/accept [post]
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
// @Summary Process invitation rejection response
// @Tags Project
// @Produce json
// @Param uuid path string true "Invitation UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/invitation/{uuid}/reject [post]
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
// @Summary Process invitation revocation request
// @Tags Project
// @Produce json
// @Param uuid path string true "Invitation UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/invitation/{uuid}/revoke [post]
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

// handleParticipantInfoUpdate process a participant info update event
// @Summary Process participant info update event, called by this FML manager's site context only
// @Tags Project
// @Produce json
// @Param project body event.ProjectParticipantUpdateEvent true "Updated participant info"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/event/participant/update [post]
func (controller *ProjectController) handleParticipantInfoUpdate(c *gin.Context) {
	if err := func() error {
		updateEvent := &event.ProjectParticipantUpdateEvent{}
		if err := c.ShouldBindJSON(updateEvent); err != nil {
			return err
		}
		return controller.projectApp.ProcessParticipantInfoUpdate(updateEvent.UUID)
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

// handleDataAssociation process a new data association
// @Summary Process new data association from site
// @Tags Project
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param project body service.ProjectDataAssociation true "Data association info"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/data/associate [post]
func (controller *ProjectController) handleDataAssociation(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		association := &service.ProjectDataAssociation{}
		if err := c.ShouldBindJSON(association); err != nil {
			return err
		}
		return controller.projectApp.ProcessDataAssociation(projectUUID, association)
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

// handleDataDismissal process project data dismissal
// @Summary Process data dismissal from site
// @Tags Project
// @Produce json
// @Param uuid path string true "Project UUID"
// @Param project body service.ProjectDataAssociationBase true "Data association info containing the data UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/data/dismiss [post]
func (controller *ProjectController) handleDataDismissal(c *gin.Context) {
	if err := func() error {
		projectUUID := c.Param("uuid")
		association := &service.ProjectDataAssociationBase{}
		if err := c.ShouldBindJSON(association); err != nil {
			return err
		}
		return controller.projectApp.ProcessDataDismissal(projectUUID, association)
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

// handleParticipantLeaving process project participant leaving
// @Summary Process participant leaving
// @Tags Project
// @Produce json
// @Param uuid     path string true "Project UUID"
// @Param siteUUID path string true "Site UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/participant/{siteUUID}/leave [post]
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

// handleParticipantDismissal process project participant dismissal
// @Summary Process participant dismissal, called by the managing site only
// @Tags Project
// @Produce json
// @Param uuid     path string true "Project UUID"
// @Param siteUUID path string true "Site UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/participant/{siteUUID}/dismiss [post]
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

// handleProjectClosing process project closing
// @Summary Process project closing
// @Tags Project
// @Produce json
// @Param uuid     path string true "Project UUID"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/close [post]
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

// list returns all projects or project related to the specified participant
// @Summary List all project
// @Tags Project
// @Produce json
// @Param participant query string false "participant uuid, if set, only returns the projects containing the participant"
// @Success 200 {object} GeneralResponse{data=map[string]service.ProjectInfoWithStatus} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project [get]
func (controller *ProjectController) list(c *gin.Context) {
	// TODO: use token to extract participant uuid and do authz check
	participantUUID := c.DefaultQuery("participant", "")
	projectList, err := controller.projectApp.ListProjectByParticipant(participantUUID)
	if err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
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

// listData returns all data association in a project
// @Summary List all data association in a project
// @Tags Project
// @Produce json
// @Param uuid path string true "Project UUID"
// @Success 200 {object} GeneralResponse{data=map[string]service.ProjectDataAssociation} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/data [get]
func (controller *ProjectController) listData(c *gin.Context) {
	// TODO: use token to verify the requester can access these info
	projectUUID := c.Param("uuid")
	dataList, err := controller.projectApp.ListDataAssociationByProject(projectUUID)
	if err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: dataList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// listParticipant returns all participants info in a project
// @Summary List all participants in a project
// @Tags Project
// @Produce json
// @Param uuid path string true "Project UUID"
// @Success 200 {object} GeneralResponse{data=map[string]service.ProjectDataAssociation} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /project/{uuid}/participant [get]
func (controller *ProjectController) listParticipant(c *gin.Context) {
	// TODO: use token to verify the requester can access these info
	projectUUID := c.Param("uuid")
	participantList, err := controller.projectApp.ListParticipantByProject(projectUUID)
	if err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
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
