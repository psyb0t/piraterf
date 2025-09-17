#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🏴‍☠️ Complete setup of the fkin' Pi"

info "🚀 Running complete setup sequence..."

# Run all setup steps in order
make pi-setup-deps
make pi-setup-ap
make pi-setup-branding
make deploy
make install
make pi-setup-branding
make pi-reboot

success "🏴‍☠️ PIrateRF is ready to do some fucking fuckage!"