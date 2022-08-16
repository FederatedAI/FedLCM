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

import "gorm.io/gorm"

// CertificateBinding is the binding relationship between certificate and the service
type CertificateBinding struct {
	gorm.Model
	UUID            string `gorm:"type:varchar(36);index;unique"`
	CertificateUUID string `gorm:"type:varchar(36);not null"`
	ParticipantUUID string `gorm:"type:varchar(36);not null"`
	ServiceType     CertificateBindingServiceType
	FederationUUID  string         `gorm:"type:varchar(36)"`
	FederationType  FederationType `gorm:"type:varchar(255)"`
}

type CertificateBindingServiceType uint8

const (
	CertificateBindingServiceTypeUnknown CertificateBindingServiceType = iota
	CertificateBindingServiceTypeATS
	CertificateBindingServiceTypePulsarServer
	CertificateBindingServiceFMLManagerServer
	CertificateBindingServiceFMLManagerClient
	CertificateBindingServiceSitePortalServer
	CertificateBindingServiceSitePortalClient
)

// openfl
const (
	CertificateBindingServiceTypeOpenFLDirector CertificateBindingServiceType = iota + 101
	CertificateBindingServiceTypeOpenFLJupyter
	CertificateBindingServiceTypeOpenFLEnvoy
)

func (t CertificateBindingServiceType) String() string {
	switch t {
	case CertificateBindingServiceTypeATS:
		return "pulsar proxy"
	case CertificateBindingServiceTypePulsarServer:
		return "pulsar server"
	case CertificateBindingServiceFMLManagerServer:
		return "fml manager server"
	case CertificateBindingServiceFMLManagerClient:
		return "fml manager client"
	case CertificateBindingServiceSitePortalServer:
		return "site portal server"
	case CertificateBindingServiceSitePortalClient:
		return "site portal client"
	case CertificateBindingServiceTypeOpenFLDirector:
		return "openfl director"
	case CertificateBindingServiceTypeOpenFLJupyter:
		return "openfl jupyter client"
	}
	return "unknown"
}
