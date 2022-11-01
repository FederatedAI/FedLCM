
# Deploy a FATE without engine

FATE deployed by default using FedLCM will include engines such as Spark HDFS and Pulsar. The default engine performance is not powerful. If there are already clusters of Spark HDFS and Pulsar outside, they can provide more power to FATE.

## Requirements

- Spark: Spark's worker nodes must contain FATE's python runtime dependencies.
- HDFS: FATE component fateflow must be able to access the HDFS name node and data node.
- Pulsar: Pulsar needs to configure the CA certificate to communicate with ATS.

## Create a FATE

First you need to deploy an exchange。[Creating Exchange](./Getting_Started_FATE.md#creating-exchange)

Then create a cluster.

choose "select certificates"  > "Skip, I will manually install certificates"

Then in the step Select Extrenal engine. Add the configuration information of the external engine.

***If using an external engine, all three engine information needs to be filled in, including Spark HDFS and Pulsar.***

After filling in, generate yaml, and then you can deploy FATE.

Next, you can use FATE for federated learning tasks.
