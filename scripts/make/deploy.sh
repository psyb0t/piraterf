#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "📦 Deploying PIrateRF files to the bastard Pi"

info "📦 Copying files to Pi..."
# Copy build tar file and deployment script to Pi
$SCP_CMD build/piraterf.tar.gz scripts/deploy.sh $PI_TARGET:/tmp/

info "📦 Running deployment script on the bastard..."
# Execute the deployment script on the Pi (handles cleanup and restart)
if $SSH_CMD -t $PI_TARGET "cd /tmp && bash deploy.sh"; then
    success "✅ Deployment fucking complete!"
else
    error "💥 Deployment failed!"
    exit 1
fi
