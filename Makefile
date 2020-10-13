SHELL := /bin/bash # Use bash syntax

# Set up variables
GO111MODULE=on

DOCKER_REPOSITORY ?= amazon/aws-controllers-k8s
DOCKER_USERNAME ?= ""
DOCKER_PASSWORD ?= ""

AWS_SERVICE=$(shell echo $(SERVICE) | tr '[:upper:]' '[:lower:]')

# Build ldflags
VERSION ?= "v0.0.0"
GITCOMMIT=$(shell git rev-parse HEAD)
BUILDDATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS=-ldflags "-X main.version=$(VERSION) \
			-X main.buildHash=$(GITCOMMIT) \
			-X main.buildDate=$(BUILDDATE)"

# We need to use the codegen tag when building and testing because the
# aws-sdk-go/private/model/api package is gated behind a build tag "codegen"...
GO_TAGS=-tags codegen

.PHONY: all build-ack-generate test clean-mocks mocks build-controller-image \
	build-controller kind-test delete-all-kind-clusters

all: test

build-ack-generate:	## Build ack-generate binary
	@echo -n "building ack-generate ... "
	@go build ${GO_TAGS} ${GO_LDFLAGS} -o bin/ack-generate cmd/ack-generate/main.go
	@echo "ok."

test: | mocks	## Run code tests
	go test ${GO_TAGS} ./...

clean-mocks:	## Remove mocks directory
	rm -rf mocks

build-controller-image:	## Build container image for SERVICE
	@./scripts/build-controller-image.sh -s $(AWS_SERVICE)

publish-controller-image:  ## docker push a container image for SERVICE
	@echo $(DOCKER_PASSWORD) | docker login -u $(DOCKER_USERNAME) --password-stdin
	./scripts/publish-controller-image.sh -r $(DOCKER_REPOSITORY) -s $(AWS_SERVICE)

build-controller: build-ack-generate ## Generate controller code for SERVICE
	@./scripts/build-controller.sh $(AWS_SERVICE)

kind-test: ## Run functional tests for SERVICE with AWS_ROLE_ARN
	@./scripts/kind-build-test.sh -s $(AWS_SERVICE) -p -r $(AWS_ROLE_ARN)

delete-all-kind-clusters:	## Delete all local kind clusters
	@kind get clusters | \
	while read name ; do \
	kind delete cluster --name $$name; \
	done
	@rm -rf build/tmp-test*

install-mockery:
	@scripts/install-mockery.sh

mocks: install-mockery ## Build mocks
	@echo -n "building mocks for pkg/types ... "
	@bin/mockery --quiet --all --tags=codegen --case=underscore --output=mocks/pkg/types --dir=pkg/types
	@echo "ok."

help:           ## Show this help.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep | sed -e 's/\\$$//' \
		| awk -F'[:#]' '{print $$1 = sprintf("%-30s", $$1), $$4}'
