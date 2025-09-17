#!/bin/bash

# Pi Zero W Standalone Access Point Setup - DHCP Only Edition
# Creates AP with DHCP, preserves other interfaces (usb0/OTG)

set -e

# Check if running as root, if not, re-run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "🔐 Need root privileges, re-running with sudo..."
    exec sudo "$0" "$@"
fi

# Always non-interactive - no fucking questions
NON_INTERACTIVE=true

# Configuration
AP_SSID="🏴‍☠️📡"
AP_PASSWORD="FUCKER!!!"
AP_CHANNEL="7"
COUNTRY_CODE="US"

# Network Configuration - Standalone AP with DHCP
WIFI_INTERFACE="wlan0"
AP_IP="192.168.4.1"
AP_SUBNET="192.168.4.1/24"
DHCP_START="192.168.4.2"
DHCP_END="192.168.4.50"
DHCP_NETMASK="255.255.255.0"
DHCP_LEASE_TIME="12h"

# Local DNS for dhcpcd config (unused by dnsmasq - DNS is disabled)
LOCAL_DNS="192.168.4.1"

# System Configuration
BACKUP_USER="fucker"
TARGET_OS_VERSION="11"
REQUIRED_PACKAGES="hostapd dnsmasq"

# Paths
BACKUP_BASE_DIR="/home/${BACKUP_USER}"
DHCPCD_CONF="/etc/dhcpcd.conf"
DNSMASQ_CONF="/etc/dnsmasq.conf"
HOSTAPD_CONF_DIR="/etc/hostapd"
HOSTAPD_CONF="${HOSTAPD_CONF_DIR}/hostapd.conf"
SYSTEMD_SERVICE_DIR="/etc/systemd/system"
DNSMASQ_SERVICE_OVERRIDE_DIR="${SYSTEMD_SERVICE_DIR}/dnsmasq.service.d"
WPA_SUPPLICANT_CONF="/etc/wpa_supplicant/wpa_supplicant.conf"

# Service Names
HOSTAPD_SERVICE="hostapd"
DNSMASQ_SERVICE="dnsmasq"
NETWORKMANAGER_SERVICE="NetworkManager"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Function to check if dhcpcd already has AP configuration
check_dhcpcd_config() {
    if grep -q "interface $WIFI_INTERFACE" "$DHCPCD_CONF" &&
       grep -q "static ip_address=$AP_SUBNET" "$DHCPCD_CONF" &&
       grep -q "nohook wpa_supplicant" "$DHCPCD_CONF"; then
        return 0  # Config exists
    fi
    return 1  # Config doesn't exist
}

# Function to check if hostapd config already exists with our settings
check_hostapd_config() {
    if [ -f "$HOSTAPD_CONF" ] &&
       grep -q "interface=$WIFI_INTERFACE" "$HOSTAPD_CONF" &&
       grep -q "ssid=$AP_SSID" "$HOSTAPD_CONF" &&
       grep -q "channel=$AP_CHANNEL" "$HOSTAPD_CONF"; then
        return 0  # Config exists
    fi
    return 1  # Config doesn't exist
}

# Function to check if dnsmasq config already exists with our settings
check_dnsmasq_config() {
    if [ -f "$DNSMASQ_CONF" ] &&
       grep -q "interface=$WIFI_INTERFACE" "$DNSMASQ_CONF" &&
       grep -q "dhcp-range=$DHCP_START,$DHCP_END,$DHCP_NETMASK,$DHCP_LEASE_TIME" "$DNSMASQ_CONF"; then
        return 0  # Config exists
    fi
    return 1  # Config doesn't exist
}

# Function to check if backup already exists
check_backup_exists() {
    if ls "$BACKUP_BASE_DIR"/ap-backup-* 1> /dev/null 2>&1; then
        return 0  # Backup exists
    fi
    return 1  # No backup exists
}

echo -e "${GREEN}🏴‍☠️ Pi Zero W Standalone AP Setup! 🏴‍☠️${NC}"
if [ "$NON_INTERACTIVE" = true ]; then
    echo -e "${YELLOW}Running in non-interactive mode, no fucking questions asked!${NC}"
fi

# Root check (handled by auto-sudo above, but keeping as safety net)
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Something went wrong with sudo elevation${NC}"
    exit 1
fi

# Bullseye version check - NO MERCY
OS_VERSION=$(cat /etc/os-release | grep VERSION_ID | cut -d'"' -f2)
if [ "$OS_VERSION" != "$TARGET_OS_VERSION" ]; then
    echo -e "${RED}Error: This shit only works on Bullseye (Debian $TARGET_OS_VERSION). You're on $OS_VERSION${NC}"
    echo -e "${RED}Get the right fucking OS and try again!${NC}"
    exit 1
fi

# Check WiFi interface exists
if ! ip link show $WIFI_INTERFACE >/dev/null 2>&1; then
    echo -e "${RED}Error: $WIFI_INTERFACE not found. Your WiFi shit is fucked!${NC}"
    exit 1
fi

# Check if NetworkManager is interfering (fucking annoying)
if systemctl is-active --quiet $NETWORKMANAGER_SERVICE; then
    echo -e "${YELLOW}$NETWORKMANAGER_SERVICE detected. This shit might fuck with our AP setup.${NC}"
    echo "Use raspi-config → Advanced → Network Config → dhcpcd to fix this crap"
    echo "Fuck it, continuing anyway"
fi

echo "🔄 Updating this fucking system..."
apt-get update && apt-get upgrade -y

echo "📦 Installing the fucking packages..."
apt-get install -y $REQUIRED_PACKAGES

echo "⏹️ Stopping services for configuration..."
systemctl stop $HOSTAPD_SERVICE $DNSMASQ_SERVICE 2>/dev/null || true
# Don't mask wpa_supplicant globally - other interfaces might need it
# The 'nohook wpa_supplicant' in dhcpcd.conf handles wlan0 isolation

# Backup configs with timestamps (only if no backup exists)
if check_backup_exists; then
    EXISTING_BACKUP=$(ls -1t "$BACKUP_BASE_DIR"/ap-backup-* | head -n1)
    echo "💾 Backup already exists: $(basename "$EXISTING_BACKUP")"
    BACKUP_DIR="$EXISTING_BACKUP"
else
    BACKUP_DIR="$BACKUP_BASE_DIR/ap-backup-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    echo "💾 Creating backup of your shit configs to $BACKUP_DIR"

    [ -f $DHCPCD_CONF ] && cp $DHCPCD_CONF "$BACKUP_DIR/"
    [ -f $DNSMASQ_CONF ] && cp $DNSMASQ_CONF "$BACKUP_DIR/"
    [ -f $HOSTAPD_CONF ] && cp $HOSTAPD_CONF "$BACKUP_DIR/" 2>/dev/null || true
fi

# CRITICAL: Set country code first (fucking requirement)
echo "🌍 Setting WiFi country code..."
raspi-config nonint do_wifi_country "$COUNTRY_CODE"

# Unblock WiFi (common fucking issue)
echo "🔓 Unblocking WiFi radio..."
rfkill unblock wifi

# Configure dhcpcd with CRITICAL nohook directive - WLAN0 ONLY
echo "⚙️ Configuring dhcpcd (network shit)..."
if check_dhcpcd_config; then
    echo "⚙️ dhcpcd already configured for AP mode - skipping"
else
    cat >> $DHCPCD_CONF << EOF

# AP Configuration - Added $(date) - WLAN0 ONLY
interface $WIFI_INTERFACE
static ip_address=$AP_SUBNET
static routers=$AP_IP
static domain_name_servers=$LOCAL_DNS
# ONLY disable wpa_supplicant for wlan0 - other interfaces can use DHCP
nohook wpa_supplicant

# Ensure other interfaces (like usb0 for OTG) still get DHCP
interface usb0
# Let usb0 use DHCP from host
EOF
fi

# Configure hostapd - NO DRIVER LINE (Pi Zero W fucking quirk)
echo "🔧 Configuring hostapd..."
mkdir -p $HOSTAPD_CONF_DIR
if check_hostapd_config; then
    echo "🔧 hostapd already configured for AP mode - skipping"
else
    cat > $HOSTAPD_CONF << EOF
# Pi Zero W hostapd config - $(date)
# NO driver line - auto-detection works better, fuck manual config
interface=$WIFI_INTERFACE
ssid=$AP_SSID
country_code=$COUNTRY_CODE
hw_mode=g
channel=$AP_CHANNEL
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0
wpa=2
wpa_passphrase=$AP_PASSWORD
wpa_key_mgmt=WPA-PSK
rsn_pairwise=CCMP
EOF
fi

# Configure dnsmasq as DHCP server only (no DNS)
echo "🌐 Configuring dnsmasq (DHCP only for wlan0 - DNS disabled)..."
if check_dnsmasq_config; then
    echo "🌐 dnsmasq already configured for AP mode - skipping"
else
    cat > $DNSMASQ_CONF << EOF
# Pi Zero W dnsmasq config - DHCP ONLY for wlan0 - $(date)
# Pure DHCP server - no DNS interference
interface=$WIFI_INTERFACE
bind-interfaces
except-interface=lo
dhcp-range=$DHCP_START,$DHCP_END,$DHCP_NETMASK,$DHCP_LEASE_TIME
# Disable DNS server entirely - only DHCP
port=0
EOF
fi


# Configure firewall for standalone AP (no internet routing for AP clients)
echo "🏴‍☠️ Configuring standalone AP (no internet routing for clients)..."
# Leave IP forwarding alone - other interfaces (OTG) might need it
echo "ℹ️  IP forwarding left unchanged to preserve other network interfaces"

# Clear only wlan0-related rules instead of nuking everything
echo "🔥 Configuring firewall for wlan0 AP (preserving other interfaces like usb0)..."

# Remove existing wlan0 rules instead of destroying all rules
iptables -S | grep "$WIFI_INTERFACE" | sed 's/^-A/-D/' | while read rule; do
    iptables $rule 2>/dev/null || true
done

# Remove existing wlan0 NAT rules ONLY
iptables -t nat -S | grep "$WIFI_INTERFACE" | sed 's/^-A/-D/' | while read rule; do
    iptables -t nat $rule 2>/dev/null || true
done

# Basic firewall rules for isolated wlan0 network only
iptables -A INPUT -i $WIFI_INTERFACE -j ACCEPT
iptables -A OUTPUT -o $WIFI_INTERFACE -j ACCEPT

# Preserve any existing usb0/OTG rules - don't touch them
echo "ℹ️  Preserving existing USB OTG (usb0) network configuration"

# Service configuration with proper dependencies
echo "⚙️ Configuring fucking services..."
systemctl unmask $HOSTAPD_SERVICE
systemctl enable $HOSTAPD_SERVICE
systemctl enable $DNSMASQ_SERVICE

# Create dnsmasq service override for proper startup order and retry logic
mkdir -p $DNSMASQ_SERVICE_OVERRIDE_DIR
cat > $DNSMASQ_SERVICE_OVERRIDE_DIR/override.conf << EOF
[Unit]
After=$HOSTAPD_SERVICE.service
Wants=$HOSTAPD_SERVICE.service

[Service]
Restart=on-failure
RestartSec=2
StartLimitBurst=5
StartLimitIntervalSec=30
EOF

systemctl daemon-reload

# No fancy management scripts needed - use make targets instead

# Final validation
echo "🔍 Running final fucking checks..."

# Check if country code is properly set
if ! grep -q "country=$COUNTRY_CODE" $WPA_SUPPLICANT_CONF 2>/dev/null; then
    echo -e "${YELLOW}Warning: Country code might not be set in wpa_supplicant shit${NC}"
fi

# Skip validation - let systemd handle any config issues on boot

# Check services are enabled
for service in $HOSTAPD_SERVICE $DNSMASQ_SERVICE; do
    if ! systemctl is-enabled "$service" >/dev/null 2>&1; then
        echo -e "${RED}Error: $service not enabled${NC}"
        exit 1
    fi
done

echo ""
echo -e "${GREEN}🏴‍☠️ PIrateRF AP Setup Fucking Complete! 🏴‍☠️${NC}"
echo "================================================"
echo "SSID: $AP_SSID"
echo "Password: $AP_PASSWORD"
echo "AP IP: $AP_IP"
echo "DHCP: $DHCP_START-$(echo $DHCP_END | cut -d'.' -f4) ($DHCP_LEASE_TIME lease)"
echo "Country: $COUNTRY_CODE"
echo "Channel: $AP_CHANNEL"
echo ""
echo "💾 Your old shit backed up to: $BACKUP_DIR"
echo ""

echo -e "${YELLOW}🚀 Setup complete! Reboot manually when ready: sudo reboot${NC}"
