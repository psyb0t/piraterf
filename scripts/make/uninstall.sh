#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🗑️ Remove the fucking PIrateRF shit from the Pi completely"

# Check if deployment exists
info "🔍 Checking if deployment exists on this fucking Pi..."
if ! $SSH_CMD "$PI_TARGET" "[ -d $DEPLOY_DIR ]"; then
    error "❌ Error: No fucking deployment found at $DEPLOY_DIR"
    info "   🤷 Nothing to fucking uninstall"
    exit 1
fi

if ! $SSH_CMD "$PI_TARGET" "[ -f $DEPLOY_DIR/uninstall.sh ]"; then
    error "❌ Error: uninstall.sh not found in deployment"
    info "   📦 Run 'make deploy' to get the fucking uninstall script"
    exit 1
fi

# Execute uninstall
info "🗑️ Executing PIrateRF uninstall on this bastard Pi..."
$SSH_CMD "$PI_TARGET" "cd $DEPLOY_DIR && sudo bash uninstall.sh"

success "✅ PIrateRF fucking uninstalled from Pi!"