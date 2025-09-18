#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🏴‍☠️ Setting up wifi AP on the bastard PI"

info "🏴‍☠️ Copying AP setup script to the fucking Pi..."
# Copy AP setup script and config to Pi
$SCP_CMD scripts/setup_ap.sh scripts/pi_config.sh $PI_TARGET:/tmp/

info "📡 Executing AP setup on this motherfucker..."
# Execute the AP setup script on the Pi with sudo and clean up
if $SSH_CMD $PI_TARGET "sudo bash /tmp/setup_ap.sh && rm -f /tmp/setup_ap.sh /tmp/pi_config.sh"; then
    success "✅ AP setup fucking complete!"
    exit 0
fi

error "💥 AP setup failed!"
# Clean up even on failure
$SSH_CMD $PI_TARGET "rm -f /tmp/setup_ap.sh /tmp/pi_config.sh" || true
exit 1