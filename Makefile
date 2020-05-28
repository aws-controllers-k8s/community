GO111MODULE=on

PKGS=$(sort $(dir $(wildcard pkg/*/*/)))
MOCKS=$(foreach x, $(PKGS), mocks/$(x))

.PHONY: all test clean-mocks mocks

all: test

test: | mocks
	go test ./...

clean-mocks:
	rm -rf mocks

mocks: ensure-mockery $(MOCKS)

$(MOCKS): mocks/% : %
	mockery -case=underscore -output=$@ -dir=$^ -all

ensure-mockery:
	@mockery -version 2>&1 >/dev/null || go get github.com/vektra/mockery/cmd/mockery
