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

package mock

import (
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
)

type ParticipantFATERepoMock struct {
	CreateFn                                 func(instance interface{}) error
	ListFn                                   func() (interface{}, error)
	DeleteByUUIDFn                           func(uuid string) error
	GetByUUIDFn                              func(uuid string) (interface{}, error)
	ListByFederationUUIDFn                   func(uuid string) (interface{}, error)
	ListByEndpointUUIDFn                     func(uuid string) (interface{}, error)
	UpdateStatusByUUIDFn                     func(instance interface{}) error
	UpdateDeploymentYAMLByUUIDFn             func(instance interface{}) error
	UpdateInfoByUUIDFn                       func(instance interface{}) error
	IsExchangeCreatedByFederationUUIDFn      func(uuid string) (bool, error)
	GetExchangeByFederationUUIDFn            func(uuid string) (interface{}, error)
	IsConflictedByFederationUUIDAndPartyIDFn func(uuid string, partyID int) (bool, error)
}

func (m *ParticipantFATERepoMock) Create(instance interface{}) error {
	if m.CreateFn != nil {
		return m.CreateFn(instance)
	}
	return nil
}

func (m *ParticipantFATERepoMock) List() (interface{}, error) {
	if m.ListFn != nil {
		return m.ListFn()
	}
	return nil, nil
}

func (m *ParticipantFATERepoMock) DeleteByUUID(uuid string) error {
	if m.DeleteByUUIDFn != nil {
		return m.DeleteByUUIDFn(uuid)
	}
	return nil
}

func (m *ParticipantFATERepoMock) GetByUUID(uuid string) (interface{}, error) {
	if m.GetByUUIDFn != nil {
		return m.GetByUUIDFn(uuid)
	}
	return &entity.ParticipantFATE{}, nil
}

func (m *ParticipantFATERepoMock) ListByFederationUUID(uuid string) (interface{}, error) {
	if m.ListByFederationUUIDFn != nil {
		return m.ListByFederationUUIDFn(uuid)
	}
	return nil, nil
}

func (m *ParticipantFATERepoMock) ListByEndpointUUID(uuid string) (interface{}, error) {
	if m.ListByEndpointUUIDFn != nil {
		return m.ListByEndpointUUIDFn(uuid)
	}
	return nil, nil
}

func (m *ParticipantFATERepoMock) UpdateStatusByUUID(instance interface{}) error {
	if m.UpdateStatusByUUIDFn != nil {
		return m.UpdateStatusByUUIDFn(instance)
	}
	return nil
}

func (m *ParticipantFATERepoMock) UpdateDeploymentYAMLByUUID(instance interface{}) error {
	if m.UpdateDeploymentYAMLByUUIDFn != nil {
		return m.UpdateDeploymentYAMLByUUIDFn(instance)
	}
	return nil
}

func (m *ParticipantFATERepoMock) UpdateInfoByUUID(instance interface{}) error {
	if m.UpdateInfoByUUIDFn != nil {
		return m.UpdateInfoByUUIDFn(instance)
	}
	return nil
}

func (m *ParticipantFATERepoMock) IsExchangeCreatedByFederationUUID(uuid string) (bool, error) {
	if m.IsExchangeCreatedByFederationUUIDFn != nil {
		return m.IsExchangeCreatedByFederationUUIDFn(uuid)
	}
	return false, nil
}

func (m *ParticipantFATERepoMock) GetExchangeByFederationUUID(uuid string) (interface{}, error) {
	if m.GetExchangeByFederationUUIDFn != nil {
		return m.GetExchangeByFederationUUIDFn(uuid)
	}
	return nil, nil
}

func (m *ParticipantFATERepoMock) IsConflictedByFederationUUIDAndPartyID(uuid string, partyID int) (bool, error) {
	if m.IsConflictedByFederationUUIDAndPartyIDFn != nil {
		return m.IsConflictedByFederationUUIDAndPartyIDFn(uuid, partyID)
	}
	return false, nil
}

var _ repo.ParticipantFATERepository = (*ParticipantFATERepoMock)(nil)
