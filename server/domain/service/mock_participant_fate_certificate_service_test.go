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

package service

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
)

type mockParticipantFATECertificateServiceInt struct {
	// TODO add stubs
}

func (m *mockParticipantFATECertificateServiceInt) DefaultCA() (*entity.CertificateAuthority, error) {
	return &entity.CertificateAuthority{
		Type:              1,
		ConfigurationJSON: `{"service_url":"https://127.0.0.1","service_cert_pem":"-----BEGIN CERTIFICATE-----\nMIIBpDCCAUmgAwIBAgIQYOQRhYETGs+ywmloTXJZ3TAKBggqhkjOPQQDAjAwMRIw\nEAYDVQQKEwlTbWFsbHN0ZXAxGjAYBgNVBAMTEVNtYWxsc3RlcCBSb290IENBMB4X\nDTIyMDExMTA5MTMwNloXDTMyMDEwOTA5MTMwNlowMDESMBAGA1UEChMJU21hbGxz\ndGVwMRowGAYDVQQDExFTbWFsbHN0ZXAgUm9vdCBDQTBZMBMGByqGSM49AgEGCCqG\nSM49AwEHA0IABL2V5CItQTxwRzoo1pZtSQ4GeT5VzTymhv/YRJtNjjQO9PrGxA1f\nedxZg2Z/VR4imTbQafFRUDCc35PpgojXP0+jRTBDMA4GA1UdDwEB/wQEAwIBBjAS\nBgNVHRMBAf8ECDAGAQH/AgEBMB0GA1UdDgQWBBQfGCQ3WZyp9lTtZRCKBVtN/oFx\nbzAKBggqhkjOPQQDAgNJADBGAiEAr8oLkTo+Nu2gyZ9NQXQslueRSIfI2ob9A7D7\nOsrGFYYCIQCRR6APcfdvzacQ52Z8iXIpRmvYxn0tBh1VZE5y8dI09Q==\n-----END CERTIFICATE-----","provisioner_name":"admin","provisioner_password":"5nPGyRJjFPzBCVf31Oyk5NCKiTe6OwCmtDB0ZHW7"}`,
	}, nil
}

func (m *mockParticipantFATECertificateServiceInt) CreateCertificateSimple(string, time.Duration, []string) (cert *entity.Certificate, pk *rsa.PrivateKey, err error) {
	cert = &entity.Certificate{
		UUID:            "",
		Name:            "",
		SerialNumberStr: "",
		PEM:             "",
		Certificate: &x509.Certificate{
			SerialNumber: &big.Int{},
			Subject: pkix.Name{
				SerialNumber: "",
				CommonName:   "",
			},
			DNSNames: []string{},
		},
	}
	pk = &rsa.PrivateKey{}
	return
}

func (m *mockParticipantFATECertificateServiceInt) CreateBinding(*entity.Certificate, entity.CertificateBindingServiceType, string, string, entity.FederationType) error {
	return nil
}

func (m *mockParticipantFATECertificateServiceInt) RemoveBinding(string) error {
	return nil
}

var _ ParticipantCertificateServiceInt = (*mockParticipantFATECertificateServiceInt)(nil)
