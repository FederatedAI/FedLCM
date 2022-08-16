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

// ParticipantOpenFLRepository contains OpenFL specific operations in addition to ParticipantRepository
type ParticipantOpenFLRepository interface {
	ParticipantRepository
	// IsDirectorCreatedByFederationUUID returns whether a director exists in the specified federation
	IsDirectorCreatedByFederationUUID(string) (bool, error)
	// CountByTokenUUID returns the number of participant using a specified token
	CountByTokenUUID(string) (int, error)
	// GetDirectorByFederationUUID returns an *entity.ParticipantOpenFL that is the director of the specified federation
	GetDirectorByFederationUUID(string) (interface{}, error)
}
