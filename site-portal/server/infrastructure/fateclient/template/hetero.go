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

const heteroTrainingHostComponentParamTemplate = `
{
  "reader_0": {
    "table": {
      "name": "%s",
      "namespace": "%s"
    }
  },
  "DataTransform_0": {
    "input_format": "dense",
    "delimitor": ",",
    "data_type": "float64",
    "exclusive_data_type": null,
    "tag_with_value": false,
    "tag_value_delimitor": ":",
    "missing_fill": false,
    "default_value": 0,
    "missing_fill_method": null,
    "missing_impute": null,
    "outlier_replace": false,
    "outlier_replace_method": null,
    "outlier_impute": null,
    "outlier_replace_value": 0,
    "with_label": false,
    "label_name": "y",
    "label_type": "int",
    "output_format": "dense",
    "with_match_id": false
  }
}
`

const heteroPredictingHostComponentParamTemplate = `
{
  "reader_0": {
    "table": {
      "name": "%s",
      "namespace": "%s"
    }
  }
}
`

const heteroPredictingJobConf = `
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
      "task_parallelism": 2,
      "eggroll_run": {
        "eggroll.session.processors.per.node": 2
      },
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      },
      "job_type": "predict",
      "model_id": "%s",
      "model_version": "%s"
    }
  },
  "component_parameters": {
    "role": {
      "host": %s,
      "guest": {
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

// HeteroTrainingParam contains parameters for a vertical job
type HeteroTrainingParam struct {
	Guest             PartyDataInfo
	Hosts             []PartyDataInfo
	LabelName         string
	ValidationEnabled bool
	ValidationPercent uint
	Type              HeteroAlgorithmType
}

// HeteroPredictingParam contains parameters for creating a predicting job for a vertical model
type HeteroPredictingParam struct {
	Guest        PartyDataInfo
	Hosts        []PartyDataInfo
	ModelID      string
	ModelVersion string
}

// HeteroAlgorithmType is the enum of vertical algorithm types
type HeteroAlgorithmType uint8

const (
	HeteroAlgorithmTypeUnknown HeteroAlgorithmType = iota
	HeteroAlgorithmTypeLR
	HeteroAlgorithmTypeSBT
)

var heteroAlgorithmTypeTemplateMap = map[HeteroAlgorithmType]map[bool][]string{
	HeteroAlgorithmTypeLR: {
		true: {
			heteroLRHeteroDataSplitDSL,
			heteroLRHeteroDataSplitConf,
		},
		false: {
			heteroLRDSL,
			heteroLRConf,
		},
	},
	HeteroAlgorithmTypeSBT: {
		true: {
			heteroSBTHeteroDataSplitDSL,
			heteroSBTHeteroDataSplitConf,
		},
		false: {
			heteroSBTDSL,
			heteroSBTConf,
		},
	},
}

// BuildHeteroTrainingConf returns the FATE job conf and dsl from the specified param
func BuildHeteroTrainingConf(param HeteroTrainingParam) (string, string, error) {
	if param.LabelName == "" {
		param.LabelName = "y"
	}
	hostParamStr, hostArrayStr, err := buildHostParams(param.Hosts, heteroTrainingHostComponentParamTemplate)
	if err != nil {
		return "", "", err
	}
	dslStr := heteroAlgorithmTypeTemplateMap[param.Type][param.ValidationEnabled][0]
	confStr := heteroAlgorithmTypeTemplateMap[param.Type][param.ValidationEnabled][1]

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
			hostParamStr,
			param.Guest.TableName,
			param.Guest.TableNamespace,
			param.LabelName,
		)
	} else {
		confStr = fmt.Sprintf(confStr,
			param.Guest.PartyID,
			param.Guest.PartyID,
			hostArrayStr,
			arbiterPartyID,
			hostParamStr,
			param.Guest.TableName,
			param.Guest.TableNamespace,
			param.LabelName,
		)
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

// BuildHeteroPredictingConf returns the FATE job conf and dsl from the specified param
func BuildHeteroPredictingConf(param HeteroPredictingParam) (string, string, error) {
	hostParamStr, hostArrayStr, err := buildHostParams(param.Hosts, heteroPredictingHostComponentParamTemplate)
	if err != nil {
		return "", "", err
	}

	arbiterPartyID := param.Guest.PartyID
	if len(param.Hosts) > 0 {
		arbiterPartyID = param.Hosts[0].PartyID
	}
	return fmt.Sprintf(heteroPredictingJobConf,
		param.Guest.PartyID,
		param.Guest.PartyID,
		hostArrayStr,
		arbiterPartyID,
		param.ModelID,
		param.ModelVersion,
		hostParamStr,
		param.Guest.TableName,
		param.Guest.TableNamespace), "{}", nil
}
