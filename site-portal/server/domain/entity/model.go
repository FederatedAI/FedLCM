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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Model is the domain entity of the model management context
type Model struct {
	gorm.Model
	UUID                   string `gorm:"type:varchar(36)"`
	Name                   string `gorm:"type:varchar(255)"`
	FATEModelID            string `gorm:"type:varchar(255);column:fate_model_id"`
	FATEModelVersion       string `gorm:"type:varchar(255);column:fate_model_version"`
	ProjectUUID            string `gorm:"type:varchar(36)"`
	ProjectName            string `gorm:"type:varchar(255)"`
	JobUUID                string `gorm:"type:varchar(36)"`
	JobName                string `gorm:"type:varchar(255)"`
	ComponentName          string `gorm:"type:varchar(255)"`
	ComponentAlgorithmType ComponentAlgorithmType
	Role                   string `gorm:"type:varchar(255)"`
	PartyID                uint
	Evaluation             valueobject.ModelEvaluation `gorm:"type:text"`
	Repo                   repo.ModelRepository        `gorm:"-"`
}

//ComponentAlgorithmType is the type enum of the algorithm
type ComponentAlgorithmType uint8

const (
	ComponentAlgorithmTypeUnknown ComponentAlgorithmType = iota
	ComponentAlgorithmTypeHomoLR
	ComponentAlgorithmTypeHomoSBT
)

// Create initializes the model and creates it in the repo
func (model *Model) Create() error {
	model.UUID = uuid.NewV4().String()
	if err := model.Repo.Create(model); err != nil {
		return errors.Wrap(err, "failed to create model")
	}
	return nil
}
