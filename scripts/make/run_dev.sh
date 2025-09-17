#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/servicepack/common.sh"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')

section "🐳 Run PIrateRF in fucking development mode"

# Create required directories
info "Creating fucking local development directories..."
# Default directory paths (matching piraterf service defaults)
PIRATERF_HTMLDIR="${PIRATERF_HTMLDIR:-./html}"
PIRATERF_STATICDIR="${PIRATERF_STATICDIR:-./static}"
PIRATERF_FILESDIR="${PIRATERF_FILESDIR:-./files}"
PIRATERF_UPLOADDIR="${PIRATERF_UPLOADDIR:-./uploads}"

# Directory permissions (matching piraterf service)
DIR_PERMS=750

# Create all required directories with proper permissions
mkdir -p "$PIRATERF_HTMLDIR" && chmod $DIR_PERMS "$PIRATERF_HTMLDIR"
mkdir -p "$PIRATERF_STATICDIR" && chmod $DIR_PERMS "$PIRATERF_STATICDIR"
mkdir -p "$PIRATERF_FILESDIR" && chmod $DIR_PERMS "$PIRATERF_FILESDIR"
mkdir -p "$PIRATERF_UPLOADDIR" && chmod $DIR_PERMS "$PIRATERF_UPLOADDIR"

# Audio directories
mkdir -p "$PIRATERF_FILESDIR/audio" && chmod $DIR_PERMS "$PIRATERF_FILESDIR/audio"
mkdir -p "$PIRATERF_FILESDIR/audio/uploads" && chmod $DIR_PERMS "$PIRATERF_FILESDIR/audio/uploads"
mkdir -p "$PIRATERF_FILESDIR/audio/sfx" && chmod $DIR_PERMS "$PIRATERF_FILESDIR/audio/sfx"

# Images directories
mkdir -p "$PIRATERF_FILESDIR/images" && chmod $DIR_PERMS "$PIRATERF_FILESDIR/images"
mkdir -p "$PIRATERF_FILESDIR/images/uploads" && chmod $DIR_PERMS "$PIRATERF_FILESDIR/images/uploads"

success "✅ All required directories fucking created"

# Build dev image first
info "🐳 Building the fucking development Docker image..."
"$SCRIPT_DIR/servicepack/docker_build_dev.sh"

info "🚀 Starting the containerized development shit..."
docker run -i --rm \
    -p 127.0.0.1:8080:8080 \
    -p 127.0.0.1:8443:8443 \
    --name "$APP_NAME-dev" \
    -v "$(pwd)/html:/app/html" \
    -v "$(pwd)/static:/app/static" \
    -v "$(pwd)/uploads:/app/uploads" \
    -v "$(pwd)/files:/app/files" \
    -v "$(pwd)/.tls:/app/.tls" \
    -e ENV=dev \
    -e LOG_LEVEL=debug \
    -e LOG_FORMAT=text \
    -e LOG_CALLER=true \
    -e SERVICES_ENABLED= \
    -e GORPITX_PATH=/home/$PI_USER/rpitx \
    -e HTTP_SERVER_LISTENADDRESS=0.0.0.0:8080 \
    -e HTTP_SERVER_TLSENABLED=true \
    -e HTTP_SERVER_TLSLISTENADDRESS=0.0.0.0:8443 \
    -e HTTP_SERVER_TLSCERTFILE=./.tls/cert.pem \
    -e HTTP_SERVER_TLSKEYFILE=./.tls/key.pem \
    -e PIRATERF_HTMLDIR=./html \
    -e PIRATERF_STATICDIR=./static \
    -e PIRATERF_FILESDIR=./files \
    -e PIRATERF_UPLOADDIR=./uploads \
    "$APP_NAME-dev" sh -c "CGO_ENABLED=1 go build -race -o ./build/$APP_NAME ./cmd/... && ./build/$APP_NAME run"

success "✅ Development container fucking finished"
