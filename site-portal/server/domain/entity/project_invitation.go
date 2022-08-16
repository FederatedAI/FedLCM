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

import "gorm.io/gorm"

// ProjectInvitation is the invitation for a project
type ProjectInvitation struct {
	gorm.Model
	UUID        string `gorm:"type:varchar(36);index;unique"`
	ProjectUUID string `gorm:"type:varchar(36)"`
	SiteUUID    string `gorm:"type:varchar(36)"`
	Status      ProjectInvitationStatus
}

// ProjectInvitationStatus is the status of the invitation
type ProjectInvitationStatus uint8

const (
	// ProjectInvitationStatusCreated means the invitation is created but hasn't been sent yet
	ProjectInvitationStatusCreated ProjectInvitationStatus = iota
	ProjectInvitationStatusSent
	ProjectInvitationStatusRevoked
	ProjectInvitationStatusAccepted
	ProjectInvitationStatusRejected
)
