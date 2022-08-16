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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/smallstep/certificates/cas/apiv1"
)

// CertificateService provides functions to work with certificate related workflows
type CertificateService struct {
	CertificateAuthorityRepo repo.CertificateAuthorityRepository
	CertificateRepo          repo.CertificateRepository
	CertificateBindingRepo   repo.CertificateBindingRepository
}

// CreateCertificateSimple just take the CN, lifetime and dnsNames and will give the automatically generated private key and the certificate using the first available CA
func (s *CertificateService) CreateCertificateSimple(commonName string, lifetime time.Duration, dnsNames []string) (cert *entity.Certificate, pk *rsa.PrivateKey, err error) {

	certificateAuthority, err := s.DefaultCA()
	if err != nil {
		return
	}
	caClient, err := certificateAuthority.Client()
	if err != nil {
		return
	}
	pk, err = rsa.GenerateKey(rand.Reader, 2048)
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: commonName,
		},
		DNSNames: dnsNames,
	}, pk)
	if err != nil {
		return
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return
	}

	resp, err := caClient.CreateCertificate(&apiv1.CreateCertificateRequest{
		CSR:      csr,
		Lifetime: lifetime,
	})
	if err != nil {
		return
	}
	cert = &entity.Certificate{
		UUID:        uuid.NewV4().String(),
		Name:        fmt.Sprintf("Certificate for %s", commonName),
		Certificate: resp.Certificate,
		Chain:       resp.CertificateChain,
	}

	if saveErr := s.CertificateRepo.Create(cert); saveErr != nil {
		err = errors.Wrap(err, "failed to save certificate")
	}
	return
}

// DefaultCA returns the default CA info
func (s *CertificateService) DefaultCA() (*entity.CertificateAuthority, error) {
	instance, err := s.CertificateAuthorityRepo.GetFirst()
	if err != nil {
		return nil, err
	}
	certificateAuthority := instance.(*entity.CertificateAuthority)
	return certificateAuthority, nil
}

// CreateBinding create a binding record of a certificate and a participant
func (s *CertificateService) CreateBinding(cert *entity.Certificate,
	serviceType entity.CertificateBindingServiceType,
	participantUUID string,
	federationUUID string,
	federationType entity.FederationType) error {
	return s.CertificateBindingRepo.Create(&entity.CertificateBinding{
		UUID:            uuid.NewV4().String(),
		CertificateUUID: cert.UUID,
		ParticipantUUID: participantUUID,
		ServiceType:     serviceType,
		FederationUUID:  federationUUID,
		FederationType:  federationType,
	})
}

// RemoveBinding deletes a bindings record and deletes the certificate if there is no bindings for it
func (s *CertificateService) RemoveBinding(participantUUID string) error {
	instanceList, err := s.CertificateBindingRepo.ListByParticipantUUID(participantUUID)
	if err != nil {
		return err
	}
	bindingList := instanceList.([]entity.CertificateBinding)
	certificateUUIDSet := map[string]interface{}{}
	for _, binding := range bindingList {
		certificateUUIDSet[binding.CertificateUUID] = nil
	}
	if err := s.CertificateBindingRepo.DeleteByParticipantUUID(participantUUID); err != nil {
		return errors.Wrapf(err, "failed to delete bindings")
	}
	for certificateUUID := range certificateUUIDSet {
		instanceList, err = s.CertificateBindingRepo.ListByCertificateUUID(certificateUUID)
		if err != nil {
			return errors.Wrapf(err, "failed to query bindings")
		}
		if instanceList != nil {
			bindingList = instanceList.([]entity.CertificateBinding)
			if len(bindingList) != 0 {
				continue
			}
		}
		log.Info().Str("certificate uuid", certificateUUID).Msgf("removing unused certificate")
		if err := s.CertificateRepo.DeleteByUUID(certificateUUID); err != nil {
			log.Err(err).Msgf("failed to remove certificate %s", certificateUUID)
		}
	}
	return nil
}
