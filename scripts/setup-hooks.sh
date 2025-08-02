#!/bin/bash

# Script to set up Git hooks for the project
# Run this script once after cloning the repository

echo "🔧 Setting up Git hooks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if we're in a Git repository
if [ ! -d ".git" ]; then
    print_error "Not a Git repository"
    exit 1
fi

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Copy hooks and make them executable
if [ -f ".githooks/pre-commit" ]; then
    cp .githooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    print_status "Pre-commit hook installed"
else
    print_error "Pre-commit hook file not found"
fi

if [ -f ".githooks/pre-push" ]; then
    cp .githooks/pre-push .git/hooks/pre-push
    chmod +x .git/hooks/pre-push
    print_status "Pre-push hook installed"
else
    print_error "Pre-push hook file not found"
fi

# Set up Git hooks path (Git 2.9+)
if git config core.hooksPath .githooks 2>/dev/null; then
    print_status "Git hooks path configured to use .githooks directory"
else
    print_warning "Could not set hooks path. Using manual copy method."
fi

echo ""
echo "🎉 Git hooks setup complete!"
echo ""
echo "The following hooks are now active:"
echo "  • pre-commit: Runs formatting and syntax checks on staged files"
echo "  • pre-push: Runs full tests and linting before pushing"
echo ""
echo "To bypass hooks temporarily:"
echo "  • git commit --no-verify"
echo "  • git push --no-verify"
echo ""
echo "💡 Make sure to install golangci-lint for full linting support:"
echo "   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"