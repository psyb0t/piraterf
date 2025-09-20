#!/bin/bash

# PIrateRF deployment script - copy and extract files on Pi

# Source Pi configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/pi_config.sh"

# Configuration
DEPLOY_USER="$PI_USER"
DEPLOY_DIR="/home/${DEPLOY_USER}/piraterf"
TAR_FILE="piraterf.tar.gz"
EXECUTABLE="piraterf.sh"

# No sudo needed - just deploying files to user directory

echo "🚀 Starting PIrateRF deployment..."

# Check if directory exists and ask about overwriting
if [ -d "$DEPLOY_DIR" ]; then
    echo "⚠️  Directory $DEPLOY_DIR already exists."
    read -p "Wanna fuckin' overwrite? [y/N]: " choice
    case $choice in
        y|Y)
            echo "💥 Nuking the existing shit..."
            sudo rm -rf "$DEPLOY_DIR"
            mkdir -p "$DEPLOY_DIR"
            cd "$DEPLOY_DIR"
            ;;
        *)
            echo "🔄 Just overwriting the files..."
            cd "$DEPLOY_DIR"
            ;;
    esac
else
    echo "📁 Creating deployment directory..."
    mkdir -p "$DEPLOY_DIR"
    cd "$DEPLOY_DIR"
fi

# Extract the tar file from /tmp
echo "📦 Extracting deployment package..."
tar -xzf "/tmp/$TAR_FILE"
rm "/tmp/$TAR_FILE"

# Make executables
chmod +x "$EXECUTABLE"
chmod +x install.sh
chmod +x uninstall.sh

echo "🧹 Cleaning up temp files..."
rm -f /tmp/deploy.sh /tmp/pi_config.sh

echo "🔄 Checking for piraterf service..."
if sudo systemctl is-enabled piraterf >/dev/null 2>&1; then
    echo "🔄 Restarting piraterf service..."
    if sudo systemctl restart piraterf; then
        echo "✅ PIrateRF deployment and restart fucking complete!"
    else
        echo "⚠️  Service restart failed, but deployment completed"
        exit 1
    fi
else
    echo "ℹ️  PIrateRF service not installed yet, skipping restart"
    echo "✅ PIrateRF deployment fucking complete!"
fi
