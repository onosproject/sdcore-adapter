# SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

# If any command in a pipe has nonzero status, return that status
SHELL = bash -o pipefail

export CGO_ENABLED=1
export GO111MODULE=on

.PHONY: build

KIND_CLUSTER_NAME           ?= kind
DOCKER_REPOSITORY           ?= onosproject/
ONOS_SDCORE_ADAPTER_VERSION ?= latest
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
	GOPRIVATE="github.com/onosproject/*" go build -v `go list ./...`
	bash -c "diff -u <(echo -n) <(git diff go.mod)"
	bash -c "diff -u <(echo -n) <(git diff go.sum)"

linters: golang-ci # @HELP examines Go source code and reports coding problems
	golangci-lint run --timeout 5m

build-tools: # @HELP install the ONOS build tools if needed
	@if [ ! -d "../build-tools" ]; then cd .. && git clone https://github.com/onosproject/build-tools.git; fi

jenkins-tools: # @HELP installs tooling needed for Jenkins
	cd .. && go get -u github.com/jstemmer/go-junit-report && go get github.com/t-yuki/gocover-cobertura

golang-ci: # @HELP install golang-ci if not present
	golangci-lint --version || curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b `go env GOPATH`/bin v1.42.0

license_check: build-tools # @HELP examine and ensure license headers exist
	./../build-tools/licensing/boilerplate.py -v --rootdir=${CURDIR} --boilerplate SPDX-Apache-2.0

# @HELP build the go binary in the cmd/sdcore-adapter package
build: local-aether-models
	go build -o build/_output/sdcore-adapter ./cmd/sdcore-adapter
	go build -o build/_output/sdcore-migrate ./cmd/sdcore-migrate

# @HELP run various tests
test: build unit-test deps license_check linters images

# @HELP run init tests
unit-test:
	go test -cover -race github.com/onosproject/sdcore-adapter/pkg/...
	go test -cover -race github.com/onosproject/sdcore-adapter/cmd/...


jenkins-test:  # @HELP run the unit tests and source code validation producing a junit style report for Jenkins
jenkins-test: build deps license_check linters images jenkins-tools
	TEST_PACKAGES=`go list github.com/onosproject/sdcore-adapter/...` ./../build-tools/build/jenkins/make-unit

sdcore-adapter-docker: local-aether-models
	docker build . -f Dockerfile \
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

jenkins-publish: build-tools jenkins-tools # @HELP Jenkins calls this to publish artifacts
	./build/bin/push-images
	../build-tools/release-merge-commit

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
