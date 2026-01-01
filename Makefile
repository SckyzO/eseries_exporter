# Variables
BINARY_NAME := eseries_exporter
MAIN_PATH := ./cmd/eseries_exporter
BUILD_DIR := ./dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS := -s -w \
           -X github.com/prometheus/common/version.Version=$(VERSION) \
           -X github.com/prometheus/common/version.Revision=$(COMMIT) \
           -X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") \
           -X github.com/prometheus/common/version.BuildUser=$(shell whoami 2>/dev/null || echo "unknown") \
           -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE)

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%%-20s\033[0m %%s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: build-all
build-all: ## Build binaries for multiple platforms
	@echo "Building binaries for Linux, Windows, and Darwin..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_linux_arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_windows_amd64.exe $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)_darwin_arm64 $(MAIN_PATH)

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, using go vet"; \
		go vet ./...; \
	fi

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

.PHONY: docker
docker: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker tag $(BINARY_NAME):$(VERSION) $(BINARY_NAME):latest

.PHONY: run
run: ## Run the application locally
	go run $(MAIN_PATH) --log.level=debug