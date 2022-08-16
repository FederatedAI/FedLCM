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

package valueobject

// ProjectCreatorInfo contains info of the creator/manager of a project
type ProjectCreatorInfo struct {
	Manager             string `json:"manager" gorm:"type:varchar(255)"`
	ManagingSiteName    string `json:"managing_site_name" gorm:"type:varchar(255)"`
	ManagingSitePartyID uint   `json:"managing_site_party_id"`
	ManagingSiteUUID    string `json:"managing_site_uuid" gorm:"type:varchar(36)"`
}
