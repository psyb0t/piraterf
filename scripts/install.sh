#!/bin/bash

# PIrateRF installation script - runs on the Pi after deployment

# Source Pi configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/pi_config.sh"

# Check if running as root, if not, re-run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "🔐 Need root privileges, re-running with sudo..."
    exec sudo "$0" "$@"
fi

# Check if already installed (now as root)
if [ -f /etc/systemd/system/piraterf.service ]; then
    echo "❌ PIrateRF service already fucking installed!"
    echo "🗑️  Run 'make uninstall' first to remove the existing installation"
    exit 1
fi

echo "🚀 Starting PIrateRF installation..."

# Deployment already handled by make install dependency

# Install systemd service
echo "⚙️  Installing systemd service..."
cat > /etc/systemd/system/piraterf.service << EOF
[Unit]
Description=PIrateRF
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/${PI_USER}/piraterf
ExecStart=/home/${PI_USER}/piraterf/piraterf.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload

# Enable and start service
echo "🔄 Enabling PIrateRF service..."
systemctl enable piraterf

echo "🚀 Starting PIrateRF service..."
systemctl start piraterf

echo "✅ PIrateRF deployment fucking complete!"
echo "📊 Service status:"
systemctl status piraterf --no-pager
