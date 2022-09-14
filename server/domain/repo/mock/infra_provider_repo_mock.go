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

import "github.com/FederatedAI/FedLCM/server/domain/repo"

type InfraProviderKubernetesRepoMock struct {
	CreateFn            func(instance interface{}) error
	ListFn              func() (interface{}, error)
	DeleteByUUIDFn      func(uuid string) error
	GetByUUIDFn         func(uuid string) (interface{}, error)
	UpdateByUUIDFn      func(instance interface{}) error
	GetByConfigSHA256Fn func(address string) (interface{}, error)
}

func (m *InfraProviderKubernetesRepoMock) ProviderExists(instance interface{}) error {
	return m.ProviderExists(instance)
}

func (m *InfraProviderKubernetesRepoMock) Create(instance interface{}) error {
	return m.CreateFn(instance)
}

func (m *InfraProviderKubernetesRepoMock) List() (interface{}, error) {
	return m.ListFn()
}

func (m *InfraProviderKubernetesRepoMock) DeleteByUUID(uuid string) error {
	return m.DeleteByUUIDFn(uuid)
}

func (m *InfraProviderKubernetesRepoMock) GetByUUID(uuid string) (interface{}, error) {
	return m.GetByUUIDFn(uuid)
}

func (m *InfraProviderKubernetesRepoMock) UpdateByUUID(instance interface{}) error {
	return m.UpdateByUUIDFn(instance)
}

func (m *InfraProviderKubernetesRepoMock) GetByConfigSHA256(address string) (interface{}, error) {
	return m.GetByConfigSHA256Fn(address)
}

var _ repo.InfraProviderRepository = (*InfraProviderKubernetesRepoMock)(nil)
