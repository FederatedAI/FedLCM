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
	"github.com/FederatedAI/FedLCM/server/domain/repo"
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
			ChartUUID:   "242bf84c-548c-43d4-9f34-15f6d4dc0f33", // from the chart test repo
			Name:        "test-exchange",
			Namespace:   "test-ns",
			ServiceType: entity.ParticipantDefaultServiceTypeLoadBalancer,
		},
		ParticipantDeploymentBaseInfo: ParticipantDeploymentBaseInfo{
			Description:  "",
			EndpointUUID: "",
			DeploymentYAML: `chartName: fate-exchange
chartVersion: v1.9.1-fedlcm-v0.2.0
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

func TestParticipantFATEService_GetClusterDeploymentYAML(t *testing.T) {
	type fields struct {
		ParticipantFATERepo repo.ParticipantFATERepository
		ParticipantService  ParticipantService
	}
	type args struct {
		req *ParticipantFATEClusterYAMLCreationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "External",
			fields: fields{
				ParticipantFATERepo: &mock.ParticipantFATERepoMock{
					GetExchangeByFederationUUIDFn: func(uuid string) (interface{}, error) {
						return &entity.ParticipantFATE{
							Participant: entity.Participant{},
							PartyID:     0,
							Type:        0,
							Status:      entity.ParticipantFATEStatusActive,
							CertConfig:  entity.ParticipantFATECertConfig{},
							AccessInfo: map[entity.ParticipantFATEServiceName]entity.ParticipantModulesAccess{
								entity.ParticipantFATEServiceNameNginx: {
									ServiceType: "",
									Host:        "127.0.1.1",
									Port:        9370,
									TLS:         false,
									FQDN:        "",
								},
								entity.ParticipantFATEServiceNameATS: {
									ServiceType: "",
									Host:        "127.0.1.2",
									Port:        6651,
									TLS:         false,
									FQDN:        "",
								},
							},
							IngressInfo: map[string]entity.ParticipantFATEIngress{},
						}, nil
					},
				},
				ParticipantService: ParticipantService{
					FederationRepo:     &mock.FederationFATERepoMock{},
					ChartRepo:          &gorm.ChartMockRepo{},
					EventService:       &mockEventServiceInt{},
					CertificateService: &mockParticipantFATECertificateServiceInt{},
					EndpointService:    &mockParticipantFATEEndpointServiceInt{},
				},
			},
			args: args{
				req: &ParticipantFATEClusterYAMLCreationRequest{
					ParticipantFATEExchangeYAMLCreationRequest: ParticipantFATEExchangeYAMLCreationRequest{
						ChartUUID:   "7a51112a-b0ad-4c26-b2c0-1e6f7eca6073", // from the chart test repo
						Name:        "test-fate",
						Namespace:   "test-fate-ns",
						ServiceType: entity.ParticipantDefaultServiceTypeNodePort,
					},
					FederationUUID:    "test",
					PartyID:           8888,
					EnablePersistence: false,
					StorageClass:      "",
					ExternalSpark: ExternalSpark{
						Enable:                true,
						Cores_per_node:        8,
						Nodes:                 1,
						Master:                "spark://127.0.0.1:7077",
						DriverHost:            "127.0.1.1",
						DriverHostType:        "NodePort",
						PortMaxRetries:        10,
						DriverStartPort:       30100,
						BlockManagerStartPort: 30200,
						PysparkPython:         "",
					},
					ExternalHDFS: ExternalHDFS{
						Enable:      true,
						Name_node:   "hdfs://127.0.0.1:9000",
						Path_prefix: "",
					},
					ExternalPulsar: ExternalPulsar{
						Enable:   true,
						Host:     "127.0.0.1",
						Mng_port: 8001,
						Port:     6650,
						SSLPort:  6651,
					},
				},
			},
			want: `name: test-fate
namespace: test-fate-ns
chartName: fate
chartVersion: v1.8.0
partyId: 8888
# imageTag: "1.8.0-release"
persistence: false
# pullPolicy:
podSecurityPolicy:
  enabled: false

# ingressClassName: nginx

modules:
  - mysql
  - python
  - fateboard
  - client
  - nginx

backend: spark_pulsar

ingress:
  fateboard:
    hosts:
    - name: test-fate.fateboard.test.example.com
  client:
    hosts:
    - name: test-fate.notebook.test.example.com

nginx:
  type: NodePort
  exchange:
    ip: 127.0.1.1
    httpPort: 9370
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort: 30093
  # grpcNodePort: 30098
pulsar:
  exchange:
    ip: 127.0.1.2
    port: 6651
    domain: test.example.com

mysql:
  size: 1Gi
  storageClass: 
  existingClaim: ""
  accessMode: ReadWriteOnce
  subPath: "mysql"
  # nodeSelector:
  # tolerations:
  # affinity:
  # ip: mysql
  # port: 3306
  # database: eggroll_meta
  # user: fate
  # password: fate_dev

python:
  size: 10Gi
  storageClass: 
  existingClaim: ""
  accessMode: ReadWriteOnce
  # httpNodePort:
  # grpcNodePort:
  # loadBalancerIP:
  # serviceAccountName: ""
  # nodeSelector:
  # tolerations:
  # affinity:
  # resources:
  # logLevel: INFO
  spark: 
    cores_per_node: 8
    nodes: 1
    master: spark://127.0.0.1:7077
    driverHost: 127.0.1.1
    driverHostType: NodePort
    portMaxRetries: 10
    driverStartPort: 30100
    blockManagerStartPort: 30200
    pysparkPython: 
  hdfs:
    name_node: hdfs://127.0.0.1:9000
    path_prefix: 
  pulsar:
    host: 127.0.0.1
    mng_port: 8001
    port: 6650
    ssl_port: 6651
  # nginx:
    # host: nginx
    # http_port: 9300
    # grpc_port: 9310

client:
  size: 1Gi
  storageClass: 
  existingClaim: ""
  accessMode: ReadWriteOnce
  subPath: "client"
  # nodeSelector:
  # tolerations:
  # affinity:`,
			wantErr: false,
		},
		{
			name: "Internal",
			fields: fields{
				ParticipantFATERepo: &mock.ParticipantFATERepoMock{
					GetExchangeByFederationUUIDFn: func(uuid string) (interface{}, error) {
						return &entity.ParticipantFATE{
							Participant: entity.Participant{},
							PartyID:     0,
							Type:        0,
							Status:      entity.ParticipantFATEStatusActive,
							CertConfig:  entity.ParticipantFATECertConfig{},
							AccessInfo: map[entity.ParticipantFATEServiceName]entity.ParticipantModulesAccess{
								entity.ParticipantFATEServiceNameNginx: {
									ServiceType: "",
									Host:        "127.0.1.1",
									Port:        9370,
									TLS:         false,
									FQDN:        "",
								},
								entity.ParticipantFATEServiceNameATS: {
									ServiceType: "",
									Host:        "127.0.1.2",
									Port:        6651,
									TLS:         false,
									FQDN:        "",
								},
							},
							IngressInfo: map[string]entity.ParticipantFATEIngress{},
						}, nil
					},
				},
				ParticipantService: ParticipantService{
					FederationRepo:     &mock.FederationFATERepoMock{},
					ChartRepo:          &gorm.ChartMockRepo{},
					EventService:       &mockEventServiceInt{},
					CertificateService: &mockParticipantFATECertificateServiceInt{},
					EndpointService:    &mockParticipantFATEEndpointServiceInt{},
				},
			},
			args: args{
				req: &ParticipantFATEClusterYAMLCreationRequest{
					ParticipantFATEExchangeYAMLCreationRequest: ParticipantFATEExchangeYAMLCreationRequest{
						ChartUUID:   "8d1b15c1-cc7e-460b-8563-fa732457a049", // from the chart test repo
						Name:        "test-fate",
						Namespace:   "test-fate-ns",
						ServiceType: entity.ParticipantDefaultServiceTypeNodePort,
					},
					FederationUUID:    "test",
					PartyID:           7777,
					EnablePersistence: false,
					StorageClass:      "",
					ExternalSpark:     ExternalSpark{},
					ExternalHDFS:      ExternalHDFS{},
					ExternalPulsar:    ExternalPulsar{},
				},
			},
			want: `name: test-fate
namespace: test-fate-ns
chartName: fate
chartVersion: v1.9.1-fedlcm-v0.2.0
partyId: 7777
persistence: false
# pullPolicy: IfNotPresent
podSecurityPolicy:
  enabled: false

modules:
  - mysql
  - python
  - fateboard
  - client
  - spark
  - hdfs
  - pulsar
  - nginx
  - frontend
  - sitePortalServer
  - postgres

computing: Spark
federation: Pulsar
storage: HDFS
algorithm: Basic
device: CPU

skippedKeys:
- route_table

ingress:
  fateboard:
    hosts:
    - name: test-fate.fateboard.test.example.com
  client:
    hosts:
    - name: test-fate.notebook.test.example.com
  spark:
    hosts:
    - name: test-fate.spark.test.example.com
  pulsar:
    hosts:
    - name: test-fate.pulsar.test.example.com

python:
  # type: ClusterIP
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP:
  # serviceAccountName: ""
  # resources:
  # nodeSelector:
  # tolerations:
  # affinity:
  # logLevel: INFO
  existingClaim: ""
  storageClass: 
  accessMode: ReadWriteOnce
  size: 10Gi
  # resources:
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"
  spark: 
    cores_per_node: 20
    nodes: 2
    master: spark://spark-master:7077
    driverHost:
    driverHostType:
    portMaxRetries:
    driverStartPort:
    blockManagerStartPort:
    pysparkPython:
  hdfs:
    name_node: hdfs://namenode:9000
    path_prefix:
  pulsar:
    host: pulsar
    mng_port: 8080
    port: 6650
  nginx:
    host: nginx
    http_port: 9300
    grpc_port: 9310

fateboard: 
  type: ClusterIP
  username: admin
  password: admin

client:
  subPath: "client"
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: 
  # nodeSelector:
  # tolerations:
  # affinity:

mysql:
  subPath: "mysql"
  size: 1Gi
  storageClass: 
  existingClaim: ""
  accessMode: ReadWriteOnce
  # nodeSelector:
  # tolerations:
  # affinity:
  # ip: mysql
  # port: 3306
  # database: eggroll_meta
  # user: fate
  # password: fate_dev
spark:
  master:
    # image: "federatedai/spark-master"
    # imageTag: "1.9.1-release"
    replicas: 1
    # resources:
      # requests:
        # cpu: "1"
        # memory: "2Gi"
      # limits:
        # cpu: "1"
        # memory: "2Gi"
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
  worker:
    # image: "federatedai/spark-worker"
    # imageTag: "1.9.1-release"
    replicas: 2
    # resources:
      # requests:
        # cpu: "2"
        # memory: "4Gi"
      # limits:
        # cpu: "4"
        # memory: "8Gi"
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
hdfs:
  namenode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: 
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: 
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
nginx:
  type: NodePort
  exchange:
    ip: 127.0.1.1
    httpPort: 9370
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort:
  # grpcNodePort:
pulsar:
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: 
  publicLB:
    enabled: true
  exchange:
    ip: 127.0.1.2
    port: 6651
    domain: test.example.com
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 
  # httpsNodePort: 
  # loadBalancerIP:
postgres:
  user: site_portal
  password: site_portal
  db: site_portal
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: 
  # type: ClusterIP
  # nodeSelector:
  # tolerations:
  # affinity:
  # user: site_portal
  # password: site_portal
  # db: site_portal
  # subPath: ""

frontend:
  type: NodePort
  type: NodePort
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
 
sitePortalServer:
  existingClaim: ""
  storageClass: 
  accessMode: ReadWriteOnce
  size: 1Gi
  # type: ClusterIP
  # nodeSelector:
  # tolerations:
  # affinity:
  # postgresHost: postgres
  # postgresPort: 5432
  # postgresDb: site_portal
  # postgresUser: site_portal
  # postgresPassword: site_portal
  # adminPassword: admin
  # userPassword: user
  # serverCert: /var/lib/site-portal/cert/server.crt
  # serverKey: /var/lib/site-portal/cert/server.key
  # clientCert: /var/lib/site-portal/cert/client.crt
  # clientKey: /var/lib/site-portal/cert/client.key
  # caCert: /var/lib/site-portal/cert/ca.crt
  # tlsEnabled: 'true'
  # tlsPort: 8443
  tlsCommonName: site-7777.server.test.example.com
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ParticipantFATEService{
				ParticipantFATERepo: tt.fields.ParticipantFATERepo,
				ParticipantService:  tt.fields.ParticipantService,
			}
			got, err := s.GetClusterDeploymentYAML(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParticipantFATEService.GetClusterDeploymentYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParticipantFATEService.GetClusterDeploymentYAML() = `%v`, want `%v`", got, tt.want)
			}
		})
	}
}

func Test_getPulsarInformationFromYAML(t *testing.T) {
	type args struct {
		yamlStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				yamlStr: `algorithm: Basic
chartName: fate
chartVersion: v1.9.0
client:
  accessMode: ReadWriteOnce
  existingClaim: ""
  size: 1Gi
  storageClass: null
  subPath: client
computing: Spark
device: CPU
federation: Pulsar
imagePullSecrets:
- name: registrykeyfate
ingress:
  client:
    hosts:
    - name: fate-10000.notebook.k8s.fate.org
  fateboard:
    hosts:
    - name: fate-10000.fateboard.k8s.fate.org
modules:
- mysql
- python
- fateboard
- client
- nginx
mysql:
  accessMode: ReadWriteOnce
  existingClaim: ""
  size: 1Gi
  storageClass: null
  subPath: mysql
name: fate-10000
namespace: fate-10000
nginx:
  exchange:
    httpPort: 30225
    ip: 192.168.0.1
  type: NodePort
partyId: 10000
persistence: false
podSecurityPolicy:
  enabled: false
pulsar:
  exchange:
    domain: k8s.fate.org
    ip: 192.168.0.1
    port: 30449
python:
  accessMode: ReadWriteOnce
  existingClaim: ""
  hdfs:
    name_node: hdfs://192.168.10.1:9000
    path_prefix: null
  pulsar:
    host: 192.168.10.1
    mng_port: 8001
    port: 6650
    ssl_port: 6651
  size: 10Gi
  spark:
    blockManagerStartPort: 31200
    cores_per_node: 8
    driverHost: 192.168.10.1
    driverHostType: NodePort
    driverStartPort: 31100
    master: spark://192.168.10.1:7077
    nodes: 1
    portMaxRetries: 32
    pysparkPython: null
  storageClass: null
storage: HDFS`,
			},
			want:    "192.168.10.1",
			want1:   6651,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getPulsarInformationFromYAML(tt.args.yamlStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPulsarInformationFromYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPulsarInformationFromYAML() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getPulsarInformationFromYAML() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
