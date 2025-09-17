#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Check if post-update scripts exist
if [ ! -f "scripts/.post-update-temp/review.sh" ]; then
    error "No pending servicepack update found."
    warning "Run 'make servicepack-update' first to create an update."
    exit 1
fi

# Execute the review script
bash scripts/.post-update-temp/review.sh