#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

section "Cleaning Build Artifacts"

info "Removing build directory..."
rm -rf build/


info "Removing coverage files..."
rm -f coverage.txt

info "Cleaning Go module cache..."
go clean -modcache

success "Clean completed successfully!"