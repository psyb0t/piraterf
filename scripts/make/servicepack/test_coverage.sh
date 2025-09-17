#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

MIN_TEST_COVERAGE=90

section "Running Tests with Coverage Check"
info "Running tests with coverage analysis..."

# Ensure cleanup on exit
trap 'rm -f coverage.txt' EXIT

# Run tests with coverage
# Run tests with coverage - need to use array for proper word splitting
readarray -t packages < <(go list ./... | grep -v /cmd | grep -v '/internal/pkg/services$' | grep -v /internal/pkg/services/hello-world)
if ! go test -race -coverprofile=coverage.txt "${packages[@]}"; then
    error "Tests failed"
    exit 1
fi

# Calculate coverage
result=$(go tool cover -func=coverage.txt | grep -oP 'total:\s+\(statements\)\s+\K\d+' || echo "0")

if [ "$result" -eq 0 ]; then
    warning "No test coverage information available"
    exit 0
elif [ "$result" -lt "$MIN_TEST_COVERAGE" ]; then
    error "Coverage $result% is less than the minimum $MIN_TEST_COVERAGE%"
    exit 1
else
    success "Coverage $result% meets the minimum requirement of $MIN_TEST_COVERAGE%"
fi