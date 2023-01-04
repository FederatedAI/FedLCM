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

const homoSBTDSL = `
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
      "DataTransform_0": {
          "module": "DataTransform",
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
              ],
              "model": [
                  "model"
              ]
          }
      },
      "HomoSecureBoost_0": {
          "module": "HomoSecureBoost",
          "input": {
              "data": {
                  "train_data": [
                      "DataTransform_0.data"
                  ]
              }
          },
          "output": {
              "data": [
                  "data"
              ],
              "model": [
                  "model"
              ]
          }
      },
      "Evaluation_0": {
          "module": "Evaluation",
          "input": {
              "data": {
                  "data": [
                      "HomoSecureBoost_0.data"
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

const homoSBTConf = `
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
      "task_parallelism": 2,
      "computing_partitions": 8,
      "eggroll_run": {
        "eggroll.session.processors.per.node": 2
      },
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      }
    }
  },
  "component_parameters": {
    "common": {
      "DataTransform_0": {
        "with_label": true,
        "output_format": "dense",
		    "label_type": "int",
    	  "label_name": "%s"
      },
      "HomoSecureBoost_0": {
          "task_type": "classification",
          "objective_param": {
              "objective": "cross_entropy"
          },
          "num_trees": 3,
          "validation_freqs": 1,
          "tree_param": {
              "max_depth": 3
          }
      },
      "Evaluation_0": {
          "eval_type": "binary"
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

const homoSBTHomoDataSplitDSL = `
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
    "DataTransform_0": {
      "module": "DataTransform",
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
        ],
        "model": [
          "model"
        ]
      }
    },
    "HomoDataSplit_0": {
      "output": {
        "data": [
          "train_data",
          "validate_data",
          "test_data"
        ]
      },
      "input": {
        "data": {
          "data": [
            "DataTransform_0.data"
          ]
        }
      },
      "module": "HomoDataSplit"
    },
    "HomoSecureBoost_0": {
        "module": "HomoSecureBoost",
        "input": {
            "data": {
              "validate_data": [
                "HomoDataSplit_0.validate_data"
              ],
              "train_data": [
                "HomoDataSplit_0.train_data"
              ]
            }
        },
        "output": {
            "data": [
                "data"
            ],
            "model": [
                "model"
            ]
        }
    },
    "Evaluation_0": {
        "module": "Evaluation",
        "input": {
            "data": {
                "data": [
                    "HomoSecureBoost_0.data"
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

const homoSBTHomoDataSplitConf = `
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
      "task_parallelism": 2,
      "computing_partitions": 8,
      "eggroll_run": {
        "eggroll.session.processors.per.node": 2
      },
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      }
    }
  },
  "component_parameters": {
    "common": {
      "HomoDataSplit_0": {
        "validate_size": %s,
        "split_points": [
          0,
          %s
        ],
        "test_size": 0,
        "stratified": true
      },
      "DataTransform_0": {
        "with_label": true,
        "output_format": "dense",
		    "label_type": "int",
    	  "label_name": "%s"
      },
      "HomoSecureBoost_0": {
        "task_type": "classification",
        "objective_param": {
          "objective": "cross_entropy"
        },
        "num_trees": 3,
        "validation_freqs": 1,
        "tree_param": {
          "max_depth": 3
        }
      },
      "Evaluation_0": {
        "eval_type": "binary"
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
