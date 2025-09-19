#!/bin/bash

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🏴‍☠️ Complete setup of the fkin' Pi"

info "🚀 Running complete setup sequence..."

# Run all setup steps in order - stop on any failure
info "📦 Setting up dependencies..."
if ! make pi-setup-deps; then
    error "❌ Dependencies setup failed, aborting..."
    exit 1
fi

info "📡 Setting up WiFi access point..."
if ! make pi-setup-ap; then
    error "❌ AP setup failed, aborting..."
    exit 1
fi

info "🏴‍☠️ Setting up system branding..."
if ! make pi-setup-branding; then
    error "❌ Branding setup failed, aborting..."
    exit 1
fi

info "🚀 Deploying PIrateRF files..."
if ! make deploy; then
    error "❌ Deployment failed, aborting..."
    exit 1
fi

info "⚙️ Installing PIrateRF service..."
if ! make install; then
    error "❌ Service installation failed, aborting..."
    exit 1
fi

info "🔄 Rebooting Pi..."
if ! make pi-reboot; then
    error "❌ Reboot failed, aborting..."
    exit 1
fi

success "🏴‍☠️ PIrateRF setup sequence fucking complete!"
