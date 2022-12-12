# FedLCM: Federation Lifecycle Manager

FedLCM是一个基于Web的联邦学习的联邦生命周期管理服务，支持联邦学习组件和应用的部署管理，联邦网络创建以及底层基础设施的安装配置等。

[文档](./doc) | [English](./README.md)

FedLCM同时包含有一个Site Portal服务，用于进行图形化的建模、数据、模型管理等功能。Site Portal可以通过FedLCM部署，也可以不部署FedLCM，直接使用Docker Compose运行 Site Portal，关于如何直接使用该服务，请参阅 [Site Portal 使用手册](./doc/Site_Portal_Manual_zh.md)。

## 使用 Docker Compose 部署 FedLCM

**系统要求**：安装有 docker 18+ 以及 docker-compose 1.28+ 的 Linux 系统

* 在 release 页面下载 `fedlcm-docker-compose-<version>.tgz` 安装包，并解压到指定文件夹。或者直接在本仓库目录下执行如下操作。
* **(可选)** 若使用自定义的镜像仓库或镜像，请修改 `.env` 文件中的镜像名称。
* 执行如下命令开启应用：

```shell
docker-compose pull
docker-compose up
```

应用成功开启后可通过主机地址及服务端口号（默认为 9080）访问 FedLCM 的网页。

## 部署至 Kubernetes 集群

* 在 release 界面下载 `fedlcm-k8s-<version>.tgz` 并解压至指定文件夹。或者直接在本仓库目录下执行如下操作。
* **(可选)** 建议使用持久化存储来避免 CA 根证书发生变化或者重启服务之后数据丢失。可以创建 `StorageClass` 以及相应的 provisioner，然后修改 `k8s_deploy.yaml` 中相关的注释内容来开启 `persistentVolumeClaim` 。注意需要替换 `storageClassName` 部分的值。
* 执行如下指令完成部署：

```shell
kubectl apply -f rbac_config.yaml
kubectl apply -f k8s_deploy.yaml
```

Web 界面默认使用 NodePort 服务并监听 30008 端口。待所有资源都成功创建并运行后，可以通过 `<NodeAddress>:30008` 访问 FedLCM 界面。如需修改相关配置，请自行调整上述 YAML 文件的内容。

## 快速开始使用 FedLCM

参见 [FATE 联邦管理指南](./doc/Getting_Started_FATE_zh.md)。

如果我们希望不通过FedLCM部署Site Portal，请参阅 [Site Portal 使用手册](./doc/Site_Portal_Manual_zh.md)。

## 本地开发

### Build

```shell
make all
```

生成的文件默认存放在 `./output` 目录下。 

### 打包并运行 Docker 镜像

* 修改 `.env` 文件中的镜像名称，然后执行：

```shell
set -a; source .env; set +a
make docker-build
```

* 可以使用如下命令快捷推送镜像

```shell
make docker-push
```

* 开启服务

```shell
docker-compose up
```

详情参见 [FedLCM 本地开发指南](./doc/Development_Guide_zh.md)。

## License

FedLCM 使用 [Apache 2 license](LICENSE).

本项目使用了有其他附加条款的开源组件，关于其官方镜像以及授权条款的详细信息参见如下链接：

* Photon OS: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)
