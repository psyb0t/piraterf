#!/bin/bash

# Pi configuration
export PI_USER="${PI_USER:-fucker}"
export PI_HOST="${PI_HOST:-piraterf.local}"
export PI_PASS="${PI_PASS:-FUCKER}"
export PI_TARGET="$PI_USER@$PI_HOST"
export DEPLOY_DIR="/home/$PI_USER/piraterf"
export SSH_CMD="sshpass -p $PI_PASS ssh"
export SCP_CMD="sshpass -p $PI_PASS scp"