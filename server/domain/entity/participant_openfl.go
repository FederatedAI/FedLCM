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

	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
)

const (
	ParticipantOpenFLSecretNameDirector = "director-cert"
	ParticipantOpenFLSecretNameJupyter  = "notebook-cert"
	ParticipantOpenFLSecretNameEnvoy    = "envoy-cert"
)

// ParticipantOpenFL represent an OpenFL type participant
type ParticipantOpenFL struct {
	Participant
	Type       ParticipantOpenFLType
	Status     ParticipantOpenFLStatus
	InfraUUID  string                            `gorm:"type:varchar(36)"`
	TokenUUID  string                            `gorm:"type:varchar(36)"`
	CertConfig ParticipantOpenFLCertConfig       `gorm:"type:text"`
	AccessInfo ParticipantOpenFLModulesAccessMap `gorm:"type:text"`
	Labels     valueobject.Labels                `gorm:"type:text"`
}

// ParticipantOpenFLType is the openfl participant type
type ParticipantOpenFLType uint8

const (
	ParticipantOpenFLTypeUnknown ParticipantOpenFLType = iota
	ParticipantOpenFLTypeDirector
	ParticipantOpenFLTypeEnvoy
)

func (t ParticipantOpenFLType) String() string {
	switch t {
	case ParticipantOpenFLTypeDirector:
		return "director"
	case ParticipantOpenFLTypeEnvoy:
		return "envoy"
	}
	return "unknown"
}

// ParticipantOpenFLStatus is the status of the openfl participant
type ParticipantOpenFLStatus uint8

const (
	ParticipantOpenFLStatusUnknown ParticipantOpenFLStatus = iota
	ParticipantOpenFLStatusActive
	ParticipantOpenFLStatusRemoving
	ParticipantOpenFLStatusFailed
	ParticipantOpenFLStatusInstallingDirector
	ParticipantOpenFLStatusConfiguringInfra
	ParticipantOpenFLStatusInstallingEndpoint
	ParticipantOpenFLStatusInstallingEnvoy
)

func (t ParticipantOpenFLStatus) String() string {
	switch t {
	case ParticipantOpenFLStatusActive:
		return "Active"
	case ParticipantOpenFLStatusRemoving:
		return "Removing"
	case ParticipantOpenFLStatusFailed:
		return "Failed"
	case ParticipantOpenFLStatusInstallingDirector:
		return "Installing Director"
	case ParticipantOpenFLStatusConfiguringInfra:
		return "Configuring Infrastructure"
	case ParticipantOpenFLStatusInstallingEndpoint:
		return "Installing Endpoint"
	case ParticipantOpenFLStatusInstallingEnvoy:
		return "Installing Envoy"
	}
	return "Unknown"
}

// ParticipantOpenFLCertConfig contains configurations for certificates in an OpenFL participant
type ParticipantOpenFLCertConfig struct {
	DirectorServerCertInfo ParticipantComponentCertInfo `json:"director_server_cert_info"`
	JupyterClientCertInfo  ParticipantComponentCertInfo `json:"jupyter_client_cert_info"`

	EnvoyClientCertInfo ParticipantComponentCertInfo `json:"envoy_client_cert_info"`
}

func (c ParticipantOpenFLCertConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *ParticipantOpenFLCertConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// ParticipantOpenFLServiceName is a enum of the exposed service names
type ParticipantOpenFLServiceName string

const (
	ParticipantOpenFLServiceNameDirector   ParticipantOpenFLServiceName = "director"
	ParticipantOpenFLServiceNameAggregator ParticipantOpenFLServiceName = "agg"
	ParticipantOpenFLServiceNameJupyter    ParticipantOpenFLServiceName = "notebook"
)

// ParticipantOpenFLModulesAccessMap contains the exposed services in an OpenFL participant
type ParticipantOpenFLModulesAccessMap map[ParticipantOpenFLServiceName]ParticipantModulesAccess

func (c ParticipantOpenFLModulesAccessMap) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *ParticipantOpenFLModulesAccessMap) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}
