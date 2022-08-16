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
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/pkg/errors"
	stepcaapiv1 "github.com/smallstep/certificates/cas/apiv1"
	"github.com/smallstep/certificates/cas/stepcas"
	"gorm.io/gorm"
)

// CertificateAuthorityClient provides interface functions to work with a CA service
type CertificateAuthorityClient interface {
	stepcaapiv1.CertificateAuthorityService
}

// CertificateAuthority represent a certificate authority
type CertificateAuthority struct {
	gorm.Model
	UUID              string `gorm:"type:varchar(36);index;unique"`
	Name              string `gorm:"type:varchar(255);not null"`
	Description       string `gorm:"type:text"`
	Type              CertificateAuthorityType
	ConfigurationJSON string `gorm:"type:text;column:config_json"`
}

// CertificateAuthorityType is the certificate authority type
type CertificateAuthorityType uint8

const (
	CertificateAuthorityTypeUnknown CertificateAuthorityType = iota
	CertificateAuthorityTypeStepCA
)

// CertificateAuthorityConfigurationStepCA contains basic configuration for a StepCA service
type CertificateAuthorityConfigurationStepCA struct {
	ServiceURL            string `json:"service_url" mapstructure:"service_url"`
	ServiceCertificatePEM string `json:"service_cert_pem" mapstructure:"service_cert_pem"`
	ProvisionerName       string `json:"provisioner_name" mapstructure:"provisioner_name"`
	ProvisionerPassword   string `json:"provisioner_password" mapstructure:"provisioner_password"`
}

// Client returns a client to work with a CA service, with basic validation executed
func (ca *CertificateAuthority) Client() (CertificateAuthorityClient, error) {
	switch ca.Type {
	case CertificateAuthorityTypeStepCA:
		var config CertificateAuthorityConfigurationStepCA
		err := json.Unmarshal([]byte(ca.ConfigurationJSON), &config)
		if err != nil {
			return nil, err
		}
		b, _ := pem.Decode([]byte(config.ServiceCertificatePEM))
		if b == nil {
			return nil, errors.Errorf("failed to decode PEM block")
		}
		return stepcas.New(context.TODO(), stepcaapiv1.Options{
			CertificateAuthority:            config.ServiceURL,
			CertificateAuthorityFingerprint: fmt.Sprintf("%x", sha256.Sum256(b.Bytes)),
			CertificateIssuer: &stepcaapiv1.CertificateIssuer{
				Type:        "jwk",
				Provisioner: config.ProvisionerName,
				Password:    config.ProvisionerPassword,
			},
		})
	}
	return nil, errors.Errorf("unknown CA type: %v", ca.Type)
}

func (ca *CertificateAuthority) RootCert() (*x509.Certificate, error) {
	var config CertificateAuthorityConfigurationStepCA
	err := json.Unmarshal([]byte(ca.ConfigurationJSON), &config)
	if err != nil {
		return nil, err
	}
	b, _ := pem.Decode([]byte(config.ServiceCertificatePEM))
	return x509.ParseCertificate(b.Bytes)
}
