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

// ChartRepository is the interface to handle chart's persistence related actions
type ChartRepository interface {
	// Create takes a *entity.Chart creates a chart info record in the repository
	Create(interface{}) error
	// List returns []entity.Chart of all saved chart
	List() (interface{}, error)
	// DeleteByUUID delete the chart of the specified uuid
	DeleteByUUID(string) error
	// GetByUUID returns an *entity.Chart of the specified uuid
	GetByUUID(string) (interface{}, error)
	// GetByNameAndNamespace returns an *entity.Chart of the specified name and version
	GetByNameAndVersion(string, string) (interface{}, error)
	// ListByType takes an entity.ChartType and returns []entity.Chart that is for the specified type
	ListByType(interface{}) (interface{}, error)
}
