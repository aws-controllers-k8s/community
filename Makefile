SHELL := /bin/bash # Use bash syntax
GO111MODULE=on

PKGS=$(sort $(dir $(wildcard pkg/*/*/)))
MOCKS=$(foreach x, $(PKGS), mocks/$(x))

MOCKERY_BIN=$(shell which mockery || "./bin/mockery")

# We need to use the codegen tag when building and testing because the
# aws-sdk-go/private/model/api package is gated behind a build tag "codegen"...
GO_TAGS=-tags codegen

.PHONY: all build-ack-generate test clean-mocks mocks build-controller-image build-controller build-kind-cluster build-kind-cluster-preserve \
        build-kind-cluster-functional kind-get-cluster-names delete-all-kind-clusters

all: test

build-ack-generate:	## Build ack-generate binary
	go build ${GO_TAGS} -o bin/ack-generate cmd/ack-generate/main.go

test: | mocks	## Run code tests
	go test ${GO_TAGS} ./...

clean-mocks:	## Remove mocks directory
	rm -rf mocks

build-controller-image:	## Build container image for SERVICE
	./scripts/build-controller-image.sh -s $(SERVICE)

build-controller: build-ack-generate	## Generate controller code for SERVICE
	./scripts/build-controller.sh $(SERVICE)

kind-cluster: test	## Run tests in a local kind cluster for SERVICE
	./scripts/kind-build-test.sh -s $(SERVICE)

kind-cluster-preserve: test	## Run tests in a local kind cluster for SERVICE and preserve cluster
	./scripts/kind-build-test.sh -s $(SERVICE) -p

kind-cluster-functional: test	## Run functional tests for SERVICE with ROLE_ARN
	./scripts/kind-build-test.sh -s $(SERVICE) -p -r $(ROLE_ARN)

kind-get-cluster-names:	## Get local kind clusters
	@kind get clusters

delete-all-kind-clusters:	## Delete all local kind clusters
	@kind get clusters | \
	while read name ; do \
	kind delete cluster --name $$name; \
	done
	@rm -rf build/tmp-test*

mocks: $(MOCKS)	## Run mock tests

$(MOCKS): mocks/% : %
	${MOCKERY_BIN} --tags=codegen --case=underscore --output=$@ --dir=$^ --all

help:           ## Show this help.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v grep | sed -e 's/\\$$//' \
		| awk -F'[:#]' '{print $$1 = sprintf("%-30s", $$1), $$4}'
