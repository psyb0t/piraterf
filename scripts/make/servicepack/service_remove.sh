#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Check if service name is provided
if [ -z "$1" ]; then
    error "Service name is required"
    echo "Usage: $0 <service-name>"
    echo "Example: $0 myservice"
    exit 1
fi

SERVICE_NAME="$1"
SERVICE_DIR="internal/pkg/services/$SERVICE_NAME"

# Check if service exists
if [ ! -d "$SERVICE_DIR" ]; then
    error "Service '$SERVICE_NAME' does not exist at $SERVICE_DIR"
    exit 1
fi

section "Removing Service"
warning "Removing service '$SERVICE_NAME'..."
rm -rf "$SERVICE_DIR"

success "Service '$SERVICE_NAME' removed from $SERVICE_DIR"

# Regenerate service registration
info "Regenerating service registration..."
make service-registration