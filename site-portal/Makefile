.PHONY: all clean format swag swag-bin server-unittest server frontend run

RELEASE_VERSION ?= ${shell git describe --tags}
TAG ?= v0.2.0

SERVER_NAME ?= federatedai/site-portal-server
SERVER_IMG ?= ${SERVER_NAME}:${TAG}

FRONTEND_NAME ?= federatedai/site-portal-frontend
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
OUTPUT_FILE = ${OUTPUT_DIR}/site-portal.exe
else
BUILD_MODE =
OUTPUT_FILE = ${OUTPUT_DIR}/site-portal
endif
OUTPUT_FRONTEND_FOLDER = ${OUTPUT_DIR}/frontend

BRANCH ?= $(shell git symbolic-ref --short HEAD)
COMMIT ?= $(shell git log --pretty=format:'%h' -n 1)
NOW ?= $(shell date "+%Y-%m-%d %T UTC%z")

LDFLAGS = "-X 'github.com/FederatedAI/FedLCM/site-portal/server/constants.Branch=$(BRANCH)' \
           -X 'github.com/FederatedAI/FedLCM/site-portal/server/constants.Commit=$(COMMIT)' \
           -X 'github.com/FederatedAI/FedLCM/site-portal/server/constants.BuildTime=$(NOW)' \
           -extldflags '-static'"


all: swag server frontend

frontend:
	rm -rf ${OUTPUT_FRONTEND_FOLDER}
	mkdir -p ${OUTPUT_DIR}
	cd frontend && npm run build --prod
	cp -rf frontend/dist/Site-portal ${OUTPUT_FRONTEND_FOLDER}

# Run go fmt & vet against code
format:
	go fmt ./...
	go vet ./...

# Build manager binary
server: format
	mkdir -p ${OUTPUT_DIR}
	CGO_ENABLED=0 go build -a --ldflags ${LDFLAGS} -o ${OUTPUT_FILE} ${BUILD_MODE} server/main.go

# Run server tests
server-unittest: format
	go test ./... -coverprofile cover.out


run: format
	env FRONTEND_DIR=${OUTPUT_FRONTEND_FOLDER} go run --ldflags ${LDFLAGS} ./server/main.go

# Generate swag API file
swag: swag-bin
	cd server && $(SWAG_BIN) init --parseDependency --parseInternal

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
	docker build . -t ${SERVER_IMG} -f make/server/Dockerfile --build-arg BRANCH=$(BRANCH) --build-arg COMMIT=$(COMMIT)
	docker build . -t ${FRONTEND_IMG} -f make/frontend/Dockerfile

docker-push:
	docker push ${SERVER_IMG}
	docker push ${FRONTEND_IMG}

clean:
	rm -rf ${OUTPUT_DIR}
	rm -rf frontend/dist

release:
	rm -rf ${RELEASE_DIR}
	mkdir -p ${RELEASE_DIR}
	tar -czvf ${RELEASE_DIR}/site-portal-docker-compose-${RELEASE_VERSION}.tgz .env docker-compose.yml docker-compose-https.yml tls
