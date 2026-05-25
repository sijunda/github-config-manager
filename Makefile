# Git Config Manager - Makefile

BINARY_NAME=gcm
MODULE=git-config-manager
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-s -w -X $(MODULE)/pkg/version.Version=$(VERSION) -X $(MODULE)/pkg/version.Commit=$(COMMIT) -X $(MODULE)/pkg/version.Date=$(DATE)

.PHONY: build build-all test lint fmt verify clean install install-system release help

## Build

build: ## Build for current platform
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/gcm

build-all: ## Cross-compile for all platforms
	@echo "Building for all platforms..."
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-amd64  ./cmd/gcm
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-darwin-arm64  ./cmd/gcm
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64   ./cmd/gcm
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64   ./cmd/gcm
	GOOS=linux   GOARCH=arm   go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm     ./cmd/gcm
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/gcm

## Test

test: ## Run tests
	go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

test-verbose: ## Run tests with verbose output
	go test -race -v -coverprofile=coverage.out ./...

test-coverage: test ## Show coverage report in browser
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

bench: ## Run benchmarks
	go test -bench=. -benchmem ./...

## Lint & Format

lint: ## Run linters
	@echo "Running linters..."
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt: ## Format code
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed, skipping"; \
	fi

verify: fmt lint test ## Format, lint, and test everything

## Install

GOBIN_DIR ?= $(shell go env GOBIN)
ifeq ($(strip $(GOBIN_DIR)),)
GOBIN_DIR := $(shell go env GOPATH)/bin
endif

install: build ## Install to GOBIN (falls back to $(go env GOPATH)/bin)
	@mkdir -p "$(GOBIN_DIR)"
	@echo "Installing $(BINARY_NAME) -> $(GOBIN_DIR)/$(BINARY_NAME)"
	cp bin/$(BINARY_NAME) "$(GOBIN_DIR)/$(BINARY_NAME)"
	@echo "Done. Make sure $(GOBIN_DIR) is on your PATH."

install-system: build ## Install to /usr/local/bin (may require sudo)
	@echo "Installing $(BINARY_NAME) -> /usr/local/bin/$(BINARY_NAME)"
	install -m 0755 bin/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)

## Release

release: ## Create release with goreleaser
	goreleaser release --clean

release-snapshot: ## Create snapshot release (no publish)
	goreleaser release --snapshot --clean

## Clean

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out coverage.html

## Development

dev: build ## Build and run
	./bin/$(BINARY_NAME)

run: ## Run without building
	go run ./cmd/gcm $(ARGS)

## Dependencies

deps: ## Download dependencies
	go mod download
	go mod tidy

## Help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
