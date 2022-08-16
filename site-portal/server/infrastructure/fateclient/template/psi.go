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
	"fmt"
)

const PSIDSL = `
{
    "components": {
        "reader_0": {
            "module": "Reader",
            "output": {
                "data": [
                    "data"
                ]
            }
        },
        "dataio_0": {
            "module": "DataIO",
            "input": {
                "data": {
                    "data": [
                        "reader_0.data"
                    ]
                }
            },
            "output": {
                "data": [
                    "data"
                ]
            }
        },
        "intersection_0": {
            "module": "Intersection",
            "input": {
                "data": {
                    "data": [
                        "dataio_0.data"
                    ]
                }
            },
            "output": {
                "data": [
                    "data"
                ]
            }
        }
    }
}
`

const PSIConf = `
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
        ]
    },
	"job_parameters": {
		"common": {
      	"job_type": "train",
      	"backend": 2,
      	"work_mode": 1,
      	"spark_run": {
        "num-executors": 1,
        "executor-cores": 1,
        "total-executor-cores": 1
      	}
    }
  },
    "component_parameters": {
        "common": {
            "intersect_0": {
                "intersect_method": "rsa",
                "sync_intersect_ids": false,
                "only_output_key": true,
                "rsa_params": {
                    "hash_method": "sha256",
                    "final_hash_method": "sha256",
                    "split_calculation": false,
                    "key_length": 2048
                }
            },
		"dataio_0": {
			"with_label": false,
			"output_format": "dense",
			"label_type": "int"
			}
        },
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

const PSIHostParamTemplate = `
{
	"reader_0": {
    	"table": {
      		"name": "%s",
      		"namespace": "%s"
    	}
  	}
}
`

// PSIParam contains parameters for a PSI job
type PSIParam struct {
	Guest PartyDataInfo
	Hosts []PartyDataInfo
}

// BuildPsiConf returns the DSL and Conf files
func BuildPsiConf(param PSIParam) (string, string, error) {
	hostParamStr, hostArrayStr, err := buildHostParams(param.Hosts, PSIHostParamTemplate)
	if err != nil {
		return "", "", err
	}
	confStr := fmt.Sprintf(PSIConf,
		param.Guest.PartyID,
		param.Guest.PartyID,
		hostArrayStr,
		hostParamStr,
		param.Guest.TableName,
		param.Guest.TableNamespace,
	)
	return confStr, PSIDSL, nil
}
