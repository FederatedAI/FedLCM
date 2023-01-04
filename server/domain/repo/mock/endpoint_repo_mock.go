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

type EndpointKubeFATERepoMock struct {
	CreateFn                              func(instance interface{}) error
	ListFn                                func() (interface{}, error)
	DeleteByUUIDFn                        func(uuid string) error
	GetByUUIDFn                           func(uuid string) (interface{}, error)
	ListByInfraProviderUUIDFn             func(uuid string) (interface{}, error)
	UpdateStatusByUUIDFn                  func(instance interface{}) error
	UpdateInfoByUUIDFn                    func(instance interface{}) error
	ListByInfraProviderUUIDAndNamespaceFn func(string, string) (interface{}, error)
}

func (m *EndpointKubeFATERepoMock) ListByInfraProviderUUIDAndNamespace(uuid string, namespace string) (interface{}, error) {
	if m.ListByInfraProviderUUIDAndNamespaceFn != nil {
		return m.ListByInfraProviderUUIDAndNamespaceFn(uuid, namespace)
	}
	return nil, nil
}

func (m *EndpointKubeFATERepoMock) Create(instance interface{}) error {
	return m.CreateFn(instance)
}

func (m *EndpointKubeFATERepoMock) List() (interface{}, error) {
	return m.ListFn()
}

func (m *EndpointKubeFATERepoMock) DeleteByUUID(uuid string) error {
	return m.DeleteByUUIDFn(uuid)
}

func (m *EndpointKubeFATERepoMock) GetByUUID(uuid string) (interface{}, error) {
	return m.GetByUUIDFn(uuid)
}

func (m *EndpointKubeFATERepoMock) ListByInfraProviderUUID(uuid string) (interface{}, error) {
	return m.ListByInfraProviderUUIDFn(uuid)
}

func (m *EndpointKubeFATERepoMock) UpdateStatusByUUID(instance interface{}) error {
	return m.UpdateStatusByUUIDFn(instance)
}

func (m *EndpointKubeFATERepoMock) UpdateInfoByUUID(instance interface{}) error {
	return m.UpdateInfoByUUIDFn(instance)
}

var _ repo.EndpointRepository = (*EndpointKubeFATERepoMock)(nil)
