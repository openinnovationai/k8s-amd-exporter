APP_VERSION?=0.1.0
BASE_IMAGE_VERSION?=0.0.1
BASE_DEV_IMAGE_VERSION?=0.0.1
COMMIT_HASH?=$(shell git describe --dirty --tags --always)
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GOCMD=go
GOBUILD=$(GOCMD) build
CONTAINERTOOL?=docker
BINARY_NAME=amd-smi-exporter
SRC_FOLDER=cmd/exporterd
IMAGE?=amd-smi-exporter
BASE_IMAGE?=amd-smi-exporter-base
BASE_DEV_IMAGE?=amd-smi-exporter-base-dev
RELEASED_IMAGE ?= openinnovationai/$(IMAGE):$(APP_VERSION)
OPENINNOVATIONAI_IMAGE_NAME ?= amd_smi_exporter_v2
REGISTRY_SERVER ?= registry.gitlab.com/openinnovationai/platform/infra/amd/amd_smi_exporter_v2
OPENINNOVATIONAI_IMAGE ?= $(REGISTRY_SERVER)/$(IMAGE):$(APP_VERSION)
OPENINNOVATIONAI_BASE_IMAGE ?= $(REGISTRY_SERVER)/$(BASE_IMAGE):$(BASE_IMAGE_VERSION)
OPENINNOVATIONAI_BASE_DEV_IMAGE ?= $(REGISTRY_SERVER)/$(BASE_DEV_IMAGE):$(BASE_DEV_IMAGE_VERSION)
GO_MODULE_NAME=github.com/openinnovationai/k8s-amd-exporter
LDFLAGS_APPLICATION_VERSION=$(GO_MODULE_NAME)/internal/application.version=${APP_VERSION}
LDFLAGS_COMMIT_HASH=$(GO_MODULE_NAME)/internal/application.commitHash=${COMMIT_HASH}
LDFLAGS_BUILD_DATE=$(GO_MODULE_NAME)/internal/application.buildDate=${BUILD_DATE}
LDFLAGS?="-X $(LDFLAGS_APPLICATION_VERSION) -X $(LDFLAGS_COMMIT_HASH) -X $(LDFLAGS_BUILD_DATE) -s -w"

.PHONY: all
all: help

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: tidy
tidy: ## Run go mod tidy to organize dependencies.
	@$(GOCMD) mod tidy

.PHONY: test
test: ## Run unit tests
	@$(GOCMD) test -count=1 -race ./...

.PHONY: run-app
run-app: ## Run app
	@$(GOCMD) run -ldflags ${LDFLAGS} $(SRC_FOLDER)/main.go

.PHONY: build-linux
build-linux: ## Build binary for Linux taking GOARCH from env
	@mkdir -p bin
	CGO_ENABLED=1 GOOS=linux ${GOBUILD} -ldflags ${LDFLAGS} -o ./bin/${BINARY_NAME} ./${SRC_FOLDER}/main.go

.PHONY: build-base-dev-image
build-base-dev-image: ## build base image for building and development locally
	@${CONTAINERTOOL} build \
	-f deploy/Dockerfile.dev.base \
	-t ${BASE_DEV_IMAGE}:local-${BASE_DEV_IMAGE_VERSION} .

.PHONY: build-base-image
build-base-image: ## build base image locally to run exporter
	@${CONTAINERTOOL} build \
	-f deploy/Dockerfile.base \
	-t ${BASE_IMAGE}:local-${BASE_IMAGE_VERSION} .

.PHONY: build-image
build-image: ## build container image locally
	@${CONTAINERTOOL} build \
	--build-arg appVersion=${APP_VERSION} \
	--build-arg buildDate=${BUILD_DATE} \
	--build-arg commitHash=${COMMIT_HASH} \
	-f deploy/Dockerfile.ubuntu \
	-t ${IMAGE}:local-${APP_VERSION} .

.PHONY: tag-base-image
tag-base-image:
	${CONTAINERTOOL} tag ${BASE_IMAGE}:local-${BASE_IMAGE_VERSION} ${OPENINNOVATIONAI_BASE_IMAGE}

.PHONY: tag-base-dev-image
tag-base-dev-image:
	${CONTAINERTOOL} tag ${BASE_DEV_IMAGE}:local-${BASE_DEV_IMAGE_VERSION} ${OPENINNOVATIONAI_BASE_DEV_IMAGE}

.PHONY: tag-image
tag-image:
	${CONTAINERTOOL} tag ${IMAGE}:local-${APP_VERSION} ${OPENINNOVATIONAI_IMAGE}

.PHONY: push-image
push-image: tag-image
	${CONTAINERTOOL} push ${OPENINNOVATIONAI_IMAGE}

.PHONY: lint
lint: ## Run lint
	@$(CONTAINERTOOL) run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.0.2-alpine golangci-lint run --fix --timeout=10m