#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

section "Building Production Docker Image"

# Check for user Dockerfile first, then fall back to servicepack version
if [ -f "Dockerfile" ]; then
    DOCKERFILE="Dockerfile"
    info "Using user Dockerfile..."
else
    DOCKERFILE="Dockerfile.servicepack"
    info "Using framework Dockerfile.servicepack..."
fi

info "Building production Docker image: $APP_NAME..."
docker build -f "$DOCKERFILE" -t "$APP_NAME" .

success "Production Docker image built successfully: $APP_NAME"