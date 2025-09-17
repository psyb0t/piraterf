#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

BACKUP_FILE="$1"

section "Backup Restore"

# If no backup file specified, use the latest one
if [ -z "$BACKUP_FILE" ]; then
    if [ -d ".backup" ] && find .backup -name "*.tar.gz" -type f -print -quit 2>/dev/null | grep -q .; then
        BACKUP_FILE=$(find .backup -name "*.tar.gz" -type f -printf '%T@ %p\n' 2>/dev/null | sort -nr | head -n1 | cut -d' ' -f2- | xargs basename)
        info "No backup specified, using latest: $BACKUP_FILE"
    else
        error "No backups found in .backup/"
        echo "Usage: make restore-backup [BACKUP=filename.tar.gz]"
        exit 1
    fi
fi

# Check if backup exists in local backup directory
LOCAL_BACKUP=".backup/$BACKUP_FILE"
if [ ! -f "$LOCAL_BACKUP" ]; then
    error "Backup file not found: $LOCAL_BACKUP"
    echo ""
    echo "Available backups:"
    if [ -d ".backup" ] && find .backup -name "*.tar.gz" -type f -print -quit 2>/dev/null | grep -q .; then
        find .backup -name "*.tar.gz" -type f -exec basename {} \; 2>/dev/null
    else
        echo "  No backups found in .backup/"
    fi
    exit 1
fi

warning "This will delete EVERYTHING in the current directory and restore from backup!"
info "Backup file: $LOCAL_BACKUP"
echo ""
read -r -p "Are you sure? Type 'yes' to continue: " confirm

if [ "$confirm" != "yes" ]; then
    warning "Restore cancelled."
    exit 0
fi

info "Nuking current directory (except .backup)..."
find . -mindepth 1 -maxdepth 1 ! -name '.backup' -exec rm -rf {} +

info "Restoring from backup..."
tar -xzf "$LOCAL_BACKUP"

success "Restore completed successfully!"
warning "Don't forget to run 'make dep' to update dependencies if needed."