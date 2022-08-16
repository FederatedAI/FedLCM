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

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	domainService "github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// ModelController handles model related APIs
type ModelController struct {
	modelApp *service.ModelApp
}

// NewModelController returns a controller instance to handle model API requests
func NewModelController(modelRepo repo.ModelRepository,
	modelDeploymentRepo repo.ModelDeploymentRepository,
	siteRepo repo.SiteRepository,
	projectRepo repo.ProjectRepository) *ModelController {
	return &ModelController{
		modelApp: &service.ModelApp{
			ModelRepo:           modelRepo,
			ModelDeploymentRepo: modelDeploymentRepo,
			SiteRepo:            siteRepo,
			ProjectRepo:         projectRepo,
		},
	}
}

// Route set up route mappings to model related APIs
func (controller *ModelController) Route(r *gin.RouterGroup) {
	model := r.Group("model")
	internal := model.Group("internal")
	if viper.GetBool("siteportal.tls.enabled") {
		internal.Use(certAuthenticator())
	}
	{
		internal.POST("/event/create", controller.create)
	}
	model.Use(authMiddleware.MiddlewareFunc())
	{
		model.GET("", controller.list)
		model.GET("/:uuid", controller.get)
		model.POST("/:uuid/publish", controller.deployModel)
		model.GET("/:uuid/supportedDeploymentTypes", controller.getSupportedDeployments)
		model.DELETE("/:uuid", controller.delete)
	}
}

// list returns a list of models
// @Summary Get model list
// @Tags Model
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.ModelListItem} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model [get]
func (controller *ModelController) list(c *gin.Context) {
	if data, err := func() ([]service.ModelListItem, error) {
		return controller.modelApp.List("")
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

// get returns detailed information of a model
// @Summary Get model's detailed info
// @Tags Model
// @Produce json
// @Param uuid path string true "Model UUID"
// @Success 200 {object} GeneralResponse{data=service.ModelDetail} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model/{uuid} [get]
func (controller *ModelController) get(c *gin.Context) {
	if data, err := func() (*service.ModelDetail, error) {
		modelUUID := c.Param("uuid")
		return controller.modelApp.Get(modelUUID)
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

// delete deletes the specified model
// @Summary Delete the model
// @Tags Model
// @Produce json
// @Param uuid path string true "Model UUID"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model/{uuid} [delete]
func (controller *ModelController) delete(c *gin.Context) {
	modelUUID := c.Param("uuid")
	if err := controller.modelApp.Delete(modelUUID); err != nil {
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

// create process model creation event
// @Summary Handle model creation event, called by the job context only
// @Tags Model
// @Produce json
// @Param request body service.ModelCreationRequest true "Creation Request"
// @Success 200 {object} GeneralResponse{} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model/internal/event/create [post]
func (controller *ModelController) create(c *gin.Context) {
	if err := func() error {
		request := &service.ModelCreationRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return err
		}
		return controller.modelApp.Create(request)
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

// deployModel publish the model to online serving system
// @Summary Publish model to online serving system
// @Tags Model
// @Produce json
// @Param uuid path string true "Model UUID"
// @Param request body service.ModelDeploymentRequest true "Creation Request"
// @Success 200 {object} GeneralResponse{data=entity.ModelDeployment} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model/{uuid}/publish [post]
func (controller *ModelController) deployModel(c *gin.Context) {
	if deployment, err := func() (*entity.ModelDeployment, error) {
		modelUUID := c.Param("uuid")
		request := &domainService.ModelDeploymentRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return nil, err
		}
		request.ModelUUID = modelUUID
		return controller.modelApp.Publish(request)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: deployment,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getSupportedDeployments returns list of the deployment type this model can use
// @Summary Get list of deployment types (KFServing, FATE-Serving, etc.) this model can use
// @Tags Model
// @Produce json
// @Param uuid path string true "Model UUID"
// @Success 200 {object} GeneralResponse{data=[]entity.ModelDeploymentType} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /model/{uuid}/supportedDeploymentTypes [get]
func (controller *ModelController) getSupportedDeployments(c *gin.Context) {
	if types, err := func() ([]entity.ModelDeploymentType, error) {
		modelUUID := c.Param("uuid")
		return controller.modelApp.GetSupportedDeploymentTypes(modelUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: types,
		}
		c.JSON(http.StatusOK, resp)
	}
}
