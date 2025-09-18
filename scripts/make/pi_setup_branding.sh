#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🏴‍☠️ Setting up branding on the bastard PI"

info "🏴‍☠️ Copying branding setup script to the fucking Pi..."
# Copy branding setup script and config to Pi
$SCP_CMD scripts/setup_branding.sh scripts/pi_config.sh $PI_TARGET:/tmp/

info "🎨 Executing branding setup on this motherfucker..."
# Execute the branding setup script on the Pi with sudo and clean up
if $SSH_CMD $PI_TARGET "sudo bash /tmp/setup_branding.sh && rm -f /tmp/setup_branding.sh /tmp/pi_config.sh"; then
    success "✅ Branding setup fucking complete!"
    exit 0
fi

error "💥 Branding setup failed!"
# Clean up even on failure
$SSH_CMD $PI_TARGET "rm -f /tmp/setup_branding.sh /tmp/pi_config.sh" || true
exit 1