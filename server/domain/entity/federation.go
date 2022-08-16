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
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"gorm.io/gorm"
)

// Federation is a logic concept that contains multiple FML participants
type Federation struct {
	gorm.Model
	UUID        string                    `gorm:"type:varchar(36);index;unique"`
	Name        string                    `gorm:"type:varchar(255);not null"`
	Description string                    `gorm:"type:text"`
	Type        FederationType            `gorm:"type:varchar(255);not null"`
	Repo        repo.FederationRepository `gorm:"-"`
}

// FederationType is the type of federation
type FederationType string

const (
	FederationTypeFATE   FederationType = "FATE"
	FederationTypeOpenFL FederationType = "OpenFL"
)
