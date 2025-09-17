#!/bin/bash

# Common configuration for PIrateRF make scripts

# Source Pi configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../pi_config.sh"

# Build Pi connection details
export PI_TARGET="$PI_USER@$PI_HOST"
export DEPLOY_DIR="/home/$PI_USER/piraterf"
export SSH_CMD="sshpass -p $PI_PASS ssh"
export SCP_CMD="sshpass -p $PI_PASS scp"