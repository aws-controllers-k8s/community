SHELL := /bin/bash # Use bash syntax
GO111MODULE=on

PKGS=$(sort $(dir $(wildcard pkg/*/*/)))
MOCKS=$(foreach x, $(PKGS), mocks/$(x))

MOCKERY_BIN=$(shell which mockery || "./bin/mockery")

# We need to use the codegen tag when building and testing because the
# aws-sdk-go/private/model/api package is gated behind a build tag "codegen"...
GO_TAGS=-tags codegen

.PHONY: all build-ack-generate test clean-mocks mocks build-controller build-kind-cluster build-kind-cluster-preserve \
        build-kind-cluster-functional kind-get-cluster-names delete-all-kind-clusters

all: test

build-ack-generate:
	go build ${GO_TAGS} -o bin/ack-generate cmd/ack-generate/main.go

test: | mocks
	go test ${GO_TAGS} ./...

clean-mocks:
	rm -rf mocks

build-controller:
	./scripts/build-controller.sh $(SERVICE)

kind-cluster: test
	./scripts/kind-build-test.sh -s $(SERVICE)

kind-cluster-preserve: test
	./scripts/kind-build-test.sh -s $(SERVICE) -p

kind-cluster-functional: test
	./scripts/kind-build-test.sh -s $(SERVICE) -p -r $(ROLE_ARN)

kind-get-cluster-names:
	@kind get clusters

delete-all-kind-clusters:
	@kind get clusters | \
	while read name ; do \
	kind delete cluster --name $$name; \
	done
	@rm -rf build/tmp-test*

mocks: $(MOCKS)

$(MOCKS): mocks/% : %
	${MOCKERY_BIN} --tags=codegen --case=underscore --output=$@ --dir=$^ --all
