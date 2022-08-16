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
