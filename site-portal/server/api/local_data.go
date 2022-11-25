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
	"net/http/httputil"
	"path/filepath"

	"github.com/FederatedAI/FedLCM/site-portal/server/application/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/gin-gonic/gin"
)

// LocalDataController handles local data related APIs
type LocalDataController struct {
	localDataApp *service.LocalDataApp
}

// NewLocalDataController returns a controller instance to handle local data API requests
func NewLocalDataController(localDataRepo repo.LocalDataRepository,
	siteRepo repo.SiteRepository,
	projectRepo repo.ProjectRepository,
	projectDataRepo repo.ProjectDataRepository) *LocalDataController {
	return &LocalDataController{
		localDataApp: &service.LocalDataApp{
			LocalDataRepo:   localDataRepo,
			SiteRepo:        siteRepo,
			ProjectRepo:     projectRepo,
			ProjectDataRepo: projectDataRepo,
		},
	}
}

// Route set up route mappings to local data related APIs
func (controller *LocalDataController) Route(r *gin.RouterGroup) {
	data := r.Group("data")
	data.Use(authMiddleware.MiddlewareFunc())
	{
		data.POST("", controller.upload)
		data.POST("associate", controller.associate)
		data.GET("", controller.list)
		data.GET("/:uuid", controller.get)
		data.GET("/:uuid/columns", controller.getColumns)
		data.GET("/:uuid/file", controller.download)
		data.DELETE("/:uuid", controller.delete)
		data.PUT("/:uuid/idmetainfo", controller.putIdMetaInfo)
	}
}

// upload uploads a local csv data
// @Summary Upload a local csv data
// @Tags LocalData
// @Produce json
// @Param file        formData file true "The csv file"
// @Param name        formData string true "Data name"
// @Param description formData string true "Data description"
// @Success 200 {object} GeneralResponse{} "Success, the data field is the data UUID"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data [post]
func (controller *LocalDataController) upload(c *gin.Context) {
	if uuid, err := func() (string, error) {
		f, err := c.FormFile("file")
		if err != nil {
			return "", err
		}
		uploadRequest := service.LocalDataUploadRequest{}
		if err = c.ShouldBind(&uploadRequest); err != nil {
			return "", err
		}
		uploadRequest.FileHeader = f
		return controller.localDataApp.Upload(&uploadRequest)
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

// list returns all data records
// @Summary List all data records
// @Tags LocalData
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.LocalDataListItem} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data [get]
func (controller *LocalDataController) list(c *gin.Context) {
	if dataList, err := controller.localDataApp.List(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
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

// get returns detailed information of a data record
// @Summary Get data record's detailed info
// @Tags LocalData
// @Produce json
// @Param uuid path string true "Data UUID"
// @Success 200 {object} GeneralResponse{data=service.LocalDataDetail} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/{uuid} [get]
func (controller *LocalDataController) get(c *gin.Context) {
	uuid := c.Param("uuid")
	if dataDetail, err := controller.localDataApp.Get(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: dataDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// download returns the original csv data file
// @Summary Download data file
// @Tags LocalData
// @Produce json
// @Param uuid path string true "Data UUID"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/{uuid}/file [get]
func (controller *LocalDataController) download(c *gin.Context) {
	uuid := c.Param("uuid")
	if path, err := controller.localDataApp.GetFilePath(uuid); err != nil {
		if req, err := controller.localDataApp.GetDataDownloadRequest(uuid); err != nil {
			resp := &GeneralResponse{
				Code:    constants.RespInternalErr,
				Message: err.Error(),
			}
			c.JSON(http.StatusInternalServerError, resp)
		} else {
			proxy := &httputil.ReverseProxy{
				Director: func(*http.Request) {
					// no-op as we don't need to change the request
				},
			}
			proxy.ServeHTTP(c.Writer, req)
		}
	} else {
		// TODO: investigate and implement "chunked Transfer-Encoding"
		c.FileAttachment(path, filepath.Base(path))
	}
}

// delete removes the data
// @Summary Delete the data file, both the local copy and the FATE table
// @Tags LocalData
// @Produce json
// @Param uuid path string true "Data UUID"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/{uuid} [delete]
func (controller *LocalDataController) delete(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.localDataApp.DeleteData(uuid); err != nil {
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

// putIdMetaInfo update data record ID meta info
// @Summary Update data record's ID meta info
// @Tags LocalData
// @Produce json
// @Param info body service.LocalDataIDMetaInfoUpdateRequest true "The meta info"
// @Param uuid path string true "Data UUID"
// @Success 200 {object} GeneralResponse "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/{uuid}/idmetainfo [put]
func (controller *LocalDataController) putIdMetaInfo(c *gin.Context) {
	if err := func() error {
		uuid := c.Param("uuid")
		info := &service.LocalDataIDMetaInfoUpdateRequest{}
		if err := c.ShouldBindJSON(info); err != nil {
			return err
		}
		if err := controller.localDataApp.UpdateIDMetaInfo(uuid, info); err != nil {
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

// getColumns returns a list of the data's headers
// @Summary Get data headers
// @Tags LocalData
// @Produce json
// @Param uuid path string true "Data UUID"
// @Success 200 {object} GeneralResponse{data=[]string} "Success"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/{uuid}/columns [get]
func (controller *LocalDataController) getColumns(c *gin.Context) {
	uuid := c.Param("uuid")
	if columns, err := controller.localDataApp.GetColumns(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: columns,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// associate creates a local data item with existing flow data table
// @Summary Associate flow data table to a local data
// @Tags LocalData
// @Produce json
// @Param project body service.LocalDataAssociateRequest true "Local data association request"
// @Success 200 {object} GeneralResponse{} "Success, the data field is the data UUID"
// @Failure 401 {object} GeneralResponse "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int} "Internal server error"
// @Router /data/associate [post]
func (controller *LocalDataController) associate(c *gin.Context) {
	if uuid, err := func() (string, error) {
		request := &service.LocalDataAssociateRequest{}
		if err := c.ShouldBindJSON(request); err != nil {
			return "", err
		}
		return controller.localDataApp.AssociateFlowTable(request)
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
