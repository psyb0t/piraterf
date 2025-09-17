#!/bin/bash
# PIrateRF uninstallation script - removes PIrateRF from the Pi

# Source Pi configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/pi_config.sh"

# Check if running as root, if not, re-run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "🔐 Need root privileges, re-running with sudo..."
    exec sudo "$0" "$@"
fi

echo "💥 Starting PIrateRF uninstallation..."

# Stop and disable the service
echo "⏹️ Stopping PIrateRF service..."
systemctl stop piraterf 2>/dev/null || true

echo "🚫 Disabling PIrateRF service..."
systemctl disable piraterf 2>/dev/null || true

# Remove service file
echo "🗑️ Removing service file..."
rm -f /etc/systemd/system/piraterf.service

# Reload systemd
echo "🔄 Reloading systemd daemon..."
systemctl daemon-reload

# Remove PIrateRF directory
echo "📂 Removing PIrateRF installation directory..."
rm -rf "/home/$PI_USER/piraterf"

# Remove any PIrateRF processes
echo "💀 Killing any remaining PIrateRF processes..."
pkill -f piraterf || true

echo "✅ PIrateRF uninstallation fucking complete!"
echo "🧹 System cleaned of all PIrateRF traces"