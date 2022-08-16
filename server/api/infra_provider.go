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

	"github.com/FederatedAI/FedLCM/server/application/service"
	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/gin-gonic/gin"
)

// InfraProviderController provides API handlers for the infra provider related APIs
type InfraProviderController struct {
	infraProviderAppService *service.InfraProviderApp
}

// NewInfraProviderController returns a controller instance to handle infra provider API requests
func NewInfraProviderController(infraProviderKubernetesRepo repo.InfraProviderRepository,
	endpointKubeFATERepo repo.EndpointRepository) *InfraProviderController {
	return &InfraProviderController{
		infraProviderAppService: &service.InfraProviderApp{
			InfraProviderKubernetesRepo: infraProviderKubernetesRepo,
			EndpointKubeFATERepo:        endpointKubeFATERepo,
		},
	}
}

// Route sets up route mappings to infra provider related APIs
func (controller *InfraProviderController) Route(r *gin.RouterGroup) {
	infraProvider := r.Group("infra")
	infraProvider.Use(authMiddleware.MiddlewareFunc())
	{
		infraProvider.GET("", controller.list)
		infraProvider.POST("", controller.create)
		infraProvider.POST("/kubernetes/connect", controller.testKubernetes)

		infraProvider.GET("/:uuid", controller.get)
		infraProvider.DELETE("/:uuid", controller.delete)
		infraProvider.PUT("/:uuid", controller.update)
	}
}

// list returns the provider list
// @Summary  Return provider list data
// @Tags     InfraProvider
// @Produce  json
// @Success  200  {object}  GeneralResponse{data=[]service.InfraProviderListItem}  "Success"
// @Failure  401  {object}  GeneralResponse                                        "Unauthorized operation"
// @Failure  500  {object}  GeneralResponse{code=int}                              "Internal server error"
// @Router   /infra [get]
func (controller *InfraProviderController) list(c *gin.Context) {
	providerList, err := controller.infraProviderAppService.GetProviderList()
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
			Data:    providerList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// create a new provider
// @Summary  Create a new infra provider
// @Tags     InfraProvider
// @Produce  json
// @Param    provider  body      service.InfraProviderCreationRequest  true  "The provider information, currently for the type field only 'Kubernetes' is supported"
// @Success  200       {object}  GeneralResponse                       "Success"
// @Failure  401       {object}  GeneralResponse                       "Unauthorized operation"
// @Failure  500       {object}  GeneralResponse{code=int}             "Internal server error"
// @Router   /infra [post]
func (controller *InfraProviderController) create(c *gin.Context) {
	if err := func() error {
		providerInfo := &service.InfraProviderCreationRequest{}
		if err := c.ShouldBindJSON(providerInfo); err != nil {
			return err
		}
		return controller.infraProviderAppService.CreateProvider(providerInfo)
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

// testKubernetes test connection to Kubernetes infra provider
// @Summary  Test connection to a Kubernetes infra provider
// @Tags     InfraProvider
// @Produce  json
// @Param    permission  body      valueobject.KubeConfig     true  "The kubeconfig content"
// @Success  200         {object}  GeneralResponse            "Success"
// @Failure  401         {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500         {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /infra/kubernetes/connect [post]
func (controller *InfraProviderController) testKubernetes(c *gin.Context) {
	if err := func() error {
		kubeconfig := &valueobject.KubeConfig{}
		if err := c.ShouldBindJSON(kubeconfig); err != nil {
			return err
		}
		return controller.infraProviderAppService.TestKubernetesConnection(kubeconfig)
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

// get returns detailed information of an infra provider
// @Summary  Get infra provider's detailed info
// @Tags     InfraProvider
// @Produce  json
// @Param    uuid  path      string                                             true  "Provider UUID"
// @Success  200   {object}  GeneralResponse{data=service.InfraProviderDetail}  "Success"
// @Failure  401   {object}  GeneralResponse                                    "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                          "Internal server error"
// @Router   /infra/{uuid} [get]
func (controller *InfraProviderController) get(c *gin.Context) {
	uuid := c.Param("uuid")
	if providerDetail, err := controller.infraProviderAppService.GetProviderDetail(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: providerDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// delete the infra provider
// @Summary  Delete the infra provider
// @Tags     InfraProvider
// @Produce  json
// @Param    uuid  path      string                     true  "Provider UUID"
// @Success  200   {object}  GeneralResponse            "Success"
// @Failure  401   {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /infra/{uuid} [delete]
func (controller *InfraProviderController) delete(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.infraProviderAppService.DeleteProvider(uuid); err != nil {
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

// update the provider configuration
// @Summary  Updates the infra provider
// @Tags     InfraProvider
// @Produce  json
// @Param    uuid      path      string                              true  "Provider UUID"
// @Param    provider  body      service.InfraProviderUpdateRequest  true  "The updated provider information"
// @Success  200       {object}  GeneralResponse                     "Success"
// @Failure  401       {object}  GeneralResponse                     "Unauthorized operation"
// @Failure  500       {object}  GeneralResponse{code=int}           "Internal server error"
// @Router   /infra/{uuid} [put]
func (controller *InfraProviderController) update(c *gin.Context) {
	if err := func() error {
		uuid := c.Param("uuid")
		providerInfo := &service.InfraProviderUpdateRequest{}
		if err := c.ShouldBindJSON(providerInfo); err != nil {
			return err
		}
		return controller.infraProviderAppService.UpdateProvider(uuid, providerInfo)
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
