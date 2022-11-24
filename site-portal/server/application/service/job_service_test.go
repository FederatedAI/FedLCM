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

package service

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getTwoSiteInput() string {
	input := `{
		"reader_0": {
			"attributeType": "diff",
			"commonAttributes": {},
			"diffAttributes": {
				"guest": {},
                "host_0": {}
			},
			"conditions": {
				"output": {
					"data": ["data"]
				}
			},
			"module": "Reader"
		},
		"DataIO_0": {
			"attributeType": "diff",
			"commonAttributes": {},
			"diffAttributes": {
				"guest": {
					"input_format": "dense",
					"delimitor": ",",
					"data_type": "float64",
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
					"with_label": false,
					"label_name": "y",
					"label_type": "int",
					"output_format": "dense"
				},
				"host_0": {
					"input_format": "dense",
					"delimitor": ",",
					"data_type": "float64",
					"exclusive_data_type": {},
					"tag_with_value": false,
					"tag_value_delimitor": ":",
					"missing_fill": false,
					"default_value": 0,
					"missing_fill_method": "",
					"missing_impute": [],
					"outlier_replace": true,
					"outlier_replace_method": "",
					"outlier_impute": [],
					"outlier_replace_value": [],
					"with_label": false,
					"label_name": "y",
					"label_type": "int",
					"output_format": "dense"
				}
			},
			"conditions": {
				"input": {
					"data": {
						"data": ["reader_0.data"]
					}
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			"module": "DataIO"
		},
		"HomoLR_0": {
			"attributeType": "diff",
			"commonAttributes": {},
			"diffAttributes": {
				"guest": {
					"penalty": "L2",
					"tol": 1e-4,
					"alpha": 1.0,
					"optimizer": "rmsprop",
					"batch_size": -1,
					"learning_rate": 0.01,
					"max_iter": 100,
					"early_stop": "diff",
					"decay": 1,
					"decay_sqrt": true,
					"encrypt_param": "{\"method\": null}",
					"predict_param": "{\"method\": null}",
					"callback_param": "{\"method\": null}",
					"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
					"multi_class": "ovr",
					"validation_freqs": "",
					"early_stopping_rounds": "",
					"metrics": "",
					"use_first_metric_only": false,
					"re_encrypt_batches": 2,
					"aggregate_iters": 1,
					"use_proximal": false,
					"mu": 0.1
				},
				"host_0": {
					"penalty": "L2",
					"tol": 1e-4,
					"alpha": 1.0,
					"optimizer": "rmsprop",
					"batch_size": -1,
					"learning_rate": 0.1,
					"max_iter": 100,
					"early_stop": "diff",
					"decay": 1,
					"decay_sqrt": true,
					"encrypt_param": "{\"method\": null}",
					"predict_param": "{\"method\": null}",
					"callback_param": "{\"method\": null}",
					"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
					"multi_class": "ovr",
					"validation_freqs": "",
					"early_stopping_rounds": "",
					"metrics": "",
					"use_first_metric_only": false,
					"re_encrypt_batches": 2,
					"aggregate_iters": 1,
					"use_proximal": false,
					"mu": 0.1
				}
			},
			"conditions": {
				"input": {
					"data": {
						"data": ["DataIO_0.data"]
					}
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			"module": "HomoLR"
		},
		"Evaluation_0": {
			"attributeType": "common",
			"commonAttributes": {
				"eval_type": "binary",
				"unfold_multi_result": false,
				"pos_label": "1",
				"need_run": true
			},
			"diffAttributes": {},
			"conditions": {
				"input": {
					"data": {
						"data": ["HomoLR_0.data"]
					}
				},
				"output": {
					"data": ["data"]
				}
			},
			"module": "Evaluation"
		}
	}`
	return input
}

func getSingleSiteInput() string {
	input := `{
	"reader_0": {
		"attributeType": "common",
		"commonAttributes": {},
		"diffAttributes": {},
		"conditions": {
			"output": {
				"data": [
					"data"
				]
			}
		},
		"module": "Reader"
	},
	"DataIO_0": {
		"attributeType": "common",
		"commonAttributes": {
			"input_format": "dense",
			"delimitor": ",",
			"data_type": "float64",
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
			"label_type": "int",
			"output_format": "dense"
		},
		"diffAttributes": {},
		"conditions": {
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
		"module": "DataIO"
	},
	"HomoSecureBoost_0": {
		"attributeType": "common",
		"commonAttributes": {
			"task_type": "classification",
			"objective_param": {
				"objective": "cross_entropy"
			},
			"learning_rate": 0.3,
			"num_trees": 5,
			"subsample_feature_rate": 1,
			"n_iter_no_change": true,
			"bin_num": 32,
			"validation_freqs": 1,
			"tree_param": {
				"max_depth": 3
			}
		},
		"diffAttributes": {},
		"conditions": {
			"input": {
				"data": {
					"train_data": [
						"DataIO_0.data"
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
		"module": "HomoSecureBoost"
	},
	"Evaluation_0": {
		"attributeType": "common",
		"commonAttributes": {
			"eval_type": "binary",
			"unfold_multi_result": false,
			"pos_label": 1,
			"need_run": true
		},
		"diffAttributes": {},
		"conditions": {
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
				],
				"model": []
			}
		},
		"module": "Evaluation"
	}
}`
	return input
}

func getThreeSiteInput() string {
	input := `{
		"reader_0": {
			"attributeType": "diff",
			"commonAttributes": {},
			"diffAttributes": {
				"guest": {},
                "host_0": {},
				"host_1": {}
			},
			"conditions": {
				"output": {
					"data": ["data"]
				}
			},
			"module": "Reader"
		},
		"DataIO_0": {
			"attributeType": "diff",
			"commonAttributes": {},
			"diffAttributes": {
				"guest": {
					"input_format": "dense",
					"delimitor": ",",
					"data_type": "float64",
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
					"with_label": false,
					"label_name": "y",
					"label_type": "int",
					"output_format": "dense"
				},
				"host_0": {
					"input_format": "dense",
					"delimitor": ",",
					"data_type": "float64",
					"exclusive_data_type": {},
					"tag_with_value": false,
					"tag_value_delimitor": ":",
					"missing_fill": false,
					"default_value": 0,
					"missing_fill_method": "",
					"missing_impute": [],
					"outlier_replace": true,
					"outlier_replace_method": "",
					"outlier_impute": [],
					"outlier_replace_value": [],
					"with_label": false,
					"label_name": "y",
					"label_type": "int",
					"output_format": "dense"
				},
				"host_1": {
					"input_format": "dense",
					"delimitor": ",",
					"data_type": "float64",
					"exclusive_data_type": {},
					"tag_with_value": false,
					"tag_value_delimitor": ":",
					"missing_fill": false,
					"default_value": 0,
					"missing_fill_method": "",
					"missing_impute": [],
					"outlier_replace": true,
					"outlier_replace_method": "",
					"outlier_impute": [],
					"outlier_replace_value": [],
					"with_label": false,
					"label_name": "y",
					"label_type": "int",
					"output_format": "dense"
				}
			},
			"conditions": {
				"input": {
					"data": {
						"data": ["reader_0.data"]
					}
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			"module": "DataIO"
		},
		"HomoLR_0": {
			"attributeType": "common",
			"diffAttributes": {},
			"commonAttributes": {
				"guest": {
					"penalty": "L2",
					"tol": 1e-4,
					"alpha": 1.0,
					"optimizer": "rmsprop",
					"batch_size": -1,
					"learning_rate": 0.01,
					"max_iter": 100,
					"early_stop": "diff",
					"decay": 1,
					"decay_sqrt": true,
					"encrypt_param": "{\"method\": null}",
					"predict_param": "{\"method\": null}",
					"callback_param": "{\"method\": null}",
					"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
					"multi_class": "ovr",
					"validation_freqs": "",
					"early_stopping_rounds": "",
					"metrics": "",
					"use_first_metric_only": false,
					"re_encrypt_batches": 2,
					"aggregate_iters": 1,
					"use_proximal": false,
					"mu": 0.1
				}
			},
			"conditions": {
				"input": {
					"data": {
						"data": ["DataIO_0.data"]
					}
				},
				"output": {
					"data": ["data"],
					"model": ["model"]
				}
			},
			"module": "HomoLR"
		},
		"Evaluation_0": {
			"attributeType": "common",
			"commonAttributes": {
				"eval_type": "binary",
				"unfold_multi_result": false,
				"pos_label": "1",
				"need_run": true
			},
			"diffAttributes": {},
			"conditions": {
				"input": {
					"data": {
						"data": ["HomoLR_0.data"]
					}
				},
				"output": {
					"data": ["data"]
				}
			},
			"module": "Evaluation"
		}
	}`
	return input
}

func TestGenerateDslFromDag(t *testing.T) {
	input := getTwoSiteInput()
	expected := `
	{
		"components": {
			"DataIO_0": {
				"input": {
					"data": {
						"data": [
							"reader_0.data"
						]
					}
				},
				"module": "DataIO",
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
				"input": {
					"data": {
						"data": [
							"HomoLR_0.data"
						]
					}
				},
				"module": "Evaluation",
				"output": {
					"data": [
						"data"
					]
				}
			},
			"HomoLR_0": {
				"input": {
					"data": {
						"data": [
							"DataIO_0.data"
						]
					}
				},
				"module": "HomoLR",
				"output": {
					"data": [
						"data"
					],
					"model": [
						"model"
					]
				}
			},
			"reader_0": {
				"module": "Reader",
				"output": {
					"data": [
						"data"
					]
				}
			}
		}
	}`
	var expectedStruct, actualStruct map[string]interface{}
	jobApp := JobApp{}
	actual, _ := jobApp.GenerateDslFromDag(input)
	json.Unmarshal([]byte(expected), &expectedStruct)
	json.Unmarshal([]byte(actual), &actualStruct)
	assert.True(t, reflect.DeepEqual(actualStruct, expectedStruct))
}

func TestBuildComponentParametersTwoSites(t *testing.T) {
	jobApp := JobApp{}
	input := getTwoSiteInput()
	expected := `{
	"component_parameters": {
		"common": {
			"Evaluation_0": {
				"eval_type": "binary",
				"need_run": true,
				"pos_label": "1",
				"unfold_multi_result": false
			}
		},
		"role": {
			"guest": {
				"0": {
					"DataIO_0": {
						"data_type": "float64",
						"default_value": 0,
						"delimitor": ",",
						"input_format": "dense",
						"label_name": "y",
						"label_type": "int",
						"missing_fill": false,
						"outlier_replace": false,
						"output_format": "dense",
						"tag_value_delimitor": ":",
						"tag_with_value": false,
						"with_label": false
					},
					"HomoLR_0": {
						"aggregate_iters": 1,
						"alpha": 1,
						"batch_size": -1,
						"callback_param": "{\"method\": null}",
						"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
						"decay": 1,
						"decay_sqrt": true,
						"early_stop": "diff",
						"encrypt_param": "{\"method\": null}",
						"learning_rate": 0.01,
						"max_iter": 100,
						"mu": 0.1,
						"multi_class": "ovr",
						"optimizer": "rmsprop",
						"penalty": "L2",
						"predict_param": "{\"method\": null}",
						"re_encrypt_batches": 2,
						"tol": 0.0001,
						"use_first_metric_only": false,
						"use_proximal": false
					}
				}
			},
			"host": {
				"0": {
					"DataIO_0": {
						"data_type": "float64",
						"default_value": 0,
						"delimitor": ",",
						"input_format": "dense",
						"label_name": "y",
						"label_type": "int",
						"missing_fill": false,
						"outlier_replace": true,
						"output_format": "dense",
						"tag_value_delimitor": ":",
						"tag_with_value": false,
						"with_label": false
					},
					"HomoLR_0": {
						"aggregate_iters": 1,
						"alpha": 1,
						"batch_size": -1,
						"callback_param": "{\"method\": null}",
						"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
						"decay": 1,
						"decay_sqrt": true,
						"early_stop": "diff",
						"encrypt_param": "{\"method\": null}",
						"learning_rate": 0.1,
						"max_iter": 100,
						"mu": 0.1,
						"multi_class": "ovr",
						"optimizer": "rmsprop",
						"penalty": "L2",
						"predict_param": "{\"method\": null}",
						"re_encrypt_batches": 2,
						"tol": 0.0001,
						"use_first_metric_only": false,
						"use_proximal": false
					}
				}
			}
		}
	}
}`
	var expectedStruct, actualStruct map[string]interface{}
	actualStruct, _ = jobApp.buildComponentParameters(input, 1)
	json.Unmarshal([]byte(expected), &expectedStruct)
	assert.True(t, reflect.DeepEqual(actualStruct, expectedStruct))
}

func TestBuildComponentParametersSingleSite(t *testing.T) {
	jobApp := JobApp{}
	input := getSingleSiteInput()
	var actualStruct, expectedStruct map[string]interface{}
	expected := `{
	"component_parameters": {
		"common": {
			"DataIO_0": {
				"data_type": "float64",
				"default_value": 0,
				"delimitor": ",",
				"input_format": "dense",
				"label_name": "y",
				"label_type": "int",
				"missing_fill": false,
				"outlier_replace": false,
				"output_format": "dense",
				"tag_value_delimitor": ":",
				"tag_with_value": false,
				"with_label": true
			},
			"Evaluation_0": {
				"eval_type": "binary",
				"need_run": true,
				"pos_label": 1,
				"unfold_multi_result": false
			},
			"HomoSecureBoost_0": {
				"bin_num": 32,
				"learning_rate": 0.3,
				"n_iter_no_change": true,
				"num_trees": 5,
				"objective_param": {
					"objective": "cross_entropy"
				},
				"subsample_feature_rate": 1,
				"task_type": "classification",
				"tree_param": {
					"max_depth": 3
				},
				"validation_freqs": 1
			},
			"reader_0": {}
		},
		"role": {
			"guest": {
				"0": {}
			},
			"host": {}
		}
	}
}`
	actualStruct, _ = jobApp.buildComponentParameters(input, 0)
	json.Unmarshal([]byte(expected), &expectedStruct)
	assert.True(t, reflect.DeepEqual(actualStruct, expectedStruct))
}

func TestBuildComponentParametersThreeSites(t *testing.T) {
	jobApp := JobApp{}
	input := getThreeSiteInput()
	var actualStruct, expectedStruct map[string]interface{}
	expected := `{
	"component_parameters": {
		"common": {
			"Evaluation_0": {
				"eval_type": "binary",
				"need_run": true,
				"pos_label": "1",
				"unfold_multi_result": false
			},
			"HomoLR_0": {
				"guest": {
					"aggregate_iters": 1,
					"alpha": 1,
					"batch_size": -1,
					"callback_param": "{\"method\": null}",
					"cv_param": "{\"n_splits\": 4, \"shuffle\": true, \"random_seed\": 33, \"need_cv\": false}",
					"decay": 1,
					"decay_sqrt": true,
					"early_stop": "diff",
					"early_stopping_rounds": "",
					"encrypt_param": "{\"method\": null}",
					"learning_rate": 0.01,
					"max_iter": 100,
					"metrics": "",
					"mu": 0.1,
					"multi_class": "ovr",
					"optimizer": "rmsprop",
					"penalty": "L2",
					"predict_param": "{\"method\": null}",
					"re_encrypt_batches": 2,
					"tol": 0.0001,
					"use_first_metric_only": false,
					"use_proximal": false,
					"validation_freqs": ""
				}
			}
		},
		"role": {
			"guest": {
				"0": {
					"DataIO_0": {
						"data_type": "float64",
						"default_value": 0,
						"delimitor": ",",
						"input_format": "dense",
						"label_name": "y",
						"label_type": "int",
						"missing_fill": false,
						"outlier_replace": false,
						"output_format": "dense",
						"tag_value_delimitor": ":",
						"tag_with_value": false,
						"with_label": false
					}
				}
			},
			"host": {
				"0": {
					"DataIO_0": {
						"data_type": "float64",
						"default_value": 0,
						"delimitor": ",",
						"input_format": "dense",
						"label_name": "y",
						"label_type": "int",
						"missing_fill": false,
						"outlier_replace": true,
						"output_format": "dense",
						"tag_value_delimitor": ":",
						"tag_with_value": false,
						"with_label": false
					}
				},
				"1": {
					"DataIO_0": {
						"data_type": "float64",
						"default_value": 0,
						"delimitor": ",",
						"input_format": "dense",
						"label_name": "y",
						"label_type": "int",
						"missing_fill": false,
						"outlier_replace": true,
						"output_format": "dense",
						"tag_value_delimitor": ":",
						"tag_with_value": false,
						"with_label": false
					}
				}
			}
		}
	}
}`
	actualStruct, _ = jobApp.buildComponentParameters(input, 2)
	json.Unmarshal([]byte(expected), &expectedStruct)
	assert.True(t, reflect.DeepEqual(actualStruct, expectedStruct))
}

func TestIsEmpty(t *testing.T) {
	jobApp := JobApp{}
	// Test String
	output := jobApp.isEmpty("")
	assert.True(t, output)
	output = jobApp.isEmpty("str")
	assert.False(t, output)

	// Test Slice
	var testSlice []interface{}
	output = jobApp.isEmpty(testSlice)
	assert.True(t, output)
	testSlice = append(testSlice, "hello")
	output = jobApp.isEmpty(testSlice)
	assert.False(t, output)

	// Test Map
	testMap := make(map[string]interface{})
	output = jobApp.isEmpty(testMap)
	assert.True(t, output)
	testMap["key"] = make([]interface{}, 0)
	output = jobApp.isEmpty(testMap)
}

func TestFilterEmptyAttributes(t *testing.T) {
	jobApp := JobApp{}
	inputStr := `
	{
		"input_format": "dense",
		"delimitor": ",",
		"data_type": "float64",
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
		"with_label": false,
		"label_name": "y",
		"label_type": "int",
		"output_format": "dense"
	}`
	expectedStr := `
	{
		"input_format": "dense",
		"delimitor": ",",
		"data_type": "float64",
		"tag_with_value": false,
		"tag_value_delimitor": ":",
		"missing_fill": false,
		"default_value": 0,
		"outlier_replace": false,
		"with_label": false,
		"label_name": "y",
		"label_type": "int",
		"output_format": "dense"
	}`
	var inputStruct, expectedStruct map[string]interface{}
	json.Unmarshal([]byte(inputStr), &inputStruct)
	json.Unmarshal([]byte(expectedStr), &expectedStruct)
	actualOutputStruct := jobApp.filterEmptyAttributes(inputStruct)
	assert.True(t, reflect.DeepEqual(expectedStruct, actualOutputStruct))
}

func TestGenerateIndentedJsonStr(t *testing.T) {
	jobApp := JobApp{}
	originalJsonStr := "{\"Evaluation_0\":{\"eval_type\":\"binary\",\"need_run\":true,\"pos_label\":\"1\",\"unfold_multi_result\":false}}"
	expectetStr := "{\n  \"Evaluation_0\": {\n    \"eval_type\": \"binary\",\n    \"need_run\": true,\n    \"pos_label\": \"1\",\n    \"unfold_multi_result\": false\n  }\n}"
	actualStr, _ := jobApp.generateIndentedJsonStr(originalJsonStr)
	assert.Equal(t, expectetStr, actualStr)
}
