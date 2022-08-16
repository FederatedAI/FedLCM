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

// CertificateController provides API handlers for the certificate related APIs
type CertificateController struct {
	certificateApp *service.CertificateApp
}

// NewCertificateController returns a controller instance to handle certificate API requests
func NewCertificateController(caRepo repo.CertificateAuthorityRepository,
	certRepo repo.CertificateRepository,
	bindingRepo repo.CertificateBindingRepository,
	participantFATERepo repo.ParticipantFATERepository,
	participantOpenFLRepo repo.ParticipantOpenFLRepository,
	federationFATERepo repo.FederationRepository,
	federationOpenFLRepo repo.FederationRepository) *CertificateController {
	return &CertificateController{
		certificateApp: &service.CertificateApp{
			CertificateAuthorityRepo: caRepo,
			CertificateRepo:          certRepo,
			CertificateBindingRepo:   bindingRepo,
			ParticipantFATERepo:      participantFATERepo,
			ParticipantOpenFLRepo:    participantOpenFLRepo,
			FederationFATERepo:       federationFATERepo,
			FederationOpenFLRepo:     federationOpenFLRepo,
		},
	}
}

// Route sets up route mappings to certificate related APIs
func (controller *CertificateController) Route(r *gin.RouterGroup) {
	certificate := r.Group("certificate")
	certificate.Use(authMiddleware.MiddlewareFunc())
	{
		certificate.GET("", controller.list)
		certificate.DELETE("/:uuid", controller.delete)

	}
}

// list returns the certificate list
// @Summary  Return issued certificate list
// @Tags     Certificate
// @Produce  json
// @Success  200  {object}  GeneralResponse{data=[]service.CertificateListItem}  "Success"
// @Failure  401  {object}  GeneralResponse                                      "Unauthorized operation"
// @Failure  500  {object}  GeneralResponse{code=int}                            "Internal server error"
// @Router   /certificate [get]
func (controller *CertificateController) list(c *gin.Context) {
	if certList, err := controller.certificateApp.List(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code:    constants.RespNoErr,
			Message: "",
			Data:    certList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// delete removes the certificate which has no participant bindings
// @Summary  Delete the certificate  which has no participant bindings
// @Tags     Certificate
// @Produce  json
// @Param    uuid  path      string                     true  "Certificate UUID"
// @Success  200   {object}  GeneralResponse            "Success"
// @Failure  401   {object}  GeneralResponse            "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}  "Internal server error"
// @Router   /certificate/{uuid} [delete]
func (controller *CertificateController) delete(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.certificateApp.DeleteCertificate(uuid); err != nil {
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
