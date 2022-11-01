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

// EventController provides API handlers for the event related APIs
type EventController struct {
	EventApp *service.EventApp
}

// NewEventController returns a controller instance to handle event API requests
func NewEventController(eventRepo repo.EventRepository) *EventController {
	return &EventController{
		EventApp: &service.EventApp{
			EventRepo: eventRepo,
		},
	}
}

// Route sets up route mappings to event related APIs
func (controller *EventController) Route(r *gin.RouterGroup) {
	event := r.Group("event")
	event.Use(authMiddleware.MiddlewareFunc())
	{
		event.GET("/:entity_uuid", controller.get)
	}
}

// list returns the event list of related entity
// @Summary Return event list of related entity
// @Tags    Event
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.EventListItem} "Success"
// @Failure 401 {object} GeneralResponse                               "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int}                     "Internal server error"
// @Router  /event/{entity_uuid} [get]
func (controller *EventController) get(c *gin.Context) {
	entity_uuid := c.Param("entity_uuid")
	eventList, err := controller.EventApp.GetEventList(entity_uuid)
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
			Data:    eventList,
		}
		c.JSON(http.StatusOK, resp)
	}
}
