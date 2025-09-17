#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

section "Building Development Docker Image"

# Check for user Dockerfile.dev first, then fall back to servicepack version
if [ -f "Dockerfile.dev" ]; then
    DOCKERFILE="Dockerfile.dev"
    info "Using user Dockerfile.dev..."
else
    DOCKERFILE="Dockerfile.servicepack.dev"
    info "Using framework Dockerfile.servicepack.dev..."
fi

info "Building development Docker image: $APP_NAME-dev..."
docker build -f "$DOCKERFILE" -t "$APP_NAME-dev" .

success "Development Docker image built successfully: $APP_NAME-dev"