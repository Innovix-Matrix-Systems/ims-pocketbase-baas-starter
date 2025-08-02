# Git Hooks Setup

This project includes Git hooks to ensure code quality and prevent broken code from being committed or pushed.

## Quick Setup

### Linux/macOS
```bash
make setup-hooks
```

### Windows
```bash
make setup-hooks-win
```

## What the Hooks Do

### Pre-commit Hook
Runs on every `git commit` and checks:
- ✅ Code formatting (gofmt)
- ✅ Syntax validation
- ✅ Go vet analysis
- ⚡ Only checks staged files (fast)

### Pre-push Hook
Runs on every `git push` and performs:
- ✅ Full test suite (`make test`)
- ✅ Linting (`make lint`) - if golangci-lint is installed
- ✅ Code formatting check
- ✅ Go vet analysis
- 🔍 Comprehensive checks (slower but thorough)

## Prerequisites

### Required
- Go (already installed)
- Git (already installed)

### Recommended
Install golangci-lint for comprehensive linting:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Bypassing Hooks

Sometimes you need to bypass hooks (use sparingly):

```bash
# Skip pre-commit hook
git commit --no-verify -m "commit message"

# Skip pre-push hook
git push --no-verify
```

## Manual Hook Installation

If the automatic setup doesn't work:

### Linux/macOS
```bash
# Make hooks executable
chmod +x .githooks/pre-commit
chmod +x .githooks/pre-push

# Copy to .git/hooks
cp .githooks/pre-commit .git/hooks/pre-commit
cp .githooks/pre-push .git/hooks/pre-push

# Or configure Git to use .githooks directory
git config core.hooksPath .githooks
```

### Windows
```bash
# Copy hooks
copy .githooks\pre-commit .git\hooks\pre-commit
copy .githooks\pre-push .git\hooks\pre-push

# Or configure Git to use .githooks directory
git config core.hooksPath .githooks
```

## Troubleshooting

### Hook not running
- Check if hooks are executable: `ls -la .git/hooks/`
- Verify Git hooks path: `git config core.hooksPath`

### golangci-lint not found
- Install it: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- Or the hook will skip linting with a warning

### Tests failing in hook
- Run `make test` manually to see detailed output
- Fix failing tests before committing/pushing

### Formatting issues
- Run `make format` to auto-fix formatting
- Or use `gofmt -w .` directly

## Hook Configuration

The hooks are located in `.githooks/` directory:
- `.githooks/pre-commit` - Pre-commit hook script
- `.githooks/pre-push` - Pre-push hook script

You can modify these scripts to customize the checks according to your team's needs.

## Benefits

✅ **Prevent broken builds** - Catch issues before they reach CI/CD
✅ **Consistent code style** - Enforce formatting standards
✅ **Early bug detection** - Find issues during development
✅ **Team productivity** - Reduce time spent on code review
✅ **CI/CD efficiency** - Fewer failed builds in CI

