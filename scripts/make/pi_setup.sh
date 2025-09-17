#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🏴‍☠️ Setup the bastard Pi for the fkin' mission"

info "🚀 Running complete Pi setup sequence..."

# Run setup components in order
make pi-setup-deps
make pi-setup-ap
make pi-setup-branding
make pi-reboot

success "🏴‍☠️ PIrateRF Pi Setup Fucking Complete!"
info "⏳ Wait for the fucking Pi to reboot..."