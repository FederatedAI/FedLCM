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

package gorm

import (
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
)

// EventRepo is the implementation of the repo.EventRepository interface
type EventRepo struct{}

var _ repo.EventRepository = (*EventRepo)(nil)

func (r *EventRepo) ListByEntityUUID(EntityUUID string) (interface{}, error) {
	var eventList []entity.Event
	err := db.Where("entity_uuid = ?", EntityUUID).Order("created_at desc").Find(&eventList).Error
	if err != nil {
		return 0, err
	}
	return eventList, nil
}

func (r *EventRepo) Create(instance interface{}) error {
	event := instance.(*entity.Event)

	if err := db.Create(event).Error; err != nil {
		return err
	}
	return nil
}

// InitTable makes sure the table is created in the db
func (r *EventRepo) InitTable() {
	if err := db.AutoMigrate(entity.Event{}); err != nil {
		panic(err)
	}
}
