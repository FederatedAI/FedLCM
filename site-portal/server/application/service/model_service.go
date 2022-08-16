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

package service

import (
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
)

// ModelApp provides interfaces for model management APIs
type ModelApp struct {
	ModelRepo           repo.ModelRepository
	ModelDeploymentRepo repo.ModelDeploymentRepository
	SiteRepo            repo.SiteRepository
	ProjectRepo         repo.ProjectRepository
}

// ModelInfoBase contains the basic info of a model
type ModelInfoBase struct {
	Name          string    `json:"name"`
	UUID          string    `json:"uuid"`
	ModelID       string    `json:"model_id"`
	ModelVersion  string    `json:"model_version"`
	ComponentName string    `json:"component_name"`
	CreateTime    time.Time `json:"create_time"`
	ProjectUUID   string    `json:"project_uuid"`
	JobUUID       string    `json:"job_uuid"`
	JobName       string    `json:"job_name"`
	Role          string    `json:"role"`
	PartyID       uint      `json:"party_id"`
}

// ModelListItem contains info necessary to show models in a list
type ModelListItem struct {
	ModelInfoBase
	ProjectName   string `json:"project_name"`
	ComponentName string `json:"component_name"`
}

// ModelDetail adds the evaluation info
type ModelDetail struct {
	ModelListItem
	Evaluation valueobject.ModelEvaluation `json:"evaluation"`
}

// ModelCreationRequest is the request struct for creating a model
type ModelCreationRequest struct {
	ModelInfoBase
	Evaluation             valueobject.ModelEvaluation   `json:"evaluation"`
	ComponentAlgorithmType entity.ComponentAlgorithmType `json:"algorithm_type"`
}

// List returns model list of the current site or of the specified project
func (app *ModelApp) List(projectUUID string) ([]ModelListItem, error) {
	var modelList []ModelListItem
	queryFunc := app.ModelRepo.GetAll
	if projectUUID != "" {
		queryFunc = func() (interface{}, error) {
			return app.ModelRepo.GetListByProjectUUID(projectUUID)
		}
	}
	modelEntityListInstance, err := queryFunc()
	if err != nil {
		return nil, err
	}
	modelEntityList := modelEntityListInstance.([]entity.Model)
	for _, modelEntity := range modelEntityList {
		modelList = append(modelList, ModelListItem{
			ModelInfoBase: ModelInfoBase{
				Name:         modelEntity.Name,
				UUID:         modelEntity.UUID,
				ModelID:      modelEntity.FATEModelID,
				ModelVersion: modelEntity.FATEModelVersion,
				CreateTime:   modelEntity.CreatedAt,
				ProjectUUID:  modelEntity.ProjectUUID,
				JobUUID:      modelEntity.JobUUID,
				JobName:      modelEntity.JobName,
				Role:         modelEntity.Role,
				PartyID:      modelEntity.PartyID,
			},
			ProjectName:   modelEntity.ProjectName,
			ComponentName: modelEntity.ComponentName,
		})
	}
	return modelList, nil
}

// Delete deletes the specified model
func (app *ModelApp) Delete(modelUUID string) error {
	return app.ModelRepo.DeleteByUUID(modelUUID)
}

// Get returns detailed info of a model
func (app *ModelApp) Get(modelUUID string) (*ModelDetail, error) {
	modelEntityInstance, err := app.ModelRepo.GetByUUID(modelUUID)
	if err != nil {
		return nil, err
	}
	modelEntity := modelEntityInstance.(*entity.Model)

	return &ModelDetail{
		ModelListItem: ModelListItem{
			ModelInfoBase: ModelInfoBase{
				Name:         modelEntity.Name,
				UUID:         modelEntity.UUID,
				ModelID:      modelEntity.FATEModelID,
				ModelVersion: modelEntity.FATEModelVersion,
				CreateTime:   modelEntity.CreatedAt,
				ProjectUUID:  modelEntity.ProjectUUID,
				JobUUID:      modelEntity.JobUUID,
				JobName:      modelEntity.JobName,
				Role:         modelEntity.Role,
				PartyID:      modelEntity.PartyID,
			},
			ProjectName:   modelEntity.ProjectName,
			ComponentName: modelEntity.ComponentName,
		},
		Evaluation: modelEntity.Evaluation,
	}, nil
}

// Create creates the model
func (app *ModelApp) Create(request *ModelCreationRequest) error {
	projectInstance, err := app.ProjectRepo.GetByUUID(request.ProjectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)
	modelEntity := &entity.Model{
		Name:                   request.Name,
		FATEModelID:            request.ModelID,
		FATEModelVersion:       request.ModelVersion,
		ProjectUUID:            request.ProjectUUID,
		ProjectName:            project.Name,
		JobUUID:                request.JobUUID,
		JobName:                request.JobName,
		ComponentName:          request.ComponentName,
		ComponentAlgorithmType: request.ComponentAlgorithmType,
		Role:                   request.Role,
		PartyID:                request.PartyID,
		Evaluation:             request.Evaluation,
		Repo:                   app.ModelRepo,
	}
	return modelEntity.Create()
}

// Publish publishes the model to an online serving system
func (app *ModelApp) Publish(request *service.ModelDeploymentRequest) (*entity.ModelDeployment, error) {
	site, err := app.loadSite()
	if err != nil {
		return nil, err
	}
	request.KubeflowConfig = site.KubeflowConfig
	request.FATEFlowContext = entity.FATEFlowContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}

	domainService := service.ModelService{
		ModelRepo:           app.ModelRepo,
		ModelDeploymentRepo: app.ModelDeploymentRepo,
	}
	return domainService.DeployModel(request)
}

// GetSupportedDeploymentTypes gets the supported deployment types this model can use
func (app *ModelApp) GetSupportedDeploymentTypes(modelUUID string) ([]entity.ModelDeploymentType, error) {
	domainService := service.ModelService{
		ModelRepo:           app.ModelRepo,
		ModelDeploymentRepo: app.ModelDeploymentRepo,
	}
	return domainService.GetSupportedDeploymentType(modelUUID)
}

// loadSite is a helper function to return site entity object
func (app *ModelApp) loadSite() (*entity.Site, error) {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load site info")
	}
	return site, nil
}
