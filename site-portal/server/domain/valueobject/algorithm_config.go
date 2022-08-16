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
	"database/sql/driver"
	"encoding/json"
)

// AlgorithmConfig contains some configurations for running a training job
type AlgorithmConfig struct {
	TrainingValidationEnabled     bool     `json:"training_validation_enabled"`
	TrainingValidationSizePercent uint     `json:"training_validation_percent"`
	TrainingComponentsToDeploy    []string `json:"training_component_list_to_deploy"`
}

func (c AlgorithmConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *AlgorithmConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}
