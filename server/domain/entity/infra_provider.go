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

// InfraProviderBase contains common information of an infra provider
type InfraProviderBase struct {
	gorm.Model
	UUID        string            `json:"uuid" gorm:"type:varchar(36);index;unique"`
	Name        string            `json:"name" gorm:"type:varchar(255);unique;not null"`
	Description string            `json:"description" gorm:"type:text"`
	Type        InfraProviderType `gorm:"type:varchar(255)"`
}

// InfraProviderType is the enum types of the provider
type InfraProviderType string

const (
	InfraProviderTypeUnknown InfraProviderType = "Unknown"
	InfraProviderTypeK8s     InfraProviderType = "Kubernetes"
)
