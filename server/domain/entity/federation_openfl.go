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
	"strings"

	"github.com/FederatedAI/FedLCM/server/domain/utils"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
)

// FederationOpenFL represents an OpenFL federation
type FederationOpenFL struct {
	Federation
	Domain                       string `gorm:"type:varchar(255);not null"`
	UseCustomizedShardDescriptor bool
	ShardDescriptorConfig        *valueobject.ShardDescriptorConfig `gorm:"type:text"`
}

// Create creates the OpenFL federation record in the repo
func (federation *FederationOpenFL) Create() error {
	federation.Type = FederationTypeOpenFL
	if !utils.IsDomainName(federation.Domain) {
		return errors.New("invalid domain name")
	}
	if federation.UseCustomizedShardDescriptor {
		for file, _ := range federation.ShardDescriptorConfig.PythonFiles {
			if strings.Contains(file, " ") {
				return errors.New("filename cannot contain space")
			}
		}
	}
	return federation.Repo.Create(federation)
}

func (FederationOpenFL) TableName() string {
	// just following the gorm convention
	return "federation_openfls"
}
