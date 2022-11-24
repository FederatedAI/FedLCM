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

const heteroSBTDSL = `
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
    "HeteroSecureBoost_0": {
      "module": "HeteroSecureBoost",
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
            "HeteroSecureBoost_0.data"
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

const heteroSBTConf = `
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
      "HeteroSecureBoost_0": {
        "task_type": "classification",
        "objective_param": {
          "objective": "cross_entropy"
        },
        "num_trees": 3,
        "validation_freqs": 1,
        "encrypt_param": {
          "method": "Paillier"
        },
        "tree_param": {
          "max_depth": 3
        }
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

const heteroSBTHeteroDataSplitDSL = `
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
            "Intersection_0.data"
          ]
        }
      },
      "module": "HeteroDataSplit"
    },
    "HeteroSecureBoost_0": {
      "module": "HeteroSecureBoost",
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
            "HeteroSecureBoost_0.data"
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

const heteroSBTHeteroDataSplitConf = `
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
      "HeteroSecureBoost_0": {
        "task_type": "classification",
        "objective_param": {
          "objective": "cross_entropy"
        },
        "num_trees": 3,
        "validation_freqs": 1,
        "encrypt_param": {
          "method": "Paillier"
        },
        "tree_param": {
          "max_depth": 3
        }
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
