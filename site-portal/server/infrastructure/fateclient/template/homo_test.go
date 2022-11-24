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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHomoLRConf(t *testing.T) {
	for _, algorithmType := range []HomoAlgorithmType{HomoAlgorithmTypeLR, HomoAlgorithmTypeSBT} {
		info := HomoTrainingParam{
			Guest: PartyDataInfo{
				PartyID:        "9999",
				TableName:      "tb1",
				TableNamespace: "tbns1",
			},
			Hosts: []PartyDataInfo{
				{
					PartyID:        "10000",
					TableName:      "tb2",
					TableNamespace: "tbns2",
				},
			},
			Type: algorithmType,
		}
		conf, dsl, err := BuildHomoTrainingConf(info)
		assert.NoError(t, err)
		assert.True(t, json.Valid([]byte(conf)))
		assert.True(t, json.Valid([]byte(dsl)))

		info.Hosts = append(info.Hosts, PartyDataInfo{
			PartyID:        "20000",
			TableName:      "tb3",
			TableNamespace: "tbns3",
		})
		conf, dsl, err = BuildHomoTrainingConf(info)
		assert.NoError(t, err)
		assert.True(t, json.Valid([]byte(conf)))
		assert.True(t, json.Valid([]byte(dsl)))

		info.Hosts = nil
		conf, dsl, err = BuildHomoTrainingConf(info)
		assert.NoError(t, err)
		assert.True(t, json.Valid([]byte(conf)))
		assert.True(t, json.Valid([]byte(dsl)))

		info.ValidationEnabled = true
		info.ValidationPercent = 80
		conf, dsl, err = BuildHomoTrainingConf(info)
		assert.NoError(t, err)
		assert.True(t, json.Valid([]byte(conf)))
		assert.True(t, json.Valid([]byte(dsl)))
	}
}

func TestBuildHomoTrainingConf(t *testing.T) {
	type args struct {
		param HomoTrainingParam
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
			name: "homo-lr-base",
			args: args{
				param: HomoTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "999",
						TableName:      "guest-table-name-999",
						TableNamespace: "guest-table-namespace-999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "1000",
							TableName:      "host-table-name-1000",
							TableNamespace: "host-table-namespace-1000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: false,
					ValidationPercent: 0,
					Type:              HomoAlgorithmTypeLR,
				},
			},
			wantErr: false,
		},
		{
			name: "homo-lr-base-validation",
			args: args{
				param: HomoTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "999",
						TableName:      "guest-table-name-999",
						TableNamespace: "guest-table-namespace-999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "1000",
							TableName:      "host-table-name-1000",
							TableNamespace: "host-table-namespace-1000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: true,
					ValidationPercent: 10,
					Type:              HomoAlgorithmTypeLR,
				},
			},
			wantErr: false,
		},
		{
			name: "homo-sbt-base",
			args: args{
				param: HomoTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "999",
						TableName:      "guest-table-name-999",
						TableNamespace: "guest-table-namespace-999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "1000",
							TableName:      "host-table-name-1000",
							TableNamespace: "host-table-namespace-1000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: false,
					ValidationPercent: 0,
					Type:              HomoAlgorithmTypeSBT,
				},
			},
			wantErr: false,
		},
		{
			name: "homo-sbt-base-validation",
			args: args{
				param: HomoTrainingParam{
					Guest: PartyDataInfo{
						PartyID:        "999",
						TableName:      "guest-table-name-999",
						TableNamespace: "guest-table-namespace-999",
					},
					Hosts: []PartyDataInfo{
						{
							PartyID:        "1000",
							TableName:      "host-table-name-1000",
							TableNamespace: "host-table-namespace-1000",
						},
					},
					LabelName:         "y",
					ValidationEnabled: true,
					ValidationPercent: 10,
					Type:              HomoAlgorithmTypeSBT,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := BuildHomoTrainingConf(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildHomoTrainingConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestBuildHomoPredictingConf(t *testing.T) {
	type args struct {
		param HomoPredictingParam
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
				param: HomoPredictingParam{
					Role:         "guest",
					ModelID:      "homo-guest-host-test",
					ModelVersion: "123456789",
					PartyDataInfo: PartyDataInfo{
						PartyID:        "9999",
						TableName:      "homo-name-test",
						TableNamespace: "homo-namespace-test",
					},
				},
			},
			want:    "",
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := BuildHomoPredictingConf(tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildHomoPredictingConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
