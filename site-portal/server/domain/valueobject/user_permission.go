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

// UserPermissionInfo holds information about user's permissions
type UserPermissionInfo struct {
	// SitePortalAccess controls whether the user can access site portal
	SitePortalAccess bool `json:"site_portal_access"`
	// FATEBoardAccess controls whether the user can access fate board
	FATEBoardAccess bool `json:"fateboard_access" gorm:"column:fateboard_access"`
	// NoteBookAccess controls whether the user can access notebook
	NotebookAccess bool `json:"notebook_access"`
}
