#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🔐 Generating Fucking TLS Certificates"

# Create TLS directory
info "📁 Creating .tls directory..."
mkdir -p ./.tls

# Generate self-signed TLS certificates
info "🔐 Generating fucking self-signed TLS certificates..."
cd .tls && openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes \
    -subj "/C=US/ST=Dev/L=Docker/O=PIrateRF-Dev/CN=localhost" \
    -addext "subjectAltName=DNS:localhost,DNS:piraterf.local,IP:127.0.0.1,IP:::1"

# Set proper permissions
info "🔒 Setting fucking proper permissions..."
chmod 600 key.pem
chmod 644 cert.pem

success "✅ TLS certificates fucking generated!"
info "   📜 Fucking Certificate: $(pwd)/cert.pem"
info "   🔑 Fucking Private Key: $(pwd)/key.pem"