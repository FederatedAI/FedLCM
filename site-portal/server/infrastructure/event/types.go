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

package event

// Event is an interface represent a event for this site
type Event interface {
	// GetUrl returns the url path for the event
	GetUrl() string
}

// ProjectParticipantUpdateEvent is an event triggered when project participant info is change
type ProjectParticipantUpdateEvent struct {
	UUID        string `json:"uuid"`
	PartyID     uint   `json:"party_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (e ProjectParticipantUpdateEvent) GetUrl() string {
	return "project/internal/event/participant/update"
}

// ModelCreationEvent is an event triggered when a modeling job is finished
type ModelCreationEvent struct {
	Name                   string            `json:"name"`
	ModelID                string            `json:"model_id"`
	ModelVersion           string            `json:"model_version"`
	ComponentName          string            `json:"component_name"`
	ProjectUUID            string            `json:"project_uuid"`
	JobUUID                string            `json:"job_uuid"`
	JobName                string            `json:"job_name"`
	Role                   string            `json:"role"`
	PartyID                uint              `json:"party_id"`
	Evaluation             map[string]string `json:"evaluation"`
	ComponentAlgorithmType uint8             `json:"algorithm_type"`
}

func (e ModelCreationEvent) GetUrl() string {
	return "model/internal/event/create"
}

// ProjectParticipantSyncEvent is an event triggered when project participant info needs to be synced from fml manager
type ProjectParticipantSyncEvent struct {
	ProjectUUID string `json:"project_uuid"`
}

func (e ProjectParticipantSyncEvent) GetUrl() string {
	return "project/internal/event/participant/sync"
}

// ProjectDataSyncEvent is an event triggered when project data info needs to be synced from fml manager
type ProjectDataSyncEvent struct {
	ProjectUUID string `json:"project_uuid"`
}

func (e ProjectDataSyncEvent) GetUrl() string {
	return "project/internal/event/data/sync"
}

// ProjectListSyncEvent is an event triggered when project list needs to be synced
type ProjectListSyncEvent struct {
}

func (e ProjectListSyncEvent) GetUrl() string {
	return "project/internal/event/list/sync"
}

// ProjectSelfUnregistrationEvent is an event triggered when this site unregistered from the fml manager
type ProjectSelfUnregistrationEvent struct {
}

func (e ProjectSelfUnregistrationEvent) GetUrl() string {
	return "project/internal/event/participant/unregister"
}
