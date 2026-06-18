BINARY_NAME=webull
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(shell go version | awk '{print $$3}')

LDFLAGS=-ldflags "-s -w \
	-X github.com/webull-inc/webull-openapi-cli/pkg/version.Version=$(VERSION) \
	-X github.com/webull-inc/webull-openapi-cli/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/webull-inc/webull-openapi-cli/pkg/version.BuildTime=$(BUILD_TIME) \
	-X github.com/webull-inc/webull-openapi-cli/pkg/version.GoVersion=$(GO_VERSION)"

.PHONY: build test lint clean install

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/webull

build-all:
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/webull
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/webull
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/webull
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/webull
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/webull

install:
	go install $(LDFLAGS) ./cmd/webull

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

all: fmt vet lint test build
