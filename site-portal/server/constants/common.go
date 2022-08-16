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

package constants

const (
	APIVersion    = "v1"
	JobComponents = `[{
		"groupName": "Data Input and Output",
		"modules": [{
				"moduleName": "DataIO",
				"parameters": {
					"input_format": {
                      "drop_down_box": ["dense", "sparse", "tag"]
                    },
					"delimitor": ",",
					"data_type": {
                      "drop_down_box": ["float", "float64", "int", "int64", "str", "long"]
                    },
					"exclusive_data_type": {},
					"tag_with_value": false,
					"tag_value_delimitor": ":",
					"missing_fill": false,
					"default_value": 0,
					"missing_fill_method": "",
					"missing_impute": [],
					"outlier_replace": false,
					"outlier_replace_method": "",
					"outlier_impute": [],
					"outlier_replace_value": [],
					"with_label": true,
					"label_name": "y",
					"label_type": {
                      "drop_down_box": ["int", "int64", "float", "float64", "str", "long"]
                    },
					"output_format": "dense"
				},
				"conditions": {
					"possible_input": ["Reader", "DataIO"],
					"can_be_endpoint": false
				},
				"input": {
					"data": ["data"],
					"model": ["model"]
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			{
				"moduleName": "HomoDataSplit",
				"parameters": {
					"random_state": "",
					"test_size": 0.0,
					"train_size": 0.8,
					"validate_size": 0.2,
					"stratified": false,
					"shuffle": true,
					"split_points": [],
					"need_run": true
				},
				"conditions": {
					"possible_input": ["DataIO", "HomoOneHotEncoder"],
					"can_be_endpoint": false
				},
				"input": {
					"data": ["data"],
					"model": []
				},
				"output": {
					"data": ["train_data", "validate_data", "test_data"],
					"model": []
				}
			}
		]
	},
	{
		"groupName": "Feature Engineering",
		"modules": [{
			"moduleName": "HomoOneHotEncoder",
			"parameters": {
				"transform_col_indexes": -1,
				"need_run": true,
				"need_alignment": true
			},
			"conditions": {
				"possible_input": ["DataIO"],
				"can_be_endpoint": false
			},
			"input": {
				"data": ["data"],
				"model": ["model"]
			},
			"output": {
				"data": ["data"],
				"model": ["model"]
			}
		}]
	},
	{
		"groupName": "Homogeneous Algorithms",
		"modules": [{
				"moduleName": "HomoLR",
				"parameters": {
					"penalty": {
                      "drop_down_box": ["L2", "L1", "None"]
                    },
					"tol": 1e-4,
					"alpha": 1.0,
					"optimizer": {
                      "drop_down_box": ["rmsprop", "sgd", "adam", "nesterov_momentum_sgd", "adagrad"]
                    },
					"batch_size": -1,
					"learning_rate": 0.01,
					"max_iter": 100,
					"early_stop": {
                      "drop_down_box": ["diff", "weight_diff", "abs"]
                    },
					"decay": 1,
					"decay_sqrt": true,
					"encrypt_param": {},
					"predict_param": {},
					"cv_param": {
						"n_splits": 4,
						"shuffle": true,
						"random_seed": 33,
						"need_cv": false
					},
					"multi_class": {
                      "drop_down_box": ["ovr"]
                    },
					"validation_freqs": "",
					"early_stopping_rounds": "",
					"metrics": "",
					"use_first_metric_only": false,
					"floating_point_precision": ""
				},
				"conditions": {
					"possible_input": ["DataIO", "HomoOneHotEncoder", "HomoDataSplit"],
					"can_be_endpoint": true
				},
				"input": {
					"data": ["data", "train_data", "validate_data"],
					"model": ["model"]
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			{
				"moduleName": "HomoSecureboost",
				"parameters": {
					"task_type": "classification",
					"objective_param": {
						"objective": "cross_entropy"
					},
					"learning_rate": 0.3,
					"num_trees": 5,
					"subsample_feature_rate": 1.0,
					"n_iter_no_change": true,
					"bin_num": 32,
					"validation_freqs": 1,
					"tree_param": {
						"max_depth": 3
					}
				},
				"conditions": {
					"possible_input": ["DataIO", "HomoOneHotEncoder", "HomoDataSplit"],
					"can_be_endpoint": true
				},
				"input": {
					"data": ["data", "train_data", "validate_data"],
					"model": ["model"]
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			}
		]
	},
	{
		"groupName": "Evaluation",
		"modules": [{
			"moduleName": "Evaluation",
			"parameters": {
				"eval_type": {
                  "drop_down_box": ["binary", "regression"]
                },
				"unfold_multi_result": false,
				"pos_label": "1",
				"need_run": true
			},
			"conditions": {
				"possible_input": ["HomoLR", "HomoSecureboost"],
				"can_be_endpoint": true
			},
			"input": {
				"data": ["data"],
				"model": []
			},
			"output": {
				"data": ["data"],
				"model": []
			}
		}]
	}
]`
)

var (
	// Branch is the source branch
	Branch string

	// Commit is the commit number
	Commit string

	// BuildTime is the compiling time
	BuildTime string
)
