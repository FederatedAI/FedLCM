.PHONY: all clean format swag swag-bin server-unittest server frontend run upgrade openfl-device-agent release

RELEASE_VERSION ?= ${shell git describe --tags}
TAG ?= v0.1.0

SERVER_NAME ?= federatedai/fedlcm-server
SERVER_IMG ?= ${SERVER_NAME}:${TAG}

FRONTEND_NAME ?= federatedai/fedlcm-frontend
FRONTEND_IMG ?= ${FRONTEND_NAME}:${TAG}

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

OUTPUT_DIR = output
RELEASE_DIR = ${OUTPUT_DIR}/release
ifeq ($(OS),Windows_NT)
BUILD_MODE = -buildmode=exe
OUTPUT_FILE = ${OUTPUT_DIR}/lifecycle-manager.exe
else
BUILD_MODE =
OUTPUT_FILE = ${OUTPUT_DIR}/lifecycle-manager
endif
OUTPUT_FRONTEND_FOLDER = ${OUTPUT_DIR}/frontend

BRANCH = $(shell git symbolic-ref --short HEAD)
COMMIT = $(shell git log --pretty=format:'%h' -n 1)
NOW = $(shell date "+%Y-%m-%d %T UTC%z")

LDFLAGS = "-X 'github.com/FederatedAI/FedLCM/server/constants.Branch=$(BRANCH)' \
           -X 'github.com/FederatedAI/FedLCM/server/constants.Commit=$(COMMIT)' \
           -X 'github.com/FederatedAI/FedLCM/server/constants.BuildTime=$(NOW)' \
           -extldflags '-static'"


all: swag server frontend openfl-device-agent

frontend:
	rm -rf ${OUTPUT_FRONTEND_FOLDER}
	mkdir -p ${OUTPUT_DIR}
	cd frontend && npm run build --prod
	cp -rf frontend/dist/lifecycle-manager ${OUTPUT_FRONTEND_FOLDER}

# Run go fmt & vet against code
format:
	go fmt ./...
	go vet ./...

# Build manager binary
server: format
	mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=0 go build -a --ldflags ${LDFLAGS} -o ${OUTPUT_FILE} ${BUILD_MODE} server/main.go

# Build the cmd line program
openfl-device-agent: format
	mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=0 go build -a --ldflags ${LDFLAGS} -o ${OUTPUT_DIR}/openfl-device-agent ${BUILD_MODE} cmd/device-agent/device-agent.go

# Run server tests
server-unittest: format
	go test ./... -coverprofile cover.out

run: format
	go run --ldflags ${LDFLAGS} ./server/main.go

# Generate swag API file
swag: swag-bin
	cd server && $(SWAG_BIN) fmt ;\
	$(SWAG_BIN) init --parseDependency --parseInternal

swag-bin:
ifeq (, $(shell which swag))
	@{ \
	set -e ;\
	SWAG_BIN_TMP_DIR=$$(mktemp -d) ;\
	cd $$SWAG_BIN_TMP_DIR ;\
	go mod init tmp ;\
	go get -u github.com/swaggo/swag/cmd/swag ;\
	rm -rf $$SWAG_BIN_TMP_DIR ;\
	}
SWAG_BIN=$(GOBIN)/swag
else
SWAG_BIN=$(shell which swag)
endif

docker-build:
	docker build . -t ${SERVER_IMG} -f make/server/Dockerfile
	docker build . -t ${FRONTEND_IMG} -f make/frontend/Dockerfile

docker-push:
	docker push ${SERVER_IMG}
	docker push ${FRONTEND_IMG}

clean:
	rm -rf ${OUTPUT_DIR}
	rm -rf frontend/dist

upgrade:
	go get -u ./...

release:
	rm -rf ${RELEASE_DIR}
	mkdir -p ${RELEASE_DIR}
	tar -czvf ${RELEASE_DIR}/fedlcm-k8s-${RELEASE_VERSION}.tgz rbac_config.yaml k8s_deploy.yaml
	tar -czvf ${RELEASE_DIR}/fedlcm-docker-compose-${RELEASE_VERSION}.tgz .env docker-compose.yml make/stepca
