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
	"github.com/FederatedAI/FedLCM/pkg/kubefate"
	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	appv1 "k8s.io/api/apps/v1"
)

type mockKubeFATEClient struct {
}

func (m *mockKubeFATEClient) CheckVersion() (string, error) {
	return "1.4.2", nil
}

func (m *mockKubeFATEClient) EnsureChartExist(string, string, []byte) error {
	return nil
}

func (m *mockKubeFATEClient) ListClusterByNamespace(string) ([]*modules.Cluster, error) {
	return nil, nil
}

func (m *mockKubeFATEClient) SubmitClusterInstallationJob(string) (string, error) {
	return "test-job-id", nil
}

func (m *mockKubeFATEClient) SubmitClusterUpdateJob(string) (string, error) {
	return "test-job-id", nil
}

func (m *mockKubeFATEClient) SubmitClusterDeletionJob(string) (string, error) {
	return "test-job-id", nil
}

func (m *mockKubeFATEClient) GetJobInfo(string) (*modules.Job, error) {
	return &modules.Job{
		ClusterId: "test-cluster-id",
		Status:    modules.JobStatusSuccess,
	}, nil
}

func (m *mockKubeFATEClient) WaitJob(string) (*modules.Job, error) {
	return &modules.Job{
		ClusterId: "test-cluster-id",
		Status:    modules.JobStatusSuccess,
	}, nil
}

func (m *mockKubeFATEClient) WaitClusterUUID(string) (string, error) {
	return "test-cluster-id", nil
}

func (m *mockKubeFATEClient) IngressAddress() string {
	return "test-ingress-address"
}

func (m *mockKubeFATEClient) IngressRuleHost() string {
	return "test-ingress-rule-host"
}

func (m *mockKubeFATEClient) StopJob(string) error {
	return nil
}

var _ kubefate.Client = (*mockKubeFATEClient)(nil) // TODO: add stubs

type mockKubeFATEManager struct {
	InstallFn   func() error
	UninstallFn func() error
	K8sClientFn func() kubernetes.Client
}

func (m *mockKubeFATEManager) Install(bool) error {
	if m.InstallFn != nil {
		return m.InstallFn()
	}
	return nil
}

func (m *mockKubeFATEManager) K8sClient() kubernetes.Client {
	if m.K8sClientFn != nil {
		return m.K8sClientFn()
	}
	return &mockK8sClient{}
}

func (m *mockKubeFATEManager) Uninstall() error {
	if m.UninstallFn != nil {
		return m.UninstallFn()
	}
	return nil
}

func (m *mockKubeFATEManager) BuildClient() (kubefate.Client, error) {
	return &mockKubeFATEClient{}, nil
}

func (m *mockKubeFATEManager) BuildPFClient() (kubefate.Client, func(), error) {
	return &mockKubeFATEClient{}, nil, nil
}

func (m *mockKubeFATEManager) InstallIngressNginxController() error {
	return nil
}

func (m *mockKubeFATEManager) GetKubeFATEDeployment() (*appv1.Deployment, error) {
	return nil, nil
}

var _ kubefate.Manager = (*mockKubeFATEManager)(nil)
