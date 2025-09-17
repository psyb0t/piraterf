#!/bin/bash

# Post-update script - runs after rsync to ensure latest framework logic
# This script itself gets updated with each framework update

set -e

# These variables are passed from the main update script
CURRENT_BRANCH="$1"
UPDATE_BRANCH="$2"
CURRENT_VERSION="$3"
LATEST_VERSION="$4"
USER_MODULE="$5"
TEMP_DIR="$6"

# Source common functions (from the newly updated framework)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

section "Restoring User Configuration"

# If user had a custom module name, restore it everywhere
if [ -n "$USER_MODULE" ]; then
    info "Restoring user module name in go.mod..."
    sed -i "1s|.*|module $USER_MODULE|" go.mod

    info "Replacing module references in all files..."
    # Get the original module name from the downloaded framework
    FRAMEWORK_MODULE=$(head -n 1 "$TEMP_DIR/go.mod" | awk '{print $2}')

    # Replace all references to framework module with user module
    find . -type f -name "*.go" -not -path "./vendor/*" -exec sed -i "s|$FRAMEWORK_MODULE|$USER_MODULE|g" {} \;
    find . -type f -name "*.mod" -not -path "./vendor/*" -exec sed -i "s|$FRAMEWORK_MODULE|$USER_MODULE|g" {} \;
fi

# Save the new version
printf "%s" "$LATEST_VERSION" > servicepack.version
success "Updated servicepack.version to: $LATEST_VERSION"

section "Updating Dependencies"
make dep

# Regenerate service registration after dependency updates
info "Regenerating service registration..."
make service-registration

section "Creating Post-Update Commands"

# Create temp directory and scripts
mkdir -p scripts/.post-update-temp

# Create review script
cat > scripts/.post-update-temp/review.sh << EOF
#!/bin/bash
echo "=== Reviewing Servicepack Update ==="
echo "Showing changes from $CURRENT_BRANCH to $UPDATE_BRANCH:"
echo ""
git diff $CURRENT_BRANCH..HEAD -- . ':!vendor'
echo ""
echo "=== Update Summary ==="
echo "Current branch: \$(git branch --show-current)"
echo "Changes ready for review. Use:"
echo "  make servicepack-update-merge   - to merge and finish update"
echo "  make servicepack-update-revert  - to cancel and revert"
EOF

# Create merge script
cat > scripts/.post-update-temp/merge.sh << EOF
#!/bin/bash
echo "=== Merging Servicepack Update ==="
git checkout $CURRENT_BRANCH
git merge $UPDATE_BRANCH
git branch -d $UPDATE_BRANCH
echo ""
echo "✅ Update merged successfully!"
echo "🧹 Cleaning up temp files..."
rm -rf scripts/.post-update-temp
echo "✅ Update complete!"
EOF

# Create revert script
cat > scripts/.post-update-temp/revert.sh << EOF
#!/bin/bash
echo "=== Reverting Servicepack Update ==="
git checkout $CURRENT_BRANCH
git branch -D $UPDATE_BRANCH
echo ""
echo "❌ Update reverted successfully!"
echo "🧹 Cleaning up temp files..."
rm -rf scripts/.post-update-temp
echo "💡 Backup is still available. Use 'make backup-restore' if needed."
EOF

# Make scripts executable
chmod +x scripts/.post-update-temp/*.sh

# Commit changes to update branch
git add -A
git commit -m "Updated servicepack from $CURRENT_VERSION to $LATEST_VERSION"

section "Finalizing Update"

# Clean up
rm -rf "$TEMP_DIR"
info "Cleaned up temporary files"

success "Framework updated successfully in branch '$UPDATE_BRANCH'!"
info "You are now on the update branch to review changes."

section "Next Steps"
echo "Update branch '$UPDATE_BRANCH' created successfully!"
echo ""
echo "Use these commands to manage the update:"
echo ""
echo "1. Review changes:"
echo -e "   ${BLUE}make servicepack-update-review${NC}"
echo ""
echo "2. Test the update:"
echo -e "   ${BLUE}make dep && make service-registration && make test${NC}"
echo ""
echo "3. If satisfied, merge the update:"
echo -e "   ${BLUE}make servicepack-update-merge${NC}"
echo ""
echo "4. If not satisfied, revert:"
echo -e "   ${BLUE}make servicepack-update-revert${NC}"
echo ""
warning "A backup was created before updating. Use 'make backup-restore' if needed."