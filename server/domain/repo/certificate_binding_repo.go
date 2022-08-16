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

package repo

// CertificateBindingRepository is the interface to handle certificate binding's persistence related actions
type CertificateBindingRepository interface {
	// Create takes a *entity.CertificateBinding and creates a record in the repository
	Create(interface{}) error
	// ListByCertificateUUID returns []entity.CertificateBinding of the specified certificate
	ListByCertificateUUID(string) (interface{}, error)
	// DeleteByParticipantUUID deletes certificate binding info with the specified participant uuid
	DeleteByParticipantUUID(string) error
	// ListByParticipantUUID returns []entity.CertificateBinding of the specified participant uuid
	ListByParticipantUUID(string) (interface{}, error)
}
