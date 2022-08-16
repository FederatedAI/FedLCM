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
	"bytes"
	"encoding/json"
	"fmt"
)

const homoHostComponentParamTemplate = `
{
  "reader_0": {
    "table": {
      "name": "%s",
      "namespace": "%s"
    }
  }
}
`

const homoPredictingJobConf = `
{
  "dsl_version": 2,
  "initiator": {
    "role": "%s",
    "party_id": %s
  },
  "role": {
    "%s": [
      %s
    ]
  },
  "job_parameters": {
    "common": {
      "work_mode": 1,
      "backend": 2,
      "job_type": "predict",
      "model_id": "%s",
      "model_version": "%s"
    }
  },
  "component_parameters": {
    "role": {
      "%s": {
        "0": {
          "reader_0": {
            "table": {
              "name": "%s",
              "namespace": "%s"
            }
          }
        }
      }
    }
  }
}
`

// HomoTrainingParam contains parameters for a horizontal job
type HomoTrainingParam struct {
	Guest             PartyDataInfo
	Hosts             []PartyDataInfo
	LabelName         string
	ValidationEnabled bool
	ValidationPercent uint
	Type              HomoAlgorithmType
}

// HomoPredictingParam contains parameters for creating a predicting job for a horizontal model
type HomoPredictingParam struct {
	Role          string
	ModelID       string
	ModelVersion  string
	PartyDataInfo PartyDataInfo
}

// HomoAlgorithmType is the enum of horizontal algorithm types
type HomoAlgorithmType uint8

const (
	HomoAlgorithmTypeUnknown HomoAlgorithmType = iota
	HomoAlgorithmTypeLR
	HomoAlgorithmTypeSBT
)

var homoAlgorithmTypeTemplateMap = map[HomoAlgorithmType]map[bool][]string{
	HomoAlgorithmTypeLR: {
		true: {
			homoLRHomoDataSplitDSL,
			homoLRHomoDataSplitConf,
		},
		false: {
			homoLRDSL,
			homoLRConf,
		},
	},
	HomoAlgorithmTypeSBT: {
		true: {
			homoSBTHomoDataSplitDSL,
			homoSBTHomoDataSplitConf,
		},
		false: {
			homoSBTDSL,
			homoSBTConf,
		},
	},
}

// BuildHomoTrainingConf returns the FATE job conf and dsl from the specified param
func BuildHomoTrainingConf(param HomoTrainingParam) (string, string, error) {
	if param.LabelName == "" {
		param.LabelName = "y"
	}
	hostParamStr, hostArrayStr, err := buildHostParams(param.Hosts, homoHostComponentParamTemplate)
	if err != nil {
		return "", "", err
	}
	dslStr := homoAlgorithmTypeTemplateMap[param.Type][param.ValidationEnabled][0]
	confStr := homoAlgorithmTypeTemplateMap[param.Type][param.ValidationEnabled][1]

	arbiterPartyID := param.Guest.PartyID
	if len(param.Hosts) > 0 {
		arbiterPartyID = param.Hosts[0].PartyID
	}

	if param.ValidationEnabled {
		validationSizeStr := fmt.Sprintf("%0.2f", float64(param.ValidationPercent)/100)
		confStr = fmt.Sprintf(confStr,
			param.Guest.PartyID,
			param.Guest.PartyID,
			hostArrayStr,
			arbiterPartyID,
			validationSizeStr,
			validationSizeStr,
			param.LabelName,
			hostParamStr,
			param.Guest.TableName,
			param.Guest.TableNamespace)
	} else {
		confStr = fmt.Sprintf(confStr,
			param.Guest.PartyID,
			param.Guest.PartyID,
			hostArrayStr,
			arbiterPartyID,
			param.LabelName,
			hostParamStr,
			param.Guest.TableName,
			param.Guest.TableNamespace)
	}
	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, []byte(confStr), "", "  "); err != nil {
		return "", "", err
	}
	confStr = prettyJson.String()

	prettyJson.Reset()
	if err := json.Indent(&prettyJson, []byte(dslStr), "", "  "); err != nil {
		return "", "", err
	}
	dslStr = prettyJson.String()
	return confStr, dslStr, nil
}

// BuildHomoPredictingConf returns the FATE job conf and dsl from the specified param
func BuildHomoPredictingConf(param HomoPredictingParam) (string, string, error) {
	return fmt.Sprintf(homoPredictingJobConf,
		param.Role,
		param.PartyDataInfo.PartyID,
		param.Role,
		param.PartyDataInfo.PartyID,
		param.ModelID,
		param.ModelVersion,
		param.Role,
		param.PartyDataInfo.TableName,
		param.PartyDataInfo.TableNamespace), "{}", nil
}
