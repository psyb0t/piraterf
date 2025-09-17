#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

section "Linting Go Files"
info "Running modernize analysis..."
go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...

info "Running golangci-lint..."
go tool golangci-lint run --timeout=30m0s ./...

success "Linting completed successfully!"