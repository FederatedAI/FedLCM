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

import "gorm.io/gorm"

// EndpointBase contains common information of an endpoint
type EndpointBase struct {
	gorm.Model
	UUID              string       `json:"uuid" gorm:"type:varchar(36);index;unique"`
	InfraProviderUUID string       `gorm:"type:varchar(36)"`
	Namespace         string       `json:"namespace" gorm:"type:varchar(255)"`
	Name              string       `json:"name" gorm:"type:varchar(255);unique;not null"`
	Description       string       `json:"description" gorm:"type:text"`
	Version           string       `json:"version" gorm:"type:varchar(255)"`
	Type              EndpointType `gorm:"type:varchar(255)"`
	Status            EndpointStatus
}

// EndpointType is the enum types of the provider
type EndpointType string

const (
	EndpointTypeUnknown  EndpointType = "Unknown"
	EndpointTypeKubeFATE EndpointType = "KubeFATE"
)

// EndpointStatus is the status of the endpoint
type EndpointStatus uint8

const (
	EndpointStatusUnknown EndpointStatus = iota
	EndpointStatusCreating
	EndpointStatusReady
	EndpointStatusDismissed
	EndpointStatusUnavailable
	EndpointStatusDeleting
)

func (s EndpointStatus) String() string {
	res := "Unknown"
	switch s {
	case EndpointStatusCreating:
		res = "Creating"
	case EndpointStatusReady:
		res = "Ready"
	case EndpointStatusDismissed:
		res = "Dismissed"
	case EndpointStatusUnavailable:
		res = "Unavailable"
	case EndpointStatusDeleting:
		res = "Deleting"
	}
	return res
}
