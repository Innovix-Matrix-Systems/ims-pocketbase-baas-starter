package logger

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// FromApp retrieves a logger instance from a core.App interface
// This function handles the type assertion from core.App to *pocketbase.PocketBase
// and returns a logger instance. If the type assertion fails, it returns nil.
func FromApp(app core.App) Logger {
	if pbApp, ok := app.(*pocketbase.PocketBase); ok {
		return GetLogger(pbApp)
	}
	return nil
}

// FromAppOrDefault retrieves a logger instance from a core.App interface
// This function handles the type assertion from core.App to *pocketbase.PocketBase
// and returns a logger instance. If the type assertion fails, it returns a no-op logger
// that doesn't store logs but still outputs to stdout.
func FromAppOrDefault(app core.App) Logger {
	if pbApp, ok := app.(*pocketbase.PocketBase); ok {
		return GetLogger(pbApp)
	}

	// Return a no-op logger that only logs to stdout
	return &noopLogger{}
}

// noopLogger is a no-op logger that only logs to stdout
type noopLogger struct{}

func (n *noopLogger) Debug(msg string, keysAndValues ...any) {
	logWithLevel(DEBUG, msg, keysAndValues...)
}

func (n *noopLogger) Info(msg string, keysAndValues ...any) {
	logWithLevel(INFO, msg, keysAndValues...)
}

func (n *noopLogger) Warn(msg string, keysAndValues ...any) {
	logWithLevel(WARN, msg, keysAndValues...)
}

func (n *noopLogger) Error(msg string, keysAndValues ...any) {
	logWithLevel(ERROR, msg, keysAndValues...)
}

func (n *noopLogger) SetStoreLogs(store bool) {
	// No-op
}

func (n *noopLogger) IsStoringLogs() bool {
	return false
}

// logWithLevel is a helper function that logs to stdout only
func logWithLevel(level LogLevel, msg string, keysAndValues ...any) {
	// Format the message with key-value pairs
	formattedMsg := formatMessage(msg, keysAndValues...)

	// Always log to stdout using Go's default logger
	log.Printf("[%s] %s", level.String(), formattedMsg)
}

// formatMessage formats the log message with key-value pairs for stdout
func formatMessage(msg string, keysAndValues ...any) string {
	if len(keysAndValues) == 0 {
		return msg
	}

	result := msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i]
			value := keysAndValues[i+1]
			result = formatAppend(result, key, value)
		}
	}
	return result
}

// formatAppend appends a key-value pair to a message
func formatAppend(msg string, key, value any) string {
	return msg + " " + formatKey(key) + "=" + formatValue(value)
}

// formatKey formats a key for logging
func formatKey(key any) string {
	return formatValue(key)
}

// formatValue formats a value for logging
func formatValue(value any) string {
	return fmt.Sprintf("%v", value)
}
