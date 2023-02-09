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

package gorm

import (
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/infrastructure/gorm/mock"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// ChartMockRepo is the implementation of the repo.ChartRepository interface
// The data is hard-coded for now as the template to generate the final yaml content
// is hard to be customized by the end users.
type ChartMockRepo struct{}

var _ repo.ChartRepository = (*ChartMockRepo)(nil)

func (r *ChartMockRepo) Create(instance interface{}) error {
	panic("implement me")
}

func (r *ChartMockRepo) List() (interface{}, error) {
	var chartList []entity.Chart
	for uuid, _ := range chartMap {
		if chartMap[uuid].Type == entity.ChartTypeOpenFLDirector ||
			chartMap[uuid].Type == entity.ChartTypeOpenFLEnvoy {
			if !viper.GetBool("lifecyclemanager.experiment.enabled") {
				continue
			}
		}
		chartList = append(chartList, chartMap[uuid])
	}
	return chartList, nil
}

func (r *ChartMockRepo) DeleteByUUID(uuid string) error {
	panic("implement me")
}

func (r *ChartMockRepo) GetByUUID(uuid string) (interface{}, error) {
	if chart, ok := chartMap[uuid]; !ok {
		return nil, errors.New("chart not found")
	} else {
		return &chart, nil
	}
}

func (r *ChartMockRepo) ListByType(instance interface{}) (interface{}, error) {
	t := instance.(entity.ChartType)
	var chartList []entity.Chart
	for uuid, chart := range chartMap {
		if t == chart.Type {
			chartList = append(chartList, chartMap[uuid])
		}
	}
	return chartList, nil
}

var (
	chartMap = map[string]entity.Chart{
		"4ad46829-a827-4632-b169-c8675360321e": {
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "4ad46829-a827-4632-b169-c8675360321e",
			Name:        "chart for FATE exchange v1.10.0",
			Description: "This chart is for deploying FATE exchange v1.10.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.10.0",
			AppVersion:  "v1.10.0",
			Chart: `apiVersion: v1
appVersion: v1.10.0
description: A Helm chart for fate exchange
name: fate-exchange
version: v1.10.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.10.0
partyId: 0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
# pullPolicy:
# persistence: false
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
modules:
  - trafficServer
  - nginx

trafficServer:
  type: {{.ServiceType}}
  route_table: 
    sni:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort:
  # loadBalancerIP:

nginx:
  type: {{.ServiceType}}
  route_table:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP: `,
			Values: `partyId: 1
partyName: fate-exchange

image:
  registry: federatedai
  isThridParty:
  tag: 1.10.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 

podSecurityPolicy:
  enabled: false

persistence:
  enabled: false
  
partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

modules:
  rollsite: 
    include: false
    ip: rollsite
    type: ClusterIP
    nodePort: 30001
    loadBalancerIP: 
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    # partyList is used to configure the cluster information of all parties that join in the exchange deployment mode. (When eggroll was used as the calculation engine at the time)
    partyList:
    # - partyId: 8888
      # partyIp: 192.168.8.1
      # partyPort: 30081
    # - partyId: 10000
      # partyIp: 192.168.10.1
      # partyPort: 30101
  nginx:
    include: false
    type: NodePort
    httpNodePort:  30003
    grpcNodePort:  30008
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table:
      # 10000: 
        # fateflow:
        # - grpc_port: 30102
          # host: 192.168.10.1
          # http_port: 30107
        # proxy:
        # - grpc_port: 30108
          # host: 192.168.10.1
          # http_port: 30103
      # 9999: 
        # fateflow:
        # - grpc_port: 30092
          # host: 192.168.9.1
          # http_port: 30097
        # proxy:
        # - grpc_port: 30098
          # host: 192.168.9.1
          # http_port: 30093
  trafficServer:
    include: false
    type: ClusterIP
    nodePort: 30007
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table: 
      # sni:
      # - fqdn: 10000.fate.org
        # tunnelRoute: 192.168.0.2:30109
      # - fqdn: 9999.fate.org
        # tunnelRoute: 192.168.0.3:30099`,
			ValuesTemplate: `partyId: {{ .partyId }}
partyName: {{ .name }}

image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    enableTLS: {{ .enableTLS | default false }}
    nodePort: {{ .nodePort }}
    partyList:
    {{- range .partyList }}
      - partyId: {{ .partyId }}
        partyIp: {{ .partyIp }}
        partyPort: {{ .partyPort }}
    {{- end }}
    {{- end }}
  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}
  trafficServer:
    include: {{ has "trafficServer" .modules }}
    {{- with .trafficServer }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    nodePort: {{ .nodePort }}
    route_table: 
      sni:
    {{- range .route_table.sni }}
      - fqdn: {{ .fqdn }}
        tunnelRoute: {{ .tunnelRoute }}
    {{- end }}
    {{- end }}`,
			ArchiveContent: nil,
			Private:        false,
		},
		"7a51112a-b0ad-4c26-b2c0-1e6f7eca6073": {
			Model: gorm.Model{
				ID:        2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "7a51112a-b0ad-4c26-b2c0-1e6f7eca6073",
			Name:        "chart for FATE cluster v1.10.0",
			Description: "This is chart for installing FATE cluster v1.10.0",
			Type:        entity.ChartTypeFATECluster,
			ChartName:   "fate",
			Version:     "v1.10.0",
			AppVersion:  "v1.10.0",
			Chart: `apiVersion: v1
appVersion: v1.10.0
description: A Helm chart for fate-training
name: fate
version: v1.10.0
home: https://fate.fedai.org
icon: https://aisp-1251170195.cos.ap-hongkong.myqcloud.com/wp-content/uploads/sites/12/2019/09/logo.png
sources:
  - https://github.com/FederatedAI/KubeFATE
  - https://github.com/FederatedAI/FATE`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.10.0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
partyId: {{.PartyID}}
persistence: {{ .EnablePersistence }}
# pullPolicy:
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
ingressClassName: nginx

modules:
  - mysql
  - python
  - fateboard
  - client
  {{- if not .EnableExternalSpark }}
  - spark
  {{- end }}
  {{- if not .EnableExternalHDFS }}
  - hdfs
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  - pulsar
  {{- end }}
  - nginx

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
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  {{- if not .EnableExternalSpark }}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}
  {{- end }}

python:
  # type: ClusterIP
  # replicas: 1
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP:
  # serviceAccountName: ""
  # nodeSelector:
  # tolerations:
  # affinity:
  # failedTaskAutoRetryTimes:
  # failedTaskAutoRetryDelay:
  # logLevel: INFO
  existingClaim: ""
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  # dependent_distribution: false
  size: 10Gi
  # resources:
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"
  {{- if .EnableExternalSpark }}
  spark: 
    cores_per_node: {{.ExternalSparkCoresPerNode}}
    nodes: {{.ExternalSparkNode}}
    master: {{.ExternalSparkMaster}}
    driverHost: {{.ExternalSparkDriverHost}}
    driverHostType: {{.ExternalSparkDriverHostType}}
    portMaxRetries: {{.ExternalSparkPortMaxRetries}}
    driverStartPort: {{.ExternalSparkDriverStartPort}}
    blockManagerStartPort: {{.ExternalSparkBlockManagerStartPort}}
    pysparkPython: {{.ExternalSparkPysparkPython}}
  {{- else }}
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
  {{- end }}
  {{- if .EnableExternalHDFS }}
  hdfs:
    name_node: {{.ExternalHDFSNamenode}}
    path_prefix: {{.ExternalHDFSPathPrefix}}
  {{- else }}
  hdfs:
    name_node: hdfs://namenode:9000
    path_prefix:
  {{- end }}
  {{- if .EnableExternalPulsar }}
  pulsar:
    host: {{.ExternalPulsarHost}}
    mng_port: {{.ExternalPulsarMngPort}}
    port: {{.ExternalPulsarPort}}
    ssl_port: {{.ExternalPulsarSSLPort}}
    topic_ttl: 3
    cluster: standalone
    tenant: fl-tenant
  {{- else }}
  pulsar:
    host: pulsar
    mng_port: 8080
    port: 6650
    topic_ttl: 3
    cluster: standalone
    tenant: fl-tenant
  {{- end }}
  nginx:
    host: nginx
    http_port: 9300
    grpc_port: 9310
  # hive:
  #   host: 127.0.0.1
  #   port: 10000
  #   auth_mechanism:
  #   username:
  #   password:

fateboard: 
  type: ClusterIP
  username: admin
  password: admin
#   nodeSelector:
#   tolerations:
#   affinity:

client:
# nodeSelector:
# tolerations:
# affinity:
  subPath: "client"
  existingClaim: ""
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  size: 1Gi
# notebook_hashed_password: ""


mysql:
  subPath: "mysql"
  size: 1Gi
  storageClass: {{ .StorageClass }}
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

{{- if not .EnableExternalSpark }}
spark:
  master:
    # image: "federatedai/spark-master"
    # imageTag: "1.10.0-release"
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
    # imageTag: "1.10.0-release"
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
{{- end }}
{{- if not .EnableExternalHDFS }}
hdfs:
  namenode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
{{- end }}
nginx:
  type: {{.ServiceType}}
  exchange:
    ip: {{.ExchangeNginxHost}}
    httpPort: {{.ExchangeNginxPort}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort:
  # grpcNodePort:

{{- if not .EnableExternalPulsar }}
pulsar:
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  publicLB:
    enabled: true
# env:
#   - name: PULSAR_MEM
#     value: "-Xms4g -Xmx4g -XX:MaxDirectMemorySize=8g"
# confs:
#     brokerDeleteInactiveTopicsFrequencySeconds: 60
#     backlogQuotaDefaultLimitGB: 10
#  
# resources:
#   requests:
#     cpu: "2"
#     memory: "4Gi"
#   limits:
#     cpu: "4"
#     memory: "8Gi" 
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 
  # httpsNodePort: 
  # loadBalancerIP:
{{- else }}
pulsar:
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
{{- end }}`,
			Values: `image:
  registry: federatedai
  isThridParty:
  tag: 1.10.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

# Computing : Eggroll, Spark, Spark_local
computing: Eggroll
# Federation: Eggroll(computing: Eggroll), Pulsar/RabbitMQ(computing: Spark/Spark_local)
federation: Eggroll
# Storage: Eggroll(computing: Eggroll), HDFS(computing: Spark), LocalFS(computing: Spark_local)
storage: Eggroll
# Algorithm: Basic, NN
algorithm: Basic
# Device: CPU, IPCL
device: IPCL

istio:
  enabled: false

podSecurityPolicy:
  enabled: false

ingressClassName: nginx

ingress:
  fateboard:
    # annotations:
    hosts:
    - name: fateboard.example.com
      path: /
    tls: []
    # - secretName: my-tls-secret
      # hosts:
        # - fateboard.example.com
  client:
    # annotations:
    hosts:
    - name: notebook.example.com
      path: /
    tls: [] 
  spark:
    # annotations:
    hosts:
    - name: spark.example.com
      path: /
    tls: [] 
  rabbitmq:
    # annotations:
    hosts:
    - name: rabbitmq.example.com
      path: /
    tls: [] 
  pulsar:
    # annotations: 
    hosts:
    - name:  pulsar.example.com
      path: /
    tls: []
    
exchange:
  partyIp: 192.168.1.1
  partyPort: 30001

exchangeList:
- id: 9991
  ip: 192.168.1.1
  port: 30910

partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

persistence:
  enabled: false

modules:
  rollsite: 
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP:
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    polling:
      enabled: false
      
      # type: client
      # server:
        # ip: 192.168.9.1
        # port: 9370
      
      # type: server
      # clientList:
      # - partID: 9999
      # concurrency: 50
      
  lbrollsite:
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP: 
    size: "2M"
    nodeSelector:
    tolerations:
    affinity:
  python: 
    include: true
    replicas: 1
    type: ClusterIP
    httpNodePort: 30097
    grpcNodePort: 30092
    loadBalancerIP: 
    serviceAccountName: 
    nodeSelector:
    tolerations:
    affinity:
    failedTaskAutoRetryTimes:
    failedTaskAutoRetryDelay:
    logLevel: INFO
    # subPath: ""
    existingClaim:
    dependent_distribution: false
    claimName: python-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    clustermanager:
      cores_per_node: 16
      nodes: 2
    spark: 
      cores_per_node: 20
      nodes: 2
      master: spark://spark-master:7077
      driverHost: fateflow
      driverHostType: 
      portMaxRetries: 
      driverStartPort: 
      blockManagerStartPort: 
      pysparkPython: 
    hdfs:
      name_node: hdfs://namenode:9000
      path_prefix:
    rabbitmq:
      host: rabbitmq
      mng_port: 15672
      port: 5672
      user: fate
      password: fate
    pulsar:
      host: pulsar
      port: 6650
      mng_port: 8080      
      topic_ttl: 3
      cluster: standalone
      tenant: fl-tenant  
    nginx:
      host: nginx
      http_port: 9300
      grpc_port: 9310
    hive:
      host:
      port:
      auth_mechanism:
      username:
      password:
  client:
    include: true
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim:
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    notebook_hashed_password: 
  clustermanager:
    include: true
    ip: clustermanager
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
  nodemanager:
    include: true
    replicas: 2
    nodeSelector:
    tolerations:
    affinity:
    sessionProcessorsPerNode: 2
    subPath: "nodemanager"
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    existingClaim:
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"


  mysql: 
    include: true
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    ip: mysql
    port: 3306
    database: eggroll_meta
    user: fate
    password: fate_dev
    subPath: "mysql"
    existingClaim:
    claimName: mysql-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi

  serving:
    ip: 192.168.9.1
    port: 30095
    useRegistry: false
    zookeeper:
      hosts:
        - serving-zookeeper.fate-serving-9999:2181
    use_acl: false
    user: fate
    password: fate

  fateboard:
    include: true
    type: ClusterIP
    username: admin
    password: admin

  spark:
    include: true
    master:
      Image: ""
      ImageTag: ""
      replicas: 1
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30977
    worker:
      Image: ""
      ImageTag: ""
      replicas: 2
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      resources:
        requests:
          cpu: "2"
          memory: "4Gi"
  hdfs:
    include: true
    namenode:
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30900
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
    datanode:
      replicas: 3
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
  nginx:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    httpNodePort: 30093
    grpcNodePort: 30098
    loadBalancerIP: 
    exchange:
      ip: 192.168.10.1
      httpPort: 30003
      grpcPort: 30008
    route_table: 
#      10000: 
#        proxy: 
#        - host: 192.168.10.1 
#          http_port: 30103
#          grpc_port: 30108
#        fateflow:
#        - host: 192.168.10.1  
#          http_port: 30107
#          grpc_port: 30102
  rabbitmq:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    nodePort: 30094
    loadBalancerIP: 
    default_user: fate
    default_pass: fate
    user: fate
    password: fate
    route_table: 
#      10000:
#        host: 192.168.10.1 
#        port: 30104

  pulsar:
    include: true
    nodeSelector:
    tolerations:
    env:
    confs:
    affinity:
    type: ClusterIP
    httpNodePort: 30094
    httpsNodePort: 30099
    loadBalancerIP:
    existingClaim:
    accessMode: ReadWriteOnce
    storageClass:
    size: 1Gi
    publicLB:
      enabled: false
    # exchange:
      # ip: 192.168.10.1
      # port: 30000
      # domain: fate.org
    route_table: 
#      10000:
#        host: 192.168.10.1
#        port: 30104
#        sslPort: 30109
#        proxy: ""
#   

# externalMysqlIp: mysql
# externalMysqlPort: 3306
# externalMysqlDatabase: eggroll_meta
# externalMysqlUser: fate
# externalMysqlPassword: fate_dev`,
			ValuesTemplate: `image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

partyId: {{ .partyId | int64 | toString }}
partyName: {{ .name }}

computing: {{ .computing }}
federation: {{ .federation }}
storage: {{ .storage }}
algorithm: {{ .algorithm }}
device: {{ .device }}

{{- $partyId := (.partyId | int64 | toString) }}

{{- with .ingress }}
ingress:
  {{- with .fateboard }}
  fateboard:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .client }}
  client:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .spark }}
  spark:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .rabbitmq }}
  rabbitmq:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .pulsar }}
  pulsar:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}

{{- end }}

{{- with .istio }}
istio:
  enabled: {{ .enabled | default false }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

ingressClassName: {{ .ingressClassName | default "nginx"}}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

exchangeList:
{{- with .lbrollsite }}
{{- range .exchangeList }}
  - id: {{ .id }}
    ip: {{ .ip }}
    port: {{ .port }}
{{- end }}
{{- end }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    enableTLS: {{ .enableTLS | default false}}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .polling }}
    polling:
      enabled: {{ .enabled }}
      type: {{ .type }}
      {{- with .server }}
      server:
        ip: {{ .ip }}
        port: {{ .port }}
      {{- end }}
      {{- with .clientList }}
      clientList:
{{ toYaml . | indent 6 }}
      {{- end }}
      concurrency: {{ .concurrency }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  lbrollsite:
    include: {{ has "lbrollsite" .modules }}
    {{- with .lbrollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    loadBalancerIP: {{ .loadBalancerIP }}
    nodePort: {{ .nodePort }}
    size: {{ .size | default "2M" }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  python: 
    include: {{ has "python" .modules }}
    {{- with .python }}
    replicas: {{ .replicas | default 1 }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    logLevel: {{ .logLevel | default "INFO" }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    grpcNodePort: {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    dependent_distribution: {{ .dependent_distribution }}
    serviceAccountName: {{ .serviceAccountName }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    failedTaskAutoRetryTimes: {{ .failedTaskAutoRetryTimes | default 5 }}
    failedTaskAutoRetryDelay: {{ .failedTaskAutoRetryDelay | default 60 }}
    existingClaim: {{ .existingClaim  }}
    claimName: {{ .claimName | default "python-data" }}
    storageClass: {{ .storageClass | default "python" }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- with .clustermanager }}
    clustermanager:
      cores_per_node: {{ .cores_per_node }}
      nodes: {{ .nodes }}
    {{- end }}
    {{- with .spark }}

    spark: 
{{ toYaml . | indent 6}}
    {{- end }}
    {{- with .hdfs }}
    hdfs:
      name_node: {{ .name_node }}
      path_prefix: {{ .path_prefix }}
    {{- end }}
    {{- with .pulsar }}
    pulsar:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
      topic_ttl: {{ .topic_ttl }}
      cluster: {{ .cluster }}
      tenant: {{ .tenant }}      
    {{- end }}
    {{- with .rabbitmq }}
    rabbitmq:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
      user: {{ .user }}
      password: {{ .password }}
    {{- end }}
    {{- with .nginx }}
    nginx:
      host: {{ .host }}
      http_port: {{ .http_port }}
      grpc_port: {{ .grpc_port }}
    {{- end }}
    {{- with .hive }}
    hive:
      host: {{ .host }}
      port: {{ .port }}
      auth_mechanism: {{ .auth_mechanism }}
      username: {{ .username }}
      password: {{ .password }}
    {{- end }}
    {{- end }}


  clustermanager: 
    include: {{ has "clustermanager" .modules }}
    {{- with .clustermanager }}
    ip: clustermanager
    type: "ClusterIP"
    enableTLS: {{ .enableTLS | default false }}
  {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  nodemanager:  
    include: {{ has "nodemanager" .modules }}
    {{- with .nodemanager }}
    sessionProcessorsPerNode: {{ .sessionProcessorsPerNode }}
    replicas: {{ .replicas | default 2 }}
    subPath: {{ .subPath }}
    storageClass: {{ .storageClass  | default "nodemanager" }}
    existingClaim: {{ .existingClaim }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  client: 
    include: {{ has "client" .modules }}
    {{- with .client }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass  | default "client" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    notebook_hashed_password: {{ .notebook_hashed_password | default "" }}
    {{- end }}


  mysql: 
    include: {{ has "mysql" .modules }}
    {{- with .mysql }}
    type: {{ .type  | default "ClusterIP" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    ip: {{ .ip | default "mysql" }}
    port: {{ .port | default "3306" }}
    database: {{ .database | default "eggroll_meta" }}
    user: {{ .user | default "fate" }}
    password: {{ .password | default "fate_dev" }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- end }}


  serving:
    ip: {{ .servingIp }}
    port: {{ .servingPort }}
    {{- with .serving }}
    useRegistry: {{ .useRegistry | default false }}
    zookeeper:
{{ toYaml .zookeeper | indent 6 }}
    {{- end}}

  fateboard:
    include: {{ has "fateboard" .modules }}
    {{- with .fateboard }}
    type: {{ .type }}
    username: {{ .username }}
    password: {{ .password }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end}}

  spark:
    include: {{ has "spark" .modules }}
    {{- with .spark }}
    {{- if .master }}
    master:
      Image: "{{ .master.Image }}"
      ImageTag: "{{ .master.ImageTag }}"
      replicas: {{ .master.replicas }}
      {{- with .master.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .master.type }}
      nodePort: {{ .master.nodePort }}
    {{- end }}
    {{- if .worker }}
    worker:
      Image: "{{ .worker.Image }}"
      ImageTag: "{{ .worker.ImageTag }}"
      replicas: {{ .worker.replicas }}
      {{- with .worker.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .worker.type | default "ClusterIP" }}
    {{- end }}
    {{- end }}


  hdfs:
    include: {{ has "hdfs" .modules }}
    {{- with .hdfs }}
    namenode:
      {{- with .namenode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .namenode.type | default "ClusterIP" }}
      nodePort: {{ .namenode.nodePort }}
      existingClaim: {{ .namenode.existingClaim }}
      storageClass: {{ .namenode.storageClass | default "" }}
      accessMode: {{ .namenode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .namenode.size | default "1Gi" }}
    datanode:
      replicas: {{ .datanode.replicas | default 3 }}
      {{- with .datanode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .datanode.type | default "ClusterIP" }}
      existingClaim: {{ .datanode.existingClaim }}
      storageClass: {{ .datanode.storageClass | default "" }}
      accessMode: {{ .datanode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .datanode.size | default "1Gi" }}
    {{- end }}


  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      httpPort: {{ .httpPort }}
      grpcPort: {{ .grpcPort }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  rabbitmq:
    include: {{ has "rabbitmq" .modules }}
    {{- with .rabbitmq }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    default_user: {{ .default_user }}
    default_pass: {{ .default_pass }}
    loadBalancerIP: {{ .loadBalancerIP }}
    user: {{ .user }}
    password: {{ .password }}
    route_table:
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  pulsar:
    include: {{ has "pulsar" .modules }}
    {{- with .pulsar }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .env }}
    env:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .confs }}
    confs:
{{ toYaml . | indent 6 }}
    {{- end }}    
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    httpsNodePort: {{ .httpsNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .publicLB}}
    publicLB:
      enabled: {{ .enabled | default false }}
    {{- end }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      port: {{ .port }}
      domain: {{ .domain | default "fate.org" }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass | default "" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce"  }}
    size: {{ .size | default "1Gi" }}
    {{- end }}
     
externalMysqlIp: {{ .externalMysqlIp }}
externalMysqlPort: {{ .externalMysqlPort }}
externalMysqlDatabase: {{ .externalMysqlDatabase }}
externalMysqlUser: {{ .externalMysqlUser }}
externalMysqlPassword: {{ .externalMysqlPassword }}`,
			ArchiveContent: nil,
			Private:        false,
		},
		"49fdaa3d-d5ad-4218-87cc-d1f023384729": {
			Model: gorm.Model{
				ID:        3,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "49fdaa3d-d5ad-4218-87cc-d1f023384729",
			Name:        "chart for FATE exchange v1.10.0 with fml-manager service v0.3.0",
			Description: "This chart is for deploying FATE exchange v1.10.0 with fml-manager v0.3.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.10.0-fedlcm-v0.3.0",
			AppVersion:  "exchangev1.10.0 & fedlcmv0.3.0",
			Chart: `apiVersion: v1
appVersion: "exchangev1.10.0 & fedlcmv0.3.0"
description: A Helm chart for fate exchange and fml-manager
name: fate-exchange
version: v1.10.0-fedlcm-v0.3.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.10.0-fedlcm-v0.3.0
partyId: 0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
# pullPolicy:
# persistence: false
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
modules:
  - trafficServer
  - nginx
  - postgres
  - fmlManagerServer

trafficServer:
  type: {{.ServiceType}}
  route_table: 
    sni:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort:
  # loadBalancerIP:

nginx:
  type: {{.ServiceType}}
  route_table:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP: 

postgres:
  user: fml_manager
  password: fml_manager
  db: fml_manager
  # nodeSelector:
  # tolerations:
  # affinity:
  # subPath: ""
  # existingClaim: ""
  # storageClass: <your-storage-class>
  # accessMode: ReadWriteOnce
  # size: 1Gi

fmlManagerServer:
  type: {{.ServiceType}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
  # postgresHost: postgres
  # postgresPort: 5432
  # postgresDb: fml_manager
  # postgresUser: fml_manager
  # postgresPassword: fml_manager
  # tlsPort: 8443
  # serverCert: /var/lib/fml_manager/cert/server.crt
  # serverKey: /var/lib/fml_manager/cert/server.key
  # clientCert: /var/lib/fml_manager/cert/client.crt
  # clientKey: /var/lib/fml_manager/cert/client.key
  # caCert: /var/lib/fml_manager/cert/ca.crt
  # tlsEnabled: 'true'`,
			Values: `partyId: 1
partyName: fate-exchange

image:
  registry: federatedai
  isThridParty:
  tag: 1.10.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets:
#  - name:

podSecurityPolicy:
  enabled: false

persistence:
  enabled: false

partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

modules:
  rollsite: 
    include: false
    ip: rollsite
    type: ClusterIP
    nodePort: 30001
    loadBalancerIP:
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    # partyList is used to configure the cluster information of all parties that join in the exchange deployment mode. (When eggroll was used as the calculation engine at the time)
    partyList:
    # - partyId: 8888
      # partyIp: 192.168.8.1
      # partyPort: 30081
    # - partyId: 10000
      # partyIp: 192.168.10.1
      # partyPort: 30101
  nginx:
    include: false
    type: NodePort
    httpNodePort:  30003
    grpcNodePort:  30008
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table:
      # 10000: 
        # fateflow:
        # - grpc_port: 30102
          # host: 192.168.10.1
          # http_port: 30107
        # proxy:
        # - grpc_port: 30108
          # host: 192.168.10.1
          # http_port: 30103
      # 9999: 
        # fateflow:
        # - grpc_port: 30092
          # host: 192.168.9.1
          # http_port: 30097
        # proxy:
        # - grpc_port: 30098
          # host: 192.168.9.1
          # http_port: 30093
  trafficServer:
    include: false
    type: ClusterIP
    nodePort: 30007
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table: 
      # sni:
      # - fqdn: 10000.fate.org
        # tunnelRoute: 192.168.0.2:30109
      # - fqdn: 9999.fate.org
        # tunnelRoute: 192.168.0.3:30099

  postgres:
    include: true
    type: ClusterIP
    image: postgres
    imageTag: 13.3
    # nodeSelector:
    # tolerations:
    # affinity:
    user: fml_manager
    password: fml_manager
    db: fml_manager
    # subPath: ""
    # existingClaim: ""
    # storageClass: ""
    # accessMode: ReadWriteOnce
    # size: 1Gi

  fmlManagerServer:
    include: true
    image: federatedai/fml-manager-server
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP: 192.168.0.1
    postgresHost: postgres
    postgresPort: 5432
    postgresDb: fml_manager
    postgresUser: fml_manager
    postgresPassword: fml_manager
    tlsPort: 8443
    serverCert: /var/lib/fml_manager/cert/server.crt
    serverKey: /var/lib/fml_manager/cert/server.key
    clientCert: /var/lib/fml_manager/cert/client.crt
    clientKey: /var/lib/fml_manager/cert/client.key
    caCert: /var/lib/fml_manager/cert/ca.crt
    tlsEnabled: 'true'`,
			ValuesTemplate: `partyId: {{ .partyId }}
partyName: {{ .name }}

image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    enableTLS: {{ .enableTLS | default false }}
    nodePort: {{ .nodePort }}
    partyList:
    {{- range .partyList }}
      - partyId: {{ .partyId }}
        partyIp: {{ .partyIp }}
        partyPort: {{ .partyPort }}
    {{- end }}
    {{- end }}
  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}
  trafficServer:
    include: {{ has "trafficServer" .modules }}
    {{- with .trafficServer }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    nodePort: {{ .nodePort }}
    route_table: 
      sni:
    {{- range .route_table.sni }}
      - fqdn: {{ .fqdn }}
        tunnelRoute: {{ .tunnelRoute }}
    {{- end }}
    {{- end }}
    
  postgres:
    include: {{ has "postgres" .modules }}
    {{- with .postgres }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    type: {{ .type | default "ClusterIP" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    user: {{ .user }}
    password: {{ .password }}
    db: {{ .db }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode }}
    size: {{ .size }}
    {{- end }}
    
  fmlManagerServer:
    include: {{ has "fmlManagerServer" .modules }}
    {{- with .fmlManagerServer }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    postgresHost: {{ .postgresHost | default "postgres" }}
    postgresPort: {{ .postgresPort | default "5432" }}
    postgresDb: {{ .postgresDb | default "fml_manager" }}
    postgresUser: {{ .postgresUser | default "fml_manager" }}
    postgresPassword: {{ .postgresPassword | default "fml_manager" }}
    tlsPort: {{ .tlsPort | default "8443" }}
    serverCert: {{ .serverCert | default "/var/lib/fml_manager/cert/server.crt" }}
    serverKey: {{ .serverKey | default "/var/lib/fml_manager/cert/server.key" }}
    clientCert: {{ .clientCert| default "/var/lib/fml_manager/cert/client.crt" }}
    clientKey: {{ .clientKey | default "/var/lib/fml_manager/cert/client.key" }}
    caCert: {{ .caCert | default "/var/lib/fml_manager/cert/ca.crt" }}
    tlsEnabled: {{ .tlsEnabled | default "true" }}
    {{- end }}`,
			ArchiveContent: mock.FATEExchange1100WithManagerChartArchiveContent,
			Private:        true,
		},
		"c5380b96-6a9f-4c3e-8991-1ddc73b5813d": {
			Model: gorm.Model{
				ID:        4,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:           "c5380b96-6a9f-4c3e-8991-1ddc73b5813d",
			Name:           "chart for FATE cluster v1.10.0 with site-portal v0.3.0",
			Description:    "This is chart for installing FATE cluster v1.10.0 with site-portal v0.3.0",
			Type:           entity.ChartTypeFATECluster,
			ChartName:      "fate",
			Version:        "v1.10.0-fedlcm-v0.3.0",
			AppVersion:     "fatev1.10.0+fedlcmv0.3.0",
			ArchiveContent: mock.FATE1100WithPortalChartArchiveContent,
			Chart: `apiVersion: v1
appVersion: "fatev1.10.0+fedlcmv0.3.0"
description: Helm chart for FATE and site-portal in FedLCM
name: fate
version: v1.10.0-fedlcm-v0.3.0
home: https://fate.fedai.org
icon: https://aisp-1251170195.cos.ap-hongkong.myqcloud.com/wp-content/uploads/sites/12/2019/09/logo.png
sources:
  - https://github.com/FederatedAI/KubeFATE
  - https://github.com/FederatedAI/FATE`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.10.0-fedlcm-v0.3.0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
partyId: {{.PartyID}}
persistence: {{.EnablePersistence}}
# pullPolicy: IfNotPresent
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
ingressClassName: nginx

modules:
  - mysql
  - python
  - fateboard
  - client
  {{- if not .EnableExternalSpark }}
  - spark
  {{- end }}
  {{- if not .EnableExternalHDFS }}
  - hdfs
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  - pulsar
  {{- end }}
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
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  {{- if not .EnableExternalSpark }}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}
  {{- end }}
  {{- if not true }}
  # TODO: This requires the front-end to pass the value, and the current front-end does not support it yet.
  # example: sitePortalServerTlsEnabled
  frontend:
    hosts:
    - name: {{.Name}}.frontend.{{.Domain}}
  {{- end }}

python:
  # type: ClusterIP
  # replicas: 1
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP:
  # serviceAccountName: ""
  # nodeSelector:
  # tolerations:
  # affinity:
  # failedTaskAutoRetryTimes:
  # failedTaskAutoRetryDelay:
  # logLevel: INFO
  existingClaim: ""
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  # dependent_distribution: false
  size: 10Gi
  # resources:
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"
  {{- if .EnableExternalSpark }}
  spark: 
    cores_per_node: {{.ExternalSparkCoresPerNode}}
    nodes: {{.ExternalSparkNode}}
    master: {{.ExternalSparkMaster}}
    driverHost: {{.ExternalSparkDriverHost}}
    driverHostType: {{.ExternalSparkDriverHostType}}
    portMaxRetries: {{.ExternalSparkPortMaxRetries}}
    driverStartPort: {{.ExternalSparkDriverStartPort}}
    blockManagerStartPort: {{.ExternalSparkBlockManagerStartPort}}
    pysparkPython: {{.ExternalSparkPysparkPython}}
  {{- else }}
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
  {{- end }}
  {{- if .EnableExternalHDFS }}
  hdfs:
    name_node: {{.ExternalHDFSNamenode}}
    path_prefix: {{.ExternalHDFSPathPrefix}}
  {{- else }}
  hdfs:
    name_node: hdfs://namenode:9000
    path_prefix:
  {{- end }}
  {{- if .EnableExternalPulsar }}
  pulsar:
    host: {{.ExternalPulsarHost}}
    mng_port: {{.ExternalPulsarMngPort}}
    port: {{.ExternalPulsarPort}}
    ssl_port: {{.ExternalPulsarSSLPort}}
    topic_ttl: 3
    cluster: standalone
    tenant: fl-tenant
  {{- else }}
  pulsar:
    host: pulsar
    mng_port: 8080
    port: 6650
    topic_ttl: 3
    cluster: standalone
    tenant: fl-tenant
  {{- end }}
  nginx:
    host: nginx
    http_port: 9300
    grpc_port: 9310
  # hive:
  #   host: 127.0.0.1
  #   port: 10000
  #   auth_mechanism:
  #   username:
  #   password:

fateboard: 
  type: ClusterIP
  username: admin
  password: admin
#   nodeSelector:
#   tolerations:
#   affinity:

client:
# nodeSelector:
# tolerations:
# affinity:
  subPath: "client"
  existingClaim: ""
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  size: 1Gi
# notebook_hashed_password: ""


mysql:
  subPath: "mysql"
  size: 1Gi
  storageClass: {{ .StorageClass }}
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

{{- if not .EnableExternalSpark }}
spark:
  master:
    # image: "federatedai/spark-master"
    # imageTag: "1.10.0-release"
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
    # imageTag: "1.10.0-release"
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
{{- end }}
{{- if not .EnableExternalHDFS }}
hdfs:
  namenode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
{{- end }}
nginx:
  type: {{.ServiceType}}
  exchange:
    ip: {{.ExchangeNginxHost}}
    httpPort: {{.ExchangeNginxPort}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort:
  # grpcNodePort:

{{- if not .EnableExternalPulsar }}
pulsar:
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  publicLB:
    enabled: true
# env:
#   - name: PULSAR_MEM
#     value: "-Xms4g -Xmx4g -XX:MaxDirectMemorySize=8g"
# confs:
#     brokerDeleteInactiveTopicsFrequencySeconds: 60
#     backlogQuotaDefaultLimitGB: 10
#  
# resources:
#   requests:
#     cpu: "2"
#     memory: "4Gi"
#   limits:
#     cpu: "4"
#     memory: "8Gi" 
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 
  # httpsNodePort: 
  # loadBalancerIP:
{{- else }}
pulsar:
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
{{- end }}
postgres:
  user: site_portal
  password: site_portal
  db: site_portal
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  # type: ClusterIP
  # nodeSelector:
  # tolerations:
  # affinity:
  # user: site_portal
  # password: site_portal
  # db: site_portal
  # subPath: ""

frontend:
  type: {{.ServiceType}}
  type: {{.ServiceType}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
 
sitePortalServer:
  existingClaim: ""
  storageClass: {{ .StorageClass }}
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
  tlsCommonName: {{.SitePortalTLSCommonName}}
`,
			Values: `
image:
  registry: federatedai
  isThridParty:
  tag: 1.10.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

# Computing : Eggroll, Spark, Spark_local
computing: Eggroll
# Federation: Eggroll(computing: Eggroll), Pulsar/RabbitMQ(computing: Spark/Spark_local)
federation: Eggroll
# Storage: Eggroll(computing: Eggroll), HDFS(computing: Spark), LocalFS(computing: Spark_local)
storage: Eggroll
# Algorithm: Basic, NN
algorithm: Basic
# Device: CPU, IPCL
device: IPCL

istio:
  enabled: false

podSecurityPolicy:
  enabled: false

ingressClassName: nginx

ingress:
  fateboard:
    # annotations:
    hosts:
    - name: fateboard.example.com
      path: /
    tls: []
    # - secretName: my-tls-secret
      # hosts:
        # - fateboard.example.com
  client:
    # annotations:
    hosts:
    - name: notebook.example.com
      path: /
    tls: [] 
  spark:
    # annotations:
    hosts:
    - name: spark.example.com
      path: /
    tls: [] 
  rabbitmq:
    # annotations:
    hosts:
    - name: rabbitmq.example.com
      path: /
    tls: [] 
  pulsar:
    # annotations: 
    hosts:
    - name:  pulsar.example.com
      path: /
    tls: []
  frontend:
    # annotations: 
    hosts:
    - name:  frontend.example.com
      path: /
    tls: []
    
exchange:
  partyIp: 192.168.1.1
  partyPort: 30001

exchangeList:
- id: 9991
  ip: 192.168.1.1
  port: 30910

partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

persistence:
  enabled: false

modules:
  rollsite: 
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP:
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    polling:
      enabled: false
      
      # type: client
      # server:
        # ip: 192.168.9.1
        # port: 9370
      
      # type: server
      # clientList:
      # - partID: 9999
      # concurrency: 50
      
  lbrollsite:
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP: 
    size: "2M"
    nodeSelector:
    tolerations:
    affinity:
  python: 
    include: true
    replicas: 1
    type: ClusterIP
    httpNodePort: 30097
    grpcNodePort: 30092
    loadBalancerIP: 
    serviceAccountName: 
    nodeSelector:
    tolerations:
    affinity:
    failedTaskAutoRetryTimes:
    failedTaskAutoRetryDelay:
    logLevel: INFO
    # subPath: ""
    existingClaim:
    dependent_distribution: false
    claimName: python-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    clustermanager:
      cores_per_node: 16
      nodes: 2
    spark: 
      cores_per_node: 20
      nodes: 2
      master: spark://spark-master:7077
      driverHost: fateflow
      driverHostType: 
      portMaxRetries: 
      driverStartPort: 
      blockManagerStartPort: 
      pysparkPython: 
    hdfs:
      name_node: hdfs://namenode:9000
      path_prefix:
    rabbitmq:
      host: rabbitmq
      mng_port: 15672
      port: 5672
      user: fate
      password: fate
    pulsar:
      host: pulsar
      port: 6650
      mng_port: 8080      
      topic_ttl: 3
      cluster: standalone
      tenant: fl-tenant  
    nginx:
      host: nginx
      http_port: 9300
      grpc_port: 9310
    hive:
      host:
      port:
      auth_mechanism:
      username:
      password:
  client:
    include: true
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim:
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    notebook_hashed_password: 
  clustermanager:
    include: true
    ip: clustermanager
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
  nodemanager:
    include: true
    replicas: 2
    nodeSelector:
    tolerations:
    affinity:
    sessionProcessorsPerNode: 2
    subPath: "nodemanager"
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    existingClaim:
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"


  mysql: 
    include: true
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    ip: mysql
    port: 3306
    database: eggroll_meta
    user: fate
    password: fate_dev
    subPath: "mysql"
    existingClaim:
    claimName: mysql-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi

  serving:
    ip: 192.168.9.1
    port: 30095
    useRegistry: false
    zookeeper:
      hosts:
        - serving-zookeeper.fate-serving-9999:2181
    use_acl: false
    user: fate
    password: fate

  fateboard:
    include: true
    type: ClusterIP
    username: admin
    password: admin

  spark:
    include: true
    master:
      Image: ""
      ImageTag: ""
      replicas: 1
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30977
    worker:
      Image: ""
      ImageTag: ""
      replicas: 2
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      resources:
        requests:
          cpu: "2"
          memory: "4Gi"
  hdfs:
    include: true
    namenode:
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30900
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
    datanode:
      replicas: 3
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
  nginx:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    httpNodePort: 30093
    grpcNodePort: 30098
    loadBalancerIP: 
    exchange:
      ip: 192.168.10.1
      httpPort: 30003
      grpcPort: 30008
    route_table: 
#      10000: 
#        proxy: 
#        - host: 192.168.10.1 
#          http_port: 30103
#          grpc_port: 30108
#        fateflow:
#        - host: 192.168.10.1  
#          http_port: 30107
#          grpc_port: 30102
  rabbitmq:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    nodePort: 30094
    loadBalancerIP: 
    default_user: fate
    default_pass: fate
    user: fate
    password: fate
    route_table: 
#      10000:
#        host: 192.168.10.1 
#        port: 30104

  pulsar:
    include: true
    nodeSelector:
    tolerations:
    env:
    confs:
    affinity:
    type: ClusterIP
    httpNodePort: 30094
    httpsNodePort: 30099
    loadBalancerIP:
    existingClaim:
    accessMode: ReadWriteOnce
    storageClass:
    size: 1Gi
    publicLB:
      enabled: false
    # exchange:
      # ip: 192.168.10.1
      # port: 30000
      # domain: fate.org
    route_table: 
#      10000:
#        host: 192.168.10.1
#        port: 30104
#        sslPort: 30109
#        proxy: ""
#   

  postgres:
    include: false
    image: postgres
    imageTag: 13.3
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP:
    user: site_portal
    password: site_portal
    db: site_portal
    # subPath: ""
    existingClaim: ""
    storageClass: ""
    accessMode: ReadWriteOnce
    size: 1Gi
  
  frontend:
    include: false
    image: federatedai/site-portal-frontend
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    
    # nodePort: 
    # loadBalancerIP:
    
  sitePortalServer:
    include: false
    image: site-portal-server
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP:
    existingClaim: ""
    storageClass: "sitePortalServer"
    accessMode: ReadWriteOnce
    size: 1Gi
    postgresHost: postgres
    postgresPort: 5432
    postgresDb: site_portal
    postgresUser: site_portal
    postgresPassword: site_portal
    adminPassword: admin
    userPassword: user
    serverCert: /var/lib/site-portal/cert/server.crt
    serverKey: /var/lib/site-portal/cert/server.key
    clientCert: /var/lib/site-portal/cert/client.crt
    clientKey: /var/lib/site-portal/cert/client.key
    caCert: /var/lib/site-portal/cert/ca.crt
    tlsEnabled: 'true'
    tlsPort: 8443
    tlsCommonName: site-1.server.example.com

# externalMysqlIp: mysql
# externalMysqlPort: 3306
# externalMysqlDatabase: eggroll_meta
# externalMysqlUser: fate
# externalMysqlPassword: fate_dev    
`,
			ValuesTemplate: `
image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

partyId: {{ .partyId | int64 | toString }}
partyName: {{ .name }}

computing: {{ .computing }}
federation: {{ .federation }}
storage: {{ .storage }}
algorithm: {{ .algorithm }}
device: {{ .device }}

{{- $partyId := (.partyId | int64 | toString) }}

{{- with .ingress }}
ingress:
  {{- with .fateboard }}
  fateboard:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .client }}
  client:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .spark }}
  spark:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .rabbitmq }}
  rabbitmq:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .pulsar }}
  pulsar:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}

  {{- with .frontend }}
  frontend:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}

{{- end }}

{{- with .istio }}
istio:
  enabled: {{ .enabled | default false }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

ingressClassName: {{ .ingressClassName | default "nginx"}}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

exchangeList:
{{- with .lbrollsite }}
{{- range .exchangeList }}
  - id: {{ .id }}
    ip: {{ .ip }}
    port: {{ .port }}
{{- end }}
{{- end }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    enableTLS: {{ .enableTLS | default false}}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .polling }}
    polling:
      enabled: {{ .enabled }}
      type: {{ .type }}
      {{- with .server }}
      server:
        ip: {{ .ip }}
        port: {{ .port }}
      {{- end }}
      {{- with .clientList }}
      clientList:
{{ toYaml . | indent 6 }}
      {{- end }}
      concurrency: {{ .concurrency }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  lbrollsite:
    include: {{ has "lbrollsite" .modules }}
    {{- with .lbrollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    loadBalancerIP: {{ .loadBalancerIP }}
    nodePort: {{ .nodePort }}
    size: {{ .size | default "2M" }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  python: 
    include: {{ has "python" .modules }}
    {{- with .python }}
    replicas: {{ .replicas | default 1 }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    logLevel: {{ .logLevel | default "INFO" }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    grpcNodePort: {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    dependent_distribution: {{ .dependent_distribution }}
    serviceAccountName: {{ .serviceAccountName }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    failedTaskAutoRetryTimes: {{ .failedTaskAutoRetryTimes | default 5 }}
    failedTaskAutoRetryDelay: {{ .failedTaskAutoRetryDelay | default 60 }}
    existingClaim: {{ .existingClaim  }}
    claimName: {{ .claimName | default "python-data" }}
    storageClass: {{ .storageClass | default "python" }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- with .clustermanager }}
    clustermanager:
      cores_per_node: {{ .cores_per_node }}
      nodes: {{ .nodes }}
    {{- end }}
    {{- with .spark }}

    spark: 
{{ toYaml . | indent 6}}
    {{- end }}
    {{- with .hdfs }}
    hdfs:
      name_node: {{ .name_node }}
      path_prefix: {{ .path_prefix }}
    {{- end }}
    {{- with .pulsar }}
    pulsar:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
      topic_ttl: {{ .topic_ttl }}
      cluster: {{ .cluster }}
      tenant: {{ .tenant }}      
    {{- end }}
    {{- with .rabbitmq }}
    rabbitmq:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
      user: {{ .user }}
      password: {{ .password }}
    {{- end }}
    {{- with .nginx }}
    nginx:
      host: {{ .host }}
      http_port: {{ .http_port }}
      grpc_port: {{ .grpc_port }}
    {{- end }}
    {{- with .hive }}
    hive:
      host: {{ .host }}
      port: {{ .port }}
      auth_mechanism: {{ .auth_mechanism }}
      username: {{ .username }}
      password: {{ .password }}
    {{- end }}
    {{- end }}


  clustermanager: 
    include: {{ has "clustermanager" .modules }}
    {{- with .clustermanager }}
    ip: clustermanager
    type: "ClusterIP"
    enableTLS: {{ .enableTLS | default false }}
  {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  nodemanager:  
    include: {{ has "nodemanager" .modules }}
    {{- with .nodemanager }}
    sessionProcessorsPerNode: {{ .sessionProcessorsPerNode }}
    replicas: {{ .replicas | default 2 }}
    subPath: {{ .subPath }}
    storageClass: {{ .storageClass  | default "nodemanager" }}
    existingClaim: {{ .existingClaim }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  client: 
    include: {{ has "client" .modules }}
    {{- with .client }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass  | default "client" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    notebook_hashed_password: {{ .notebook_hashed_password | default "" }}
    {{- end }}


  mysql: 
    include: {{ has "mysql" .modules }}
    {{- with .mysql }}
    type: {{ .type  | default "ClusterIP" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    ip: {{ .ip | default "mysql" }}
    port: {{ .port | default "3306" }}
    database: {{ .database | default "eggroll_meta" }}
    user: {{ .user | default "fate" }}
    password: {{ .password | default "fate_dev" }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- end }}


  serving:
    ip: {{ .servingIp }}
    port: {{ .servingPort }}
    {{- with .serving }}
    useRegistry: {{ .useRegistry | default false }}
    zookeeper:
{{ toYaml .zookeeper | indent 6 }}
    {{- end}}

  fateboard:
    include: {{ has "fateboard" .modules }}
    {{- with .fateboard }}
    type: {{ .type }}
    username: {{ .username }}
    password: {{ .password }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end}}

  spark:
    include: {{ has "spark" .modules }}
    {{- with .spark }}
    {{- if .master }}
    master:
      Image: "{{ .master.Image }}"
      ImageTag: "{{ .master.ImageTag }}"
      replicas: {{ .master.replicas }}
      {{- with .master.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .master.type }}
      nodePort: {{ .master.nodePort }}
    {{- end }}
    {{- if .worker }}
    worker:
      Image: "{{ .worker.Image }}"
      ImageTag: "{{ .worker.ImageTag }}"
      replicas: {{ .worker.replicas }}
      {{- with .worker.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .worker.type | default "ClusterIP" }}
    {{- end }}
    {{- end }}


  hdfs:
    include: {{ has "hdfs" .modules }}
    {{- with .hdfs }}
    namenode:
      {{- with .namenode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .namenode.type | default "ClusterIP" }}
      nodePort: {{ .namenode.nodePort }}
      existingClaim: {{ .namenode.existingClaim }}
      storageClass: {{ .namenode.storageClass | default "" }}
      accessMode: {{ .namenode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .namenode.size | default "1Gi" }}
    datanode:
      replicas: {{ .datanode.replicas | default 3 }}
      {{- with .datanode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .datanode.type | default "ClusterIP" }}
      existingClaim: {{ .datanode.existingClaim }}
      storageClass: {{ .datanode.storageClass | default "" }}
      accessMode: {{ .datanode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .datanode.size | default "1Gi" }}
    {{- end }}


  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      httpPort: {{ .httpPort }}
      grpcPort: {{ .grpcPort }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  rabbitmq:
    include: {{ has "rabbitmq" .modules }}
    {{- with .rabbitmq }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    default_user: {{ .default_user }}
    default_pass: {{ .default_pass }}
    loadBalancerIP: {{ .loadBalancerIP }}
    user: {{ .user }}
    password: {{ .password }}
    route_table:
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  pulsar:
    include: {{ has "pulsar" .modules }}
    {{- with .pulsar }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .env }}
    env:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .confs }}
    confs:
{{ toYaml . | indent 6 }}
    {{- end }}    
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    httpsNodePort: {{ .httpsNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .publicLB}}
    publicLB:
      enabled: {{ .enabled | default false }}
    {{- end }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      port: {{ .port }}
      domain: {{ .domain | default "fate.org" }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass | default "" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce"  }}
    size: {{ .size | default "1Gi" }}
    {{- end }}

  postgres:
    include: {{ has "postgres" .modules }}
    {{- with .postgres }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    user: {{ .user }}
    password: {{ .password }}
    db: {{ .db }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode }}
    size: {{ .size }}
    {{- end }}

  frontend:
    include: {{ has "frontend" .modules }}
    {{- with .frontend }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- end }}

  sitePortalServer:
    include: {{ has "sitePortalServer" .modules }}
    {{- with .sitePortalServer }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode | default "ReadWriteOnce"  }}
    size: {{ .size  | default "1Gi" }}
    postgresHost: {{ .postgresHost | default "postgres" }}
    postgresPort: {{ .postgresPort | default "5432" }}
    postgresDb: {{ .postgresDb | default "site_portal" }}
    postgresUser: {{ .postgresUser | default "site_portal" }}
    postgresPassword: {{ .postgresPassword | default "site_portal" }}
    adminPassword: {{ .adminPassword | default "admin" }}
    userPassword: {{ .userPassword | default "user" }}
    serverCert: {{ .serverCert| default "/var/lib/site-portal/cert/server.crt" }}
    serverKey: {{ .serverKey | default "/var/lib/site-portal/cert/server.key" }}
    clientCert: {{ .clientCert | default "/var/lib/site-portal/cert/client.crt" }}
    clientKey: {{ .clientKey | default "/var/lib/site-portal/cert/client.key" }}
    caCert: {{ .caCert | default "/var/lib/site-portal/cert/ca.crt" }}
    tlsEnabled: {{ .tlsEnabled | default "'true'" }}
    tlsPort: {{ .tlsPort | default "8443" }}
    tlsCommonName: {{ .tlsCommonName | default "site-1.server.example.com" }}
    {{- end }}

externalMysqlIp: {{ .externalMysqlIp }}
externalMysqlPort: {{ .externalMysqlPort }}
externalMysqlDatabase: {{ .externalMysqlDatabase }}
externalMysqlUser: {{ .externalMysqlUser }}
externalMysqlPassword: {{ .externalMysqlPassword }}
`,
			Private: true,
		},
		"242bf84c-548c-43d4-9f34-15f6d4dc0f33": {
			Model: gorm.Model{
				ID:        5,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "242bf84c-548c-43d4-9f34-15f6d4dc0f33",
			Name:        "chart for FATE exchange v1.9.1 with fml-manager service v0.2.0",
			Description: "This chart is for deploying FATE exchange v1.9.1 with fml-manager v0.2.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.9.1-fedlcm-v0.2.0",
			AppVersion:  "exchangev1.9.1 & fedlcmv0.2.0",
			Chart: `apiVersion: v1
appVersion: "exchangev1.9.1 & fedlcmv0.2.0"
description: A Helm chart for fate exchange and fml-manager
name: fate-exchange
version: v1.9.1-fedlcm-v0.2.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.9.1-fedlcm-v0.2.0
partyId: 0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
# pullPolicy:
# persistence: false
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
modules:
  - trafficServer
  - nginx
  - postgres
  - fmlManagerServer

trafficServer:
  type: {{.ServiceType}}
  route_table: 
    sni:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort:
  # loadBalancerIP:

nginx:
  type: {{.ServiceType}}
  route_table:
  # replicas: 1
  # nodeSelector:
  # tolerations:
  # affinity:
  # httpNodePort: 
  # grpcNodePort: 
  # loadBalancerIP: 

postgres:
  user: fml_manager
  password: fml_manager
  db: fml_manager
  # nodeSelector:
  # tolerations:
  # affinity:
  # subPath: ""
  # existingClaim: ""
  # storageClass: <your-storage-class>
  # accessMode: ReadWriteOnce
  # size: 1Gi

fmlManagerServer:
  type: {{.ServiceType}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
  # postgresHost: postgres
  # postgresPort: 5432
  # postgresDb: fml_manager
  # postgresUser: fml_manager
  # postgresPassword: fml_manager
  # tlsPort: 8443
  # serverCert: /var/lib/fml_manager/cert/server.crt
  # serverKey: /var/lib/fml_manager/cert/server.key
  # clientCert: /var/lib/fml_manager/cert/client.crt
  # clientKey: /var/lib/fml_manager/cert/client.key
  # caCert: /var/lib/fml_manager/cert/ca.crt
  # tlsEnabled: 'true'`,
			Values: `partyId: 1
partyName: fate-exchange

image:
  registry: federatedai
  isThridParty:
  tag: 1.9.1-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

podSecurityPolicy:
  enabled: false

persistence:
  enabled: false

partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

modules:
  rollsite: 
    include: false
    ip: rollsite
    type: ClusterIP
    nodePort: 30001
    loadBalancerIP:
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    # partyList is used to configure the cluster information of all parties that join in the exchange deployment mode. (When eggroll was used as the calculation engine at the time)
    partyList:
    # - partyId: 8888
      # partyIp: 192.168.8.1
      # partyPort: 30081
    # - partyId: 10000
      # partyIp: 192.168.10.1
      # partyPort: 30101
  nginx:
    include: false
    type: NodePort
    httpNodePort:  30003
    grpcNodePort:  30008
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table:
      # 10000: 
        # fateflow:
        # - grpc_port: 30102
          # host: 192.168.10.1
          # http_port: 30107
        # proxy:
        # - grpc_port: 30108
          # host: 192.168.10.1
          # http_port: 30103
      # 9999: 
        # fateflow:
        # - grpc_port: 30092
          # host: 192.168.9.1
          # http_port: 30097
        # proxy:
        # - grpc_port: 30098
          # host: 192.168.9.1
          # http_port: 30093
  trafficServer:
    include: false
    type: ClusterIP
    nodePort: 30007
    loadBalancerIP: 
    nodeSelector: 
    tolerations:
    affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table: 
      # sni:
      # - fqdn: 10000.fate.org
        # tunnelRoute: 192.168.0.2:30109
      # - fqdn: 9999.fate.org
        # tunnelRoute: 192.168.0.3:30099

  postgres:
    include: true
    type: ClusterIP
    image: postgres
    imageTag: 13.3
    # nodeSelector:
    # tolerations:
    # affinity:
    user: fml_manager
    password: fml_manager
    db: fml_manager
    # subPath: ""
    # existingClaim: ""
    # storageClass: ""
    # accessMode: ReadWriteOnce
    # size: 1Gi

  fmlManagerServer:
    include: true
    image: federatedai/fml-manager-server
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP: 192.168.0.1
    postgresHost: postgres
    postgresPort: 5432
    postgresDb: fml_manager
    postgresUser: fml_manager
    postgresPassword: fml_manager
    tlsPort: 8443
    serverCert: /var/lib/fml_manager/cert/server.crt
    serverKey: /var/lib/fml_manager/cert/server.key
    clientCert: /var/lib/fml_manager/cert/client.crt
    clientKey: /var/lib/fml_manager/cert/client.key
    caCert: /var/lib/fml_manager/cert/ca.crt
    tlsEnabled: 'true'`,
			ValuesTemplate: `partyId: {{ .partyId }}
partyName: {{ .name }}

image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    enableTLS: {{ .enableTLS | default false }}
    nodePort: {{ .nodePort }}
    partyList:
    {{- range .partyList }}
      - partyId: {{ .partyId }}
        partyIp: {{ .partyIp }}
        partyPort: {{ .partyPort }}
    {{- end }}
    {{- end }}
  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}
  trafficServer:
    include: {{ has "trafficServer" .modules }}
    {{- with .trafficServer }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type }}
    replicas: {{ .replicas }}
    nodePort: {{ .nodePort }}
    route_table: 
      sni:
    {{- range .route_table.sni }}
      - fqdn: {{ .fqdn }}
        tunnelRoute: {{ .tunnelRoute }}
    {{- end }}
    {{- end }}
    
  postgres:
    include: {{ has "postgres" .modules }}
    {{- with .postgres }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    type: {{ .type | default "ClusterIP" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    user: {{ .user }}
    password: {{ .password }}
    db: {{ .db }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode }}
    size: {{ .size }}
    {{- end }}
    
  fmlManagerServer:
    include: {{ has "fmlManagerServer" .modules }}
    {{- with .fmlManagerServer }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    postgresHost: {{ .postgresHost | default "postgres" }}
    postgresPort: {{ .postgresPort | default "5432" }}
    postgresDb: {{ .postgresDb | default "fml_manager" }}
    postgresUser: {{ .postgresUser | default "fml_manager" }}
    postgresPassword: {{ .postgresPassword | default "fml_manager" }}
    tlsPort: {{ .tlsPort | default "8443" }}
    serverCert: {{ .serverCert | default "/var/lib/fml_manager/cert/server.crt" }}
    serverKey: {{ .serverKey | default "/var/lib/fml_manager/cert/server.key" }}
    clientCert: {{ .clientCert| default "/var/lib/fml_manager/cert/client.crt" }}
    clientKey: {{ .clientKey | default "/var/lib/fml_manager/cert/client.key" }}
    caCert: {{ .caCert | default "/var/lib/fml_manager/cert/ca.crt" }}
    tlsEnabled: {{ .tlsEnabled | default "true" }}
    {{- end }}`,
			ArchiveContent: mock.FATEExchange191WithManagerChartArchiveContent,
			Private:        true,
		},
		"8d1b15c1-cc7e-460b-8563-fa732457a049": {
			Model: gorm.Model{
				ID:        6,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:           "8d1b15c1-cc7e-460b-8563-fa732457a049",
			Name:           "chart for FATE cluster v1.9.1 with site-portal v0.2.0",
			Description:    "This is chart for installing FATE cluster v1.9.1 with site-portal v0.2.0",
			Type:           entity.ChartTypeFATECluster,
			ChartName:      "fate",
			Version:        "v1.9.1-fedlcm-v0.2.0",
			AppVersion:     "fatev1.9.1+fedlcmv0.2.0",
			ArchiveContent: mock.FATE191WithPortalChartArchiveContent,
			Chart: `apiVersion: v1
appVersion: "fatev1.9.1+fedlcmv0.2.0"
description: Helm chart for FATE and site-portal in FedLCM
name: fate
version: v1.9.1-fedlcm-v0.2.0
icon: https://aisp-1251170195.cos.ap-hongkong.myqcloud.com/wp-content/uploads/sites/12/2019/09/logo.png
sources:
  - https://github.com/FederatedAI/KubeFATE
  - https://github.com/FederatedAI/FATE`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.9.1-fedlcm-v0.2.0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
partyId: {{.PartyID}}
persistence: {{.EnablePersistence}}
# pullPolicy: IfNotPresent
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}

modules:
  - mysql
  - python
  - fateboard
  - client
  {{- if not .EnableExternalSpark }}
  - spark
  {{- end }}
  {{- if not .EnableExternalHDFS }}
  - hdfs
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  - pulsar
  {{- end }}
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
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  {{- if not .EnableExternalSpark }}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  {{- end }}
  {{- if not .EnableExternalPulsar }}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}
  {{- end }}
  {{- if not true }}
  # TODO: This requires the front-end to pass the value, and the current front-end does not support it yet.
  # example: sitePortalServerTlsEnabled
  frontend:
    hosts:
    - name: {{.Name}}.frontend.{{.Domain}}
  {{- end }}

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
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  size: 10Gi
  # resources:
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"
  {{- if .EnableExternalSpark }}
  spark: 
    cores_per_node: {{.ExternalSparkCoresPerNode}}
    nodes: {{.ExternalSparkNode}}
    master: {{.ExternalSparkMaster}}
    driverHost: {{.ExternalSparkDriverHost}}
    driverHostType: {{.ExternalSparkDriverHostType}}
    portMaxRetries: {{.ExternalSparkPortMaxRetries}}
    driverStartPort: {{.ExternalSparkDriverStartPort}}
    blockManagerStartPort: {{.ExternalSparkBlockManagerStartPort}}
    pysparkPython: {{.ExternalSparkPysparkPython}}
  {{- else }}
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
  {{- end }}
  {{- if .EnableExternalHDFS }}
  hdfs:
    name_node: {{.ExternalHDFSNamenode}}
    path_prefix: {{.ExternalHDFSPathPrefix}}
  {{- else }}
  hdfs:
    name_node: hdfs://namenode:9000
    path_prefix:
  {{- end }}
  {{- if .EnableExternalPulsar }}
  pulsar:
    host: {{.ExternalPulsarHost}}
    mng_port: {{.ExternalPulsarMngPort}}
    port: {{.ExternalPulsarPort}}
    ssl_port: {{.ExternalPulsarSSLPort}}
  {{- else }}
  pulsar:
    host: pulsar
    mng_port: 8080
    port: 6650
  {{- end }}
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
  storageClass: {{ .StorageClass }}
  # nodeSelector:
  # tolerations:
  # affinity:

mysql:
  subPath: "mysql"
  size: 1Gi
  storageClass: {{ .StorageClass }}
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

{{- if not .EnableExternalSpark }}
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
{{- end }}
{{- if not .EnableExternalHDFS }}
hdfs:
  namenode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
{{- end }}
nginx:
  type: {{.ServiceType}}
  exchange:
    ip: {{.ExchangeNginxHost}}
    httpPort: {{.ExchangeNginxPort}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort:
  # grpcNodePort:

{{- if not .EnableExternalPulsar }}
pulsar:
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  publicLB:
    enabled: true
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 
  # httpsNodePort: 
  # loadBalancerIP:
{{- else }}
pulsar:
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
{{- end }}
postgres:
  user: site_portal
  password: site_portal
  db: site_portal
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  # type: ClusterIP
  # nodeSelector:
  # tolerations:
  # affinity:
  # user: site_portal
  # password: site_portal
  # db: site_portal
  # subPath: ""

frontend:
  type: {{.ServiceType}}
  type: {{.ServiceType}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
 
sitePortalServer:
  existingClaim: ""
  storageClass: {{ .StorageClass }}
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
  tlsCommonName: {{.SitePortalTLSCommonName}}
`,
			Values: `
image:
  registry: federatedai
  isThridParty:
  tag: 1.9.1-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

# Computing : Eggroll, Spark, Spark_local
computing: Eggroll
# Federation: Eggroll(computing: Eggroll), Pulsar/RabbitMQ(computing: Spark/Spark_local)
federation: Eggroll
# Storage: Eggroll(computing: Eggroll), HDFS(computing: Spark), LocalFS(computing: Spark_local)
storage: Eggroll
# Algorithm: Basic, NN
algorithm: Basic
# Device: CPU, IPCL
device: IPCL

istio:
  enabled: false

podSecurityPolicy:
  enabled: false

ingressClassName: nginx

ingress:
  fateboard:
    # annotations:
    hosts:
    - name: fateboard.example.com
      path: /
    tls: []
    # - secretName: my-tls-secret
      # hosts:
        # - fateboard.example.com
  client:
    # annotations:
    hosts:
    - name: notebook.example.com
      path: /
    tls: [] 
  spark:
    # annotations:
    hosts:
    - name: spark.example.com
      path: /
    tls: [] 
  rabbitmq:
    # annotations:
    hosts:
    - name: rabbitmq.example.com
      path: /
    tls: [] 
  pulsar:
    # annotations: 
    hosts:
    - name:  pulsar.example.com
      path: /
    tls: []
  frontend:
    # annotations: 
    hosts:
    - name:  frontend.example.com
      path: /
    tls: []
    
exchange:
  partyIp: 192.168.1.1
  partyPort: 30001

exchangeList:
- id: 9991
  ip: 192.168.1.1
  port: 30910

partyList:
- partyId: 8888
  partyIp: 192.168.8.1
  partyPort: 30081
- partyId: 10000
  partyIp: 192.168.10.1
  partyPort: 30101

persistence:
  enabled: false

modules:
  rollsite: 
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP:
    enableTLS: false
    nodeSelector:
    tolerations:
    affinity:
    polling:
      enabled: false
      
      # type: client
      # server:
        # ip: 192.168.9.1
        # port: 9370
      
      # type: server
      # clientList:
      # - partID: 9999
      # concurrency: 50
      
  lbrollsite:
    include: true
    ip: rollsite
    type: ClusterIP
    nodePort: 30091
    loadBalancerIP: 
    size: "2M"
    nodeSelector:
    tolerations:
    affinity:
  python: 
    include: true
    type: ClusterIP
    httpNodePort: 30097
    grpcNodePort: 30092
    loadBalancerIP: 
    serviceAccountName: 
    nodeSelector:
    tolerations:
    affinity:
    logLevel: INFO
    # subPath: ""
    existingClaim:
    claimName: python-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    clustermanager:
      cores_per_node: 16
      nodes: 2
    spark: 
      cores_per_node: 20
      nodes: 2
      master: spark://spark-master:7077
      driverHost: fateflow
      driverHostType: 
      portMaxRetries: 
      driverStartPort: 
      blockManagerStartPort: 
      pysparkPython: 
    hdfs:
      name_node: hdfs://namenode:9000
      path_prefix:
    rabbitmq:
      host: rabbitmq
      mng_port: 15672
      port: 5672
      user: fate
      password: fate
    pulsar:
      host: pulsar
      mng_port: 8080
      port: 6650
    nginx:
      host: nginx
      http_port: 9300
      grpc_port: 9310
  client:
    include: true
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim:
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
  clustermanager: 
    include: true
    ip: clustermanager
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
  nodemanager:  
    include: true
    replicas: 2
    nodeSelector:
    tolerations:
    affinity:
    sessionProcessorsPerNode: 2
    subPath: "nodemanager"
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
    existingClaim:
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"

  client: 
    include: true
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim:
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi

  mysql: 
    include: true
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    ip: mysql
    port: 3306
    database: eggroll_meta
    user: fate
    password: fate_dev
    subPath: "mysql"
    existingClaim:
    claimName: mysql-data
    storageClass:
    accessMode: ReadWriteOnce
    size: 1Gi
  serving:
    ip: 192.168.9.1
    port: 30095
    useRegistry: false
    zookeeper:
      hosts:
        - serving-zookeeper.fate-serving-9999:2181
      use_acl: false
  fateboard:
    include: true
    type: ClusterIP
    username: admin
    password: admin

  spark:
    include: true
    master:
      Image: ""
      ImageTag: ""
      replicas: 1
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30977
    worker:
      Image: ""
      ImageTag: ""
      replicas: 2
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      resources:
        requests:
          cpu: "2"
          memory: "4Gi"
  hdfs:
    include: true
    namenode:
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30900
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
    datanode:
      replicas: 3
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      existingClaim:
      storageClass:
      accessMode: ReadWriteOnce
      size: 1Gi
  nginx:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    httpNodePort: 30093
    grpcNodePort: 30098
    loadBalancerIP: 
    exchange:
      ip: 192.168.10.1
      httpPort: 30003
      grpcPort: 30008
    route_table: 
#      10000: 
#        proxy: 
#        - host: 192.168.10.1 
#          http_port: 30103
#          grpc_port: 30108
#        fateflow:
#        - host: 192.168.10.1  
#          http_port: 30107
#          grpc_port: 30102
  rabbitmq:
    include: true
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    nodePort: 30094
    loadBalancerIP: 
    default_user: fate
    default_pass: fate
    user: fate
    password: fate
    route_table: 
#      10000:
#        host: 192.168.10.1 
#        port: 30104

  pulsar:
    include: true
    nodeSelector:
    tolerations:
    env:
    confs:
    affinity:
    type: ClusterIP
    httpNodePort: 30094
    httpsNodePort: 30099
    loadBalancerIP:
    existingClaim:
    accessMode: ReadWriteOnce
    storageClass:
    size: 1Gi
    publicLB:
      enabled: false
    # exchange:
      # ip: 192.168.10.1
      # port: 30000
      # domain: fate.org
    route_table: 
#      10000:
#        host: 192.168.10.1
#        port: 30104
#        sslPort: 30109
#        proxy: ""
#   

  postgres:
    include: false
    image: postgres
    imageTag: 13.3
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP:
    user: site_portal
    password: site_portal
    db: site_portal
    # subPath: ""
    existingClaim: ""
    storageClass: ""
    accessMode: ReadWriteOnce
    size: 1Gi
  
  frontend:
    include: false
    image: federatedai/site-portal-frontend
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    
    # nodePort: 
    # loadBalancerIP:
    
  sitePortalServer:
    include: false
    image: site-portal-server
    imageTag: v0.2.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    # nodePort: 
    # loadBalancerIP:
    existingClaim: ""
    storageClass: "sitePortalServer"
    accessMode: ReadWriteOnce
    size: 1Gi
    postgresHost: postgres
    postgresPort: 5432
    postgresDb: site_portal
    postgresUser: site_portal
    postgresPassword: site_portal
    adminPassword: admin
    userPassword: user
    serverCert: /var/lib/site-portal/cert/server.crt
    serverKey: /var/lib/site-portal/cert/server.key
    clientCert: /var/lib/site-portal/cert/client.crt
    clientKey: /var/lib/site-portal/cert/client.key
    caCert: /var/lib/site-portal/cert/ca.crt
    tlsEnabled: 'true'
    tlsPort: 8443
    tlsCommonName: site-1.server.example.com

# externalMysqlIp: mysql
# externalMysqlPort: 3306
# externalMysqlDatabase: eggroll_meta
# externalMysqlUser: fate
# externalMysqlPassword: fate_dev    
`,
			ValuesTemplate: `
image:
  registry: {{ .registry | default "federatedai" }}
  isThridParty: {{ empty .registry | ternary  "false" "true" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

partyId: {{ .partyId | int64 | toString }}
partyName: {{ .name }}

computing: {{ .computing }}
federation: {{ .federation }}
storage: {{ .storage }}
algorithm: {{ .algorithm }}
device: {{ .device }}

{{- $partyId := (.partyId | int64 | toString) }}

{{- with .ingress }}
ingress:
  {{- with .fateboard }}
  fateboard:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .client }}
  client:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .spark }}
  spark:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .rabbitmq }}
  rabbitmq:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
  
  {{- with .pulsar }}
  pulsar:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}

  {{- with .frontend }}
  frontend:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls: 
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}

{{- end }}

{{- with .istio }}
istio:
  enabled: {{ .enabled | default false }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

ingressClassName: {{ .ingressClassName | default "nginx"}}

exchange:
{{- with .rollsite }}
{{- with .exchange }}
  partyIp: {{ .ip }}
  partyPort: {{ .port }}
{{- end }}
{{- end }}

exchangeList:
{{- with .lbrollsite }}
{{- range .exchangeList }}
  - id: {{ .id }}
    ip: {{ .ip }}
    port: {{ .port }}
{{- end }}
{{- end }}

partyList:
{{- with .rollsite }}
{{- range .partyList }}
  - partyId: {{ .partyId }}
    partyIp: {{ .partyIp }}
    partyPort: {{ .partyPort }}
{{- end }}
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

modules:
  rollsite: 
    include: {{ has "rollsite" .modules }}
    {{- with .rollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    enableTLS: {{ .enableTLS | default false}}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .polling }}
    polling:
      enabled: {{ .enabled }}
      type: {{ .type }}
      {{- with .server }}
      server:
        ip: {{ .ip }}
        port: {{ .port }}
      {{- end }}
      {{- with .clientList }}
      clientList:
{{ toYaml . | indent 6 }}
      {{- end }}
      concurrency: {{ .concurrency }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  lbrollsite:
    include: {{ has "lbrollsite" .modules }}
    {{- with .lbrollsite }}
    ip: rollsite
    type: {{ .type | default "ClusterIP" }}
    loadBalancerIP: {{ .loadBalancerIP }}
    nodePort: {{ .nodePort }}
    size: {{ .size | default "2M" }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  python: 
    include: {{ has "python" .modules }}
    {{- with .python }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    logLevel: {{ .logLevel | default "INFO" }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    grpcNodePort: {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    serviceAccountName: {{ .serviceAccountName }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    existingClaim: {{ .existingClaim  }}
    claimName: {{ .claimName | default "python-data" }}
    storageClass: {{ .storageClass | default "python" }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- with .clustermanager }}
    clustermanager:
      cores_per_node: {{ .cores_per_node }}
      nodes: {{ .nodes }}
    {{- end }}
    {{- with .spark }}

    spark: 
{{ toYaml . | indent 6}}
    {{- end }}
    {{- with .hdfs }}
    hdfs:
      name_node: {{ .name_node }}
      path_prefix: {{ .path_prefix }}
    {{- end }}
    {{- with .pulsar }}
    pulsar:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
    {{- end }}
    {{- with .rabbitmq }}
    rabbitmq:
      host: {{ .host }}
      mng_port: {{ .mng_port }}
      port: {{ .port }}
      user: {{ .user }}
      password: {{ .password }}
    {{- end }}
    {{- with .nginx }}
    nginx:
      host: {{ .host }}
      http_port: {{ .http_port }}
      grpc_port: {{ .grpc_port }}
    {{- end }}
    {{- end }}


  clustermanager: 
    include: {{ has "clustermanager" .modules }}
    {{- with .clustermanager }}
    ip: clustermanager
    type: "ClusterIP"
    enableTLS: {{ .enableTLS | default false }}
  {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  nodemanager:  
    include: {{ has "nodemanager" .modules }}
    {{- with .nodemanager }}
    sessionProcessorsPerNode: {{ .sessionProcessorsPerNode }}
    replicas: {{ .replicas | default 2 }}
    subPath: {{ .subPath }}
    storageClass: {{ .storageClass  | default "client" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  client: 
    include: {{ has "client" .modules }}
    {{- with .client }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass  | default "client" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce" }}
    size: {{ .size  | default "1Gi" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}


  mysql: 
    include: {{ has "mysql" .modules }}
    {{- with .mysql }}
    type: {{ .type  | default "ClusterIP" }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    ip: {{ .ip | default "mysql" }}
    port: {{ .port | default "3306" }}
    database: {{ .database | default "eggroll_meta" }}
    user: {{ .user | default "fate" }}
    password: {{ .password | default "fate_dev" }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode | default "ReadWriteOnce" }}
    size: {{ .size | default "1Gi" }}
    {{- end }}


  serving:
    ip: {{ .servingIp }}
    port: {{ .servingPort }}
    {{- with .serving }}
    useRegistry: {{ .useRegistry | default false }}
    zookeeper:
{{ toYaml .zookeeper | indent 6 }}
    {{- end}}

  fateboard:
    include: {{ has "fateboard" .modules }}
    {{- with .fateboard }}
    type: {{ .type }}
    username: {{ .username }}
    password: {{ .password }}
    {{- end}}

  spark:
    include: {{ has "spark" .modules }}
    {{- with .spark }}
    {{- if .master }}
    master:
      Image: "{{ .master.Image }}"
      ImageTag: "{{ .master.ImageTag }}"
      replicas: {{ .master.replicas }}
      {{- with .master.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .master.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .master.type }}
      nodePort: {{ .master.nodePort }}
    {{- end }}
    {{- if .worker }}
    worker:
      Image: "{{ .worker.Image }}"
      ImageTag: "{{ .worker.ImageTag }}"
      replicas: {{ .worker.replicas }}
      {{- with .worker.resources }}
      resources:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .worker.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .worker.type | default "ClusterIP" }}
    {{- end }}
    {{- end }}


  hdfs:
    include: {{ has "hdfs" .modules }}
    {{- with .hdfs }}
    namenode:
      {{- with .namenode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .namenode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .namenode.type | default "ClusterIP" }}
      nodePort: {{ .namenode.nodePort }}
      existingClaim: {{ .namenode.existingClaim }}
      storageClass: {{ .namenode.storageClass | default "" }}
      accessMode: {{ .namenode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .namenode.size | default "1Gi" }}
    datanode:
      replicas: {{ .datanode.replicas | default 3 }}
      {{- with .datanode.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end }}
      {{- with .datanode.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end }}
      type: {{ .datanode.type | default "ClusterIP" }}
      existingClaim: {{ .datanode.existingClaim }}
      storageClass: {{ .datanode.storageClass | default "" }}
      accessMode: {{ .datanode.accessMode  | default "ReadWriteOnce"  }}
      size: {{ .datanode.size | default "1Gi" }}
    {{- end }}


  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
    {{- with .nodeSelector }}
    nodeSelector: 
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort:  {{ .httpNodePort }}
    grpcNodePort:  {{ .grpcNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      httpPort: {{ .httpPort }}
      grpcPort: {{ .grpcPort }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  rabbitmq:
    include: {{ has "rabbitmq" .modules }}
    {{- with .rabbitmq }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    default_user: {{ .default_user }}
    default_pass: {{ .default_pass }}
    loadBalancerIP: {{ .loadBalancerIP }}
    user: {{ .user }}
    password: {{ .password }}
    route_table:
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    {{- end }}


  pulsar:
    include: {{ has "pulsar" .modules }}
    {{- with .pulsar }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .env }}
    env:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .confs }}
    confs:
{{ toYaml . | indent 6 }}
    {{- end }}    
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    httpNodePort: {{ .httpNodePort }}
    httpsNodePort: {{ .httpsNodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- with .publicLB}}
    publicLB:
      enabled: {{ .enabled | default false }}
    {{- end }}
    {{- with .exchange }}
    exchange:
      ip: {{ .ip }}
      port: {{ .port }}
      domain: {{ .domain | default "fate.org" }}
    {{- end }}
    route_table: 
      {{- range $key, $val := .route_table }}
      {{ $key }}: 
{{ toYaml $val | indent 8 }}
      {{- end }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass | default "" }}
    accessMode: {{ .accessMode  | default "ReadWriteOnce"  }}
    size: {{ .size | default "1Gi" }}
    {{- end }}

  postgres:
    include: {{ has "postgres" .modules }}
    {{- with .postgres }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    user: {{ .user }}
    password: {{ .password }}
    db: {{ .db }}
    subPath: {{ .subPath }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode }}
    size: {{ .size }}
    {{- end }}

  frontend:
    include: {{ has "frontend" .modules }}
    {{- with .frontend }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- end }}

  sitePortalServer:
    include: {{ has "sitePortalServer" .modules }}
    {{- with .sitePortalServer }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "ClusterIP" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    existingClaim: {{ .existingClaim }}
    storageClass: {{ .storageClass }}
    accessMode: {{ .accessMode | default "ReadWriteOnce"  }}
    size: {{ .size  | default "1Gi" }}
    postgresHost: {{ .postgresHost | default "postgres" }}
    postgresPort: {{ .postgresPort | default "5432" }}
    postgresDb: {{ .postgresDb | default "site_portal" }}
    postgresUser: {{ .postgresUser | default "site_portal" }}
    postgresPassword: {{ .postgresPassword | default "site_portal" }}
    adminPassword: {{ .adminPassword | default "admin" }}
    userPassword: {{ .userPassword | default "user" }}
    serverCert: {{ .serverCert| default "/var/lib/site-portal/cert/server.crt" }}
    serverKey: {{ .serverKey | default "/var/lib/site-portal/cert/server.key" }}
    clientCert: {{ .clientCert | default "/var/lib/site-portal/cert/client.crt" }}
    clientKey: {{ .clientKey | default "/var/lib/site-portal/cert/client.key" }}
    caCert: {{ .caCert | default "/var/lib/site-portal/cert/ca.crt" }}
    tlsEnabled: {{ .tlsEnabled | default "'true'" }}
    tlsPort: {{ .tlsPort | default "8443" }}
    tlsCommonName: {{ .tlsCommonName | default "site-1.server.example.com" }}
    {{- end }}

externalMysqlIp: {{ .externalMysqlIp }}
externalMysqlPort: {{ .externalMysqlPort }}
externalMysqlDatabase: {{ .externalMysqlDatabase }}
externalMysqlUser: {{ .externalMysqlUser }}
externalMysqlPassword: {{ .externalMysqlPassword }}
`,
			Private: true,
		},
		"516a10e2-0b96-417c-812b-ee45ed197e81": {
			Model: gorm.Model{
				ID:        101,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "516a10e2-0b96-417c-812b-ee45ed197e81",
			Name:        "chart for FedLCM private OpenFL Director v0.1.0",
			Description: "This chart is for deploying FedLCM's OpenFL director v0.1.0 services, based from OpenFL 1.3 release",
			Type:        entity.ChartTypeOpenFLDirector,
			ChartName:   "openfl-director",
			Version:     "v0.1.0",
			AppVersion:  "openfl-director-v0.1.0",
			Chart: `apiVersion: v1
appVersion: "openfl-director-v0.1.0"
description: A Helm chart for openfl director, based on official OpenFL container image
name: openfl-director
version: v0.1.0
sources:
  - https://github.com/FederatedAI/KubeFATE.git`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: openfl-director
chartVersion: v0.1.0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
# pullPolicy:
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
modules:
  - director
  - notebook

ingress:
  notebook:
  # annotations:
    hosts:
    - name:  {{.Name}}.notebook.{{.Domain}}
      path: /
    # tls:

director:
  image: fedlcm-openfl
  imageTag: v0.1.0
  type: {{.ServiceType}}
  sampleShape: "{{.SampleShape}}"
  targetShape: "{{.TargetShape}}"
#  envoyHealthCheckPeriod: 60
#  nodeSelector:
#  tolerations:
#  affinity:
#  nodePort:
#  loadBalancerIp:

notebook:
  image: fedlcm-openfl
  imageTag: v0.1.0
  type: {{.ServiceType}}
  password: {{.JupyterPassword}}
#  nodeSelector:
#  tolerations:
#  affinity:
#  nodePort:
#  loadBalancerIp:`,
			Values: `image:
  registry: federatedai
  pullPolicy: IfNotPresent
  imagePullSecrets:
#  - name:

podSecurityPolicy:
  enabled: false

ingress:
  notebook:
    # annotations:
    hosts:
      - name: notebook.openfl.example.com
        path: /
    tls: []

modules:
  director:
    image: fedlcm-openfl
    imageTag: v0.1.0
    sampleShape: "['1']"
    targetShape: "['1']"
    envoyHealthCheckPeriod: 60
    # nodeSelector:
    # tolerations:
    # affinity:
    type: NodePort
    # nodePort:
    # loadBalancerIP:

  notebook:
    image: fedlcm-openfl
    imageTag: v0.1.0
    # password:
    # nodeSelector:
    # tolerations:
    # affinity:
    type: NodePort
    # nodePort:
    # loadBalancerIP:`,
			ValuesTemplate: `name: {{ .name }}

image:
  registry: {{ .registry | default "federatedai" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

{{- with .ingress }}
ingress:
  {{- with .notebook }}
  notebook:
    {{- with .annotations }}
    annotations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .hosts }}
    hosts:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tls }}
    tls:
{{ toYaml . | indent 6 }}
    {{- end }}
  {{- end }}
{{- end }}

modules:
  director:
    {{- with .director }}
    image: {{ .image }}
    sampleShape: "{{ .sampleShape }}"
    targetShape: "{{ .targetShape }}"
    envoyHealthCheckPeriod: {{ .envoyHealthCheckPeriod | default "60"}}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "NodePort" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- end }}

  notebook:
    {{- with .notebook}}
    image: {{ .image }}
    password: {{ .password }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    type: {{ .type | default "NodePort" }}
    nodePort: {{ .nodePort }}
    loadBalancerIP: {{ .loadBalancerIP }}
    {{- end }}`,
			ArchiveContent: mock.FedLCMOpenFLDirector010ChartArchiveContent,
			Private:        true,
		},
		"c62b27a6-bf0f-4515-840a-2554ed63aa56": {
			Model: gorm.Model{
				ID:        102,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "c62b27a6-bf0f-4515-840a-2554ed63aa56",
			Name:        "chart for FedLCM private OpenFL Envoy v0.1.0",
			Description: "This chart is for deploying OpenFL envoy built for FedLCM, based from OpenFL 1.3 release",
			Type:        entity.ChartTypeOpenFLEnvoy,
			ChartName:   "openfl-envoy",
			Version:     "v0.1.0",
			AppVersion:  "openfl-envoy-v0.1.0",
			Chart: `apiVersion: v1
appVersion: "openfl-envoy-v0.1.0"
description: A Helm chart for openfl envoy
name: openfl-envoy
version: v0.1.0
sources:
  - https://github.com/FederatedAI/KubeFATE.git`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: openfl-envoy
chartVersion: v0.1.0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}
modules:
  - envoy

envoy:
  image: fedlcm-openfl
  imageTag: v0.1.0
  directorFqdn: {{.DirectorFQDN}}
  directorIp: {{.DirectorIP}}
  directorPort: {{.DirectorPort}}
  aggPort: {{.AggPort}}
  envoyConfigs:
{{ .EnvoyConfig | trimAll "\n" | indent 4 }}
#  nodeSelector:
#  tolerations:
#  affinity:`,
			Values: `name: envoy-1

image:
  registry: federatedai
  pullPolicy: IfNotPresent
  imagePullSecrets:
#  - name:

podSecurityPolicy:
  enabled: false

modules:
  envoy:
    image: fedlcm-openfl
    imageTag: v0.1.0
    directorFqdn: director
    directorIp: 192.168.1.1
    directorPort: 50051
    aggPort: 50052
    envoyConfigs:
    # nodeSelector:
    # tolerations:
    # affinity:`,
			ValuesTemplate: `name: {{ .name }}

image:
  registry: {{ .registry | default "federatedai" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

modules:
  envoy:
    {{- with .envoy }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    directorFqdn: {{ .directorFqdn }}
    directorIp: {{ .directorIp }}
    directorPort: {{ .directorPort }}
    aggPort: {{ .aggPort }}
    envoyConfigs:
{{ toYaml .envoyConfigs | indent 6 }}
    {{- with .nodeSelector }}
    nodeSelector:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .tolerations }}
    tolerations:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- with .affinity }}
    affinity:
{{ toYaml . | indent 6 }}
    {{- end }}
    {{- end }}`,
			ArchiveContent: mock.FedLCMOpenFLEnvoy010ChartArchiveContent,
			Private:        true,
		},
	}
)
