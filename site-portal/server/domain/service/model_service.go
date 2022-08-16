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
	"encoding/json"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// ModelService provides domain service functions to work with trained models
type ModelService struct {
	ModelRepo           repo.ModelRepository
	ModelDeploymentRepo repo.ModelDeploymentRepository
}

// ModelDeploymentRequest is a request to deploy a model
type ModelDeploymentRequest struct {
	ServiceName        string                     `json:"service_name"`
	Type               entity.ModelDeploymentType `json:"deployment_type"`
	UserParametersJson string                     `json:"parameters_json"`
	ModelUUID          string                     `json:"-"`
	FATEFlowContext    entity.FATEFlowContext     `json:"-"`
	KubeflowConfig     valueobject.KubeflowConfig `json:"-"`
}

// DeployModel deploy the requested model to the requested platform
func (s *ModelService) DeployModel(request *ModelDeploymentRequest) (*entity.ModelDeployment, error) {
	model, err := s.loadModel(request.ModelUUID)
	if err != nil {
		return nil, err
	}

	// validate
	supportedTypes, err := s.GetSupportedDeploymentType(request.ModelUUID)
	if err != nil {
		return nil, err
	}
	if supported := func(t entity.ModelDeploymentType) bool {
		for _, supportedType := range supportedTypes {
			if t == supportedType {
				return true
			}
		}
		return false
	}(request.Type); !supported {
		return nil, errors.Errorf("cannot deploy model to the specified deployment type: %v", request.Type)
	}

	if err := request.KubeflowConfig.Validate(); err != nil {
		return nil, errors.Wrapf(err, "failed to validate kubeflow configuration")
	}

	// create the deployment object
	requestJsonByte, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, err
	}
	requestJson := string(requestJsonByte)

	deploymentParamsJson, err := valueobject.GetKFServingDeploymentParametersJson(request.UserParametersJson, request.KubeflowConfig)
	if err != nil {
		return nil, err
	}

	modelDeployment := &entity.ModelDeployment{
		UUID:                     uuid.NewV4().String(),
		ServiceName:              request.ServiceName,
		ModelUUID:                request.ModelUUID,
		Type:                     request.Type,
		Status:                   entity.ModelDeploymentStatusCreated,
		DeploymentParametersJson: deploymentParamsJson,
		RequestJson:              requestJson,
		ResultJson:               "",
		Repo:                     s.ModelDeploymentRepo,
	}
	if err := s.ModelDeploymentRepo.Create(modelDeployment); err != nil {
		return nil, err
	}

	// deploy it!
	if err := modelDeployment.Deploy(entity.ModelDeploymentContext{
		Model:           model,
		FATEFlowContext: request.FATEFlowContext,
	}); err != nil {
		modelDeployment.Status = entity.ModelDeploymentStatusFailed
		if updateErr := s.ModelDeploymentRepo.UpdateStatusByUUID(modelDeployment); updateErr != nil {
			log.Err(updateErr).Msg("failed to update deployment status")
		}
		return nil, err
	}
	return modelDeployment, nil
}

// GetSupportedDeploymentType returns a list of entity.ModelDeploymentType that the specified model can be deployed to
func (s *ModelService) GetSupportedDeploymentType(modelUUID string) ([]entity.ModelDeploymentType, error) {
	model, err := s.loadModel(modelUUID)
	if err != nil {
		return nil, err
	}
	switch model.ComponentAlgorithmType {
	case entity.ComponentAlgorithmTypeHomoSBT, entity.ComponentAlgorithmTypeHomoLR:
		return []entity.ModelDeploymentType{entity.ModelDeploymentTypeKFServing}, nil
	default:
		return nil, errors.Errorf("unsupported component type: %v", model.ComponentAlgorithmType)
	}
}

func (s *ModelService) loadModel(modelUUID string) (*entity.Model, error) {
	modelInstance, err := s.ModelRepo.GetByUUID(modelUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query model")
	}
	model := modelInstance.(*entity.Model)
	model.Repo = s.ModelRepo
	return model, nil
}
