BINARY_NAME=etoro-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build build-all clean install test

## Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

## Cross-compile for all platforms
build-all: clean
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .

## Install to $GOPATH/bin
install:
	go install $(LDFLAGS) .

## Run tests
test:
	go test ./...

## Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf dist/
