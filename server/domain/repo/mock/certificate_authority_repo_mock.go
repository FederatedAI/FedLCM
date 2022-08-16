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
	"github.com/FederatedAI/FedLCM/server/domain/repo"
)

type CertificateAuthorityRepoMock struct {
	CreateFn       func(instance interface{}) error
	UpdateByUUIDFn func(instance interface{}) error
	GetFirstFn     func() (interface{}, error)
}

func (m *CertificateAuthorityRepoMock) Create(instance interface{}) error {
	return m.CreateFn(instance)
}

func (m *CertificateAuthorityRepoMock) UpdateByUUID(instance interface{}) error {
	return m.UpdateByUUIDFn(instance)
}

func (m *CertificateAuthorityRepoMock) GetFirst() (interface{}, error) {
	return m.GetFirstFn()
}

var _ repo.CertificateAuthorityRepository = (*CertificateAuthorityRepoMock)(nil)
