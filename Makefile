# Makefile for Packer Plugin VergeIO

# Default shell
SHELL := /bin/bash

# Plugin information
PLUGIN_NAME := packer-plugin-vergeio
MODULE_PATH := github.com/verge-io/packer-plugin-vergeio

# Version information
VERSION := $(shell cat version/version.go | grep 'Version.*=' | head -1 | cut -d'"' -f2)
VERSION_PRERELEASE := dev

# Build flags
LDFLAGS := -X $(MODULE_PATH)/version.Version=$(VERSION)
LDFLAGS += -X $(MODULE_PATH)/version.VersionPrerelease=$(VERSION_PRERELEASE)

# Go build flags
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# Binary name
BINARY := $(PLUGIN_NAME)_v$(VERSION)_$(VERSION_PRERELEASE)_$(GOOS)_$(GOARCH)

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: ## Build the plugin binary
	@echo "Building $(PLUGIN_NAME)..."
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

.PHONY: dev
dev: build install ## Build and install plugin for local development
	@echo "Development build complete. Plugin installed for local use."

.PHONY: install
install: ## Install the plugin locally
	@echo "Installing plugin locally..."
	packer plugins install --path $(BINARY) github.com/verge-io/vergeio

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(PLUGIN_NAME)*
	rm -f packer-plugin-vergeio*

##@ Testing

.PHONY: test
test: ## Run unit tests
	@echo "Running unit tests..."
	go test ./... -v

.PHONY: testacc
testacc: ## Run acceptance tests (requires PACKER_ACC=1)
	@echo "Running acceptance tests..."
	@if [ -z "$(PACKER_ACC)" ]; then \
		echo "PACKER_ACC must be set to run acceptance tests"; \
		echo "Usage: PACKER_ACC=1 make testacc"; \
		exit 1; \
	fi
	PACKER_ACC=1 go test -count 1 -v ./... -timeout=120m

##@ Quality Assurance

.PHONY: plugin-check
plugin-check: ## Run packer plugin compatibility check (required by goreleaser)
	@echo "Running plugin compatibility check..."
	@if ! command -v packer &> /dev/null; then \
		echo "Error: packer command not found. Please install Packer first."; \
		exit 1; \
	fi
	@echo "Building plugin for compatibility check..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="-X $(MODULE_PATH)/version.Version=$(VERSION) -X $(MODULE_PATH)/version.VersionPrerelease=" \
		-o $(PLUGIN_NAME)_v$(VERSION)_x5.0_linux_amd64 .
	@echo "Plugin compatibility check completed!"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint (if available)
	@if command -v golangci-lint &> /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping lint check"; \
	fi

##@ Dependency Management

.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

##@ Generation

.PHONY: generate
generate: ## Run go generate to update generated files
	@echo "Running go generate..."
	go generate ./...

##@ Release

.PHONY: release-snapshot
release-snapshot: ## Create a snapshot release (no upload)
	@if ! command -v goreleaser &> /dev/null; then \
		echo "Error: goreleaser not found. Please install goreleaser first."; \
		exit 1; \
	fi
	goreleaser release --snapshot --rm-dist

.PHONY: ci-release-docs
ci-release-docs: ## Generate documentation for release (used in CI)
	@echo "Generating release documentation..."
	@if [ -d ".web-docs" ]; then \
		cd .web-docs && zip -r ../docs.zip .; \
	else \
		echo "No .web-docs directory found"; \
		exit 1; \
	fi

##@ Information

.PHONY: version
version: ## Show version information
	@echo "Plugin: $(PLUGIN_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Prerelease: $(VERSION_PRERELEASE)"
	@echo "Module: $(MODULE_PATH)"
	@echo "Binary: $(BINARY)"

.PHONY: info
info: version ## Show build environment information
	@echo ""
	@echo "Build Environment:"
	@echo "GOOS: $(GOOS)"
	@echo "GOARCH: $(GOARCH)"
	@echo "Go Version: $(shell go version)"
	@echo "Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Git Branch: $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"

# Default target
.DEFAULT_GOAL := help