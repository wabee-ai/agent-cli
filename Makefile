# Wabee CLI Makefile

# Build variables
BINARY_NAME=wabee
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/wabee-ai/wabee-cli/cmd.Version=$(VERSION) -X github.com/wabee-ai/wabee-cli/cmd.Commit=$(COMMIT) -X github.com/wabee-ai/wabee-cli/cmd.BuildDate=$(BUILD_DATE)"

# Go variables
GOBIN?=$(shell go env GOPATH)/bin
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: all build install clean test lint fmt help

all: build

## build: Build the binary for the current platform
build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

## install: Install the binary to GOBIN
install:
	go install $(LDFLAGS) .

## build-all: Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .

## clean: Remove build artifacts
clean:
	rm -rf bin/
	go clean

## test: Run tests
test:
	go test -v ./...

## test-coverage: Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format code
fmt:
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## tidy: Tidy go modules
tidy:
	go mod tidy

## deps: Download dependencies
deps:
	go mod download

## release: Create a release using GoReleaser
release:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not installed. Run: go install github.com/goreleaser/goreleaser@latest"; \
	fi

## release-snapshot: Create a snapshot release (no publish)
release-snapshot:
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Run: go install github.com/goreleaser/goreleaser@latest"; \
	fi

## run: Build and run with arguments
run: build
	./bin/$(BINARY_NAME) $(ARGS)

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
