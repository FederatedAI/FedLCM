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
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
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
	if viper.GetBool("fmlmanager.tls.enabled") {
		site.Use(certAuthenticator())
	}
	{
		site.GET("", controller.getSite)
		site.POST("", controller.postSite)
		site.DELETE(":uuid", controller.deleteSite)
	}
}

// getSite returns the sites list
//	@Summary	Return sites list
//	@Tags		Site
//	@Produce	json
//	@Success	200	{object}	GeneralResponse{data=[]entity.Site}	"Success"
//	@Failure	500	{object}	GeneralResponse{code=int}			"Internal server error"
//	@Router		/site [get]
func (controller *SiteController) getSite(c *gin.Context) {
	siteList, err := controller.siteAppService.GetSiteList()
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
			Data:    siteList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// postSite creates or updates site information
//	@Summary	Create or update site info
//	@Tags		Site
//	@Produce	json
//	@Param		site	body		entity.Site					true	"The site information"
//	@Success	200		{object}	GeneralResponse				"Success"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/site [post]
func (controller *SiteController) postSite(c *gin.Context) {
	if err := func() error {
		updatedSiteInfo := &entity.Site{}
		if err := c.ShouldBindJSON(updatedSiteInfo); err != nil {
			return err
		}
		return controller.siteAppService.RegisterSite(updatedSiteInfo)
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

// deleteSite removes a site
//	@Summary	Remove a site, all related projects will be impacted
//	@Tags		Site
//	@Produce	json
//	@Param		uuid	path		string						true	"The site UUID"
//	@Success	200		{object}	GeneralResponse				"Success"
//	@Failure	500		{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/site/{uuid} [delete]
func (controller *SiteController) deleteSite(c *gin.Context) {
	if err := func() error {
		siteUUID := c.Param("uuid")
		return controller.siteAppService.UnregisterSite(siteUUID)
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
