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
      "output": {
        "data": [
          "data"
        ]
      },
      "module": "Reader"
    },
    "evaluation_0": {
      "output": {
        "data": [
          "data"
        ]
      },
      "input": {
        "data": {
          "data": [
            "HomoSecureboost_0.data"
          ]
        }
      },
      "module": "Evaluation"
    },
    "dataio_0": {
      "output": {
        "data": [
          "data"
        ],
        "model": [
          "model"
        ]
      },
      "input": {
        "data": {
          "data": [
            "reader_0.data"
          ]
        }
      },
      "module": "DataIO"
    },
    "HomoSecureboost_0": {
      "output": {
        "data": [
          "data"
        ],
        "model": [
          "model"
        ]
      },
      "input": {
        "data": {
          "train_data": [
            "dataio_0.data"
          ]
        }
      },
      "module": "HomoSecureboost"
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
      "backend": 2,
      "work_mode": 1,
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      }
    }
  },
  "component_parameters": {
    "common": {
      "evaluation_0": {
        "eval_type": "binary"
      },
      "HomoSecureboost_0": {
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
      "dataio_0": {
        "with_label": true,
        "output_format": "dense",
        "label_type": "int",
        "label_name": "%s"
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
    "homo_data_split_0": {
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
            "dataio_0.data"
          ]
        }
      },
      "module": "HomoDataSplit"
    },
    "reader_0": {
      "output": {
        "data": [
          "data"
        ]
      },
      "module": "Reader"
    },
    "evaluation_0": {
      "output": {
        "data": [
          "data"
        ]
      },
      "input": {
        "data": {
          "data": [
            "HomoSecureboost_0.data"
          ]
        }
      },
      "module": "Evaluation"
    },
    "dataio_0": {
      "output": {
        "data": [
          "data"
        ],
        "model": [
          "model"
        ]
      },
      "input": {
        "data": {
          "data": [
            "reader_0.data"
          ]
        }
      },
      "module": "DataIO"
    },
    "HomoSecureboost_0": {
      "output": {
        "data": [
          "data"
        ],
        "model": [
          "model"
        ]
      },
      "input": {
        "data": {
          "validate_data": [
            "homo_data_split_0.validate_data"
          ],
          "train_data": [
            "homo_data_split_0.train_data"
          ]
        }
      },
      "module": "HomoSecureboost"
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
      "backend": 2,
      "work_mode": 1,
      "spark_run": {
        "num-executors": 2,
        "executor-cores": 1,
        "total-executor-cores": 2
      }
    }
  },
  "component_parameters": {
    "common": {
      "homo_data_split_0": {
        "validate_size": %s,
        "split_points": [
          0,
          %s
        ],
        "test_size": 0,
        "stratified": true
      },
      "dataio_0": {
        "with_label": true,
        "output_format": "dense",
		"label_type": "int",
		"label_name": "%s"
      },
      "HomoSecureboost_0": {
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
      "evaluation_0": {
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
