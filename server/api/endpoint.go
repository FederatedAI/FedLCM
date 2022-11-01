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

	"github.com/FederatedAI/FedLCM/server/application/service"
	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// EndpointController provides API handlers for the endpoint related APIs
type EndpointController struct {
	endpointAppService *service.EndpointApp
}

// NewEndpointController returns a controller instance to handle endpoint API requests
func NewEndpointController(infraProviderKubernetesRepo repo.InfraProviderRepository,
	endpointKubeFATERepo repo.EndpointRepository,
	participantFATERepo repo.ParticipantFATERepository,
	participantOpenFLRepo repo.ParticipantOpenFLRepository,
	eventRepo repo.EventRepository) *EndpointController {
	return &EndpointController{
		endpointAppService: &service.EndpointApp{
			InfraProviderKubernetesRepo: infraProviderKubernetesRepo,
			EndpointKubeFAETRepo:        endpointKubeFATERepo,
			ParticipantFATERepo:         participantFATERepo,
			ParticipantOpenFLRepo:       participantOpenFLRepo,
			EventRepo:                   eventRepo,
		},
	}
}

// Route sets up route mappings to endpoint related APIs
func (controller *EndpointController) Route(r *gin.RouterGroup) {
	endpoint := r.Group("endpoint")
	endpoint.Use(authMiddleware.MiddlewareFunc())
	{
		endpoint.GET("", controller.list)
		endpoint.GET("/:uuid", controller.get)
		endpoint.GET("/kubefate/yaml", controller.getKubeFATEDeploymentYAML)

		endpoint.DELETE("/:uuid", controller.delete)

		endpoint.POST("/scan", controller.scan)
		endpoint.POST("/:uuid/kubefate/check", controller.checkKubeFATE)

		endpoint.POST("", controller.create)
	}
}

// list returns the endpoints list
// @Summary Return endpoints list data
// @Tags    Endpoint
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.EndpointListItem} "Success"
// @Failure 401 {object} GeneralResponse                                  "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int}                        "Internal server error"
// @Router  /endpoint [get]
func (controller *EndpointController) list(c *gin.Context) {
	endpointList, err := controller.endpointAppService.GetEndpointList()
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
			Data:    endpointList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// get returns detailed information of an endpoint
// @Summary Get endpoint's detailed info
// @Tags    Endpoint
// @Produce json
// @Param   uuid path     string                                       true "Endpoint UUID"
// @Success 200  {object} GeneralResponse{data=service.EndpointDetail} "Success"
// @Failure 401  {object} GeneralResponse                              "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int}                    "Internal server error"
// @Router  /endpoint/{uuid} [get]
func (controller *EndpointController) get(c *gin.Context) {
	uuid := c.Param("uuid")
	if endpointDetail, err := controller.endpointAppService.GetEndpointDetail(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: endpointDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// delete removes the endpoint
// @Summary Delete the endpoint
// @Tags    Endpoint
// @Produce json
// @Param   uuid      path     string                    true  "Endpoint UUID"
// @Param   uninstall query    bool                      false "if set to true, the endpoint installation will be removed too"
// @Success 200       {object} GeneralResponse           "Success"
// @Failure 401       {object} GeneralResponse           "Unauthorized operation"
// @Failure 500       {object} GeneralResponse{code=int} "Internal server error"
// @Router  /endpoint/{uuid} [delete]
func (controller *EndpointController) delete(c *gin.Context) {
	uuid := c.Param("uuid")
	uninstall, err := strconv.ParseBool(c.DefaultQuery("uninstall", "false"))
	if err != nil {
		uninstall = false
	}
	if err := controller.endpointAppService.DeleteEndpoint(uuid, uninstall); err != nil {
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

// scan finds the endpoint installation status of an infra provider
// @Summary Scan the endpoints in an infra provider
// @Tags    Endpoint
// @Produce json
// @Param   provider body     service.EndpointScanRequest                      true "Provider UUID and endpoint type"
// @Success 200      {object} GeneralResponse{data=[]service.EndpointListItem} "Success"
// @Failure 401      {object} GeneralResponse                                  "Unauthorized operation"
// @Failure 500      {object} GeneralResponse{code=int}                        "Internal server error"
// @Router  /endpoint/scan [post]
func (controller *EndpointController) scan(c *gin.Context) {
	if endpointDetail, err := func() ([]service.EndpointScanItem, error) {
		request := &service.EndpointScanRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return nil, errors.Wrapf(err, "invalid request")
		}
		return controller.endpointAppService.ScanEndpoint(request)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: endpointDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// checkKubeFATE test connection to a KubeFATE endpoint
// @Summary Test connection to KubeFATE endpoint
// @Tags    Endpoint
// @Produce json
// @Param   uuid path     string                    true "Endpoint UUID"
// @Success 200  {object} GeneralResponse           "Success"
// @Failure 401  {object} GeneralResponse           "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int} "Internal server error"
// @Router  /endpoint/{uuid}/kubefate/check [post]
func (controller *EndpointController) checkKubeFATE(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.endpointAppService.CheckKubeFATEConnection(uuid); err != nil {
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

// getKubeFATEDeploymentYAML returns the yaml content for deploying KubeFATE
// @Summary Get KubeFATE installation YAML content
// @Tags    Endpoint
// @Produce json
// @Param   service_username    query    string                       true "username of the created KubeFATE service"
// @Param   service_password    query    string                       true "password of the created KubeFATE service"
// @Param   hostname            query    string                       true "hostname domain name for the KubeFATE ingress object"
// @Param   use_registry        query    bool                         true "use_registry is to choose to use registry or not"
// @Param   registry            query    string                       true "registry is registry address"
// @Param   use_registry_secret query    bool                         true "use_registry_secret is to choose to use registry secret or not"
// @Param   registry_server_url query    string                       true "registry_server_url is registry's server url"
// @Param   registry_username   query    string                       true "registry_username is registry's username"
// @Param   registry_password   query    string                       true "registry_password is registry's password"
// @Success 200                 {object} GeneralResponse{data=string} "Success"
// @Failure 401                 {object} GeneralResponse              "Unauthorized operation"
// @Failure 500                 {object} GeneralResponse{code=int}    "Internal server error"
// @Router  /endpoint/kubefate/yaml [get]
func (controller *EndpointController) getKubeFATEDeploymentYAML(c *gin.Context) {
	if yaml, err := func() (string, error) {
		namespace := c.DefaultQuery("namespace", "")
		serviceUsername := c.DefaultQuery("service_username", "admin")
		servicePassword := c.DefaultQuery("service_password", "admin")
		hostname := c.DefaultQuery("hostname", "kubefate.net")
		if serviceUsername == "" || servicePassword == "" || hostname == "" {
			return "", errors.New("missing necessary parameters")
		}
		useRegistry, err := strconv.ParseBool(c.DefaultQuery("use_registry", "false"))
		if err != nil {
			return "", err
		}
		registry := c.DefaultQuery("registry", "")
		if useRegistry && registry == "" {
			return "", errors.New("missing registry")
		}
		useRegistrySecret, err := strconv.ParseBool(c.DefaultQuery("use_registry_secret", "false"))
		if err != nil {
			return "", err
		}
		registryServerURL := c.DefaultQuery("registry_server_url", "")
		registryUsername := c.DefaultQuery("registry_username", "")
		registryPassword := c.DefaultQuery("registry_password", "")
		if useRegistrySecret && (registryServerURL == "" || registryUsername == "" || registryPassword == "") {
			return "", errors.New("missing registry secret credentials")
		}
		return controller.endpointAppService.GetKubeFATEDeploymentYAML(namespace, serviceUsername, servicePassword, hostname, valueobject.KubeRegistryConfig{
			UseRegistry:       useRegistry,
			Registry:          registry,
			UseRegistrySecret: useRegistrySecret,
			RegistrySecretConfig: valueobject.KubeRegistrySecretConfig{
				ServerURL: registryServerURL,
				Username:  registryUsername,
				Password:  registryPassword,
			},
		})
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: yaml,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// create a new endpoint
// @Summary Create a new endpoint by install a new one or add an existing one
// @Tags    Endpoint
// @Produce json
// @Param   provider body     service.EndpointCreationRequest true "The endpoint information, currently for the type field only 'KubeFATE' is supported"
// @Success 200      {object} GeneralResponse                 "Success, the returned data contains the created endpoint"
// @Failure 401      {object} GeneralResponse                 "Unauthorized operation"
// @Failure 500      {object} GeneralResponse{code=int}       "Internal server error"
// @Router  /endpoint [post]
func (controller *EndpointController) create(c *gin.Context) {
	if uuid, err := func() (string, error) {
		req := &service.EndpointCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		return controller.endpointAppService.CreateEndpoint(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: uuid,
		}
		c.JSON(http.StatusOK, resp)
	}
}
