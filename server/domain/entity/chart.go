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
	"gorm.io/gorm"
)

// Chart is the helm chart
type Chart struct {
	gorm.Model
	UUID                string `gorm:"type:varchar(36);index;unique"`
	Name                string `gorm:"type:varchar(255);not null"`
	Description         string `gorm:"type:text"`
	Type                ChartType
	ChartName           string `gorm:"type:varchar(255)"`
	Version             string `gorm:"type:varchar(32);not null"`
	AppVersion          string `gorm:"type:varchar(32);not null"`
	Chart               string `gorm:"type:text;not null"`
	InitialYamlTemplate string `gorm:"type:text;not null"`
	Values              string `gorm:"type:text;not null"`
	ValuesTemplate      string `gorm:"type:text;not null"`
	ArchiveContent      []byte `gorm:"type:mediumblob"`
	Private             bool
}

// ChartType is the supported deployment type
type ChartType uint8

const (
	ChartTypeUnknown ChartType = iota
	ChartTypeFATEExchange
	ChartTypeFATECluster
	ChartTypeOpenFLDirector
	ChartTypeOpenFLEnvoy
)
