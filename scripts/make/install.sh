#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🔧 Installing the fucking PIrateRF to the bastard Pi"

# Check if deployment exists
info "🔍 Checking if deployment exists on this fucking Pi..."
if ! $SSH_CMD "$PI_TARGET" "[ -d $DEPLOY_DIR ]"; then
    error "❌ Error: No fucking deployment found at $DEPLOY_DIR"
    info "   📦 Run 'make deploy' first to deploy the fucking files"
    exit 1
fi

if ! $SSH_CMD "$PI_TARGET" "[ -f $DEPLOY_DIR/install.sh ]"; then
    error "❌ Error: install.sh not found in deployment"
    info "   📦 Run 'make deploy' to refresh the fucking deployment"
    exit 1
fi

# Execute install
info "🔧 Running installation script on this bastard..."
$SSH_CMD "$PI_TARGET" "cd $DEPLOY_DIR && chmod +x install.sh && ./install.sh"

success "📡 Installation fucking complete! System fucking compromised and running!"