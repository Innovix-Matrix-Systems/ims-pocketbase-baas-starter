# Logger Package

This package provides a unified logging interface for the application that follows the singleton pattern. It can optionally store logs in the database using PocketBase's built-in logger or use Go's standard logger.

## Features

- Singleton pattern - only one logger instance per application lifecycle
- Unified interface for all log levels (Debug, Info, Warn, Error)
- Optional database storage of logs through PocketBase's logger
- Fallback to Go's standard logger for console output
- Structured logging with key-value pairs

## Usage

### Initializing the Logger

```go
import (
    "github.com/pocketbase/pocketbase"
    "ims-pocketbase-baas-starter/pkg/logger"
)

app := pocketbase.New()
logger := logger.GetLogger(app)
```

### Logging Messages

```go
// Info level logging
logger.Info("User logged in", "user_id", userId, "ip", clientIP)

// Error level logging
logger.Error("Database connection failed", "error", err, "host", dbHost)

// Warning level logging
logger.Warn("Deprecated API used", "endpoint", "/api/v1/users", "user_id", userId)

// Debug level logging
logger.Debug("Processing request", "request_id", requestId, "method", method)
```

### Configuring Log Storage

By default, logs are stored in the database using PocketBase's logger. You can disable this behavior:

```go
// Disable database storage of logs
logger.SetStoreLogs(false)

// Enable database storage of logs (default)
logger.SetStoreLogs(true)

// Check current storage setting
isStoring := logger.IsStoringLogs()
```

## Log Levels

The logger supports four log levels:

1. **DEBUG** - Detailed information, typically of interest only when diagnosing problems
2. **INFO** - Confirmation that things are working as expected
3. **WARN** - An indication that something unexpected happened, but the application can continue
4. **ERROR** - An error occurred that prevented a function from completing

## Integration with PocketBase

The logger integrates with PocketBase's built-in logging system. When `SetStoreLogs(true)` is set (default), all log messages are sent to PocketBase's logger which stores them in the database for later retrieval and analysis.