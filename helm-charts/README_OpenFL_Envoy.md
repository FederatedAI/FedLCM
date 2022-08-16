Charts for OpenFL Envoy deployment. The lifecycle-manager service uses it internally. Please contact the maintainers of this project for its detailed usage.

### Chart package

```bash
helm package ./openfl-envoy
```

Using the command will generate a chart package of `*.tgz`.

### Chart verify

```bash
$ helm lint ./openfl-envoy
```

After modifying chart and before submitting the code, run verification to check whether there are obvious errors.

# Installation Guide

Use the 'kubefate' command line to upload chart and then deploy the application. For kubefate installation and use, refer to [this](https://github.com/FederatedAI/KubeFATE/tree/v1.6.1/k8s-deploy).

### Clone

```bash
$ git clone <this project>
```

### Package Charts

```bash
$ cd cd <project root>/helm-charts/charts
$ helm package ./openfl-envoy
```

### Upload Chart

```bash
$ kubefate chart upload -f openfl-envoy-${version}.tgz
```

### Generate certificates
This example uses example steps from KubeFATE project for pulsar deployment. Although the application is different, the steps are similar.

**If you have already created the root CA certificate, during deploying the Director, you can jump directly to create Envoy certificate**
#### Generate the secret key
Preparations:
```bash
$ mkdir my-ca
$ cd my-ca
$ wget https://raw.githubusercontent.com/apache/pulsar/master/site2/website/static/examples/openssl.cnf
$ export CA_HOME=$(pwd)
$ mkdir certs crl newcerts private
$ chmod 700 private/
```
Generate the private key for root certificate:
```bash
$ openssl genrsa -aes256 -out private/ca.key.pem 4096
```
Need to add a password for the private key, then:
```bash
$ chmod 400 private/ca.key.pem
```
The database files for the certificate management:
```bash
$ touch index.txt
$ echo 1000 > serial
```
Generate the root certificate:
```bash
$ openssl req -config openssl.cnf -key private/ca.key.pem \
    -new -x509 -days 7300 -sha256 -extensions v3_ca \
    -out certs/ca.cert.pem
```
No need to fill anything about the prompt questions, then:
```bash
$ chmod 444 certs/ca.cert.pem
```
Once the above commands are completed, the CA-related certificates and keys have been generated, they are:

* certs/ca.cert.pem: the certification of CA
* private/ca.key.pem: the key of CA

The next step is to generate another of private keys and certificates for Envoy:
```bash
$ mkdir envoy-1
$ openssl genrsa -out envoy-1/priv.key 2048
$ openssl req -config openssl.cnf -key envoy-1/priv.key -new -sha256 -out envoy-1/envoy.csr
```
When it prompts "common name", it should be a unique name, such as "envoy-1".
```bash
openssl ca -config openssl.cnf -days 10000 -notext -md sha256 -in envoy-1/envoy.csr -out envoy-1/envoy.crt
```

After doing all the generating, run below command to add above generated files to kubernetes cluster as a secret:
```bash
$ kubectl create namespace openfl-envoy-1
$ kubectl -n openfl-envoy-1 create secret generic envoy-cert --from-file=envoy.crt=envoy-1/envoy.crt --from-file=priv.key=envoy-1/priv.key --from-file=root_ca.crt=certs/ca.cert.pem
```

### Prepare the envoy shard descriptor file
Every Envoy is matched to one shard descriptor in order to run. When the Director starts an experiment, the Envoy accepts the experiment workspace, prepares the environment, and starts a Collaborator.
The shard descriptor is a python file. In this document, we use one example from [here](https://github.com/intel/openfl/blob/develop/openfl-tutorials/interactive_api/Tensorflow_MNIST/envoy/mnist_shard_descriptor.py), as the example in the Director document is also using this one.
For more information, please check [here](https://openfl.readthedocs.io/en/latest/running_the_federation.html?highlight=shard%20descriptor%20#step-2-start-the-envoy), and also search for `shard_descriptor`
on this [page](https://openfl.readthedocs.io/en/latest/running_the_federation.html?highlight=shard%20descriptor%20#run-the-federation).

The user is responsible for preparing a directory called `python` and put the shard descriptor python file as well as an optional `requirements.txt` file containing necessary python libs the python code needs into it, and add it as a configmap:
```bash
$ kubectl create configmap envoy-python-configs --from-file=python/ -n openfl-envoy-1
```

**NOTE**: due to the limitation of configmap rules, all filenames can only contain alphanumeric characters, `-`, `_` or `.`, other characters, including `space`, are not allowed. 

### Edit cluster configuration

We can create a file called `oe_cluster.yaml` and add below contents. For the [Tensorflow_MNIST](https://github.com/intel/openfl/tree/develop/openfl-tutorials/interactive_api/Tensorflow_MNIST) example, the content would be:
```
name: envoy-1
namespace: openfl-envoy-1
chartName: openfl-envoy
chartVersion: v0.1.0
registry: ""
pullPolicy: IfNotPresent
podSecurityPolicy:
  enabled: false
modules:
  - envoy

envoy:
  image: fedlcm-openfl
  imageTag: v0.1.0
  directorFqdn: director
  directorIp: 192.168.1.1
  directorPort: 50051
  aggPort: 50052
  envoyConfigs:
    params:
      cuda_devices: []
    optional_plugin_components: {}
    shard_descriptor:
      template: mnist_shard_descriptor.MnistShardDescriptor
      params:
        rank_worldsize: 1, 2
```

There are several things to note:
* The `name` field must be the `common name` configured in the certificate generated above.
* The `directorFqdn` field must be the `common name` configured in the director's certificate.
* The `directorIp` is the director service's exposed IP address. This can vary based on the director service's type.
* The `directorPort` and `aggPort` are also the exposed ports in the director's service. In general, these settings are the access information of the director service.
* The content for `envoyConfigs` is the yaml config that will be feed to envoy.

### Deploy and check status

Now we can install above cluster:
```bash
$ kubefate cluster install -f oe-cluster.yaml
```

When you deploy the cluster, you will get a `job_UUID`

```bash
# View deployment status
$ kubefate job describe ${job_UUID}
```

When the job status is` Success`, it indicates that the deployment is successful. 
The envoy will automatically register to the director. Check the logs of both the director and envoy to see if that step succeeded. 

### Next steps
You can add more envoys with similar steps as above, with new certificate and new names. And once you have all the envoys deployed and ready, you can run the jupyter notebook as suggested in the director's document to start the training.
