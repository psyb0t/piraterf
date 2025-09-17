#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🔄 Rebooting this fucking Pi"

info "🔄 Rebooting Pi at $PI_HOST..."
$SSH_CMD "$PI_TARGET" "sudo reboot" || true

success "✅ Pi reboot fucking initiated!"