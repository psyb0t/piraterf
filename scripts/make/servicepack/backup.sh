#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

APP_NAME=$(head -n 1 go.mod | awk '{print $2}' | awk -F'/' '{print $NF}')
TIMESTAMP=$(date +%Y_%m_%d_%H_%M_%S)
BACKUP_NAME="${TIMESTAMP}_${APP_NAME}.tar.gz"
TMP_BACKUP="/tmp/servicepack-backup"
LOCAL_BACKUP=".backup"

section "Creating Backup"
info "Backup name: $BACKUP_NAME"

# Create backup directories
mkdir -p "$TMP_BACKUP"
mkdir -p "$LOCAL_BACKUP"

# Create the backup archive
info "Archiving project..."
tar --exclude='.backup' \
    --exclude='build' \
    --exclude='coverage.txt' \
    -czf "$TMP_BACKUP/$BACKUP_NAME" .

# Copy to local backup directory
cp "$TMP_BACKUP/$BACKUP_NAME" "$LOCAL_BACKUP/"

success "Backup created:"
echo "  - /tmp: $TMP_BACKUP/$BACKUP_NAME"
echo "  - local: $LOCAL_BACKUP/$BACKUP_NAME"

# Keep only latest 6 backups in local directory
info "Cleaning old backups (keeping latest 6)..."
cd "$LOCAL_BACKUP"
find . -name "*.tar.gz" -type f -printf '%T@ %p\n' 2>/dev/null | sort -nr | tail -n +7 | cut -d' ' -f2- | xargs -r rm -f
cd - > /dev/null

BACKUP_COUNT=$(find "$LOCAL_BACKUP" -name "*.tar.gz" -type f 2>/dev/null | wc -l)
info "Local backups: $BACKUP_COUNT/6"

success "Backup completed successfully!"