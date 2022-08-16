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

const homoLRDSL = `
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
        ],
        "model": [
          "model"
        ]
      }
    },
    "HomoLR_0": {
      "module": "HomoLR",
      "input": {
        "data": {
          "train_data": [
            "dataio_0.data"
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
    "evaluation_0": {
      "module": "Evaluation",
      "input": {
        "data": {
          "data": [
            "HomoLR_0.data"
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

const homoLRConf = `
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
      "use_encrypt": false,
      "spark_run": {
        "num-executors": 1,
        "executor-cores": 1,
        "total-executor-cores": 1
      }
    }
  },
  "component_parameters": {
    "common": {
      "dataio_0": {
        "with_label": true,
        "output_format": "dense",
		"label_type": "int",
    	"label_name": "%s"
      },
      "HomoLR_0": {
        "penalty": "L2",
        "tol": 0.00001,
        "alpha": 0.01,
        "optimizer": "sgd",
        "batch_size": -1,
        "learning_rate": 0.15,
        "init_param": {
          "init_method": "zeros"
        },
        "max_iter": 30,
        "early_stop": "diff",
        "encrypt_param": {
          "method": null
        },
        "cv_param": {
          "n_splits": 4,
          "shuffle": true,
          "random_seed": 33,
          "need_cv": false
        },
        "decay": 1,
        "decay_sqrt": true
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

const homoLRHomoDataSplitDSL = `
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
            "HomoLR_0.data"
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
    "HomoLR_0": {
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
      "module": "HomoLR"
    }
  }
}
`

const homoLRHomoDataSplitConf = `
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
        "num-executors": 1,
        "executor-cores": 1,
        "total-executor-cores": 1
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
      "HomoLR_0": {
        "penalty": "L2",
        "tol": 0.00001,
        "alpha": 0.01,
        "optimizer": "sgd",
        "batch_size": -1,
        "learning_rate": 0.15,
        "init_param": {
          "init_method": "zeros"
        },
        "max_iter": 30,
        "early_stop": "diff",
        "encrypt_param": {
          "method": null
        },
        "cv_param": {
          "n_splits": 4,
          "shuffle": true,
          "random_seed": 33,
          "need_cv": false
        },
        "decay": 1,
        "decay_sqrt": true
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
