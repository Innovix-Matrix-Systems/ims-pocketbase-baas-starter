package logger

import (
	"fmt"
	"log"
	"sync"

	"github.com/pocketbase/pocketbase"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of a LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger interface defines the methods for our custom logger
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	SetStoreLogs(store bool)
	IsStoringLogs() bool
}

// loggerImpl implements the Logger interface
type loggerImpl struct {
	pbApp     *pocketbase.PocketBase
	storeLogs bool
}

// singleton instance
var (
	instance Logger
	once     sync.Once
)

// GetLogger returns the singleton logger instance
func GetLogger(app *pocketbase.PocketBase) Logger {
	once.Do(func() {
		instance = &loggerImpl{
			pbApp:     app,
			storeLogs: true, // Default to storing logs in DB
		}
	})
	return instance
}

// SetStoreLogs enables or disables storing logs in the database
func (l *loggerImpl) SetStoreLogs(store bool) {
	l.storeLogs = store
}

// IsStoringLogs returns whether logs are being stored in the database
func (l *loggerImpl) IsStoringLogs() bool {
	return l.storeLogs
}

// logWithLevel is a helper method that handles logging at different levels
func (l *loggerImpl) logWithLevel(level LogLevel, msg string, keysAndValues ...any) {
	// Format the message with key-value pairs
	formattedMsg := l.formatMessage(msg, keysAndValues...)

	// Always log to stdout using Go's default logger
	log.Printf("[%s] %s", level.String(), formattedMsg)

	// If storing logs is enabled and we have a PocketBase app, use its logger
	if l.storeLogs && l.pbApp != nil {
		switch level {
		case DEBUG:
			l.pbApp.Logger().Debug(msg, keysAndValues...)
		case INFO:
			l.pbApp.Logger().Info(msg, keysAndValues...)
		case WARN:
			l.pbApp.Logger().Warn(msg, keysAndValues...)
		case ERROR:
			l.pbApp.Logger().Error(msg, keysAndValues...)
		}
	}
}

// formatMessage formats the log message with key-value pairs for stdout
func (l *loggerImpl) formatMessage(msg string, keysAndValues ...any) string {
	if len(keysAndValues) == 0 {
		return msg
	}

	result := msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i]
			value := keysAndValues[i+1]
			result = fmt.Sprintf("%s %v=%v", result, key, value)
		}
	}
	return result
}

// Debug logs a message at DEBUG level
func (l *loggerImpl) Debug(msg string, keysAndValues ...any) {
	l.logWithLevel(DEBUG, msg, keysAndValues...)
}

// Info logs a message at INFO level
func (l *loggerImpl) Info(msg string, keysAndValues ...any) {
	l.logWithLevel(INFO, msg, keysAndValues...)
}

// Warn logs a message at WARN level
func (l *loggerImpl) Warn(msg string, keysAndValues ...any) {
	l.logWithLevel(WARN, msg, keysAndValues...)
}

// Error logs a message at ERROR level
func (l *loggerImpl) Error(msg string, keysAndValues ...any) {
	l.logWithLevel(ERROR, msg, keysAndValues...)
}

// LogToDB directly logs to the PocketBase database (if needed for custom use cases)
func (l *loggerImpl) LogToDB(level LogLevel, msg string, keysAndValues ...any) error {
	if l.pbApp == nil {
		return fmt.Errorf("pocketbase app not initialized")
	}

	// This would require creating a custom logs collection in PocketBase
	// For now, we'll just use the PocketBase logger which automatically stores logs
	l.logWithLevel(level, msg, keysAndValues...)
	return nil
}
