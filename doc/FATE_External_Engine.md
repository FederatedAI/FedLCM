# Deploy a FATE Cluster Using External Engines

FATE deployed by default using FedLCM will include engines such as Spark HDFS and Pulsar. Sometimes the default engine performance is not powerful as we typically use them for experimental practice. If there are already clusters of Spark, HDFS and Pulsar outside, it is recommended to use them to gain more performance benefits.

## Requirements

- Spark: Spark's worker nodes must contain FATE's python runtime dependencies.
- HDFS: FATE component fateflow must be able to access the HDFS name node and data node.
- Pulsar: Pulsar needs to configure the CA certificate to communicate with ATS.

## Create a FATE

First you need to deploy an exchange。[Creating Exchange](./Getting_Started_FATE.md#creating-exchange)

Then create a cluster.

And when following the steps, choose "select certificates"  > "Skip, I will manually install certificates".

Then in the step "Whether to Use Existing External Engine Service". Turn on the switch for which we want to use existing external engine, and add the configuration information of the external engine.

> By default, all the switches are off, meaning FedLCM will deploy all this services.

***If using an external engine, all three engine information needs to be filled in, including Spark HDFS and Pulsar. If any of them is still to be deployed by FedLCM, please open an Issue to get information on how to do the additional configuration***

For more information on how to configure these parameters, please refer to [this KubeFATE wiki](https://github.com/FederatedAI/KubeFATE/wiki/FATE-On-Spark---Leverage-the-external-cluster)

After filling in, generate yaml, and then you can deploy FATE.

Next, you can use FATE for federated learning tasks.
