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
		"2c8f974d-b822-4719-bcaf-f1ef608c4923": {
			Model: gorm.Model{
				ID:        5,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "2c8f974d-b822-4719-bcaf-f1ef608c4923",
			Name:        "chart for FATE exchange v1.9.0",
			Description: "This chart is for deploying FATE exchange v1.9.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.9.0",
			AppVersion:  "v1.9.0",
			Chart: `apiVersion: v1
appVersion: v1.9.0
description: A Helm chart for fate exchange
name: fate-exchange
version: v1.9.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.9.0
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
  tag: 1.9.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

podSecurityPolicy:
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
  tag: {{ .imageTag | default "1.9.0-release" }}
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
		"c32411c7-3744-46ee-bb74-046d99ce3385": {
			Model: gorm.Model{
				ID:        6,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "c32411c7-3744-46ee-bb74-046d99ce3385",
			Name:        "chart for FATE cluster v1.9.0",
			Description: "This is chart for installing FATE cluster v1.9.0",
			Type:        entity.ChartTypeFATECluster,
			ChartName:   "fate",
			Version:     "v1.9.0",
			AppVersion:  "v1.9.0",
			Chart: `apiVersion: v1
appVersion: v1.9.0
description: A Helm chart for fate-training
name: fate
version: v1.9.0
home: https://fate.fedai.org
icon: https://aisp-1251170195.cos.ap-hongkong.myqcloud.com/wp-content/uploads/sites/12/2019/09/logo.png
sources:
  - https://github.com/FederatedAI/KubeFATE
  - https://github.com/FederatedAI/FATE`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.9.0
partyId: {{.PartyID}}
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
persistence: {{ .EnablePersistence }}
# pullPolicy:
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}

# ingressClassName: nginx

modules:
  - mysql
  - python
  - fateboard
  - client
  - spark
  - hdfs
  - nginx
  - pulsar

computing: Spark
federation: Pulsar
storage: HDFS
algorithm: Basic
device: CPU

ingress:
  fateboard:
    hosts:
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}

nginx:
  type: {{.ServiceType}}
  exchange:
    ip: {{.ExchangeNginxHost}}
    httpPort: {{.ExchangeNginxPort}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort: 30093
  # grpcNodePort: 30098

pulsar:
  publicLB:
    enabled: true
  env:
    - name: PULSAR_MEM
      value: "-Xms4g -Xmx4g -XX:MaxDirectMemorySize=8g"
  confs:
      brokerDeleteInactiveTopicsFrequencySeconds: 60
      backlogQuotaDefaultLimitGB: 10
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
  size: 1Gi
  storageClass: {{ .StorageClass }}
  existingClaim: ""
  accessMode: ReadWriteOnce
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 30094
  # httpsNodePort: 30099
  # loadBalancerIP:
  # resources:
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"

mysql:
  size: 1Gi
  storageClass: {{ .StorageClass }}
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
  storageClass: {{ .StorageClass }}
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
    # requests:
      # cpu: "2"
      # memory: "4Gi"
    # limits:
      # cpu: "4"
      # memory: "8Gi"
  # logLevel: INFO
  # spark: 
    # cores_per_node: 20
    # nodes: 2
    # master: spark://spark-master:7077
    # driverHost: 
    # driverHostType: 
    # portMaxRetries: 
    # driverStartPort: 
    # blockManagerStartPort: 
    # pysparkPython: 
  # hdfs:
    # name_node: hdfs://namenode:9000
    # path_prefix:
  # pulsar:
    # host: pulsar
    # mng_port: 8080
    # port: 6650
  # nginx:
    # host: nginx
    # http_port: 9300
    # grpc_port: 9310

client:
  size: 1Gi
  storageClass: {{ .StorageClass }}
  existingClaim: ""
  accessMode: ReadWriteOnce
  subPath: "client"
  # nodeSelector:
  # tolerations:
  # affinity:

hdfs:
  namenode:
    storageClass: {{ .StorageClass }}
    size: 3Gi
    existingClaim: ""
    accessMode: ReadWriteOnce
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    size: 10Gi
    storageClass: {{ .StorageClass }}
    existingClaim: ""
    accessMode: ReadWriteOnce
    # replicas: 3
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP

spark:
  # master:
    # replicas: 1
    # resources:
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30977
  worker:
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
`,
			Values: `image:
  registry: federatedai
  isThridParty:
  tag: 1.9.0-release
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
    
externalMysqlIp: {{ .externalMysqlIp }}
externalMysqlPort: {{ .externalMysqlPort }}
externalMysqlDatabase: {{ .externalMysqlDatabase }}
externalMysqlUser: {{ .externalMysqlUser }}
externalMysqlPassword: {{ .externalMysqlPassword }}`,
			ArchiveContent: nil,
			Private:        false,
		},
		"4ad46829-a827-4632-b169-c8675360321e": {
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "4ad46829-a827-4632-b169-c8675360321e",
			Name:        "chart for FATE exchange v1.8.0",
			Description: "This chart is for deploying FATE exchange v1.8.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.8.0",
			AppVersion:  "v1.8.0",
			Chart: `apiVersion: v1
appVersion: v1.8.0
description: A Helm chart for fate exchange
name: fate-exchange
version: v1.8.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.8.0
partyId: 0
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
imageTag: "1.8.0-release"
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
  tag: 1.8.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

podSecurityPolicy:
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
  tag: {{ .imageTag | default "1.8.0-release" }}
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
			Name:        "chart for FATE cluster v1.8.0",
			Description: "This is chart for installing FATE cluster v1.8.0",
			Type:        entity.ChartTypeFATECluster,
			ChartName:   "fate",
			Version:     "v1.8.0",
			AppVersion:  "v1.8.0",
			Chart: `apiVersion: v1
appVersion: v1.8.0
description: A Helm chart for fate-training
name: fate
version: v1.8.0
home: https://fate.fedai.org
icon: https://aisp-1251170195.cos.ap-hongkong.myqcloud.com/wp-content/uploads/sites/12/2019/09/logo.png
sources:
  - https://github.com/FederatedAI/KubeFATE
  - https://github.com/FederatedAI/FATE`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.8.0
partyId: {{.PartyID}}
{{- if .UseRegistry}}
registry: {{.Registry}}
{{- end }}
# imageTag: "1.8.0-release"
persistence: {{ .EnablePersistence }}
# pullPolicy:
podSecurityPolicy:
  enabled: {{.EnablePSP}}
{{- if .UseImagePullSecrets}}
imagePullSecrets:
  - name: {{.ImagePullSecretsName}}
{{- end }}

# ingressClassName: nginx

modules:
  - mysql
  - python
  - fateboard
  - client
  - spark
  - hdfs
  - nginx
  - pulsar

backend: spark_pulsar

ingress:
  fateboard:
    hosts:
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}

nginx:
  type: {{.ServiceType}}
  exchange:
    ip: {{.ExchangeNginxHost}}
    httpPort: {{.ExchangeNginxPort}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # loadBalancerIP:
  # httpNodePort: 30093
  # grpcNodePort: 30098

pulsar:
  publicLB:
    enabled: true
  exchange:
    ip: {{.ExchangeATSHost}}
    port: {{.ExchangeATSPort}}
    domain: {{.Domain}}
  size: 1Gi
  storageClass: {{ .StorageClass }}
  existingClaim: ""
  accessMode: ReadWriteOnce
  # nodeSelector:
  # tolerations:
  # affinity:
  # type: ClusterIP
  # httpNodePort: 30094
  # httpsNodePort: 30099
  # loadBalancerIP:

mysql:
  size: 1Gi
  storageClass: {{ .StorageClass }}
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
  storageClass: {{ .StorageClass }}
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
  # spark: 
    # cores_per_node: 20
    # nodes: 2
    # master: spark://spark-master:7077
    # driverHost: 
    # driverHostType: 
    # portMaxRetries: 
    # driverStartPort: 
    # blockManagerStartPort: 
    # pysparkPython: 
  # hdfs:
    # name_node: hdfs://namenode:9000
    # path_prefix:
  # pulsar:
    # host: pulsar
    # mng_port: 8080
    # port: 6650
  # nginx:
    # host: nginx
    # http_port: 9300
    # grpc_port: 9310

client:
  size: 1Gi
  storageClass: {{ .StorageClass }}
  existingClaim: ""
  accessMode: ReadWriteOnce
  subPath: "client"
  # nodeSelector:
  # tolerations:
  # affinity:

hdfs:
  namenode:
    storageClass: {{ .StorageClass }}
    size: 3Gi
    existingClaim: ""
    accessMode: ReadWriteOnce
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30900
  datanode:
    size: 10Gi
    storageClass: {{ .StorageClass }}
    existingClaim: ""
    accessMode: ReadWriteOnce
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP

spark:
  # master:
    # replicas: 1
    # resources:
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP
    # nodePort: 30977
  worker:
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
`,
			Values: `image:
  registry: federatedai
  isThridParty:
  tag: 1.8.0-release
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

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
    backend: eggroll
    enabledNN: false
    logLevel: INFO
    # subPath: ""
    existingClaim: ""
    claimName: python-data
    storageClass: "python"
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
    existingClaim: ""
    storageClass: "nodemanager-0"
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
    list:
    - name: nodemanager-0
      nodeSelector:
      tolerations:
      affinity:
      sessionProcessorsPerNode: 2
      subPath: "nodemanager-0"
      existingClaim: ""
      storageClass: "nodemanager-0"
      accessMode: ReadWriteOnce
      size: 1Gi
    - name: nodemanager-1
      nodeSelector:
      tolerations:
      affinity:
      sessionProcessorsPerNode: 2
      subPath: "nodemanager-1"
      existingClaim: ""
      storageClass: "nodemanager-1"
      accessMode: ReadWriteOnce
      size: 1Gi

  client: 
    include: true
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim: ""
    storageClass: "client"
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
    existingClaim: ""
    claimName: mysql-data
    storageClass: "mysql"
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
      resources:
        requests:
          cpu: "2"
          memory: "4Gi"
        limits:
          cpu: "4"
          memory: "8Gi"
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
  hdfs:
    include: true
    namenode:
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30900
    datanode:
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
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
    affinity:
    type: ClusterIP
    httpNodePort: 30094
    httpsNodePort: 30099
    loadBalancerIP: 
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
  tag: {{ .imageTag | default "1.8.0-release" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

partyId: {{ .partyId | int64 | toString }}
partyName: {{ .name }}

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
    {{- end }}


  python: 
    include: {{ has "python" .modules }}
    backend: {{ default "eggroll" .backend }}
    {{- with .python }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
    logLevel: {{ .logLevel }}
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
    enabledNN: {{ .enabledNN | default false }}
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


  nodemanager:  
    include: {{ has "nodemanager" .modules }}
    {{- with .nodemanager }}
    list:
    {{- $nodemanager := . }}
    {{- range .count | int | until }}
    - name: nodemanager-{{ . }}
      {{- with $nodemanager.nodeSelector }}
      nodeSelector: 
{{ toYaml . | indent 8 }}
      {{- end}}
      {{- with $nodemanager.tolerations }}
      tolerations:
{{ toYaml . | indent 8 }}
      {{- end}}
      {{- with $nodemanager.affinity }}
      affinity:
{{ toYaml . | indent 8 }}
      {{- end}}
      sessionProcessorsPerNode: {{ $nodemanager.sessionProcessorsPerNode }}
      subPath: "nodemanager-{{ . }}"
      existingClaim: ""
      storageClass: "{{ $nodemanager.storageClass }}"
      accessMode: {{ $nodemanager.accessMode }}
      size: {{ $nodemanager.size }}
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
      size: {{ .namenode.size }}
    datanode:
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
      size: {{ .datanode.size }}
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
    size: {{ .size }}
    {{- end }}
    
externalMysqlIp: {{ .externalMysqlIp }}
externalMysqlPort: {{ .externalMysqlPort }}
externalMysqlDatabase: {{ .externalMysqlDatabase }}
externalMysqlUser: {{ .externalMysqlUser }}
externalMysqlPassword: {{ .externalMysqlPassword }}`,
			ArchiveContent: nil,
			Private:        false,
		},
		"3ce13cb2-5543-4b01-a5e4-9e4c4baa5973": {
			Model: gorm.Model{
				ID:        3,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:        "3ce13cb2-5543-4b01-a5e4-9e4c4baa5973",
			Name:        "chart for FATE exchange v1.6.1 with fml-manager service v0.1.0",
			Description: "This chart is for deploying FATE exchange v1.6.1 with fml-manager v0.1.0",
			Type:        entity.ChartTypeFATEExchange,
			ChartName:   "fate-exchange",
			Version:     "v1.6.1-fedlcm-v0.1.0",
			AppVersion:  "exchangev1.6.1 & fedlcmv0.1.0",
			Chart: `apiVersion: v1
appVersion: "exchangev1.6.1 & fedlcmv0.1.0"
description: A Helm chart for fate exchange and fml-manager
name: fate-exchange
version: v1.6.1-fedlcm-v0.1.0`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate-exchange
chartVersion: v1.6.1-fedlcm-v0.1.0
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
{{- if .UseRegistry}}
  image: trafficserver
{{- else }}
  image: federatedai/trafficserver
{{- end }}
  imageTag: latest
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
{{- if .UseRegistry}}
  image: nginx
{{- else }}
  image: federatedai/nginx
{{- end }}
  imageTag: 1.6.1-release
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
  image: postgres
  imageTag: 13.3
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
{{- if .UseRegistry}}
  image: fml-manager-server
{{- else }}
  image: federatedai/fml-manager-server
{{- end }}
  imageTag: v0.1.0
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
  registry: 
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name:

podSecurityPolicy:
  enabled: false

persistence:
  enabled: false

modules:
  nginx:
    include: true
    image: federatedai/nginx
    imageTag: 1.6.1-release
    type: ClusterIP
    # httpNodePort:  30003
    # grpcNodePort:  30008
    # loadBalancerIP: 
    # nodeSelector: 
    # tolerations:
    # affinity:
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
    image: federatedai/trafficserver
    imageTag: latest
    include: true
    type: ClusterIP
    # nodePort: 30007
    # loadBalancerIP: 
    # nodeSelector: 
    # tolerations:
    # affinity:
    # route_table is used to configure the cluster information of all parties that join in the exchange deployment mode. (When Spark was used as the calculation engine at the time)
    route_table: 
      sni:
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
    imageTag: v0.1.0
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
  registry: {{ .registry | default "" }}
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
  nginx:
    include: {{ has "nginx" .modules }}
    {{- with .nginx }}
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
			ArchiveContent: mock.FATEExchange161WithManagerChartArchiveContent,
			Private:        true,
		},
		"09356325-b1d9-4598-b952-1b2bb73baf51": {
			Model: gorm.Model{
				ID:        4,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			UUID:           "09356325-b1d9-4598-b952-1b2bb73baf51",
			Name:           "chart for FATE cluster v1.6.1 with site-portal v0.1.0",
			Description:    "This is chart for installing FATE cluster v1.6.1 with site-portal v0.1.0",
			Type:           entity.ChartTypeFATECluster,
			ChartName:      "fate",
			Version:        "v1.6.1-fedlcm-v0.1.0",
			AppVersion:     "fatev1.6.1+fedlcmv0.1.0",
			ArchiveContent: mock.FATE161WithPortalChartArchiveContent,
			Chart: `apiVersion: v1
appVersion: "fatev1.6.1+fedlcmv0.1.0"
description: Helm chart for FATE and site-portal in FedLCM
name: fate
version: v1.6.1-fedlcm-v0.1.0
sources:
  - https://github.com/FederatedAI/KubeFATE.git
  - https://github.com/FederatedAI/FATE.git`,
			InitialYamlTemplate: `name: {{.Name}}
namespace: {{.Namespace}}
chartName: fate
chartVersion: v1.6.1-fedlcm-v0.1.0
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
  - spark
  - hdfs
  - nginx
  - pulsar
  - frontend
  - sitePortalServer
  - postgres

backend: spark

ingress:
  fateboard:
    hosts:
    - name: {{.Name}}.fateboard.{{.Domain}}
  client:
    hosts:
    - name: {{.Name}}.notebook.{{.Domain}}
  spark:
    hosts:
    - name: {{.Name}}.spark.{{.Domain}}
  pulsar:
    hosts:
    - name: {{.Name}}.pulsar.{{.Domain}}

python:
{{- if .UseRegistry}}
  image: python-spark
{{- else }}
  image: federatedai/python-spark
{{- end }}
  imageTag: 1.6.1-release
  existingClaim: ""
  storageClass: {{ .StorageClass }}
  accessMode: ReadWriteOnce
  size: 10Gi
  # type: ClusterIP
  # httpNodePort: 
  # grpcNodePort: 
  # resources:
  # nodeSelector:
  # tolerations:
  # affinity:
  # enabledNN: true
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
{{- if .UseRegistry}}
  image: fateboard
{{- else }}
  image: federatedai/fateboard
{{- end }}
  imageTag: 1.6.1-release
  type: ClusterIP
  username: admin
  password: admin

client:
{{- if .UseRegistry}}
  image: client
{{- else }}
  image: federatedai/client
{{- end }}
  imageTag: 1.6.1-release
  subPath: "client"
  existingClaim: ""
  accessMode: ReadWriteOnce
  size: 1Gi
  storageClass: {{ .StorageClass }}
  # nodeSelector:
  # tolerations:
  # affinity:

mysql:
  image: mysql
  imageTag: 8
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

spark:
  master:
{{- if .UseRegistry}}
    image: "spark-master"
{{- else }}
    image: "federatedai/spark-master"
{{- end }}
    imageTag: "1.6.1-release"
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
{{- if .UseRegistry}}
    image: "spark-worker"
{{- else }}
    image: "federatedai/spark-worker"
{{- end }}
    imageTag: "1.6.1-release"
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
{{- if .UseRegistry}}
    image: hadoop-namenode
{{- else }}
    image: federatedai/hadoop-namenode
{{- end }}
    imageTag: 2.0.0-hadoop2.7.4-java8
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
{{- if .UseRegistry}}
    image: hadoop-datanode
{{- else }}
    image: federatedai/hadoop-datanode
{{- end }}
    imageTag: 2.0.0-hadoop2.7.4-java8
    existingClaim: ""
    accessMode: ReadWriteOnce
    size: 1Gi
    storageClass: {{ .StorageClass }}
    # nodeSelector:
    # tolerations:
    # affinity:
    # type: ClusterIP

nginx:
  {{- if .UseRegistry}}
  image: nginx
  {{- else }}
  image: federatedai/nginx
  {{- end }}
  imageTag: 1.6.1-release
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

pulsar:
  {{- if .UseRegistry}}
  image: pulsar
  {{- else }}
  image: federatedai/pulsar
  {{- end }}
  imageTag: 2.7.0
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

postgres:
  image: postgres
  imageTag: 13.3
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
  {{- if .UseRegistry}}
  image: site-portal-frontend
  {{- else }}
  image: federatedai/site-portal-frontend
  {{- end }}
  imageTag: v0.1.0
  type: {{.ServiceType}}
  # nodeSelector:
  # tolerations:
  # affinity:
  # nodePort: 
  # loadBalancerIP:
 
sitePortalServer:
  {{- if .UseRegistry}}
  image: site-portal-server
  {{- else }}
  image: federatedai/site-portal-server
  {{- end }}
  imageTag: v0.1.0
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
  registry: 
  pullPolicy: IfNotPresent
  imagePullSecrets: 
#  - name: 
  
partyId: 9999
partyName: fate-9999

istio:
  enabled: false

podSecurityPolicy:
  enabled: false

ingress:
  fateboard: 
    # annotations: 
    hosts:
    - name: fateboard.kubefate.net
      path: /
    tls: []
    # - secretName: my-tls-secret
      # hosts:
        # - fateboard.kubefate.net
  client: 
    # annotations: 
    hosts:
    - name: notebook.kubefate.net
      path: /
    tls: [] 
  spark: 
    # annotations: 
    hosts:
    - name: spark.kubefate.net
      path: /
    tls: [] 
  pulsar: 
    # annotations: 
    hosts:
    - name:  pulsar.kubefate.net
      path: /
    tls: []
  frontend: 
    # annotations: 
    hosts:
    - name:  frontend.example.com
      path: /
    tls: []

persistence:
  enabled: false

modules:
  python: 
    image: federatedai/python-spark
    imageTag: 1.6.1-release
    include: true
    type: ClusterIP
    httpNodePort: 30097
    grpcNodePort: 30092
    loadBalancerIP: 
    serviceAccountName: 
    nodeSelector:
    tolerations:
    affinity:
    backend: eggroll
    enabledNN: false
    # subPath: ""
    existingClaim: ""
    claimName: python-data
    storageClass: "python"
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

  fateboard:
    include: true
    image: federatedai/fateboard
    imageTag: 1.6.1-release
    type: ClusterIP
    username: admin
    password: admin

  client: 
    include: true
    image: federatedai/client
    imageTag: 1.6.1-release
    ip: client
    type: ClusterIP
    nodeSelector:
    tolerations:
    affinity:
    subPath: "client"
    existingClaim: ""
    storageClass: "client"
    accessMode: ReadWriteOnce
    size: 1Gi

  mysql: 
    include: true
    image: mysql
    imageTag: 8
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
    existingClaim: ""
    claimName: mysql-data
    storageClass: "mysql"
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

  spark:
    include: true
    master:
      image: "federatedai/spark-master"
      imageTag: "1.6.1-release"
      replicas: 1
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30977
    worker:
      image: "federatedai/spark-worker"
      imageTag: "1.6.1-release"
      replicas: 2
      resources:
        requests:
          cpu: "1"
          memory: "2Gi"
        limits:
          cpu: "2"
          memory: "4Gi"
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
  hdfs:
    include: true
    namenode:
      image: federatedai/hadoop-namenode
      imageTag: 2.0.0-hadoop2.7.4-java8
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      nodePort: 30900
      existingClaim: ""
      accessMode: ReadWriteOnce
      size: 1Gi
      storageClass: hdfs
    datanode:
      image: federatedai/hadoop-datanode
      imageTag: 2.0.0-hadoop2.7.4-java8
      nodeSelector:
      tolerations:
      affinity:
      type: ClusterIP
      existingClaim: ""
      accessMode: ReadWriteOnce
      size: 1Gi
      storageClass: hdfs
  nginx:
    include: true
    image: federatedai/nginx
    imageTag: 1.6.1-release
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

  pulsar:
    include: true
    image: federatedai/pulsar
    imageTag: 2.7.0
    nodeSelector:
    tolerations:
    affinity:
    type: ClusterIP
    httpNodePort: 30094
    httpsNodePort: 30099
    loadBalancerIP: 
    publicLB:
      enabled: false
    existingClaim: ""
    storageClass: "pulsar"
    accessMode: ReadWriteOnce
    size: 1Gi
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
    include: true
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
    include: true
    image: federatedai/site-portal-frontend
    imageTag: v0.1.0
    # nodeSelector:
    # tolerations:
    # affinity:
    type: ClusterIP
    
    # nodePort: 
    # loadBalancerIP:
    
  sitePortalServer:
    include: true
    image: federatedai/site-portal-server
    imageTag: v0.1.0
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
# externalMysqlPassword: fate_dev`,
			ValuesTemplate: `
image:
  registry: {{ .registry | default "" }}
  pullPolicy: {{ .pullPolicy | default "IfNotPresent" }}
  {{- with .imagePullSecrets }}
  imagePullSecrets:
{{ toYaml . | indent 2 }}
  {{- end }}

partyId: {{ .partyId | int64 | toString }}
partyName: {{ .name }}

{{- with .istio }}
istio:
  enabled: {{ .enabled | default false }}
{{- end }}

{{- with .podSecurityPolicy }}
podSecurityPolicy:
  enabled: {{ .enabled | default false }}
{{- end }}

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
  
  {{- if not .tlsEnabled}}
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
{{- end }}

persistence:
  enabled: {{ .persistence | default "false" }}

modules:


  python: 
    include: {{ has "python" .modules }}
    backend: {{ default "spark" .backend }}
    {{- with .python }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    {{- with .resources }}
    resources:
{{ toYaml . | indent 6 }}
    {{- end }}
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
    enabledNN: {{ .enabledNN | default false }}
    existingClaim: {{ .existingClaim  }}
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


  fateboard:
    include: {{ has "fateboard" .modules }}
    {{- with .fateboard }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
    type: {{ .type }}
    username: {{ .username }}
    password: {{ .password }}
    {{- end}}


  client: 
    include: {{ has "client" .modules }}
    {{- with .client }}
    image: {{ .image }}
    imageTag: {{ .imageTag }}
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
    image: {{ .image }}
    imageTag: {{ .imageTag }}
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

  spark:
    include: {{ has "spark" .modules }}
    {{- with .spark }}
    {{- if .master }}
    master:
      image: "{{ .master.image }}"
      imageTag: "{{ .master.imageTag }}"
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
      type: {{ .master.type | default "ClusterIP" }}
    {{- end }}
    {{- if .worker }}
    worker:
      image: "{{ .worker.image }}"
      imageTag: "{{ .worker.imageTag }}"
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
      image: {{ .namenode.image }}
      imageTag: {{ .namenode.imageTag }}
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
      image: {{ .datanode.image }}
      imageTag: {{ .datanode.imageTag }}
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


  pulsar:
    include: {{ has "pulsar" .modules }}
    {{- with .pulsar }}
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
