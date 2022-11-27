# 如何部署使用外部已存在引擎的FATE集群

FedLCM在部署FATE系统时，默认会同时部署Spark、HDFS、Pulsar这些基础引擎组件，这些组件的部署形式是简单的Kubernetes的Pod，从性能上或许不能完全满足实际生产环境的需求，同时，我们的环境中可能已经有部署好的适用于生产环境的这些服务，在这种情况下，我们推荐直接使用这些已存在服务，目前FedLCM支持在部署FATE时指定如何使用这些服务。本文简单介绍这个过程。

## 对外部基础引擎的要求

- Spark：Spark的worker node必须已经包含有运行FATE的python相关依赖。
- HDFS：FATE-Flow必须能够访问到HDFS的name node和data node。
- Pulsar：Pulsar服务必须启用TLS并配置好相关证书，能够与Exchange里的ATS组件进行TLS连接。

## 创建使用外部引擎的FATE集群

首先，我们仍然需要部署一个Exchange，如[文档所介绍](./Getting_Started_FATE_zh.md#创建-exchange)。

之后在创建FATE Cluster时，在证书选择处选择"跳过，我将手动安装"。

之后在"选择外部的引擎"这里，配置如何连接外部引擎的参数。

***我们建议三个引擎都为外部，或都为内部，如果其中部分引擎为内部部署，需要额外的配置，如有此类需求，请开Issue获得支持***

关于具体参数的配置信息，可以参阅KubeFATE的[相关介绍](https://github.com/FederatedAI/KubeFATE/wiki/FATE-On-Spark---Leverage-the-external-cluster)

之后，继续使用FedLCM产生的yaml部署即可。
