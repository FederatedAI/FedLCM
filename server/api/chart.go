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
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/gin-gonic/gin"
)

// ChartController provides API handlers for the chart related APIs
type ChartController struct {
	chartApp *service.ChartApp
}

// NewChartController returns a controller instance to handle chart API requests
func NewChartController(chartRepo repo.ChartRepository) *ChartController {
	return &ChartController{
		chartApp: &service.ChartApp{
			ChartRepo: chartRepo,
		},
	}
}

// Route sets up route mappings to chart related APIs
func (controller *ChartController) Route(r *gin.RouterGroup) {
	chart := r.Group("chart")
	chart.Use(authMiddleware.MiddlewareFunc())
	{
		chart.GET("", controller.list)
		chart.GET("/:uuid", controller.get)
		// TODO: support add/delete
	}
}

// list returns the chart list
// @Summary  Return chart list, optionally with the specified type
// @Tags     Chart
// @Produce  json
// @Param    type  query     uint8                                          false  "if set, it should be the chart type"
// @Success  200   {object}  GeneralResponse{data=[]service.ChartListItem}  "Success"
// @Failure  401   {object}  GeneralResponse                                "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                      "Internal server error"
// @Router   /chart [get]
func (controller *ChartController) list(c *gin.Context) {
	if charList, err := func() ([]service.ChartListItem, error) {
		t64, err := strconv.ParseUint(c.DefaultQuery("type", "0"), 10, 8)
		if err != nil {
			return nil, err
		}
		t := entity.ChartType(t64)
		return controller.chartApp.List(t)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code:    constants.RespNoErr,
			Message: "",
			Data:    charList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// get returns detailed information of a chart
// @Summary  Get chart's detailed info
// @Tags     Chart
// @Produce  json
// @Param    uuid  path      string                                     true  "Chart UUID"
// @Success  200   {object}  GeneralResponse{data=service.ChartDetail}  "Success"
// @Failure  401   {object}  GeneralResponse                            "Unauthorized operation"
// @Failure  500   {object}  GeneralResponse{code=int}                  "Internal server error"
// @Router   /chart/{uuid} [get]
func (controller *ChartController) get(c *gin.Context) {
	uuid := c.Param("uuid")
	if chartDetail, err := controller.chartApp.GetDetail(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: chartDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}
