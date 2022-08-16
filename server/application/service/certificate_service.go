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
	"fmt"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/pkg/errors"
)

// CertificateApp provides functions to manage the certificates
type CertificateApp struct {
	CertificateAuthorityRepo repo.CertificateAuthorityRepository
	CertificateRepo          repo.CertificateRepository
	CertificateBindingRepo   repo.CertificateBindingRepository
	ParticipantFATERepo      repo.ParticipantFATERepository
	ParticipantOpenFLRepo    repo.ParticipantOpenFLRepository
	FederationFATERepo       repo.FederationRepository
	FederationOpenFLRepo     repo.FederationRepository
}

// CertificateListItem contains basic info of a certificate
type CertificateListItem struct {
	Name           string                       `json:"name"`
	UUID           string                       `json:"uuid"`
	SerialNumber   string                       `json:"serial_number"`
	ExpirationDate time.Time                    `json:"expiration_date"`
	CommonName     string                       `json:"common_name"`
	Bindings       []CertificateBindingListItem `json:"bindings"`
}

// CertificateBindingListItem contains binding information of a certificate
type CertificateBindingListItem struct {
	ServiceType        entity.CertificateBindingServiceType `json:"service_type"`
	ParticipantUUID    string                               `json:"participant_uuid"`
	ParticipantName    string                               `json:"participant_name"`
	ParticipantType    string                               `json:"participant_type"`
	ServiceDescription string                               `json:"service_description"`
	FederationUUID     string                               `json:"federation_uuid"`
	FederationName     string                               `json:"federation_name"`
	FederationType     entity.FederationType                `json:"federation_type"`
}

// List returns the currently managed certificates
func (app *CertificateApp) List() (certificateList []CertificateListItem, err error) {
	instanceList, err := app.CertificateRepo.List()
	if err != nil {
		return
	}
	domainCertList := instanceList.([]entity.Certificate)
	for _, domainCert := range domainCertList {
		cert := CertificateListItem{
			Name:           domainCert.Name,
			UUID:           domainCert.UUID,
			SerialNumber:   domainCert.SerialNumber.String(),
			ExpirationDate: domainCert.NotAfter,
			CommonName:     domainCert.Subject.CommonName,
			Bindings:       []CertificateBindingListItem{},
		}
		instanceList, err = app.CertificateBindingRepo.ListByCertificateUUID(domainCert.UUID)
		if err != nil {
			return
		}
		bindingList := instanceList.([]entity.CertificateBinding)
		for _, domainBinding := range bindingList {
			binding := CertificateBindingListItem{
				ServiceType:        domainBinding.ServiceType,
				ParticipantUUID:    domainBinding.ParticipantUUID,
				ParticipantName:    "Unknown",
				ServiceDescription: "Unknown",
				FederationUUID:     domainBinding.FederationUUID,
				FederationType:     domainBinding.FederationType,
			}
			switch domainBinding.FederationType {
			case entity.FederationTypeFATE:
				instance, err := app.ParticipantFATERepo.GetByUUID(binding.ParticipantUUID)
				if err == nil {
					participant := instance.(*entity.ParticipantFATE)
					binding.ParticipantName = participant.Name
					binding.ParticipantType = participant.Type.String()
					binding.ServiceDescription = getBindingFATEServiceDescription(domainBinding.ServiceType, participant)
				}
				federationInstance, err := app.FederationFATERepo.GetByUUID(domainBinding.FederationUUID)
				if err == nil {
					federation := federationInstance.(*entity.FederationFATE)
					binding.FederationName = federation.Name
				}
			case entity.FederationTypeOpenFL:
				instance, err := app.ParticipantOpenFLRepo.GetByUUID(binding.ParticipantUUID)
				if err == nil {
					participant := instance.(*entity.ParticipantOpenFL)
					binding.ParticipantName = participant.Name
					binding.ParticipantType = participant.Type.String()
					binding.ServiceDescription = getBindingOpenFLServiceDescription(domainBinding.ServiceType, participant)
				}
				federationInstance, err := app.FederationOpenFLRepo.GetByUUID(domainBinding.FederationUUID)
				if err == nil {
					federation := federationInstance.(*entity.FederationOpenFL)
					binding.FederationName = federation.Name
				}
			}
			cert.Bindings = append(cert.Bindings, binding)
		}
		certificateList = append(certificateList, cert)
	}
	return
}

func getBindingFATEServiceDescription(serviceType entity.CertificateBindingServiceType, participant *entity.ParticipantFATE) string {
	return fmt.Sprintf("%v service of FATE %v", serviceType, participant.Type)
}

func getBindingOpenFLServiceDescription(serviceType entity.CertificateBindingServiceType, participant *entity.ParticipantOpenFL) string {
	return fmt.Sprintf("%v service of OpenFL %v", serviceType, participant.Type)
}

// DeleteCertificate deletes the certificate which has no participants bindings
func (app *CertificateApp) DeleteCertificate(uuid string) error {
	instanceList, err := app.CertificateBindingRepo.ListByCertificateUUID(uuid)
	if err != nil {
		return errors.Errorf("unable to get the certificate bindings")
	}
	bindingList := instanceList.([]entity.CertificateBinding)
	if len(bindingList) > 0 {
		return errors.Errorf("unable to delete certificate: there is(are) %v particicant(s) still binding to this certificate", len(bindingList))
	}
	return app.CertificateRepo.DeleteByUUID(uuid)
}
