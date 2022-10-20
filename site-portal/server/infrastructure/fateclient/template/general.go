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
	"fmt"
	"github.com/rs/zerolog/log"
	"strconv"
)

const trainingGeneralConf = `
{
  "dsl_version": 2,
  "initiator": {
    "role": "guest",
    "party_id": %s
  },
  "role": {
    "guest": [
      %s
    ],
    "host": [
      %s
    ],
    "arbiter": [
      %s
    ]
  },
  "job_parameters": {
    "common": {
      "job_type": "train",
      "backend": 2,
      "work_mode": 1,
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      }
    }
  }
}`

// GeneralTrainingParam contains common parameters for creating all kins of jobs
type GeneralTrainingParam struct {
	Guest PartyDataInfo
	Hosts []PartyDataInfo
}

// BuildTrainingConfGeneralStr returns a json string, which is a part of the final FATE job conf file. Including
// "dsl_version", "initiator", "role" and "job_parameters".
func BuildTrainingConfGeneralStr(param GeneralTrainingParam) (string, error) {
	hostNum := len(param.Hosts)
	if param.Hosts == nil || hostNum == 0 {
		log.Debug().Msg("Build training conf for a guest only training job")
		return fmt.Sprintf(trainingGeneralConf, param.Guest.PartyID, param.Guest.PartyID, "",
			param.Guest.PartyID), nil
	}
	hostArrayStr := param.Hosts[0].PartyID
	for i := 1; i < hostNum; i++ {
		hostArrayStr += ", " + param.Hosts[i].PartyID
	}
	return fmt.Sprintf(trainingGeneralConf, param.Guest.PartyID, param.Guest.PartyID, hostArrayStr,
		param.Hosts[0].PartyID), nil
}

// buildHostParams is a helper function which can help generate the parameter string for the hosts
func buildHostParams(hostInfos []PartyDataInfo, template string) (string, string, error) {
	hostArrayStr := ""
	hostParamMap := make(map[string]interface{})
	for index, host := range hostInfos {
		if hostArrayStr == "" {
			hostArrayStr = host.PartyID
		} else {
			hostArrayStr += ", " + host.PartyID
		}
		indexStr := strconv.Itoa(index)
		readerConfStr := fmt.Sprintf(template, host.TableName, host.TableNamespace)
		var readerConf map[string]interface{}
		if err := json.Unmarshal([]byte(readerConfStr), &readerConf); err != nil {
			return "", "", err
		}
		hostParamMap[indexStr] = readerConf
	}
	hostParamBytes, err := json.Marshal(hostParamMap)
	if err != nil {
		return "", "", err
	}
	hostParamStr := string(hostParamBytes)
	return hostParamStr, hostArrayStr, err
}
