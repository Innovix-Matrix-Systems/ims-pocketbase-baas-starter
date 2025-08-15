# Makefile Commands Guide

This document provides a comprehensive overview of all available Makefile commands for development, production, and utility operations.

## Command Overview

| Development                              | Production                                    | Utility                                  |
| ---------------------------------------- | --------------------------------------------- | ---------------------------------------- |
| `dev` - Start dev environment            | `build` - Build production image              | `help` - Show all commands               |
| `dev-build` - Build dev image            | `start` - Start containers                    | `generate-key` - Generate encryption key |
| `dev-logs` - Show dev logs               | `stop` - Stop containers                      | `setup-env` - Setup environment file     |
| `dev-stop` - Stop dev containers         | `restart` - Restart containers                | `test` - Run tests                       |
| `dev-clean` - Clean dev env              | `down` - Stop and remove containers           | `lint` - Run linter                      |
| `dev-data-clean` - Clean dev data        | `logs` - Show container logs                  | `format` - Format Go code                |
| `dev-start` - Alias for dev              | `clean` - Remove containers, networks, images | `status` - Show container status         |
| `dev-status` - Show dev container status | `clean-data` - Remove only volumes            | `prod-start` - Alias for start           |

## Development Commands

### `make dev`
Starts the development environment with hot reload support using Docker Compose.

```bash
make dev
```

### `make dev-build`
Builds the development Docker image with all dependencies.

```bash
make dev-build
```

### `make dev-logs`
Shows logs from the development containers in real-time.

```bash
make dev-logs
```

### `make dev-stop`
Stops the development containers without removing them.

```bash
make dev-stop
```

### `make dev-clean`
Stops and removes development containers, networks, and images.

```bash
make dev-clean
```

### `make dev-data-clean`
Removes development data volumes (database, cache, etc.).

```bash
make dev-data-clean
```

### `make dev-status`
Shows the status of development containers.

```bash
make dev-status
```

## Production Commands

### `make build`
Builds the production Docker image optimized for deployment.

```bash
make build
```

### `make start` / `make prod-start`
Starts the production containers using docker-compose.yml.

```bash
make start
# or
make prod-start
```

### `make stop`
Stops the production containers.

```bash
make stop
```

### `make restart`
Restarts the production containers.

```bash
make restart
```

### `make down`
Stops and removes production containers and networks.

```bash
make down
```

### `make logs`
Shows logs from production containers.

```bash
make logs
```

### `make clean`
Removes all containers, networks, and images.

```bash
make clean
```

### `make clean-data`
Removes only the data volumes.

```bash
make clean-data
```

## Utility Commands

### `make help`
Shows all available commands with descriptions.

```bash
make help
```

### `make generate-key`
Generates a secure encryption key for PocketBase.

```bash
make generate-key
```

Copy the generated key to your `.env` file as `PB_ENCRYPTION_KEY`.

### `make setup-env`
Creates a `.env` file from `.env.example` template.

```bash
make setup-env
```

### `make test`
Runs the Go test suite.

```bash
make test
```

### `make lint`
Runs Go linting tools to check code quality.

```bash
make lint
```

### `make format`
Formats Go code using `gofmt`.

```bash
make format
```

### `make status`
Shows the status of all containers.

```bash
make status
```

## Common Workflows

### Starting Development
```bash
# First time setup
make setup-env
make generate-key
# Edit .env file with your configuration

# Start development
make dev
```

### Production Deployment
```bash
# Build and deploy
make build
make start

# Monitor
make logs
make status
```

### Cleaning Up
```bash
# Clean development environment
make dev-clean
make dev-data-clean

# Clean production environment
make clean
make clean-data
```

## Tips

- Use `make dev-logs` to monitor application output during development
- Run `make test` before committing changes
- Use `make dev-clean` if you encounter Docker issues
- Check `make status` to verify container health
- Use `make help` to see all available commands