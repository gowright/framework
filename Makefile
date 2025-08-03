# Gowright Testing Framework Makefile

# Variables
BINARY_NAME=gowright
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X github/gowright/framework/pkg/gowright.Version=$(VERSION) -X github/gowright/framework/pkg/gowright.GitCommit=$(GIT_COMMIT) -X github/gowright/framework/pkg/gowright.BuildDate=$(BUILD_TIME)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Directories
PKG_DIR=./pkg/gowright
EXAMPLES_DIR=./examples
REPORTS_DIR=./reports
COVERAGE_DIR=./coverage

.PHONY: all build clean test test-coverage test-integration test-performance lint fmt deps help

# Default target
all: clean deps fmt lint test build

# Build the project
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(REPORTS_DIR)
	rm -rf $(COVERAGE_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOCMD) run integration_test_runner.go

# Run performance benchmarks
test-performance:
	@echo "Running performance benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

# Run all tests (unit, integration, performance)
test-all: test test-integration test-performance

# Lint the code
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Check formatting
fmt-check:
	@echo "Checking code formatting..."
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "Code is not formatted. Run 'make fmt' to fix."; \
		$(GOFMT) -l .; \
		exit 1; \
	fi

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Verify dependencies
deps-verify:
	@echo "Verifying dependencies..."
	$(GOMOD) verify

# Generate documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation server started at http://localhost:6060"

# Run examples
examples:
	@echo "Running examples..."
	@for example in $(EXAMPLES_DIR)/*.go; do \
		echo "Running $$example..."; \
		$(GOCMD) run $$example; \
	done

# Install development tools
install-tools:
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Check for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	$(GOCMD) list -json -m all | nancy sleuth

# Release preparation
release-prepare:
	@echo "Preparing release..."
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make release-prepare TAG=v1.0.0"; \
		exit 1; \
	fi
	@echo "Preparing release $(TAG)..."
	git tag -a $(TAG) -m "Release $(TAG)"
	@echo "Tag $(TAG) created. Push with: git push origin $(TAG)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# CI/CD pipeline simulation
ci: deps fmt-check lint test-coverage test-integration

# Development setup
dev-setup: install-tools deps
	@echo "Development environment setup complete!"

# Show help
help:
	@echo "Gowright Testing Framework - Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          Build the binary"
	@echo "  build-all      Build for multiple platforms"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  test           Run unit tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-integration Run integration tests"
	@echo "  test-performance Run performance benchmarks"
	@echo "  test-all       Run all tests"
	@echo ""
	@echo "Code quality targets:"
	@echo "  fmt            Format code"
	@echo "  fmt-check      Check code formatting"
	@echo "  lint           Run linter"
	@echo "  security       Run security scan"
	@echo "  vuln-check     Check for vulnerabilities"
	@echo ""
	@echo "Dependency targets:"
	@echo "  deps           Install dependencies"
	@echo "  deps-update    Update dependencies"
	@echo "  deps-verify    Verify dependencies"
	@echo ""
	@echo "Development targets:"
	@echo "  dev-setup      Setup development environment"
	@echo "  install-tools  Install development tools"
	@echo "  docs           Generate documentation"
	@echo "  examples       Run examples"
	@echo ""
	@echo "Release targets:"
	@echo "  release-prepare TAG=v1.0.0  Prepare release"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo ""
	@echo "CI/CD targets:"
	@echo "  ci             Run CI pipeline"
	@echo ""
	@echo "Usage: make <target>"