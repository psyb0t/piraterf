#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

# PIrateRF deployment script - copy and extract files on Pi
TAR_FILE="piraterf.tar.gz"
EXECUTABLE="piraterf.sh"

# No sudo needed - just deploying files to user directory

section "📦 Deploying PIrateRF files to this fucking Pi"

info "🚀 Starting PIrateRF deployment..."

# Check if directory exists and ask about overwriting
if [ -d "$DEPLOY_DIR" ]; then
    info "⚠️  Directory $DEPLOY_DIR already exists."
    read -r -p "Wanna fuckin' overwrite? [y/N]: " choice
    case $choice in
        y|Y)
            info "💥 Nuking the existing shit..."
            sudo rm -rf "$DEPLOY_DIR"
            mkdir -p "$DEPLOY_DIR"
            cd "$DEPLOY_DIR" || exit
            ;;
        *)
            info "🔄 Just overwriting the files..."
            cd "$DEPLOY_DIR" || exit
            ;;
    esac
else
    info "📁 Creating deployment directory..."
    mkdir -p "$DEPLOY_DIR"
    cd "$DEPLOY_DIR" || exit
fi

# Extract the tar file from /tmp
info "📦 Extracting deployment package..."
tar -xzf "/tmp/$TAR_FILE"
rm "/tmp/$TAR_FILE"

# Make executables
chmod +x "$EXECUTABLE"
chmod +x install.sh
chmod +x uninstall.sh

success "✅ PIrateRF deployment fucking complete!"
info "📁 Files deployed to: $DEPLOY_DIR"
info "🚀 Ready to install with: make install"
