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
	"github.com/FederatedAI/FedLCM/server/domain/utils"
	"github.com/pkg/errors"
)

// FederationFATE represents a FATE federation
type FederationFATE struct {
	Federation
	Domain string `gorm:"type:varchar(255);not null"`
}

// Create creates the FATE federation record in the repo
func (federation *FederationFATE) Create() error {
	federation.Type = FederationTypeFATE
	if !utils.IsDomainName(federation.Domain) {
		return errors.New("invalid domain name")
	}
	return federation.Repo.Create(federation)
}
