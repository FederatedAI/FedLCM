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

package repo

// EventRepository is the interface to handle event's persistence related actions
type EventRepository interface {
	// Create takes a *entity.Event and creates an event record in the repository
	Create(interface{}) error
	// ListByEntityUUID returns []entity.Event instances list that contain the specified entity uuid
	ListByEntityUUID(string) (interface{}, error)
}
