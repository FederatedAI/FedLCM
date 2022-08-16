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

package service

import (
	"encoding/json"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
)

// EventApp provide functions to manage the events
type EventApp struct {
	EventRepo repo.EventRepository
}

// EventListItem contains basic information of an event
type EventListItem struct {
	UUID       string            `json:"uuid"`
	Type       entity.EventType  `json:"type"`
	CreatedAt  time.Time         `json:"created_at"`
	EntityUUID string            `json:"entity_uuid"`
	EntityType entity.EntityType `json:"entity_type"`
	Data       entity.EventData  `json:"data"`
}

// GetEventList returns events of a entity
func (app *EventApp) GetEventList(entity_uuid string) ([]EventListItem, error) {
	eventInstanceList, err := app.EventRepo.ListByEntityUUID(entity_uuid)
	if err != nil {
		return nil, err
	}
	domainEventList := eventInstanceList.([]entity.Event)

	var eventList []EventListItem
	for _, domainEvent := range domainEventList {
		var data entity.EventData
		err = json.Unmarshal([]byte(domainEvent.Data), &data)
		if err != nil {
			return nil, err
		}
		eventList = append(eventList, EventListItem{
			UUID:       domainEvent.UUID,
			Type:       domainEvent.Type,
			CreatedAt:  domainEvent.CreatedAt,
			EntityUUID: domainEvent.EntityUUID,
			EntityType: domainEvent.EntityType,
			Data:       data,
		})
	}
	return eventList, nil
}
