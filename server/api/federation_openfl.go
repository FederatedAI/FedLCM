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
	domainService "github.com/FederatedAI/FedLCM/server/domain/service"
	"github.com/gin-gonic/gin"
)

// createOpenFL creates a new OpenFL federation
// @Summary  Create a new OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    federation  body      service.FederationOpenFLCreationRequest  true  "The federation info"
// @Success  200         {object}  GeneralResponse                          "Success, the data field is the created federation's uuid"
// @Failure  401         {object}  GeneralResponse                          "Unauthorized operation"
// @Failure  500         {object}  GeneralResponse{code=int}                "Internal server error"
// @Router   /federation/openfl [post]
func (controller *FederationController) createOpenFL(c *gin.Context) {
	if uuid, err := func() (string, error) {
		creationInfo := &service.FederationOpenFLCreationRequest{}
		if err := c.ShouldBindJSON(creationInfo); err != nil {
			return "", err
		}
		return controller.federationApp.CreateOpenFLFederation(creationInfo)
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

// getOpenFL returns detailed information of an OpenFL federation
// @Summary  Get specific info of an OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    uuid  path      string                                                true  "federation UUID"
// @Success  200   {object}  GeneralResponse{data=service.FederationOpenFLDetail}  "Success"
// @Failure  401   {object}  GeneralResponse                                       "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                             "Internal server error"
// @Router   /federation/openfl/{uuid} [get]
func (controller *FederationController) getOpenFL(c *gin.Context) {
	uuid := c.Param("uuid")
	if info, err := controller.federationApp.GetOpenFLFederation(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: info,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// deleteOpenFL deletes the specified federation
// @Summary  Delete an OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    uuid  path      string                     true  "federation UUID"
// @Success  200   {object}  GeneralResponse            "Success"
// @Failure  401   {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /federation/openfl/{uuid} [delete]
func (controller *FederationController) deleteOpenFL(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.federationApp.DeleteOpenFLFederation(uuid); err != nil {
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

// listOpenFLToken returns token list of the specified federation
// @Summary  Get registration token list of the specified OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    uuid  path      string                                                           true  "federation UUID"
// @Success  200   {object}  GeneralResponse{data=[]service.RegistrationTokenOpenFLListItem}  "Success"
// @Failure  401   {object}  GeneralResponse                                                  "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                                        "Internal server error"
// @Router   /federation/openfl/{uuid}/token [get]
func (controller *FederationController) listOpenFLToken(c *gin.Context) {
	if tokens, err := func() ([]service.RegistrationTokenOpenFLListItem, error) {
		federationUUID := c.Param("uuid")
		return controller.federationApp.ListOpenFLToken(federationUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: tokens,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createOpenFLToken creates a new registration token for an OpenFL federation
// @Summary  Create a new registration token for an OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    uuid   path      string                                    true  "federation UUID"
// @Param    token  body      service.RegistrationTokenOpenFLBasicInfo  true  "The federation info"
// @Success  200    {object}  GeneralResponse                           "Success"
// @Failure  401    {object}  GeneralResponse                           "Unauthorized operation"
// @Failure  500    {object}  GeneralResponse{code=int}                 "Internal server error"
// @Router   /federation/openfl/{uuid}/token [post]
func (controller *FederationController) createOpenFLToken(c *gin.Context) {
	if err := func() error {
		creationInfo := &service.RegistrationTokenOpenFLBasicInfo{}
		if err := c.ShouldBindJSON(creationInfo); err != nil {
			return err
		}
		federationUUID := c.Param("uuid")
		return controller.federationApp.GeneratedOpenFLToken(creationInfo, federationUUID)
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

// deleteOpenFLToken deletes the specified OpenFL federation registration token
// @Summary  Delete an OpenFL federation registration token
// @Tags     Federation
// @Produce  json
// @Param    uuid       path      string                     true  "federation UUID"
// @Param    tokenUUID  path      string                     true  "token UUID"
// @Success  200        {object}  GeneralResponse            "Success"
// @Failure  401        {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500        {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /federation/openfl/{uuid}/token/{tokenUUID} [delete]
func (controller *FederationController) deleteOpenFLToken(c *gin.Context) {
	uuid := c.Param("tokenUUID")
	federationUUID := c.Param("uuid")
	if err := controller.federationApp.DeleteOpenFLToken(uuid, federationUUID); err != nil {
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

// getOpenFLDirectorDeploymentYAML returns deployment yaml content for deploying OpenFL director
// @Summary  Get OpenFL director deployment yaml
// @Tags     Federation
// @Produce  json
// @Param    chart_uuid           query     string                        true  "the chart uuid"
// @Param    federation_uuid      query     string                        true  "the federation uuid"
// @Param    name                 query     string                        true  "name of the deployment"
// @Param    namespace            query     string                        true  "namespace of the deployment"
// @Param    service_type         query     int                           true  "type of the service to be exposed 1: LoadBalancer 2: NodePort"
// @Param    jupyter_password     query     string                        true  "password to access the Jupyter Notebook"
// @Param    registry             query     string                        true  "customized registry address"
// @Param    use_registry         query     bool                          true  "choose if use the customized registry config"
// @Param    use_registry_secret  query     bool                          true  "choose if use the customized registry secret"
// @Param    enable_psp           query     bool                          true  "choose if enable the podSecurityPolicy"
// @Success  200                  {object}  GeneralResponse{data=string}  "Success, the data field is the yaml content"
// @Failure  401                  {object}  GeneralResponse               "Unauthorized operation"
// @Failure  500                  {object}  GeneralResponse{code=int}     "Internal server error"
// @Router   /federation/openfl/director/yaml [get]
func (controller *FederationController) getOpenFLDirectorDeploymentYAML(c *gin.Context) {
	if yaml, err := func() (string, error) {
		req := &domainService.ParticipantOpenFLDirectorYAMLCreationRequest{}
		if err := c.ShouldBindQuery(req); err != nil {
			return "", err
		}
		if err := c.ShouldBindQuery(&req.RegistryConfig); err != nil {
			return "", err
		}
		return controller.participantAppService.GetOpenFLDirectorDeploymentYAML(req)
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

// createOpenFLDirector creates a new OpenFL director
// @Summary  Create a new OpenFL director
// @Tags     Federation
// @Produce  json
// @Param    uuid             path      string                                            true  "federation UUID"
// @Param    creationRequest  body      service.ParticipantOpenFLDirectorCreationRequest  true  "The creation requests"
// @Success  200              {object}  GeneralResponse                                   "Success, the data field is the created director's uuid"
// @Failure  401              {object}  GeneralResponse                                   "Unauthorized operation"
// @Failure  500              {object}  GeneralResponse{code=int}                         "Internal server error"
// @Router   /federation/openfl/{uuid}/director [post]
func (controller *FederationController) createOpenFLDirector(c *gin.Context) {
	if uuid, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantOpenFLDirectorCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.CreateOpenFLDirector(req)
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

// deleteOpenFLDirector deletes the specified OpenFL director
// @Summary  Delete an OpenFL director
// @Tags     Federation
// @Produce  json
// @Param    uuid          path      string                     true   "federation UUID"
// @Param    directorUUID  path      string                     true   "director UUID"
// @Param    force         query     bool                       false  "if set to true, will try to remove the director forcefully"
// @Success  200           {object}  GeneralResponse            "Success"
// @Failure  401           {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500           {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /federation/openfl/{uuid}/director/{directorUUID} [delete]
func (controller *FederationController) deleteOpenFLDirector(c *gin.Context) {
	directorUUID := c.Param("directorUUID")
	if err := func() error {
		force, err := strconv.ParseBool(c.DefaultQuery("force", "false"))
		if err != nil {
			return err
		}
		return controller.participantAppService.RemoveOpenFLDirector(directorUUID, force)
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

// getOpenFLParticipant returns participant list of the specified federation
// @Summary  Get participant list of the specified OpenFL federation
// @Tags     Federation
// @Produce  json
// @Param    uuid  path      string                                                           true  "federation UUID"
// @Success  200   {object}  GeneralResponse{data=service.ParticipantOpenFLListInFederation}  "Success"
// @Failure  401   {object}  GeneralResponse                                                  "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                                        "Internal server error"
// @Router   /federation/openfl/{uuid}/participant [get]
func (controller *FederationController) getOpenFLParticipant(c *gin.Context) {
	if participants, err := func() (*service.ParticipantOpenFLListInFederation, error) {
		federationUUID := c.Param("uuid")
		return controller.participantAppService.GetOpenFLParticipantList(federationUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: participants,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getOpenFLDirector returns detailed information of a OpenFL director
// @Summary  Get specific info of OpenFL director
// @Tags     Federation
// @Produce  json
// @Param    uuid          path      string                                              true  "federation UUID"
// @Param    directorUUID  path      string                                              true  "director UUID"
// @Success  200           {object}  GeneralResponse{data=service.OpenFLDirectorDetail}  "Success"
// @Failure  401           {object}  GeneralResponse                                     "Unauthorized operation"
// @Failure  500           {object}  GeneralResponse{code=int}                           "Internal server error"
// @Router   /federation/openfl/{uuid}/director/{directorUUID} [get]
func (controller *FederationController) getOpenFLDirector(c *gin.Context) {
	directorUUID := c.Param("directorUUID")
	if directorDetail, err := controller.participantAppService.GetOpenFLDirectorDetail(directorUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: directorDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// registerOpenFLEnvoy handles envoy registration request
// @Summary  Process Envoy registration request
// @Tags     Federation
// @Produce  json
// @Param    uuid                 path      string                                             true  "federation UUID"
// @Param    registrationRequest  body      service.ParticipantOpenFLEnvoyRegistrationRequest  true  "The creation requests"
// @Success  200                  {object}  GeneralResponse                                    "Success, the data field is the created director's uuid"
// @Failure  401                  {object}  GeneralResponse                                    "Unauthorized operation"
// @Failure  500                  {object}  GeneralResponse{code=int}                          "Internal server error"
// @Router   /federation/openfl/envoy/register [post]
func (controller *FederationController) registerOpenFLEnvoy(c *gin.Context) {
	if uuid, err := func() (string, error) {
		req := &domainService.ParticipantOpenFLEnvoyRegistrationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		return controller.participantAppService.HandleOpenFLEnvoyRegistration(req)
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

// deleteOpenFLEnvoy deletes the specified OpenFL envoy
// @Summary  Delete an OpenFL envoy
// @Tags     Federation
// @Produce  json
// @Param    uuid       path      string                     true   "federation UUID"
// @Param    envoyUUID  path      string                     true   "envoy UUID"
// @Param    force      query     bool                       false  "if set to true, will try to envoy the director forcefully"
// @Success  200        {object}  GeneralResponse            "Success"
// @Failure  401        {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500        {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /federation/openfl/{uuid}/envoy/{envoyUUID} [delete]
func (controller *FederationController) deleteOpenFLEnvoy(c *gin.Context) {
	envoyUUID := c.Param("envoyUUID")
	if err := func() error {
		force, err := strconv.ParseBool(c.DefaultQuery("force", "false"))
		if err != nil {
			return err
		}
		return controller.participantAppService.RemoveOpenFLEnvoy(envoyUUID, force)
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

// getOpenFLEnvoy returns detailed information of a OpenFL envoy
// @Summary  Get specific info of OpenFL envoy
// @Tags     Federation
// @Produce  json
// @Param    uuid       path      string                                           true  "federation UUID"
// @Param    envoyUUID  path      string                                           true  "envoy UUID"
// @Success  200        {object}  GeneralResponse{data=service.OpenFLEnvoyDetail}  "Success"
// @Failure  401        {object}  GeneralResponse                                  "Unauthorized operation"
// @Failure  500        {object}  GeneralResponse{code=int}                        "Internal server error"
// @Router   /federation/openfl/{uuid}/envoy/{envoyUUID} [get]
func (controller *FederationController) getOpenFLEnvoy(c *gin.Context) {
	envoyUUID := c.Param("envoyUUID")
	if envoyDetail, err := controller.participantAppService.GetOpenFLEnvoyDetail(envoyUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: envoyDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getOpenFLEnvoyWithToken returns detailed information of a OpenFL envoy, if the provided token is valid
// @Summary  Get specific info of OpenFL envoy, by providing the envoy uuid and token string
// @Tags     Federation
// @Produce  json
// @Param    uuid   path      string                                           true  "envoy UUID"
// @Param    token  query     string                                           true  "token string"
// @Success  200    {object}  GeneralResponse{data=service.OpenFLEnvoyDetail}  "Success"
// @Failure  401    {object}  GeneralResponse                                  "Unauthorized operation"
// @Failure  500    {object}  GeneralResponse{code=int}                        "Internal server error"
// @Router   /federation/openfl/envoy/{uuid} [get]
func (controller *FederationController) getOpenFLEnvoyWithToken(c *gin.Context) {
	token := c.Query("token")
	envoyUUID := c.Param("uuid")
	if envoyDetail, err := controller.participantAppService.GetOpenFLEnvoyDetailWithTokenVerification(envoyUUID, token); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: envoyDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}
