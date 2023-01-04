# FedLCM 本地开发指南

## 环境要求

本地搭建 FedLCM 开发环境所需工具最低版本要求以及推荐版本如下：

| 开发工具   | 下载链接                                                                      | 最低版本    | 推荐版本    |
|--------|---------------------------------------------------------------------------|---------|---------|
| npm    | [link](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) | >= 6    | >= 7    |
| NodeJS | [link](https://nodejs.org/en/)                                            | >= 14   | >= 16   |
| golang | [link](https://go.dev/dl/)                                                | >= 1.19 | >= 1.19 |

## 其他依赖服务

运行 FedLCM 需要提前部署 [PostgreSQL](https://www.postgresql.org/docs/current/) 数据库服务和 [StepCA](https://smallstep.com/docs/step-ca/getting-started) 证书服务。详情参见相关文档。

## 快速开始

将代码仓库克隆到本地:
<!-- TODO -->
```shell
git clone $URL
```

### 启动后端服务

配置连接数据库所需的环境变量：

| 变量名               | 备注             | 是否必需 |
|-------------------|----------------|------|
| POSTGRES_HOST     | postgres 数据库地址 | 是    |
| POSTGRES_PORT     | postgres 数据库端口 | 是    |
| POSTGRES_USER     | 数据库用户名         | 是    |
| POSTGRES_PASSWORD | 数据库密码          | 是    |
| POSTGRES_DB       | 数据库名称          | 是    |

在命令行切换到 `./server` 目录并执行如下指令:

```shell
go run main.go
```

后端服务成功运行后默认监听 `8080` 端口。

除使用命令行外，也可以用你熟悉的 IDE 来开启后端服务。

### 启动前端服务

1. 在 `./frontend` 目录下执行 `npm install` 指令。
2. 在 `./frontend` 目录下创建“proxy.config.json”文件并将其中的“target”项替换为后端服务的地址。

```json
 {
    "/api/v1": {
      "target": "http://localhost:8080",
      "secure": false,
      "changeOrigin": true,
      "logLevel": "debug",
        "headers": {
            "Connection": "keep-alive"
        }
    }
  }
```

1. 执行 `ng serve` 指令开启前端服务，打开 `http://localhost:4200/` 即可访问。如果源文件发生变化，服务会自动重启。
2. 执行 `ng build` 指令构建前端项目。生成的文件默认存放在 `dist/` 目录下。

> 默认用户名为`Admin`，密码为`admin`。

### 其他常用指令

```shell
# 将前后端服务打包至“./output”目录下
make all

# 运行单元测试
make server-unittest

# 运行前端测试
cd frontend && npm test

# 构建 docker 镜像
make docker-build

# 推送 docker 镜像
make docker-push
```

## 其他环境变量

| 变量名                                     | 备注                           | 是否必需                     |
|-----------------------------------------|------------------------------|--------------------------|
| POSTGRES_DEBUG                          | 是否开启 postgres 数据库 debug 级别日志 | 否，默认为 false              |
| POSTGRES_SSLMODE                        | 是否使用 ssl 连接数据库               | 否，默认为 false              |
| LIFECYCLEMANAGER_INITIAL_ADMIN_PASSWORD | 初始 admin 账户密码，只在第一次启动时生效     | 否，默认为 "admin"            |
| LIFECYCLEMANAGER_SECRETKEY              | 加密数据库敏感数据所用密钥                | 否，默认为 "passphrase123456" |
| LIFECYCLEMANAGER_DEBUG                  | 是否开启 debug 级别日志              | 否，默认为 false              |
| LIFECYCLEMANAGER_EXPERIMENT_ENABLED     | 是否开启 OpenFL 管理服务             | 否，默认为 false              |
| LIFECYCLEMANAGER_JWT_KEY                | 生成 JWT token 的密钥             | 否，默认为随机值                 |

## 技术栈简介

### 前端项目

前端基于 [Clarity](https://clarity.design/) 和 [Angular](https://angular.io/) 实现。

### 后端项目

使用 [Gin framework](https://github.com/gin-gonic/gin) 处理 API 请求，使用 [GORM](https://gorm.io/index.html) 持久化数据。

代码目录结构如下:

```C
pkg
├── kubefate         // KubeFATE management and client code
├── kubernetes       // K8s client code
└── utils            // some basic util functions
server
├── api              // Gin route and API handlinig
├── application      // App services called by the handlers in api level
├── constants        // some constants
├── docs             // swagger docs
├── domain           // domain driven design implementation of the main workflow
├── infrastructure   // client to other system, GORM logics etc. that can be used by other layers
└── main.go          // entry point
```
