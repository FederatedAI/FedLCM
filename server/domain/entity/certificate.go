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
	"crypto/x509"
	"encoding/pem"
	"gorm.io/gorm"
)

// Certificate is the certificate managed by this service
type Certificate struct {
	gorm.Model
	UUID              string              `gorm:"type:varchar(36);index;unique"`
	Name              string              `gorm:"type:varchar(255);not null"`
	SerialNumberStr   string              `gorm:"type:varchar(255)"`
	PEM               string              `gorm:"type:text"`
	ChainPEM          string              `gorm:"type:text"`
	Chain             []*x509.Certificate `gorm:"-"`
	*x509.Certificate `gorm:"-"`
}

func (c *Certificate) BeforeSave(tx *gorm.DB) error {
	c.SerialNumberStr = c.SerialNumber.String()
	_, err := c.EncodePEM()
	return err
}

func (c *Certificate) AfterFind(tx *gorm.DB) error {
	b, _ := pem.Decode([]byte(c.PEM))
	certificate, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return err
	}
	c.Certificate = certificate
	c.Chain = nil
	rest := []byte(c.ChainPEM)
	var block *pem.Block
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			chainCert, err := x509.ParseCertificate(b.Bytes)
			if err != nil {
				return err
			}
			c.Chain = append(c.Chain, chainCert)
		}
		if len(rest) == 0 {
			break
		}
	}
	return nil
}

// EncodePEM saves the PEM content in the related fields and returns the complete PEM content
func (c *Certificate) EncodePEM() ([]byte, error) {
	if c.PEM == "" {
		c.PEM = string(pem.EncodeToMemory(&pem.Block{
			Type:    "CERTIFICATE",
			Headers: nil,
			Bytes:   c.Raw,
		}))
	}
	certBytes := []byte(c.PEM)
	if c.ChainPEM == "" && len(c.Chain) > 0 {
		for _, chainCert := range c.Chain {
			c.ChainPEM += string(pem.EncodeToMemory(&pem.Block{
				Type:    "CERTIFICATE",
				Headers: nil,
				Bytes:   chainCert.Raw,
			}))
		}
	}
	certBytes = append(certBytes, c.ChainPEM...)
	return certBytes, nil
}
