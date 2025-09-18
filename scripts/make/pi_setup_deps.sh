#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "📡 Setting up dependencies on the fucking Pi"

info "📡 Copying dependencies setup script to the fucking Pi..."
# Copy setup script and config to Pi
$SCP_CMD scripts/setup_deps.sh scripts/pi_config.sh $PI_TARGET:/tmp/

info "🔧 Executing dependencies setup on this bastard Pi..."
# Execute the setup script on the Pi and clean up
if $SSH_CMD $PI_TARGET "bash /tmp/setup_deps.sh && rm -f /tmp/setup_deps.sh /tmp/pi_config.sh"; then
    success "✅ Dependencies setup fucking complete!"
    exit 0
fi

error "💥 Dependencies setup failed!"
# Clean up even on failure
$SSH_CMD $PI_TARGET "rm -f /tmp/setup_deps.sh /tmp/pi_config.sh" || true
exit 1
