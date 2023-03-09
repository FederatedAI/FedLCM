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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/gin-gonic/gin"
)

// SiteController provides API handlers for the site related APIs
type SiteController struct {
	siteAppService *service.SiteApp
}

// NewSiteController returns a controller instance to handle site API requests
func NewSiteController(repo repo.SiteRepository) *SiteController {
	return &SiteController{
		siteAppService: &service.SiteApp{
			SiteRepo: repo,
		},
	}
}

// Route set up route mappings to site related APIs
func (controller *SiteController) Route(r *gin.RouterGroup) {
	site := r.Group("site")
	site.Use(authMiddleware.MiddlewareFunc())
	{
		site.GET("", controller.getSite)
		site.PUT("", controller.putSite)
		site.POST("/fateflow/connect", controller.connectFATEFlow)
		site.POST("/kubeflow/connect", controller.connectKubeflow)
		site.POST("/fmlmanager/connect", controller.connectFMLManager)
		site.POST("/fmlmanager/unregister", controller.unregisterSite)
	}
}

// getSite returns the site data
//	@Summary	Return site data
//	@Tags		Site
//	@Produce	json
//	@Success	200	{object}	GeneralResponse{data=entity.Site}	"Success"
//	@Failure	401	{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500	{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/site [get]
func (controller *SiteController) getSite(c *gin.Context) {
	site, err := controller.siteAppService.GetSite()
	if err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code:    constants.RespNoErr,
			Message: "",
			Data:    site,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// putSite update site related information
//	@Summary	Update site information
//	@Tags		Site
//	@Produce	json
//	@Param		site	body		entity.Site					true	"The site information, some info like id, UUID, connected status cannot be changed"
//	@Success	200		{object}	GeneralResponse				"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/site [put]
func (controller *SiteController) putSite(c *gin.Context) {
	if err := func() error {
		updatedSiteInfo := &entity.Site{}
		if err := c.ShouldBindJSON(updatedSiteInfo); err != nil {
			return err
		}
		return controller.siteAppService.UpdateSite(updatedSiteInfo)
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

// connectFATEFlow test connection to fate flow
//	@Summary	Test site connection to fate flow service
//	@Tags		Site
//	@Produce	json
//	@Param		connectInfo	body		service.FATEFlowConnectionInfo	true	"The fate flow connection info"
//	@Success	200			{object}	GeneralResponse					"Success"
//	@Failure	401			{object}	GeneralResponse					"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}		"Internal server error"
//	@Router		/site/fateflow/connect [post]
func (controller *SiteController) connectFATEFlow(c *gin.Context) {
	if err := func() error {
		connectionInfo := &service.FATEFlowConnectionInfo{}
		if err := c.ShouldBindJSON(connectionInfo); err != nil {
			return err
		}
		return controller.siteAppService.TestFATEFlowConnection(connectionInfo)
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

// connectKubeflow test connection to Kubeflow
//	@Summary	Test site connection to Kubeflow, including MinIO and KFServing
//	@Tags		Site
//	@Produce	json
//	@Param		config	body		valueobject.KubeflowConfig	true	"The Kubeflow config info"
//	@Success	200		{object}	GeneralResponse				"Success"
//	@Failure	401		{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/site/kubeflow/connect [post]
func (controller *SiteController) connectKubeflow(c *gin.Context) {
	if err := func() error {
		connectionInfo := &valueobject.KubeflowConfig{}
		if err := c.ShouldBindJSON(connectionInfo); err != nil {
			return err
		}
		return controller.siteAppService.TestKubeflowConnection(connectionInfo)
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

// connectFMLManager registers the current site to fml manager
//	@Summary	Connect to fml manager and register itself
//	@Tags		Site
//	@Produce	json
//	@Param		connectInfo	body		service.FMLManagerConnectionInfo	true	"The FML Manager endpoint"
//	@Success	200			{object}	GeneralResponse						"Success"
//	@Failure	401			{object}	GeneralResponse						"Unauthorized operation"
//	@Failure	500			{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/site/fmlmanager/connect [post]
func (controller *SiteController) connectFMLManager(c *gin.Context) {
	if err := func() error {
		connectionInfo := &service.FMLManagerConnectionInfo{}
		if err := c.ShouldBindJSON(connectionInfo); err != nil {
			return err
		}
		return controller.siteAppService.RegisterToFMLManager(connectionInfo)
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

// unregisterSite unregisters the current site from fml manager
//	@Summary	Unregister from the fml manager
//	@Tags		Site
//	@Produce	json
//	@Success	200	{object}	GeneralResponse				"Success"
//	@Failure	401	{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500	{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/site/fmlmanager/unregister [post]
func (controller *SiteController) unregisterSite(c *gin.Context) {
	if err := func() error {
		return controller.siteAppService.UnregisterFromFMLManager()
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
