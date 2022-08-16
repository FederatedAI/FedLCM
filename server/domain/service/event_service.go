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

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	uuid "github.com/satori/go.uuid"
)

// EventServiceInt provides functions to work with core entities' lifecycle events
type EventServiceInt interface {
	// CreateEvent creates a new event record
	CreateEvent(eventType entity.EventType, entityType entity.EntityType, entityUUID string, description string, level entity.EventLogLevel) error
}

// EventService provides functions to work with core entities' lifecycle events
type EventService struct {
	EventRepo repo.EventRepository
}

func (s *EventService) CreateEvent(eventType entity.EventType, entityType entity.EntityType, entityUUID string, description string, level entity.EventLogLevel) error {
	data := &entity.EventData{
		Description: description,
		LogLevel:    level.String(),
	}

	eventData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	event := &entity.Event{
		UUID:       uuid.NewV4().String(),
		Type:       eventType,
		EntityUUID: entityUUID,
		EntityType: entityType,
		Data:       string(eventData),
	}
	return s.EventRepo.Create(event)
}
