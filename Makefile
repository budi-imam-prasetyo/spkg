BINARY_NAME=spkg
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed -E 's/-[0-9]+-g[0-9a-f]+(-dirty)?//; s/-dirty$$//' || echo "v0.1.0")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME) coverage.out

install:
	sudo cp $(BINARY_NAME) /usr/local/bin/

uninstall:
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
