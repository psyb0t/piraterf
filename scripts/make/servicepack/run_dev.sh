#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

section "Running in Development Container"

# Build dev image first
info "Building development Docker image..."
make docker-build-dev

info "Starting containerized development environment..."
docker run -i --rm \
    --name "$APP_NAME-dev" \
    "$APP_NAME-dev" sh -c "CGO_ENABLED=1 go build -race -o ./build/$APP_NAME ./cmd/... && ./build/$APP_NAME run"

success "Development container finished"