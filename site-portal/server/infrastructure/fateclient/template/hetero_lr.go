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

const heteroLRDSL = `
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
    "Intersection_0": {
      "module": "Intersection",
      "input": {
        "data": {
          "data": [
            "DataTransform_0.data"
          ]
        }
      },
      "output": {
        "data": [
          "data"
        ]
      }
    },
    "HeteroLR_0": {
      "module": "HeteroLR",
      "input": {
        "data": {
          "train_data": [
            "Intersection_0.data"
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
            "HeteroLR_0.data"
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

const heteroLRConf = `
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
      "HeteroLR_0": {
        "penalty": "L2",
        "tol": 0.0001,
        "alpha": 0.01,
        "optimizer": "rmsprop",
        "batch_size": -1,
        "learning_rate": 0.15,
        "init_param": {
          "init_method": "zeros"
        },
        "max_iter": 30,
        "early_stop": "diff",
        "cv_param": {
          "n_splits": 5,
          "shuffle": false,
          "random_seed": 103,
          "need_cv": false
        },
        "sqn_param": {
          "update_interval_L": 3,
          "memory_M": 5,
          "sample_size": 5000,
          "random_seed": null
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
            "with_label": true,
            "label_name": "%s",
            "label_type": "int",
            "output_format": "dense",
            "with_match_id": false
          }
        }
      }
    }
  }
}
`

const heteroLRHeteroDataSplitDSL = `
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
    "Intersection_0": {
      "module": "Intersection",
      "input": {
        "data": {
          "data": [
            "DataTransform_0.data"
          ]
        }
      },
      "output": {
        "data": [
          "data"
        ]
      }
    },
    "HeteroDataSplit_0": {
      "module": "HeteroDataSplit",
      "input": {
        "data": {
          "data": [
            "Intersection_0.data"
          ]
        }
      },
      "output": {
        "data": [
          "train_data",
          "validate_data",
          "test_data"
        ]
      }
    },
    "HeteroLR_0": {
      "module": "HeteroLR",
      "input": {
        "data": {
          "validate_data": [
            "HeteroDataSplit_0.validate_data"
          ],
          "train_data": [
            "HeteroDataSplit_0.train_data"
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
            "HeteroLR_0.data"
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

const heteroLRHeteroDataSplitConf = `
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
      "HeteroDataSplit_0": {
        "validate_size": %s,
        "split_points": [
          0,
            %s
        ],
        "test_size": 0,
        "stratified": true
      },
      "HeteroLR_0": {
        "penalty": "L2",
        "tol": 0.0001,
        "alpha": 0.01,
        "optimizer": "rmsprop",
        "batch_size": -1,
        "learning_rate": 0.15,
        "init_param": {
          "init_method": "zeros"
        },
        "max_iter": 30,
        "early_stop": "diff",
        "cv_param": {
          "n_splits": 5,
          "shuffle": false,
          "random_seed": 103,
          "need_cv": false
        },
        "decay": 1,
        "decay_sqrt": true,
        "multi_class": "ovr",
        "sqn_param": {
          "update_interval_L": 3,
          "memory_M": 5,
          "sample_size": 5000,
          "random_seed": null
        },
        "use_first_metric_only": false
      },
      "Evaluation_0": {
        "eval_type": "binary",
        "need_run": true,
        "pos_label": 1,
        "unfold_multi_result": false
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
            "with_label": true,
            "label_name": "%s",
            "label_type": "int",
            "output_format": "dense",
            "with_match_id": false
          }
        }
      }
    }
  }
}
`
