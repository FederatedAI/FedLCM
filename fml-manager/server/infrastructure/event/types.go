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

// Event is an interface representing an event to be handled
type Event interface {
	// GetUrl returns the base path of the event
	GetUrl() string
}

// ProjectParticipantUpdateEvent is triggered when a site info is updated
type ProjectParticipantUpdateEvent struct {
	UUID        string `json:"uuid"`
	PartyID     uint   `json:"party_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (e ProjectParticipantUpdateEvent) GetUrl() string {
	return "project/event/participant/update"
}

// ProjectParticipantUnregistrationEvent is triggered when a site is unregistered
type ProjectParticipantUnregistrationEvent struct {
	SiteUUID string `json:"siteUUID"`
}

func (e ProjectParticipantUnregistrationEvent) GetUrl() string {
	return "project/event/participant/unregister"
}
