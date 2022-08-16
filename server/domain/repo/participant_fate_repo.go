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

// ParticipantFATERepository contains FATE participant specific operations in addition to ParticipantRepository
type ParticipantFATERepository interface {
	ParticipantRepository
	// IsExchangeCreatedByFederationUUID returns whether an exchange exists in the specified federation
	IsExchangeCreatedByFederationUUID(string) (bool, error)
	// GetExchangeByFederationUUID returns an *entity.ParticipantFATE that is the exchange of the specified federation
	GetExchangeByFederationUUID(string) (interface{}, error)
	// IsConflictedByFederationUUIDAndPartyID returns whether a party id in a federation is already used
	IsConflictedByFederationUUIDAndPartyID(string, int) (bool, error)
}
