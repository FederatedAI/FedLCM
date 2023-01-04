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

package template

import (
	"testing"
)

func TestBuildHeteroTrainingConf(t *testing.T) {
	type args struct {
		param HeteroTrainingParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		// Check json format and basic functions
		{
			name: "hetero-lr-base",
			args: args{
				param: HeteroTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "guest-name-9999",
						TableNamespace: "guest-namespace-9999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "10000",
							TableName:      "host-name-10000",
							TableNamespace: "host-namespace-10000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: false,
					ValidationPercent: 0,
					Type:              HeteroAlgorithmTypeLR,
				},
			},
			wantErr: false,
		},
		{
			name: "hetero-lr-base-validation",
			args: args{
				param: HeteroTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "guest-name-9999",
						TableNamespace: "guest-namespace-9999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "10000",
							TableName:      "host-name-10000",
							TableNamespace: "host-namespace-10000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: true,
					ValidationPercent: 10,
					Type:              HeteroAlgorithmTypeLR,
				},
			},
			wantErr: false,
		},
		{
			name: "hetero-sbt-base",
			args: args{
				param: HeteroTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "guest-name-9999",
						TableNamespace: "guest-namespace-9999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "10000",
							TableName:      "host-name-10000",
							TableNamespace: "host-namespace-10000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: false,
					ValidationPercent: 0,
					Type:              HeteroAlgorithmTypeSBT,
				},
			},
			wantErr: false,
		},
		{
			name: "hetero-sbt-base-validation",
			args: args{
				param: HeteroTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "guest-name-9999",
						TableNamespace: "guest-namespace-9999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "10000",
							TableName:      "host-name-10000",
							TableNamespace: "host-namespace-10000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: true,
					ValidationPercent: 10,
					Type:              HeteroAlgorithmTypeSBT,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := BuildHeteroTrainingConf(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildHeteroTrainingConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBuildHeteroPredictingConf(t *testing.T) {
	type args struct {
		param HeteroPredictingParam
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "hetero-predicting-base",
			args: args{
				param: HeteroPredictingParam{
					Guest: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "hetero-name-guest",
						TableNamespace: "hetero-namespace-guest",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "10000",
							TableName:      "hetero-name-host",
							TableNamespace: "hetero-namespace-host",
						},
					},
					ModelID:      "hetero-guest-host-test",
					ModelVersion: "123456789",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := BuildHeteroPredictingConf(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildHeteroPredictingConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
