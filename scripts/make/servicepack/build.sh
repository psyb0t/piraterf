#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

section "Building Application"
info "Building $APP_NAME binary using Docker..."

# Create build directory
mkdir -p ./build

# Build using Docker
docker run --rm \
    -v "$(pwd)":/app \
    -w /app \
    -e USER_UID="$(id -u)" \
    -e USER_GID="$(id -g)" \
    golang:1.24.6-alpine \
    sh -c "apk add --no-cache gcc musl-dev && \
        CGO_ENABLED=0 go build -a \
        -ldflags '-extldflags \"-static\" -X main.appName=$APP_NAME' \
        -o ./build/$APP_NAME ./cmd/... && \
        chown \$USER_UID:\$USER_GID ./build/$APP_NAME"

success "Binary built successfully: ./build/$APP_NAME"
