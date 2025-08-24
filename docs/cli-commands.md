# CLI Commands Guide

This document explains how to use and extend the CLI command system in the IMS PocketBase BaaS Starter.

## Overview

The CLI command system allows you to execute administrative and maintenance tasks directly through the command line interface. Commands follow the same pattern as routes, crons, and middleware registration.

## Available Commands

### Built-in Commands

The application includes built-in commands:

#### `health` - Application Health Check
```bash
./main health
```
Performs a basic health check to verify database connectivity and other core services.

#### `sync-permissions` - Sync Hardcoded Permissions
```bash
./main sync-permissions
```
Syncs all hardcoded permissions defined in the codebase to the database, creating new ones and skipping existing ones.

## Running Commands

### Development Environment

```bash
# Run commands directly with go run
go run ./cmd/server health
go run ./cmd/server sync-permissions

# Or build and run
go build -o server ./cmd/server
./server health
./server sync-permissions
```

### Production Environment

```bash
# Run commands directly in container
docker exec <container_name> ./main health
docker exec <container_name> ./main sync-permissions

# Using docker-compose
docker-compose exec pocketbase ./main health
docker-compose exec pocketbase ./main sync-permissions
```

## Adding New Commands

### 1. Create the Command Handler

Create a new handler function in `internal/handlers/command/`:

```go
// internal/handlers/command/custom_commands.go
package command

import (
    "github.com/pocketbase/pocketbase"
    "github.com/spf13/cobra"
    "ims-pocketbase-baas-starter/pkg/logger"
)

// HandleMyCommand example command handler
func HandleMyCommand(app *pocketbase.PocketBase, cmd *cobra.Command, args []string) {
    log := logger.GetLogger(app)
    log.Info("My command executed")
    // Your command logic here
}
```

### 2. Register the Command

Add your command to `internal/commands/commands.go`:

```go
// Add to the commands array
{
    ID:      "my-command",
    Use:     "my-command",
    Short:   "Description of my command",
    Long:    "Detailed description of what my command does",
    Handler: command.HandleMyCommand,
    Enabled: true,
},
```

## Best Practices

1. **Use the Application Logger**: Always use `logger.GetLogger(app)` for consistent logging
2. **Handle Errors Gracefully**: Log errors and return appropriately
3. **Validate Input**: Check arguments and flags before processing
4. **Keep Commands Focused**: Each command should have a single, clear purpose

## Troubleshooting

### Common Issues

#### Command Not Found
Ensure the command is registered in `internal/commands/commands.go` and the application has been rebuilt.

#### Permission Denied
Make the binary executable: `chmod +x ./main`

### Debugging Tips

Check available commands: `./main --help`