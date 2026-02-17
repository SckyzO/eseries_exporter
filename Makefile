# Variables
BINARY_NAME := eseries_exporter
IMAGE_NAME := sckyzo/eseries_exporter
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_USER := $(shell whoami 2>/dev/null || echo "unknown")

# Docker build arguments
DOCKER_BUILD_ARGS := \
	--build-arg VERSION=$(VERSION) \
	--build-arg COMMIT=$(COMMIT) \
	--build-arg BUILD_DATE=$(BUILD_DATE) \
	--build-arg BUILD_USER=$(BUILD_USER)

# Docker platforms for multi-arch builds
PLATFORMS := linux/amd64,linux/arm64

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Run linter via Docker
	@echo "Running golangci-lint..."
	@docker build --target linter -t $(IMAGE_NAME):lint .

.PHONY: test
test: ## Run tests via Docker
	@echo "Running tests..."
	@docker build --target tester -t $(IMAGE_NAME):test .
	@docker run --rm $(IMAGE_NAME):test cat /app/coverage.out > coverage.out
	@echo "Coverage report saved to coverage.out"

.PHONY: build
build: ## Build Docker image (full pipeline: lint + test + build)
	@echo "Building $(IMAGE_NAME):$(VERSION)..."
	@docker build $(DOCKER_BUILD_ARGS) -t $(IMAGE_NAME):$(VERSION) -t $(IMAGE_NAME):latest .

.PHONY: build-fast
build-fast: ## Build Docker image without linting/testing (for quick iterations)
	@echo "Fast building $(IMAGE_NAME):$(VERSION)..."
	@docker build $(DOCKER_BUILD_ARGS) --target builder -t $(IMAGE_NAME):builder .

.PHONY: build-all
build-all: ## Build multi-architecture images (requires buildx)
	@echo "Building multi-arch images..."
	@docker buildx build $(DOCKER_BUILD_ARGS) \
		--platform $(PLATFORMS) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest \
		--push .

.PHONY: extract-binary
extract-binary: ## Extract binary from Docker image to dist/
	@echo "Extracting binary from Docker image..."
	@mkdir -p dist
	@docker create --name tmp-exporter $(IMAGE_NAME):$(VERSION)
	@docker cp tmp-exporter:/eseries_exporter dist/$(BINARY_NAME)
	@docker rm tmp-exporter
	@echo "Binary extracted to dist/$(BINARY_NAME)"

.PHONY: run
run: ## Run exporter locally via Docker
	@echo "Running $(IMAGE_NAME):$(VERSION)..."
	@docker run --rm -p 9313:9313 \
		-v $(PWD)/eseries_exporter.yaml:/eseries_exporter.yaml:ro \
		$(IMAGE_NAME):$(VERSION) \
		--config.file=/eseries_exporter.yaml \
		--log.level=debug

.PHONY: shell
shell: ## Open shell in development container
	@docker run --rm -it \
		-v $(PWD):/app \
		-w /app \
		golang:1.24-alpine sh

.PHONY: clean
clean: ## Clean build artifacts and Docker images
	@echo "Cleaning..."
	@rm -rf dist coverage.out
	@docker rmi $(IMAGE_NAME):$(VERSION) $(IMAGE_NAME):latest $(IMAGE_NAME):lint $(IMAGE_NAME):test 2>/dev/null || true

.PHONY: push
push: ## Push Docker image to registry
	@echo "Pushing $(IMAGE_NAME):$(VERSION)..."
	@docker push $(IMAGE_NAME):$(VERSION)
	@docker push $(IMAGE_NAME):latest

.PHONY: scan
scan: ## Scan Docker image for vulnerabilities
	@echo "Scanning $(IMAGE_NAME):$(VERSION) for vulnerabilities..."
	@docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		aquasec/trivy image $(IMAGE_NAME):$(VERSION)

.PHONY: fmt
fmt: ## Format Go code via Docker
	@echo "Formatting code..."
	@docker run --rm \
		-v $(PWD):/app \
		-w /app \
		golang:1.24-alpine \
		sh -c "go fmt ./..."

.PHONY: deps
deps: ## Update Go dependencies
	@echo "Updating dependencies..."
	@docker run --rm \
		-v $(PWD):/app \
		-w /app \
		golang:1.24-alpine \
		sh -c "go get -u ./... && go mod tidy"

.PHONY: version
version: ## Show version information
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Build User: $(BUILD_USER)"

.PHONY: ci
ci: lint test build ## Run full CI pipeline (lint + test + build)
	@echo "CI pipeline completed successfully!"
