#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

section "Linting and Fixing Go Files"
info "Running modernize analysis with fixes..."
go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix -test ./...

info "Running golangci-lint with fixes..."
go tool golangci-lint run --fix --timeout=30m0s ./...

success "Linting and fixing completed successfully!"