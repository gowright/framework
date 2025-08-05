#!/bin/bash

# GoWright CI/CD Test Script
# This script replicates the test steps from the GitHub Actions CI/CD pipeline

set -e  # Exit on any error

echo "🚀 Starting GoWright test suite..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}📋 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
REQUIRED_VERSION="1.22"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    print_warning "Go version $GO_VERSION detected. Recommended version is $REQUIRED_VERSION or later."
fi

print_step "Installing dependencies..."
go mod download
go mod verify
print_success "Dependencies installed and verified"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    print_warning "golangci-lint not found. Installing..."
    # Install golangci-lint
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

print_step "Running linter..."
if golangci-lint run --timeout=5m --max-issues-per-linter=10 --max-same-issues=3; then
    print_success "Linting passed"
else
    print_error "Linting failed"
    exit 1
fi

print_step "Checking code formatting..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    print_error "Code is not formatted. Run 'gofmt -w .' to fix:"
    echo "$UNFORMATTED"
    exit 1
else
    print_success "Code formatting is correct"
fi

print_step "Running unit tests with race detection and coverage..."
if go test -v -race -coverprofile=coverage.out ./...; then
    print_success "Unit tests passed"
    
    # Display coverage summary
    if command -v go &> /dev/null; then
        COVERAGE=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
        echo -e "${BLUE}📊 Total coverage: $COVERAGE${NC}"
    fi
else
    print_error "Unit tests failed"
    exit 1
fi

print_step "Running integration tests..."
# Check if databases are available (optional for local testing)
DB_AVAILABLE=true

# Check PostgreSQL
if ! nc -z localhost 5432 2>/dev/null; then
    print_warning "PostgreSQL not available on localhost:5432"
    DB_AVAILABLE=false
fi

# Check MySQL
if ! nc -z localhost 3306 2>/dev/null; then
    print_warning "MySQL not available on localhost:3306"
    DB_AVAILABLE=false
fi

if [ "$DB_AVAILABLE" = true ]; then
    export POSTGRES_URL="postgres://postgres:postgres@localhost:5432/testdb?sslmode=disable"
    export MYSQL_URL="root:root@tcp(localhost:3306)/testdb"
    print_step "Running integration tests with database connections..."
else
    print_warning "Running integration tests without database connections..."
fi

if go run integration_test_runner.go; then
    print_success "Integration tests passed"
else
    print_error "Integration tests failed"
    exit 1
fi

print_step "Running performance benchmarks..."
if go test -bench=. -benchmem ./pkg/... > benchmark_results.txt; then
    print_success "Performance benchmarks completed"
    echo -e "${BLUE}📈 Benchmark results saved to benchmark_results.txt${NC}"
    
    # Show a summary of benchmark results
    if [ -f benchmark_results.txt ]; then
        echo -e "${BLUE}📊 Benchmark Summary:${NC}"
        grep -E "^Benchmark" benchmark_results.txt | head -5 || echo "No benchmark results found"
    fi
else
    print_warning "Performance benchmarks completed with warnings"
fi

print_success "🎉 All tests completed successfully!"

# Summary
echo ""
echo -e "${BLUE}📋 Test Summary:${NC}"
echo "✅ Dependencies installed and verified"
echo "✅ Code linting passed"
echo "✅ Code formatting verified"
echo "✅ Unit tests passed with race detection"
echo "✅ Integration tests completed"
echo "✅ Performance benchmarks completed"

if [ -f coverage.out ]; then
    echo -e "${BLUE}📁 Generated files:${NC}"
    echo "  - coverage.out (test coverage report)"
    echo "  - benchmark_results.txt (performance benchmarks)"
fi

echo ""
echo -e "${GREEN}🚀 Ready for deployment!${NC}"