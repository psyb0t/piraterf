#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

LOCAL_BACKUP=".backup"

if [ ! -d "$LOCAL_BACKUP" ]; then
    warning "No backup directory found."
    exit 0
fi

BACKUP_COUNT=$(find "$LOCAL_BACKUP" -name "*.tar.gz" -type f 2>/dev/null | wc -l)

if [ "$BACKUP_COUNT" -eq 0 ]; then
    warning "No backup files found."
    exit 0
fi

section "Backup Files"
info "Found $BACKUP_COUNT backup files:"
find "$LOCAL_BACKUP" -name "*.tar.gz" -type f -exec basename {} \; 2>/dev/null

echo ""
read -r -p "Delete ALL backup files? Type 'yes' to continue: " confirm

if [ "$confirm" != "yes" ]; then
    warning "Backup clear cancelled."
    exit 0
fi

info "Deleting backup files..."
rm -f "$LOCAL_BACKUP"/*.tar.gz

success "All backup files deleted."