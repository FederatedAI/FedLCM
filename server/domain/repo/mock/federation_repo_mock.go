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

type FederationFATERepoMock struct {
	CreateFn       func(instance interface{}) error
	ListFn         func() (interface{}, error)
	DeleteByUUIDfn func(uuid string) error
	GetByUUIDfn    func(uuid string) (interface{}, error)
}

func (m *FederationFATERepoMock) Create(instance interface{}) error {
	if m.CreateFn != nil {
		return m.CreateFn(instance)
	}
	return nil
}

func (m *FederationFATERepoMock) List() (interface{}, error) {
	if m.ListFn != nil {
		return m.ListFn()
	}
	return nil, nil
}

func (m *FederationFATERepoMock) DeleteByUUID(uuid string) error {
	if m.DeleteByUUIDfn != nil {
		return m.DeleteByUUIDfn(uuid)
	}
	return nil
}

func (m *FederationFATERepoMock) GetByUUID(uuid string) (interface{}, error) {
	if m.GetByUUIDfn != nil {
		return m.GetByUUIDfn(uuid)
	}
	return &entity.FederationFATE{
		Domain: "test.example.com",
	}, nil
}

var _ repo.FederationRepository = (*FederationFATERepoMock)(nil)
