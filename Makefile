# Modern Makefile for eseries_exporter
# Supports modern Go development workflow with multiple targets

# Variables
BINARY_NAME := eseries_exporter
MAIN_PATH := ./cmd/eseries_exporter
BUILD_DIR := ./dist
COVERAGE_FILE := coverage.txt
GOLANGCI_LINT_VERSION := v1.24.0

# Docker settings (updated for new repo)
DOCKER_ARCHS ?= amd64 armv7 arm64 ppc64le s390x
DOCKER_REPO ?= sckyzo
DOCKER_IMAGE_NAME ?= $(BINARY_NAME)

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS := -s -w \
           -X github.com/prometheus/common/version.Version=$(VERSION) \
           -X github.com/prometheus/common/version.Revision=$(COMMIT) \
           -X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") \
           -X github.com/prometheus/common/version.BuildUser=$(shell whoami 2>/dev/null || echo "unknown") \
           -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE)

# Go build tags for different targets
TAGS ?=

# Include common targets (if Makefile.common exists)
-include Makefile.common

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
.PHONY: build
build: ## Build the binary for current platform
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) for $(GOOS)/$(GOARCH)..."
	go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-linux
build-linux: ## Build the binary for Linux
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) for Linux..."
	GOOS=linux GOARCH=amd64 go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(MAIN_PATH)

.PHONY: build-windows
build-windows: ## Build the binary for Windows
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) for Windows..."
	GOOS=windows GOARCH=amd64 go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(MAIN_PATH)

.PHONY: build-darwin
build-darwin: ## Build the binary for macOS
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) for macOS..."
	GOOS=darwin GOARCH=amd64 go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(TAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(MAIN_PATH)

.PHONY: build-all
build-all: build-linux build-windows build-darwin ## Build binaries for all platforms
	@echo "All binaries built successfully!"
	@ls -la $(BUILD_DIR)/

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts and temporary files
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@rm -f *.out
	@echo "Cleaned build artifacts"

.PHONY: clean-all
clean-all: clean ## Clean everything including Docker images
	@docker rmi $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):latest 2>/dev/null || true
	@echo "Cleaned everything"

# Testing targets
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: test-short
test-short: ## Run short tests only
	@echo "Running short tests..."
	go test -short ./...

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage and generate report
	@echo "Running tests with coverage..."
	go test -race -coverpkg=./... -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@echo "Coverage report:"
	@go tool cover -func=$(COVERAGE_FILE)

.PHONY: test-html
test-html: test-coverage ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting and formatting targets
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: golangci-lint
golangci-lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi
	golangci-lint run

.PHONY: lint
lint: fmt vet golangci-lint ## Run all linting checks

.PHONY: mod-tidy
mod-tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	go mod tidy

.PHONY: mod-verify
mod-verify: ## Verify Go modules
	@echo "Verifying Go modules..."
	go mod verify

# Security targets
.PHONY: security-scan
security-scan: ## Run security scanning tools
	@echo "Running security scans..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/gosec@latest"; \
	fi

.PHONY: dependabot-check
dependabot-check: ## Check for outdated dependencies
	@echo "Checking for outdated dependencies..."
	go list -u -m all

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):latest .

.PHONY: docker-build-multi
docker-build-multi: ## Build multi-architecture Docker images
	@echo "Building multi-architecture Docker images..."
	docker buildx build --platform $(DOCKER_ARCHS) -t $(DOCKER_REPO)/$(DOCKER_IMAGE_NAME):latest --push .

.PHONY: docker-run
docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run -d --name $(BINARY_NAME) -p 9313:9313 $(DOCKER_REPO)/$(BINARY_NAME):latest

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(BINARY_NAME) 2>/dev/null || true
	docker rm $(BINARY_NAME) 2>/dev/null || true

.PHONY: docker-clean
docker-clean: ## Clean Docker resources
	@echo "Cleaning Docker resources..."
	docker system prune -f
	docker volume prune -f

# Development targets
.PHONY: run
run: ## Run the application in development mode
	@echo "Running $(BINARY_NAME) in development mode..."
	go run $(MAIN_PATH)

.PHONY: run-with-config
run-with-config: ## Run with custom config file
	@if [ -z "$(CONFIG)" ]; then \
		echo "Usage: make run-with-config CONFIG=/path/to/config.yaml"; \
		exit 1; \
	fi
	go run $(MAIN_PATH) --config.file=$(CONFIG)

.PHONY: watch
watch: ## Run with file watcher (requires entr)
	@echo "Running with file watcher..."
	@if ! command -v entr >/dev/null 2>&1; then \
		echo "entr not found. Install with: apt-get install entr or brew install entr"; \
		exit 1; \
	fi
	find . -name "*.go" -not -path "./vendor/*" | entr -r make run

# Release targets
.PHONY: release-check
release-check: ## Check release readiness
	@echo "Checking release readiness..."
	@$(MAKE) mod-verify
	@$(MAKE) lint
	@$(MAKE) test-coverage
	@$(MAKE) build-all
	@echo "Release check completed!"

.PHONY: changelog
changelog: ## Generate changelog (requires git-cliff)
	@echo "Generating changelog..."
	@if command -v git-cliff >/dev/null 2>&1; then \
		git-cliff --output CHANGELOG.md; \
	else \
		echo "git-cliff not found. Install with: cargo install git-cliff"; \
	fi

.PHONY: tag
tag: ## Create and push git tag for current version
	@echo "Creating git tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "Tag created. Push with: git push origin $(VERSION)"

# CI/CD targets (for local testing)
.PHONY: ci-local
ci-local: ## Run CI checks locally
	@echo "Running CI checks locally..."
	@$(MAKE) clean
	@$(MAKE) mod-tidy
	@$(MAKE) mod-verify
	@$(MAKE) lint
	@$(MAKE) test-race
	@$(MAKE) test-coverage
	@$(MAKE) build-all
	@$(MAKE) docker-build
	@echo "CI checks completed successfully!"

# Install targets
.PHONY: install
install: ## Install binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install $(TAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install github.com/securecodewarrior/gosec/v2/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Documentation targets
.PHONY: docs
docs: ## Generate documentation (if swag is configured)
	@if [ -f "docs.go" ] || [ -f "main.go" ]; then \
		echo "Generating API documentation..."; \
		swag init --dir $(MAIN_PATH) --generalInfo main.go --output docs; \
	else \
		echo "No swagger configuration found. Skipping docs generation."; \
	fi

# Health and debugging targets
.PHONY: version
version: ## Show version information
	@echo "$(BINARY_NAME) version $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build date: $(BUILD_DATE)"

.PHONY: env
env: ## Show Go environment
	@echo "Go version: $(shell go version)"
	@echo "Go environment:"
	@go env

.PHONY: deps-check
deps-check: ## Check for dependency issues
	@echo "Checking for dependency issues..."
	go mod graph | grep -v "github.com/sckyzo/eseries_exporter" | head -20

# Development setup
.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@$(MAKE) install-tools
	@$(MAKE) mod-tidy
	@echo "Development environment setup complete!"

# Quick development workflow
.PHONY: dev
dev: lint test run ## Quick development: lint, test, and run

# Quality gates
.PHONY: quality-gate
quality-gate: lint test-coverage ## Run quality gates before committing
	@echo "Running quality gates..."
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}'
	@echo "Quality gates passed!"

# Help targets for specific workflows
.PHONY: help-dev
help-dev: ## Development workflow help
	@echo "Development workflow:"
	@echo "  make setup           - Setup development environment"
	@echo "  make dev             - Quick development (lint, test, run)"
	@echo "  make ci-local        - Run CI checks locally"
	@echo "  make quality-gate    - Run quality gates"

.PHONY: help-build
help-build: ## Build workflow help
	@echo "Build workflow:"
	@echo "  make build           - Build for current platform"
	@echo "  make build-all       - Build for all platforms"
	@echo "  make docker-build    - Build Docker image"
	@echo "  make clean           - Clean build artifacts"

.PHONY: help-test
help-test: ## Testing workflow help
	@echo "Testing workflow:"
	@echo "  make test            - Run all tests"
	@echo "  make test-coverage   - Run tests with coverage"
	@echo "  make test-race       - Run tests with race detector"

# The all target does everything needed for a clean release
.PHONY: all
all: clean lint test-coverage build-all docker-build ## Complete release preparation
	@echo "Release preparation complete!"

# Targets that do not create files
.PHONY: .FORCE
.FORCE:

# Include version info in help
version-info:
	@echo "$(BINARY_NAME) $(VERSION) - Commit: $(COMMIT) - Date: $(BUILD_DATE)"

# Display version in default target if no arguments
.DEFAULT_GOAL := help
