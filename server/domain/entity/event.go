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

package entity

import (
	"gorm.io/gorm"
)

// Event records events related to a certain entity
type Event struct {
	gorm.Model
	UUID       string `gorm:"type:varchar(36);index;unique"`
	Type       EventType
	EntityUUID string `gorm:"type:varchar(36);column:entity_uuid"`
	EntityType EntityType
	Data       string `gorm:"type:text" `
}

// EventType is the supported Event type
type EventType uint8

const (
	EventTypeUnknown EventType = iota
	EventTypeLogMessage
)

// EventData is detail info of an event
type EventData struct {
	Description string `json:"description"`
	LogLevel    string `json:"log_level"`
}

// EventLogLevel is the level of the log event
type EventLogLevel uint8

const (
	EventLogLevelUnknown EventLogLevel = iota
	EventLogLevelInfo
	EventLogLevelError
)

// EntityType is the entity which records events
type EntityType uint8

const (
	EntityTypeUnknown EntityType = iota
	EntityTypeEndpoint
	EntityTypeExchange
	EntityTypeCluster
)

// openfl
const (
	EntityTypeOpenFLDirector EntityType = iota + 101
	EntityTypeOpenFLEnvoy
)

func (t EventLogLevel) String() string {
	switch t {
	case EventLogLevelInfo:
		return "Info"
	case EventLogLevelError:
		return "Error"
	}
	return "Unknown"
}

func (t EntityType) String() string {
	switch t {
	case EntityTypeEndpoint:
		return "Endpoint"
	case EntityTypeExchange:
		return "Exchange"
	case EntityTypeCluster:
		return "Cluster"
	case EntityTypeOpenFLDirector:
		return "OpenFL Director"
	}
	return "Unknown"
}
