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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
)

// ModelDeploymentRepo implements repo.ModelDeploymentRepository using gorm and PostgreSQL
type ModelDeploymentRepo struct{}

// make sure ModelRepo implements the repo.ModelRepository interface
var _ repo.ModelDeploymentRepository = (*ModelDeploymentRepo)(nil)

func (r *ModelDeploymentRepo) Create(instance interface{}) error {
	newDeployment := instance.(*entity.ModelDeployment)
	// Add records
	return db.Model(&entity.ModelDeployment{}).Create(newDeployment).Error
}

func (r *ModelDeploymentRepo) UpdateStatusByUUID(instance interface{}) error {
	deployment := instance.(*entity.ModelDeployment)
	return db.Model(&entity.ModelDeployment{}).Where("uuid = ?", deployment.UUID).
		Update("status", deployment.Status).Error
}

func (r *ModelDeploymentRepo) UpdateResultJsonByUUID(instance interface{}) error {
	deployment := instance.(*entity.ModelDeployment)
	return db.Model(&entity.Job{}).Where("uuid = ?", deployment.UUID).
		Update("result_json", deployment.ResultJson).Error
}

// InitTable make sure the table is created in the db
func (r *ModelDeploymentRepo) InitTable() {
	if err := db.AutoMigrate(&entity.ModelDeployment{}); err != nil {
		panic(err)
	}
}
