#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🔒 SSH into the fucking Pi"

info "🔗 Connecting to Pi at $PI_HOST..."
$SSH_CMD "$PI_TARGET"