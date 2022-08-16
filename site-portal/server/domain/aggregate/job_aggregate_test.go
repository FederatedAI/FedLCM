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

package aggregate

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/stretchr/testify/assert"
)

func getJobAggregate() *JobAggregate {
	return &JobAggregate{
		Initiator: &entity.JobParticipant{
			SitePartyID:        1,
			DataTableName:      "guest-tablename-0",
			DataTableNamespace: "guest-tablens-0",
		},
		Participants: map[string]*entity.JobParticipant{
			"hostuuid1": {
				SitePartyID:        2,
				DataTableName:      "host-tablename-0",
				DataTableNamespace: "host-tablens-0"},
			"hostuuid2": {
				SitePartyID:        3,
				DataTableName:      "host-tablename-1",
				DataTableNamespace: "host-tablens-1"},
		},
	}
}

func TestGenerateReaderConfigMaps(t *testing.T) {
	jobAggregate := getJobAggregate()
	hostUuidList := []string{"hostuuid1", "hostuuid2"}
	hostMap, guestMap := jobAggregate.GenerateReaderConfigMaps(hostUuidList)
	expectedHostMap := map[string]map[string]map[string]map[string]string{
		"0": {
			"reader_0": {
				"table": {
					"name":      "host-tablename-0",
					"namespace": "host-tablens-0",
				},
			},
		},
		"1": {
			"reader_0": {
				"table": {
					"name":      "host-tablename-1",
					"namespace": "host-tablens-1",
				},
			},
		},
	}
	expectedGuestMap := map[string]map[string]map[string]map[string]string{
		"0": {
			"reader_0": {
				"table": {
					"name":      "guest-tablename-0",
					"namespace": "guest-tablens-0",
				},
			},
		},
	}
	assert.Equal(t, fmt.Sprintln(expectedGuestMap), fmt.Sprintln(guestMap))
	assert.Equal(t, fmt.Sprintln(expectedHostMap), fmt.Sprintln(hostMap))
}

func TestGenerateGeneralTrainingConf(t *testing.T) {
	jobAggregate := getJobAggregate()
	// Make below list in the reverser order
	hostUuidList := []string{"hostuuid2", "hostuuid1"}
	actualRes, _ := jobAggregate.GenerateGeneralTrainingConf(hostUuidList)
	actualResStruct := map[string]interface{}{}
	json.Unmarshal([]byte(actualRes), &actualResStruct)
	role := actualResStruct["role"].(map[string]interface{})
	host := role["host"].([]interface{})
	// hostUuid2 matches party id 3
	assert.Equal(t, int(host[0].(float64)), 3)
}
