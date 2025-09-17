#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

section "Getting Dependencies"
info "Running go mod tidy..."
go mod tidy

info "Running go mod vendor..."
go mod vendor

success "Dependencies updated successfully!"