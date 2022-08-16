Charts for FATE deployment by the lifecycle-manager. The lifecycle-manager service uses it internally. It is unnecessary to use these charts directly unless advanced changes are required.
### Chart package

```bash
$ helm package ./fate
$ helm package ./fate-exchange
```

Using the command will generate a chart package of `*.tgz`.

### Chart verify

```bash
$ helm lint ./fate
$ helm lint ./fate-exchange
```

After modifying chart and before submitting the code, run verification to check whether there are obvious errors.

# Installation Guide

Use the 'kubefate' command line to upload chart and then deploy site-portal. For kubefate installation and use, refer to [this](https://github.com/FederatedAI/KubeFATE/tree/v1.6.1/k8s-deploy).

#### Clone

```bash
$ git clone <this project>
```

#### Package Charts

```bash
$ cd <project root>/helm-charts/charts
$ helm package ./fate
$ helm package ./fate-exchange
```

#### Upload Chart

```bash
$ kubefate chart upload -f fate-${version}.tgz
```

```bash
$ kubefate chart upload -f fate-exchange-${version}.tgz
```

#### Edit cluster configuration

Modify the cluster configuration `site-portal.yaml` and `fml-manager.yaml`

```bash
$ vi site-portal.yaml
```

```bash
$ vi fml-manager.yaml
```

#### Generate secret key

The secret of both parties needs to be generated before deployment, refer [here](https://github.com/FederatedAI/KubeFATE/blob/v1.6.1/docs/FATE_On_Spark_With_Pulsar.md#certificate-generation). In the steps of generating the secret key, you can get the secret key file of the corresponding party, including `certs/ca.cert.pem`, for the party it is `${party_id}.fate.org/broker.cert.pem` and `${party_id }.fate.org/broker.key-pk8.pem`, the corresponding exchange is `proxy.fate.org/proxy.cert.pem` and `proxy.fate.org/proxy.key.pem`.

```bash
# for party, import the certificate to k8s
$ kubectl create secret generic pulsar-cert -n {namespace} \
	--from-file=broker.cert.pem=9999.fate.org/broker.cert.pem \
	--from-file=broker.key-pk8.pem=9999.fate.org/broker.key-pk8.pem \
	--from-file=ca.cert.pem=certs/ca.cert.pem

```

```bash
# for exchange, import the certificate to k8s
$ kubectl create secret generic traffic-server-cert -n {namespace} \
	--from-file=proxy.cert.pem=proxy.fate.org/proxy.cert.pem \
	--from-file=proxy.key.pem=proxy.fate.org/proxy.key.pem \
	--from-file=ca.cert.pem=certs/ca.cert.pem
```

#### Deploy

```bash
# party
$ kubefate cluster install -f site-portal.yaml
```
```bash
# exchange
$ kubefate cluster install -f fml-manager.yaml
```

#### Check status

When you deploy the cluster, you will get a `job_UUID`

```bash
# View deployment status
$ kubefate job describe ${job_UUID}
```

When the job status is `Success`, it indicates that the deployment is successful, and then configure the host file as `${URL}` according to `ingress.frontend.hosts[0].name` in `site-portal.yaml`. 

It can also be obtained by command `kubectl get ingress -n ${FATE_namespace}`.

```bash
$ echo "${Ingress_IP} ${URL}"  >> /etc/hosts
```

> `${Ingress_IP}` is the IP to deploy ingress controller. It may be Node IP or Load Balancer IP depending on the deployment method.

Then open `${URL}` through the browser to use the site-portal.

When using site-portal, you need the IP and port of FATE-flow, which can be retrieved via `kubectl` command

```bash
$ kubectl get pod -n ${FATE_namespace} -l fateMoudle=python -o wide
```

This command can get the cluster IP of fate flow Pod. Port is the default 9380.

Follow the site-portal document to configure the service and start using it for FML management.