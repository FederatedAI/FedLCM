# Guide for Local Development of FedLCM

## Requirements

In this section you can have a glance of the minimum and recommended versions of the tools needed to build/debug FedLCM.

These tools and services needs to be installed:

| Tool   | Link                                                                      | Minimum | Recommended |
|--------|---------------------------------------------------------------------------|---------|-------------|
| npm    | [link](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) | >= 6    | >= 7        |
| NodeJS | [link](https://nodejs.org/en/)                                            | >= 14   | >= 16       |
| golang | [link](https://go.dev/dl/)                                                | >= 1.19 | >= 1.19     |

## Service Dependency

FedLCM depends on a [PostgreSQL](https://www.postgresql.org/docs/current/) Database and a [StepCA](https://smallstep.com/docs/step-ca/getting-started) service. Refer to their documents on how to setup one.

## Quick Setup and Run

Clone the repository:
<!-- TODO -->
```shell
git clone $URL
```

### Start Backend Server
Configure necessary environment variables for connecting with PostgreSQL:

| Name              | Description                             | Required |
|-------------------|-----------------------------------------|----------|
| POSTGRES_HOST     | the address of the postgres db          | Yes      |
| POSTGRES_PORT     | the port of the postgres db             | Yes      |
| POSTGRES_USER     | the username of the postgres connection | Yes      |
| POSTGRES_PASSWORD | the password of the postgres connection | Yes      |
| POSTGRES_DB       | the db name of the postgres             | Yes      |

Go to `./server` directory and run:

```shell
go run main.go
```

If the process goes well, backend server will listen and serving HTTP on port `8080`.

Alternatively, you can configure your favorite IDE to start this service.

### Start Frontend Server

1. Run `npm install` under "frontend" directory.
2. create "proxy.config.json" file under "frontend" directory, and replace the URL of target with your available backend server.
```
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
3. Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The app will automatically reload if you change any of the source files.
4. Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.

> Default username is `Admin` with password `admin`.

### Other Useful Commands

```shell
# Build both frontend and backend into `./output` directory.
make all

# Run server tests
make server-unittest

# Run frontend tests
cd frontend && npm test

# Build docker image
make docker-build

# Push docker image
make docker-push
```

## Other configurable environment variables

| Name                                    | Description                                                    | Required                          |
|-----------------------------------------|----------------------------------------------------------------|-----------------------------------|
| POSTGRES_DEBUG                          | whether or not to enable postgres debug level logs             | No, default to false              |
| POSTGRES_SSLMODE                        | whether or not to enable ssl connection to DB                  | No, default to false              |
| LIFECYCLEMANAGER_INITIAL_ADMIN_PASSWORD | initial admin user password, only takes effect on first start  | No, default to "admin"            |
| LIFECYCLEMANAGER_SECRETKEY              | a string of secret key for encrypting sensitive data in the DB | No, default to "passphrase123456" |
| LIFECYCLEMANAGER_DEBUG                  | true or false to enable debug log                              | No, default to false              |
| LIFECYCLEMANAGER_EXPERIMENT_ENABLED     | true of false to enable OpenFL management                      | No, default to false              |
| LIFECYCLEMANAGER_JWT_KEY                | a string of secret key for generating JWT token                | No, default to a random one       |

## Development

### Frontend

The frontend is based on [Clarity](https://clarity.design/) and [Angular](https://angular.io/).

### Backend

We use [Gin framework](https://github.com/gin-gonic/gin) to handle API requests and [GORM](https://gorm.io/index.html) to persist data.

The code are organized as below:

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