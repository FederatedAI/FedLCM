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
	"database/sql/driver"
	"encoding/json"
)

// EndpointKubeFATE is an KubeFATE type of endpoint
type EndpointKubeFATE struct {
	EndpointBase
	Config                KubeFATEConfig `gorm:"type:text"`
	DeploymentYAML        string         `gorm:"type:text"`
	IngressControllerYAML string         `gorm:"type:text"`
}

// KubeFATEConfig records basic info of a KubeFATE config
type KubeFATEConfig struct {
	IngressAddress    string `json:"ingress_address"`
	IngressRuleHost   string `json:"ingress_rule_host"`
	UsePortForwarding bool   `json:"use_port_forwarding"`
}

func (c KubeFATEConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *KubeFATEConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// EndpointKubeFATEIngressControllerServiceMode is the service mode of the ingress controller
type EndpointKubeFATEIngressControllerServiceMode uint8

const (
	// EndpointKubeFATEIngressControllerServiceModeSkip means there is an ingress controller in the infra, and we skip installing it by ourselves
	EndpointKubeFATEIngressControllerServiceModeSkip EndpointKubeFATEIngressControllerServiceMode = iota
	EndpointKubeFATEIngressControllerServiceModeLoadBalancer
	EndpointKubeFATEIngressControllerServiceModeModeNodePort
	// EndpointKubeFATEIngressControllerServiceModeModeNonexistent means there is no ingress controller at all, we will use other method to access the KubeFATE service
	EndpointKubeFATEIngressControllerServiceModeModeNonexistent
)
