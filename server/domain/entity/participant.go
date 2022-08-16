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

	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
)

// Participant represent a federation participant
type Participant struct {
	gorm.Model
	UUID           string                    `gorm:"type:varchar(36);index;unique"`
	Name           string                    `gorm:"type:varchar(255);not null"`
	Description    string                    `gorm:"type:text"`
	FederationUUID string                    `gorm:"type:varchar(36)"`
	EndpointUUID   string                    `gorm:"type:varchar(36)"`
	ChartUUID      string                    `gorm:"type:varchar(36)"`
	Namespace      string                    `gorm:"type:varchar(255)"`
	ClusterUUID    string                    `gorm:"type:varchar(36)"`
	JobUUID        string                    `gorm:"type:varchar(36)"`
	DeploymentYAML string                    `gorm:"type:text"`
	IsManaged      bool                      `gorm:"type:bool"`
	ExtraAttribute ParticipantExtraAttribute `gorm:"type:text"`
}

// ParticipantExtraAttribute record some extra attributes of the participant
type ParticipantExtraAttribute struct {
	IsNewNamespace    bool `json:"is_new_namespace"`
	UseRegistrySecret bool `json:"use_registry_secret"`
}

func (a ParticipantExtraAttribute) Value() (driver.Value, error) {
	bJson, err := json.Marshal(a)
	return bJson, err
}

func (a *ParticipantExtraAttribute) Scan(v interface{}) error {
	// ignore any errors
	_ = json.Unmarshal([]byte(v.(string)), a)
	return nil
}

// ParticipantDefaultServiceType is the default service type of the exposed services in the participant
type ParticipantDefaultServiceType uint8

const (
	ParticipantDefaultServiceTypeUnknown ParticipantDefaultServiceType = iota
	ParticipantDefaultServiceTypeLoadBalancer
	ParticipantDefaultServiceTypeNodePort
)

func (t ParticipantDefaultServiceType) String() string {
	switch t {
	case ParticipantDefaultServiceTypeNodePort:
		return "NodePort"
	case ParticipantDefaultServiceTypeLoadBalancer:
		return "LoadBalancer"
	}
	return "Unknown"
}

// ParticipantComponentCertInfo contains certificate information of a component in participant
type ParticipantComponentCertInfo struct {
	BindingMode ParticipantCertBindingMode `json:"binding_mode"`
	UUID        string                     `json:"uuid"`
	CommonName  string                     `json:"common_name"`
}

// ParticipantCertBindingMode is the certificate binding mode
type ParticipantCertBindingMode uint8

const (
	CertBindingModeUnknown ParticipantCertBindingMode = iota
	CertBindingModeSkip
	CertBindingModeReuse
	CertBindingModeCreate
)

// ParticipantModulesAccess contains access info of a participant service
type ParticipantModulesAccess struct {
	ServiceType corev1.ServiceType `json:"service_type"`
	Host        string             `json:"host"`
	Port        int                `json:"port"`
	TLS         bool               `json:"tls"`
	FQDN        string             `json:"fqdn"`
}
