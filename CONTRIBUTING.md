# Contributing to IMS PocketBase BaaS Starter

Thank you for considering contributing to IMS PocketBase BaaS Starter! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Environment details** (Go version, PocketBase version, OS)
- **Code samples** that demonstrate the issue

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear title and detailed description**
- **Use case** explaining why this enhancement would be useful
- **Possible implementation** details if you have ideas

### Pull Requests

1. **Fork** the repository
2. **Create a feature branch** from `main`
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Ensure all tests pass**
6. **Update documentation** if needed
7. **Submit a pull request**

## Development Setup

```bash
# Clone your fork
git clone https://github.com/Innovix-Matrix-Systems/ims-pocketbase-baas-starter

cd ims-pocketbase-baas-starter

# Setup environment
make setup-env

# Start development environment
make dev

# Run tests
make test
```

## Coding Standards

- Follow **Go standard coding conventions**
- Use **meaningful variable and function names**
- Add **type declarations** where possible
- Write **comprehensive tests** for new features
- Keep **backward compatibility** in mind

### Code Style

We use Go's standard formatting tools:

```bash
# Format code
make format

# Run linter
make lint
```

## Testing

All contributions must include appropriate tests:

```bash
# Run all tests
make test
```

### Writing Tests

- Use Go's standard testing package
- Test both **happy path and edge cases**
- Mock external dependencies when appropriate

Example test structure:
```go
func TestFeature(t *testing.T) {
    // Setup
    app := setupTestApp()
    
    // Test case
    result, err := app.DoSomething()
    
    // Assertions
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if result != expectedResult {
        t.Errorf("Expected %v, got %v", expectedResult, result)
    }
}
```

## Database Migrations

When making changes that require database schema modifications:

1. Follow the [Database Migrations Guide](docs/migrations.md)
2. Create properly numbered migration files
3. Include both forward and rollback migrations
4. Test migrations thoroughly

## Documentation

- Update **README.md** for new features
- Add **inline documentation** for complex functions
- Include **usage examples** in comments
- Update **configuration examples** when needed

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add support for custom middleware
fix: resolve authentication issue
docs: update migration instructions
test: add tests for RBAC system
refactor: improve error handling
```

Prefix types:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions/changes
- `refactor:` Code refactoring
- `style:` Code style changes
- `chore:` Maintenance tasks

## Review Process

1. **Automated checks** must pass (tests, code style)
2. **Manual review** by maintainers
3. **Discussion** if changes are needed
4. **Approval** and merge

## Docker Development

For Docker-based development:

```bash
# Build development image
make dev-build

# Start development environment
make dev

# View logs
make dev-logs

# Clean up
make dev-clean
```

## Getting Help

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: Contact maintainers directly for sensitive issues

## Recognition

Contributors will be acknowledged in:
- **CHANGELOG.md** for their contributions
- **README.md** contributors section
- **GitHub contributors** page

Thank you for helping make IMS PocketBase BaaS Starter better! 🚀