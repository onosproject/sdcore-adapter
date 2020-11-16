# SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

# If any command in a pipe has nonzero status, return that status
SHELL = bash -o pipefail

export CGO_ENABLED=0
export GO111MODULE=on

.PHONY: build

KIND_CLUSTER_NAME           ?= kind
DOCKER_REPOSITORY           ?= onosproject/
ONOS_SDCORE_ADAPTER_VERSION ?= latest
ONOS_BUILD_VERSION          ?= v0.6.0
LOCAL_AETHER_MODELS         ?=

all: build images

images: # @HELP build simulators image
images: sdcore-adapter-docker

.PHONY: local-aether-models
local-aether-models:
ifdef LOCAL_AETHER_MODELS
	rm -rf ./local-aether-models
	cp -a ${LOCAL_AETHER_MODELS} ./local-aether-models
endif

deps: # @HELP ensure that the required dependencies are in place
	go build -v ./...
	bash -c "diff -u <(echo -n) <(git diff go.mod)"
	bash -c "diff -u <(echo -n) <(git diff go.sum)"

linters: # @HELP examines Go source code and reports coding problems
	golangci-lint run

license_check: # @HELP examine and ensure license headers exist
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR} LicenseRef-ONF-Member-1.0


# @HELP build the go binary in the cmd/sdcore-adapter package
build: local-aether-models
	go build -o build/_output/sdcore-adapter ./cmd/sdcore-adapter

test: build deps license_check linters unit-tests

go-junit-report:
	which go-junit-report || go get -u github.com/jstemmer/go-junit-report

unit-tests: go-junit-report
	@mkdir -p ./tests/results
	@go test -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	go-junit-report < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	exit $$RETURN

sdcore-adapter-docker: local-aether-models
	docker build . -f Dockerfile \
	--build-arg ONOS_BUILD_VERSION=${ONOS_BUILD_VERSION} \
	--build-arg LOCAL_AETHER_MODELS=${LOCAL_AETHER_MODELS} \
	-t ${DOCKER_REPOSITORY}sdcore-adapter:${ONOS_SDCORE_ADAPTER_VERSION}

kind: # @HELP build Docker images and add them to the currently configured kind cluster
kind: images kind-only

kind-only: # @HELP deploy the image without rebuilding first
kind-only:
	@if [ "`kind get clusters`" = '' ]; then echo "no kind cluster found" && exit 1; fi
	kind load docker-image --name ${KIND_CLUSTER_NAME} ${DOCKER_REPOSITORY}sdcore-adapter:${ONOS_SDCORE_ADAPTER_VERSION}

publish: # @HELP publish version on github and dockerhub
	./../build-tools/publish-version ${VERSION} onosproject/sdcore-adapter

clean: # @HELP remove all the build artifacts
	rm -rf ./build/_output
	rm -rf ./vendor
	rm -rf ./cmd/sdcore-adapter/sdcore-adapter

help:
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST) \
    | sort \
    | awk ' \
        BEGIN {FS = ": *# *@HELP"}; \
        {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}; \
    '
