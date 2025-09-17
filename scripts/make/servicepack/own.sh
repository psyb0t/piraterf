#!/bin/bash

set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

MODNAME="$1"

if [ -z "$MODNAME" ]; then
    error "Module name is required"
    echo "Usage: $0 <module-name>"
    echo "Example: $0 github.com/foo/bar"
    exit 1
fi

section "Making Project Your Own"
info "Module name: $MODNAME"

section "Go Version Check"

# Get the current Go version from go.mod
REQUIRED_GO_VERSION=$(grep "^go " go.mod | awk '{print $2}')
info "Required Go version from current go.mod: $REQUIRED_GO_VERSION"

# Get the user's Go version
USER_GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
info "Your Go version: $USER_GO_VERSION"

# Compare versions (simple string comparison works for Go versions)
if [ "$(printf '%s\n' "$REQUIRED_GO_VERSION" "$USER_GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_GO_VERSION" ]; then
    error "Your Go version ($USER_GO_VERSION) is less than the required version ($REQUIRED_GO_VERSION)"
    warning "Please upgrade to Go $REQUIRED_GO_VERSION or higher before proceeding."
    exit 1
fi

success "Go version check passed!"

section "Cleaning Project"

# Remove .git directory
if [ -d ".git" ]; then
    info "Removing .git directory..."
    rm -rf .git
fi

# Get the old module name before removing go.mod
OLD_MODULE=$(grep "^module " go.mod | awk '{print $2}')
info "Old module name: $OLD_MODULE"

# Remove go.sum
if [ -f "go.sum" ]; then
    info "Removing go.sum..."
    rm -f go.sum
fi

# Remove go.mod
if [ -f "go.mod" ]; then
    info "Removing go.mod..."
    rm -f go.mod
fi

section "Creating New Module"

# Create new go.mod with the provided module name and current Go version
info "Creating new go.mod with module name: $MODNAME"
cat > go.mod << EOF
module $MODNAME

go $REQUIRED_GO_VERSION
EOF

success "New go.mod created successfully!"

section "Replacing Module References"

# Replace all occurrences of old module name with new module name in all files
info "Replacing all references from '$OLD_MODULE' to '$MODNAME' in all files..."
find . -type f -name "*.go" -exec sed -i "s|$OLD_MODULE|$MODNAME|g" {} \;
find . -type f -name "*.mod" -exec sed -i "s|$OLD_MODULE|$MODNAME|g" {} \;
find . -type f -name "*.md" -exec sed -i "s|$OLD_MODULE|$MODNAME|g" {} \;

# Replace README.md with just the project name
PROJECT_NAME=$(echo "$MODNAME" | awk -F'/' '{print $NF}')
info "Creating new README.md for project: $PROJECT_NAME"
cat > README.md << EOF
# $PROJECT_NAME

---

*Built with spite using https://github.com/psyb0t/servicepack*
EOF

# Get current servicepack version and save it
if [ -f "servicepack.version" ]; then
    CURRENT_VERSION=$(cat servicepack.version)
    info "Preserving servicepack version: $CURRENT_VERSION"
else
    # Get latest commit from servicepack repo to set as current version
    LATEST_VERSION=$(git ls-remote https://github.com/psyb0t/servicepack HEAD | cut -f1)
    printf "%s" "$LATEST_VERSION" > servicepack.version
    info "Set servicepack version to: $LATEST_VERSION"
fi

success "Module name replacement completed!"

section "Installing Tools"

# Install golangci-lint tool
info "Installing golangci-lint tool..."
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4

# Install gofindimpl tool
info "Installing gofindimpl tool..."
go get -tool github.com/psyb0t/gofindimpl

# Install goimports tool
info "Installing goimports tool..."
go get -tool golang.org/x/tools/cmd/goimports

section "Updating Dependencies"
make dep

section "Initializing Git Repository"

# Initialize git repository
info "Initializing git repository..."
git init

# Rename branch to main
info "Renaming branch to main..."
git branch -m main

# Add all files and create initial commit
info "Creating initial commit..."
git add -A
git commit -m "Initial commit"

success "Project setup completed!"
