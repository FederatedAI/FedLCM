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

package service

import (
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
)

// ChartApp provide functions to manage the available helm charts
type ChartApp struct {
	ChartRepo repo.ChartRepository
}

// ChartListItem contains basic info of a chart that can be used in a "list" view
type ChartListItem struct {
	UUID                  string           `json:"uuid"`
	Name                  string           `json:"name"`
	ChartName             string           `json:"chart_name"`
	Version               string           `json:"version"`
	Description           string           `json:"description"`
	Type                  entity.ChartType `json:"type"`
	CreatedAt             time.Time        `json:"created_at"`
	ContainPortalServices bool             `json:"contain_portal_services"`
}

// ChartDetail contains detailed info of a chart
type ChartDetail struct {
	ChartListItem
	About          string `json:"about"`
	Values         string `json:"values"`
	ValuesTemplate string `json:"values_template"`
}

// List returns the currently installed chart
func (app *ChartApp) List(t entity.ChartType) ([]ChartListItem, error) {
	var chartList []ChartListItem
	var domainChartList []entity.Chart
	if t == entity.ChartTypeUnknown {
		instanceList, err := app.ChartRepo.List()
		if err != nil {
			return nil, err
		}
		domainChartList = instanceList.([]entity.Chart)
	} else {
		instanceList, err := app.ChartRepo.ListByType(t)
		if err != nil {
			return nil, err
		}
		domainChartList = instanceList.([]entity.Chart)
	}
	for _, domainChart := range domainChartList {
		chartList = append(chartList, ChartListItem{
			UUID:                  domainChart.UUID,
			Name:                  domainChart.Name,
			ChartName:             domainChart.ChartName,
			Version:               domainChart.Version,
			Description:           domainChart.Description,
			Type:                  domainChart.Type,
			CreatedAt:             domainChart.CreatedAt,
			ContainPortalServices: domainChart.Private,
		})
	}
	return chartList, nil
}

// GetDetail returns detailed info of a chart
func (app *ChartApp) GetDetail(uuid string) (*ChartDetail, error) {
	instance, err := app.ChartRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	domainChart := instance.(*entity.Chart)
	return &ChartDetail{
		ChartListItem: ChartListItem{
			UUID:                  domainChart.UUID,
			Name:                  domainChart.Name,
			ChartName:             domainChart.ChartName,
			Version:               domainChart.Version,
			Description:           domainChart.Description,
			Type:                  domainChart.Type,
			CreatedAt:             domainChart.CreatedAt,
			ContainPortalServices: domainChart.Private,
		},
		About:          domainChart.Chart,
		Values:         domainChart.Values,
		ValuesTemplate: domainChart.ValuesTemplate,
	}, nil
}
