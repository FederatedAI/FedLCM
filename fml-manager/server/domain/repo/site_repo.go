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

import "github.com/pkg/errors"

// ErrSiteNameConflict means new site cannot be created due to the existence of the same-name site
var ErrSiteNameConflict = errors.New("a site with the same name but different UUID is already registered")

// SiteRepository is the interface to handle site related information in the repo
type SiteRepository interface {
	// GetSiteList returns all sites in []entity.Site
	GetSiteList() (interface{}, error)
	// Save creates a site info record in the repository
	Save(instance interface{}) (interface{}, error)
	// ExistByUUID returns whether the site with the uuid exists
	ExistByUUID(uuid string) (bool, error)
	// UpdateByUUID updates sites info by uuid
	UpdateByUUID(instance interface{}) error
	// DeleteByUUID delete sites info with the specified uuid
	DeleteByUUID(uuid string) error
	// GetByUUID returns an *entity.Site of the specified site
	GetByUUID(string) (interface{}, error)
}
