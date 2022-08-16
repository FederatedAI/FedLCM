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

package kubernetes

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type kubernetesClient struct {
	dynamicClient dynamic.Interface
	clientSet     *kubernetes.Clientset
}

// NewKubernetesClient creates a new client instance to work with a kubernetes cluster
func NewKubernetesClient(kubeconfigPath string) (*kubernetesClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &kubernetesClient{
		dynamicClient: dynamicClient,
		clientSet:     clientSet,
	}, nil
}

// GetInferenceServiceList uses dynamic client to retrieve the isvc
func (c *kubernetesClient) GetInferenceServiceList() (*unstructured.UnstructuredList, error) {
	isvcRes := schema.GroupVersionResource{
		Group:    "serving.kubeflow.org",
		Version:  "v1beta1",
		Resource: "inferenceservices",
	}
	return c.dynamicClient.Resource(isvcRes).List(context.TODO(), v1.ListOptions{})
}
