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

package valueobject

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// KFServingParameter is the parameters used for deploying to KFServing
type KFServingParameter struct {
	ProtocolVersion         string `json:"protocol_version"`
	KubeconfigContent       string `json:"config_file_content"`
	Namespace               string `json:"namespace"`
	Replace                 bool   `json:"replace"`
	SKipCreateStorageSecret bool   `json:"skip_create_storage_secret"`
	ModelStorageType        string `json:"model_storage_type"`
}

// KFServingWithMinIOParameter is the parameters used for deploying to KFServing
type KFServingWithMinIOParameter struct {
	KFServingParameter
	ModelStorageParameters MinIOStorageMParameters `json:"model_storage_parameters"`
}

// MinIOStorageMParameters is the parameters for the minio storage
type MinIOStorageMParameters struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Secure    bool   `json:"secure"`
	Region    string `json:"region"`
}

// GetKFServingDeploymentParametersJson returns a deployment parameter json
func GetKFServingDeploymentParametersJson(userParametersJson string, kubeflowConfig KubeflowConfig) (string, error) {
	var userSpecifiedParam KFServingParameter
	if userParametersJson != "" {
		if err := json.Unmarshal([]byte(userParametersJson), &userSpecifiedParam); err != nil {
			return "", err
		}
		if userSpecifiedParam.ModelStorageType != "" && userSpecifiedParam.ModelStorageType != "minio" {
			return "", errors.Errorf("not supported storage type: %s", userSpecifiedParam.ModelStorageType)
		}
	}

	// the default param
	defaultParam := &KFServingWithMinIOParameter{
		KFServingParameter: KFServingParameter{
			ProtocolVersion:         "v1",
			KubeconfigContent:       kubeflowConfig.KubeConfig,
			Namespace:               "default",
			Replace:                 false,
			SKipCreateStorageSecret: false,
			ModelStorageType:        "minio",
		},
		ModelStorageParameters: MinIOStorageMParameters{
			Endpoint:  kubeflowConfig.MinIOEndpoint,
			AccessKey: kubeflowConfig.MinIOAccessKey,
			SecretKey: kubeflowConfig.MinIOSecretKey,
			Secure:    kubeflowConfig.MinIOSSLEnabled,
			Region:    kubeflowConfig.MinIORegion,
		},
	}

	// merge the user input
	mergedJson, err := json.Marshal(defaultParam)
	if err != nil {
		return "", err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(mergedJson, &m); err != nil {
		return "", err
	}
	if userParametersJson != "" {
		if err := json.Unmarshal([]byte(userParametersJson), &m); err != nil {
			return "", errors.Wrapf(err, "failed to merge the user specified params")
		}
		mergedJson, err = json.MarshalIndent(m, "", " ")
		if err != nil {
			return "", err
		}
	}
	return string(mergedJson), nil
}
