#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

section "🔨 Building the fucking PIrateRF beast"

# Check for TLS certificates
info "🔨 Checking for fucking TLS certificates..."
if [ ! -f .tls/cert.pem ] || [ ! -f .tls/key.pem ]; then
    info "🔐 TLS certificates missing, generating the fucking certs..."
    make tls
else
    success "✅ TLS certificates already exist, fucking good!"
fi

# Build using Docker
info "🔨 Compiling this fucking beast for ARM/Pi Zero..."
mkdir -p ./build

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

docker run --rm \
    -v "$(pwd)":/app \
    -w /app \
    -e USER_UID="$(id -u)" \
    -e USER_GID="$(id -g)" \
    golang:1.24.6-alpine \
    sh -c "apk add --no-cache gcc musl-dev && \
        GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build -a \
        -ldflags '-extldflags \"-static\" -X main.appName=$APP_NAME' \
        -o ./build/$APP_NAME ./cmd/main.go && \
        chown \$USER_UID:\$USER_GID ./build/$APP_NAME"

# Pack deployment archive
info "📦 Packing up the deployment shit..."
cd build && tar -czf piraterf.tar.gz \
    "$APP_NAME" \
    ../html \
    ../static \
    ../files/audio/sfx \
    ../.tls \
    --transform 's|.*scripts/||' \
    ../scripts/piraterf.sh \
    ../scripts/install.sh \
    ../scripts/uninstall.sh \
    ../scripts/pi_config.sh

# Remove binary, keep only archive
rm "$APP_NAME"

success "✅ Build fucking complete, ready to deploy this shit!"
