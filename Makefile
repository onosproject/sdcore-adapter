# SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

# If any command in a pipe has nonzero status, return that status
SHELL = bash -o pipefail

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

VERSION                     ?= $(shell cat ./VERSION)

KIND_CLUSTER_NAME           ?= kind
DOCKER_REPOSITORY           ?= onosproject/
ONOS_SDCORE_ADAPTER_VERSION ?= latest
LOCAL_AETHER_MODELS         ?=

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL        ?= $(shell git remote get-url $(shell git remote))
DOCKER_LABEL_VCS_REF        = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE     ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE    = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_EXTRA_ARGS           ?=
DOCKER_BUILD_ARGS ?= \
        ${DOCKER_EXTRA_ARGS} \
        --build-arg org_label_schema_version="${VERSION}" \
        --build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
        --build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
        --build-arg org_label_schema_build_date="${DOCKER_LABEL_BUILD_DATE}" \
        --build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
        --build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}" \
		--build-arg LOCAL_AETHER_MODELS=${LOCAL_AETHER_MODELS}

all: build images

build-tools:=$(shell if [ ! -d "./build/build-tools" ]; then cd build && git clone https://github.com/onosproject/build-tools.git; fi)
include ./build/build-tools/make/onf-common.mk

images: # @HELP build simulators image
images: sdcore-adapter-docker

.PHONY: local-aether-models
local-aether-models:
ifdef LOCAL_AETHER_MODELS
	rm -rf ./local-aether-models
	cp -a ${LOCAL_AETHER_MODELS} ./local-aether-models
endif

# @HELP build the go binary in the cmd/sdcore-adapter package
build: local-aether-models
	go build -o build/_output/sdcore-adapter ./cmd/sdcore-adapter
	go build -o build/_output/sdcore-migrate ./cmd/sdcore-migrate

# @HELP run various tests
test: build unit-test deps license linters images

# @HELP run init tests
unit-test:
	go test -cover -race github.com/onosproject/sdcore-adapter/pkg/...
	go test -cover -race github.com/onosproject/sdcore-adapter/cmd/...

jenkins-test:  # @HELP run the unit tests and source code validation producing a junit style report for Jenkins
jenkins-test: build deps license linters images jenkins-tools
	TEST_PACKAGES=`go list github.com/onosproject/sdcore-adapter/...` ./build/build-tools/build/jenkins/make-unit

sdcore-adapter-docker: local-aether-models
	docker build . -f Dockerfile \
	$(DOCKER_BUILD_ARGS) \
	-t ${DOCKER_REPOSITORY}sdcore-adapter:${ONOS_SDCORE_ADAPTER_VERSION}

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images kind-only

kind-only: # @HELP deploy the image without rebuilding first
kind-only:
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image --name ${KIND_CLUSTER_NAME} ${DOCKER_REPOSITORY}sdcore-adapter:${ONOS_SDCORE_ADAPTER_VERSION}

publish: # @HELP publish version on github and dockerhub
	./build/build-tools/publish-version ${VERSION} onosproject/sdcore-adapter

jenkins-publish: # @HELP Jenkins calls this to publish artifacts
	./build/bin/push-images
	./build/build-tools/release-merge-commit

clean:: # @HELP remove all the build artifacts
	rm -rf ./build/_output
	rm -rf ./vendor
	rm -rf ./cmd/sdcore-adapter/sdcore-adapter
