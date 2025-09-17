#!/bin/bash

# Check if running as root, if not, re-run with sudo
if [ "$EUID" -ne 0 ]; then
    echo "🔐 Need root privileges, re-running with sudo..."
    exec sudo "$0" "$@"
fi

export ENV=prod
export LOG_LEVEL=warn
export LOG_FORMAT=json
export LOG_CALLER=true
export GORPITX_PATH=/home/fucker/rpitx
export HTTP_SERVER_LISTENADDRESS=0.0.0.0:80

# Create TLS directory and generate certificates if they don't exist
TLS_DIR="./.tls"
CERT_FILE="$TLS_DIR/cert.pem"
KEY_FILE="$TLS_DIR/key.pem"

export HTTP_SERVER_TLSENABLED=true
export HTTP_SERVER_TLSLISTENADDRESS=0.0.0.0:443
export HTTP_SERVER_TLSCERTFILE="$CERT_FILE"
export HTTP_SERVER_TLSKEYFILE="$KEY_FILE"

if [ ! -d "$TLS_DIR" ] || [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
    echo "🔒 TLS certificates not found, generating self-signed certificates..."

    # Create TLS directory if it doesn't exist
    mkdir -p "$TLS_DIR"

    # Generate self-signed certificate
    openssl req -x509 -newkey rsa:4096 -keyout "$KEY_FILE" -out "$CERT_FILE" \
        -days 365 -nodes -subj "/C=US/ST=State/L=City/O=PIrateRF/CN=localhost"

    # Set appropriate permissions
    chmod 600 "$KEY_FILE"
    chmod 644 "$CERT_FILE"

    echo "✅ Self-signed TLS certificates generated successfully"
fi

export PIRATERF_HTMLDIR=./html
export PIRATERF_STATICDIR=./static
export PIRATERF_FILESDIR=./files
export PIRATERF_UPLOADDIR=./uploads

./piraterf run
