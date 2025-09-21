#!/bin/bash

# Pi Zero W Dependencies Setup - RF Transmission Library + Audio Processing
# Installs rpitx for RF signal generation and sox for audio conversion

set -e

# Source configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/pi_config.sh"

# Configuration
RPITX_REPO="https://github.com/psyb0t/rpitx.git"
INSTALL_DIR="/home/$PI_USER/rpitx"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}📡 Pi Zero W Dependencies Installation - Let's fucking transmit! 📡${NC}"
echo "Installing rpitx for RF signal generation and sox for audio conversion..."

# Check if directory exists first before doing any system updates
if [ -d "$INSTALL_DIR" ]; then
    echo "⚠️  Directory $INSTALL_DIR already exists."
    read -p "Wanna fuckin' reinstall? [y/N]: " choice
    case $choice in
        y|Y)
            echo "🗑️ Removing existing rpitx directory..."
            rm -rf "$INSTALL_DIR"
            ;;
        *)
            echo "❌ Installation cancelled."
            exit 0
            ;;
    esac
fi

echo "🔄 Updating this fucking system..."
sudo apt-get update

echo "📦 Installing the fucking dependencies..."
sudo apt-get install -y \
    git \
    sox \
    libsox-fmt-all \
    ffmpeg \
    openssl \
    minimodem \
    pulseaudio

echo "📥 Cloning the fucking rpitx repository..."

git clone "$RPITX_REPO" "$INSTALL_DIR"
cd "$INSTALL_DIR"

echo "⚙️ Running the fucking rpitx installation..."
./install.sh

echo ""
echo -e "${GREEN}✅ Dependencies installation fucking complete!${NC}"
