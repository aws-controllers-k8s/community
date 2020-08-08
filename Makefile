GO111MODULE=on

PKGS=$(sort $(dir $(wildcard pkg/*/*/)))
MOCKS=$(foreach x, $(PKGS), mocks/$(x))

# We need to use the codegen tag when building and testing because the
# aws-sdk-go/private/model/api package is gated behind a build tag "codegen"...
GO_TAGS=-tags codegen

.PHONY: all build-ack-generate test clean-mocks mocks

all: test

build-ack-generate:
	go build ${GO_TAGS} -o bin/ack-generate cmd/ack-generate/main.go

test: | mocks
	go test ${GO_TAGS} ./...

clean-mocks:
	rm -rf mocks

mocks: $(MOCKS)

$(MOCKS): mocks/% : %
	./bin/mockery --tags=codegen --case=underscore --output=$@ --dir=$^ --all
