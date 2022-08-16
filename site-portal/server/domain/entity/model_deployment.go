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
	"encoding/json"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ModelDeployment represents a deployment operation for a model
type ModelDeployment struct {
	gorm.Model
	UUID                     string                         `json:"uuid" gorm:"type:varchar(36)"`
	ServiceName              string                         `json:"service_name" gorm:"type:varchar(255)"`
	ModelUUID                string                         `json:"model_uuid" gorm:"type:varchar(36)"`
	Type                     ModelDeploymentType            `json:"type"`
	Status                   ModelDeploymentStatus          `json:"status"`
	DeploymentParametersJson string                         `json:"parameters_json" gorm:"type:text"`
	RequestJson              string                         `json:"request_json" gorm:"type:text"`
	ResultJson               string                         `json:"result_json" gorm:"type:text"`
	Repo                     repo.ModelDeploymentRepository `json:"-" gorm:"-"`
}

// ModelDeploymentType is a enum of the types of the target deployment runtime
// We use uint instead of uint8 here because json marshalling will convert []uint8 slice to base64-encoded string
type ModelDeploymentType uint

const (
	ModelDeploymentTypeUnknown ModelDeploymentType = iota
	ModelDeploymentTypeKFServing
)

func (t ModelDeploymentType) String() string {
	switch t {
	case ModelDeploymentTypeKFServing:
		return "kfserving"
	default:
		return "unknown"
	}
}

// ModelDeploymentStatus is a enum of the status of the deployment action
type ModelDeploymentStatus uint8

const (
	ModelDeploymentStatusUnknown ModelDeploymentStatus = iota
	ModelDeploymentStatusCreated
	ModelDeploymentStatusFailed
	ModelDeploymentStatusSucceeded
)

// ModelDeploymentContext contains the context needed to perform a deployment action
type ModelDeploymentContext struct {
	Model              *Model
	FATEFlowContext    FATEFlowContext
	KubeflowConfig     valueobject.KubeflowConfig
	UserParametersJson string
}

var minHomoDeploymentFATEVersion = func() *version.Version {
	version, _ := version.NewVersion("1.7.0")
	return version
}()

// Deploy deploys the model to FATE
func (d *ModelDeployment) Deploy(context ModelDeploymentContext) error {
	fateClient := fateclient.NewFATEFlowClient(context.FATEFlowContext.FATEFlowHost, context.FATEFlowContext.FATEFlowPort, context.FATEFlowContext.FATEFlowIsHttps)
	versionStr, err := fateClient.GetFATEVersion()
	if err != nil {
		return err
	}
	currentVersion, err := version.NewVersion(versionStr)
	if err != nil {
		return err
	}
	if currentVersion.LessThan(minHomoDeploymentFATEVersion) {
		return errors.Errorf("current FATE version (%s) is lower than the supportted version (%s)", currentVersion.String(), minHomoDeploymentFATEVersion.String())
	}

	basicModelInfo := fateclient.HomoModelConversionRequest{
		ModelInfo: fateclient.ModelInfo{
			ModelID:      context.Model.FATEModelID,
			ModelVersion: context.Model.FATEModelVersion,
		},
		PartyID: context.Model.PartyID,
		Role:    context.Model.Role,
	}
	if err := fateClient.ConvertHomoModel(basicModelInfo); err != nil {
		return errors.Wrapf(err, "failed to convert model")
	}

	var parameterInstance map[string]interface{}
	if err := json.Unmarshal([]byte(d.DeploymentParametersJson), &parameterInstance); err != nil {
		return err
	}
	d.ResultJson, err = fateClient.DeployHomoModel(fateclient.HomoModelDeploymentRequest{
		HomoModelConversionRequest: basicModelInfo,
		ServiceID:                  d.ServiceName,
		ComponentName:              context.Model.ComponentName,
		DeploymentType:             d.Type.String(),
		DeploymentParameters:       parameterInstance,
	})
	if err != nil {
		d.Status = ModelDeploymentStatusFailed
	} else {
		d.Status = ModelDeploymentStatusSucceeded
	}
	if err := d.Repo.UpdateStatusByUUID(d); err != nil {
		log.Err(err).Msg("failed to update deployment status")
	}
	if err := d.Repo.UpdateResultJsonByUUID(d); err != nil {
		log.Err(err).Msg("failed to update deployment result json")
	}
	if err != nil {
		return errors.Wrapf(err, "failed to deploy model")
	}
	return nil
}
