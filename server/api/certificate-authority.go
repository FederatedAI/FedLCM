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
	"github.com/gin-gonic/gin"
)

// CertificateAuthorityController provides API handlers for the certificate authority related APIs
type CertificateAuthorityController struct {
	certificateAuthorityApp *service.CertificateAuthorityApp
}

// NewCertificateAuthorityController returns a controller instance to handle certificate authority API requests
func NewCertificateAuthorityController(caRepo repo.CertificateAuthorityRepository) *CertificateAuthorityController {
	return &CertificateAuthorityController{
		certificateAuthorityApp: &service.CertificateAuthorityApp{
			CertificateAuthorityRepo: caRepo,
		},
	}
}

// Route sets up route mappings to certificate-authority related APIs
func (controller *CertificateAuthorityController) Route(r *gin.RouterGroup) {
	ca := r.Group("certificate-authority")
	ca.Use(authMiddleware.MiddlewareFunc())
	{
		ca.GET("", controller.get)
		ca.POST("", controller.create)
		ca.PUT("/:uuid", controller.update)
		ca.GET("/built-in-ca", controller.getBuiltInCAConfig)
	}
}

// get returns the certificate authority info
// @Summary  Return certificate authority info
// @Tags     CertificateAuthority
// @Produce  json
// @Success  200  {object}  GeneralResponse{data=service.CertificateAuthorityDetail}  "Success"
// @Failure  401  {object}  GeneralResponse                                           "Unauthorized operation"
// @Failure  500  {object}  GeneralResponse{code=int}                                 "Internal server error"
// @Router   /certificate-authority [get]
func (controller *CertificateAuthorityController) get(c *gin.Context) {
	if caInfo, err := controller.certificateAuthorityApp.Get(); err != nil {
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
			Data:    caInfo,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// create a new certificate authority
// @Summary  Create a new certificate authority
// @Tags     CertificateAuthority
// @Produce  json
// @Param    certificateAuthority  body      service.CertificateAuthorityEditableItem  true  "The CA information, currently for the type field only '1(StepCA)'  is  supported"
// @Success  200                   {object}  GeneralResponse                           "Success"
// @Failure  401                   {object}  GeneralResponse                           "Unauthorized operation"
// @Failure  500                   {object}  GeneralResponse{code=int}                 "Internal server error"
// @Router   /certificate-authority [post]
func (controller *CertificateAuthorityController) create(c *gin.Context) {
	if err := func() error {
		caInfo := &service.CertificateAuthorityEditableItem{}
		if err := c.ShouldBindJSON(caInfo); err != nil {
			return err
		}
		return controller.certificateAuthorityApp.CreateCA(caInfo)
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

// update the CA configuration
// @Summary  Updates the certificate authority
// @Tags     CertificateAuthority
// @Produce  json
// @Param    uuid                  path      string                                    true  "certificate authority UUID"
// @Param    certificateAuthority  body      service.CertificateAuthorityEditableItem  true  "The updated CA information"
// @Success  200                   {object}  GeneralResponse                           "Success"
// @Failure  401                   {object}  GeneralResponse                           "Unauthorized operation"
// @Failure  500                   {object}  GeneralResponse{code=int}                 "Internal server error"
// @Router   /certificate-authority/{uuid} [put]
func (controller *CertificateAuthorityController) update(c *gin.Context) {
	if err := func() error {
		uuid := c.Param("uuid")
		caInfo := &service.CertificateAuthorityEditableItem{}
		if err := c.ShouldBindJSON(caInfo); err != nil {
			return err
		}
		return controller.certificateAuthorityApp.Update(uuid, caInfo)
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

// getBuiltInCAConfig returns the built-in certificate authority config
// @Summary  Return the built-in certificate authority config
// @Tags     CertificateAuthority
// @Produce  json
// @Success  200  {object}  GeneralResponse{data=entity.CertificateAuthorityConfigurationStepCA}  "Success"
// @Failure  401  {object}  GeneralResponse                                                       "Unauthorized operation"
// @Failure  500  {object}  GeneralResponse{code=int}                                             "Internal server error"
// @Router   /certificate-authority/built-in-ca [get]
func (controller *CertificateAuthorityController) getBuiltInCAConfig(c *gin.Context) {
	caConfig, err := controller.certificateAuthorityApp.GetBuiltInCAConfig()
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
			Data:    caConfig,
		}
		c.JSON(http.StatusOK, resp)
	}
}
