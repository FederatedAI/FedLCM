# FedLCM: Federation Lifecycle Manager

A web application to manage FML deployments lifecycles

[DOC](./doc) | [中文](./README_zh.md)

## Deploy & Run Using Docker Compose
**System requirements:** On a Linux host with docker 18+ and docker-compose 1.28+ .
* Download the `fedlcm-docker-compose-<version>.tgz` package from the releases pages, unzip the files into a folder. Or clone the whole project and execute in the project root folder.
* **(Optional)** modify `.env` file to change the image names. Only do this if you want to use your customized registry or images.
* Then
```
docker-compose pull
docker-compose up
```
By default, the web UI is exposed on the 9080 port of the machine. Access that address from a browser to open the FedLCM's UI.

## Deploy Into Kubernetes
* Download the `fedlcm-k8s-<version>.tgz` package from the releases pages, and unzip the files into a folder. Or clone the whole project and execute in the project root folder.
* **(Optional)** We recommend using persistent storage to avoid the change of CA root certificate and the loss of data when restarting the deployment. To enable persistent volume, you should create your `StorageClass` and corresponding provisioner first. Then modify the commented section in the `k8s_deploy.yaml` to use `persistentVolumeClaim`. Replace with your storage class name at `storageClassName`.
* Apply the yaml files in the follow order
```
kubectl apply -f rbac_config.yaml
kubectl apply -f k8s_deploy.yaml
```
By default, the web UI is exposed via a NodePort service that listens on port 30008 of the nodes. After all resources are successfully created and running, you can enter into your FedLCM service use `<NodeAddress>:30008`. Alternatively you can change the Service definition in the yaml to use your preferred service type and exposed port.

## Getting Started

Refer to the [getting started guide](./doc/Getting_Started_FATE.md) for FATE federation management.

## Development
### Build
```
make all
```
The generated deliverables are placed in the `output` folder.

### Build & Run Docker Image
* Modify `.env` file to change the image name, and then
```
set -a; source .env; set +a
make docker-build
```
* Optionally push the image
```
make docker-push
```
* Start the service
```
docker-compose up
```

Refer to the [development guide](./doc/Development_Guide.md) for more development related topics.

## License

FedLCM is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Photon OS: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)