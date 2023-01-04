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
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ParticipantFATE represent a FATE type participant
type ParticipantFATE struct {
	Participant
	PartyID     int
	Type        ParticipantFATEType
	Status      ParticipantFATEStatus
	CertConfig  ParticipantFATECertConfig       `gorm:"type:text"`
	AccessInfo  ParticipantFATEModulesAccessMap `gorm:"type:text"`
	IngressInfo ParticipantFATEIngressMap       `gorm:"type:text"`
}

// GetSitePortalAdminPassword returns the admin password of the deployed site portal service
func (p ParticipantFATE) GetSitePortalAdminPassword() (string, error) {
	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(p.DeploymentYAML), &m)
	if err != nil {
		return "", errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}
	password := "admin"
	if serverConfigInt, ok := m["sitePortalServer"]; ok {
		serverConfig := serverConfigInt.(map[string]interface{})
		if passwordInt, ok := serverConfig["adminPassword"]; ok {
			password = passwordInt.(string)
		}
	}
	return password, err
}

// ParticipantFATECertConfig contains all the certificate configuration of a FATE participant
type ParticipantFATECertConfig struct {
	ProxyServerCertInfo      ParticipantComponentCertInfo `json:"proxy_server_cert_info"`
	FMLManagerServerCertInfo ParticipantComponentCertInfo `json:"fml_manager_server_cert_info"`
	FMLManagerClientCertInfo ParticipantComponentCertInfo `json:"fml_manager_client_cert_info"`

	PulsarServerCertInfo     ParticipantComponentCertInfo `json:"pulsar_server_cert_info"`
	SitePortalServerCertInfo ParticipantComponentCertInfo `json:"site_portal_server_cert_info"`
	SitePortalClientCertInfo ParticipantComponentCertInfo `json:"site_portal_client_cert_info"`
}

func (c ParticipantFATECertConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *ParticipantFATECertConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// ParticipantFATEServiceName is all the exposed FATE participant service name
type ParticipantFATEServiceName string

const (
	ParticipantFATEServiceNameNginx  ParticipantFATEServiceName = "nginx"
	ParticipantFATEServiceNameATS    ParticipantFATEServiceName = "traffic-server"
	ParticipantFATEServiceNamePulsar ParticipantFATEServiceName = "pulsar-public-tls"
	ParticipantFATEServiceNamePortal ParticipantFATEServiceName = "frontend"
	ParticipantFATEServiceNameFMLMgr ParticipantFATEServiceName = "fml-manager-server"
)

const (
	ParticipantFATESecretNameATS    = "traffic-server-cert"
	ParticipantFATESecretNamePulsar = "pulsar-cert"
	ParticipantFATESecretNameFMLMgr = "fml-manager-cert"
	ParticipantFATESecretNamePortal = "site-portal-cert"
)

// ParticipantFATEModulesAccessMap contains the exposed services access information
type ParticipantFATEModulesAccessMap map[ParticipantFATEServiceName]ParticipantModulesAccess

func (c ParticipantFATEModulesAccessMap) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *ParticipantFATEModulesAccessMap) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// ParticipantFATEType is the participant type
type ParticipantFATEType uint8

const (
	ParticipantFATETypeUnknown ParticipantFATEType = iota
	ParticipantFATETypeExchange
	ParticipantFATETypeCluster
)

func (t ParticipantFATEType) String() string {
	switch t {
	case ParticipantFATETypeExchange:
		return "exchange"
	case ParticipantFATETypeCluster:
		return "cluster"
	}
	return "unknown"
}

// ParticipantFATEStatus is the status of the fate participant
type ParticipantFATEStatus uint8

const (
	ParticipantFATEStatusUnknown ParticipantFATEStatus = iota
	ParticipantFATEStatusActive
	ParticipantFATEStatusInstalling
	ParticipantFATEStatusRemoving
	ParticipantFATEStatusReconfiguring
	ParticipantFATEStatusFailed
)

func (t ParticipantFATEStatus) String() string {
	switch t {
	case ParticipantFATEStatusActive:
		return "Active"
	case ParticipantFATEStatusInstalling:
		return "Installing"
	case ParticipantFATEStatusRemoving:
		return "Removing"
	case ParticipantFATEStatusReconfiguring:
		return "Reconfiguring"
	case ParticipantFATEStatusFailed:
		return "Failed"
	}
	return "Unknown"
}

type ParticipantFATEIngressMap map[string]ParticipantFATEIngress

func (c ParticipantFATEIngressMap) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *ParticipantFATEIngressMap) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// ParticipantFATEIngress contains ingress info of a FATE participant module
type ParticipantFATEIngress struct {
	Hosts     []string `json:"hosts"`
	Addresses []string `json:"addresses"`
	TLS       bool     `json:"tls"`
}
