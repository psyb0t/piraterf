#!/bin/bash

# PIrateRF installation script - runs on the Pi after deployment

# Check if running as root, if not, re-run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "🔐 Need root privileges, re-running with sudo..."
    exec sudo "$0" "$@"
fi

echo "🚀 Starting PIrateRF installation..."

# Deployment already handled by make install dependency

# Stop existing service if running
echo "⏹️  Stopping existing PIrateRF service..."
systemctl stop piraterf 2>/dev/null || true

# Install systemd service
echo "⚙️  Installing systemd service..."
cat > /etc/systemd/system/piraterf.service << 'EOF'
[Unit]
Description=PIrateRF
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/home/fucker/piraterf
ExecStart=/home/fucker/piraterf/piraterf.sh
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
