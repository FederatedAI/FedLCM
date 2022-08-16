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
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo/mock"
	"github.com/FederatedAI/FedLCM/server/infrastructure/gorm"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgo "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
)

var getServiceAccessWithFallbackOrig = getServiceAccessWithFallback

func TestCreateExchange_PosWithNewCert(t *testing.T) {
	// stub the util functions that will be called
	ensureNSExisting = func(client kubernetes.Client, namespace string) (bool, error) {
		return true, nil
	}
	getServiceAccessWithFallback = func(client kubernetes.Client, namespace, serviceName, portName string, lbFallbackToNodePort bool) (serviceType corev1.ServiceType, host string, port int, err error) {
		serviceType = corev1.ServiceTypeLoadBalancer
		host = "test-host"
		port = 8080
		return
	}
	createATSSecret = func(client kubernetes.Client, namespace string, caCert *x509.Certificate, serverCert *entity.Certificate, privateKey *rsa.PrivateKey) error {
		return nil
	}
	createTLSSecret = func(client kubernetes.Client, namespace string, serverCert *entity.Certificate, serverPrivateKey *rsa.PrivateKey, clientCert *entity.Certificate, clientPrivateKey *rsa.PrivateKey, caCert *x509.Certificate, secretName string) error {
		return nil
	}

	// restore
	defer func() {
		getServiceAccessWithFallback = getServiceAccessWithFallbackOrig
	}()

	service := ParticipantFATEService{
		ParticipantFATERepo: &mock.ParticipantFATERepoMock{},
		ParticipantService: ParticipantService{
			FederationRepo:     &mock.FederationFATERepoMock{},
			ChartRepo:          &gorm.ChartMockRepo{},
			EventService:       &mockEventServiceInt{},
			CertificateService: &mockParticipantFATECertificateServiceInt{},
			EndpointService:    &mockParticipantFATEEndpointServiceInt{},
		},
	}

	exchange, wg, err := service.CreateExchange(&ParticipantFATEExchangeCreationRequest{
		ParticipantFATEExchangeYAMLCreationRequest: ParticipantFATEExchangeYAMLCreationRequest{
			ChartUUID:   "3ce13cb2-5543-4b01-a5e4-9e4c4baa5973", // from the chart test repo
			Name:        "test-exchange",
			Namespace:   "test-ns",
			ServiceType: entity.ParticipantDefaultServiceTypeLoadBalancer,
		},
		ParticipantDeploymentBaseInfo: ParticipantDeploymentBaseInfo{
			Description:  "",
			EndpointUUID: "",
			DeploymentYAML: `chartName: fate-exchange
chartVersion: v1.6.1-fedlcm-v0.2.0
fmlManagerServer:
  image: federatedai/fml-manager-server
  imageTag: v0.2.0
  type: NodePort
modules:
- trafficServer
- nginx
- postgres
- fmlManagerServer
name: test-exchange
namespace: fate-exchange-test1
nginx:
  route_table: null
  type: NodePort
partyId: 0
podSecurityPolicy:
  enabled: true
trafficServer:
  route_table:
    sni: null
  type: NodePort`,
		},
		FederationUUID: "",
		ProxyServerCertInfo: entity.ParticipantComponentCertInfo{
			BindingMode: entity.CertBindingModeCreate,
			UUID:        "",
			CommonName:  "",
		},
		FMLManagerServerCertInfo: entity.ParticipantComponentCertInfo{
			BindingMode: entity.CertBindingModeCreate,
			UUID:        "",
			CommonName:  "",
		},
		FMLManagerClientCertInfo: entity.ParticipantComponentCertInfo{
			BindingMode: entity.CertBindingModeCreate,
			UUID:        "",
			CommonName:  "",
		},
	})
	assert.NoError(t, err, "positive test should return with no error")
	wg.Wait()
	assert.Equal(t, entity.ParticipantFATEStatusActive, exchange.Status, "exchange status should be active")
	assert.Equal(t, 3, len(exchange.AccessInfo), "exchange with fml-manager should expose 3 services")
}

func TestGetServiceAccessWithFallback_PosFallbackToNodePort(t *testing.T) {
	mockClient := &mockK8sClient{
		GetClientSetFn: func() clientgo.Interface {
			fakeClientSet := &fake.Clientset{}
			fakeClientSet.AddReactor("get", "services", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:     "test-port",
								Protocol: corev1.ProtocolTCP,
								Port:     80,
								NodePort: 30080,
							},
						},
						ClusterIP: "10.0.0.1",
						Type:      corev1.ServiceTypeLoadBalancer,
					},
				}, nil
			})
			fakeClientSet.AddReactor("list", "nodes", func(action clientgotesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.NodeList{
					Items: []corev1.Node{
						{
							Status: corev1.NodeStatus{
								Addresses: []corev1.NodeAddress{
									{
										Type:    corev1.NodeExternalIP,
										Address: "127.0.0.1",
									},
								},
							},
						},
					},
				}, nil
			})
			return fakeClientSet
		},
	}
	svcType, _, _, err := getServiceAccessWithFallbackOrig(mockClient, "test-ns", "test-svc", "test-port", true)
	assert.Equal(t, corev1.ServiceTypeNodePort, svcType, "should be a NodePort service")
	assert.NoError(t, err, "positive test should return no error")
}
