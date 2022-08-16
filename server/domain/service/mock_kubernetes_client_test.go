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
	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	clientgo "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type mockK8sClient struct {
	GetClientSetFn      func() clientgo.Interface
	GetConfigFn         func() (*rest.Config, error)
	ApplyOrDeleteYAMLFn func(yamlStr string, delete bool) error
}

func (m *mockK8sClient) GetClientSet() clientgo.Interface {
	if m.GetClientSetFn != nil {
		return m.GetClientSetFn()
	}
	return &fake.Clientset{}
}

func (m *mockK8sClient) GetConfig() (*rest.Config, error) {
	if m.GetConfigFn != nil {
		return m.GetConfigFn()
	}
	//TODO implement me
	panic("implement me")
}

func (m *mockK8sClient) ApplyOrDeleteYAML(yamlStr string, delete bool) error {
	if m.ApplyOrDeleteYAMLFn != nil {
		return m.ApplyOrDeleteYAMLFn(yamlStr, delete)
	}
	return nil
}

var _ kubernetes.Client = (*mockK8sClient)(nil)
