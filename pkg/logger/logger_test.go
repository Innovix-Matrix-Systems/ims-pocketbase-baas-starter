package logger

import (
	"testing"

	"github.com/pocketbase/pocketbase"
)

func TestLoggerSingleton(t *testing.T) {
	app1 := pocketbase.New()
	logger1 := GetLogger(app1)

	app2 := pocketbase.New()
	logger2 := GetLogger(app2)

	// Both should return the same instance
	if logger1 != logger2 {
		t.Error("Expected same logger instance, got different instances")
	}
}

func TestLoggerLevels(t *testing.T) {
	app := pocketbase.New()
	logger := GetLogger(app)

	// Test all log levels
	logger.Debug("Debug message", "key", "value")
	logger.Info("Info message", "count", 42)
	logger.Warn("Warning message", "warning", "test")
	logger.Error("Error message", "error", "test")

	// Test that the logger is storing logs by default
	if !logger.IsStoringLogs() {
		t.Error("Expected logger to store logs by default")
	}

	// Test disabling log storage
	logger.SetStoreLogs(false)
	if logger.IsStoringLogs() {
		t.Error("Expected logger to not store logs after disabling")
	}

	// Re-enable log storage
	logger.SetStoreLogs(true)
	if !logger.IsStoringLogs() {
		t.Error("Expected logger to store logs after re-enabling")
	}
}

func TestLogLevelString(t *testing.T) {
	if DEBUG.String() != "DEBUG" {
		t.Errorf("Expected DEBUG.String() to be 'DEBUG', got %s", DEBUG.String())
	}

	if INFO.String() != "INFO" {
		t.Errorf("Expected INFO.String() to be 'INFO', got %s", INFO.String())
	}

	if WARN.String() != "WARN" {
		t.Errorf("Expected WARN.String() to be 'WARN', got %s", WARN.String())
	}

	if ERROR.String() != "ERROR" {
		t.Errorf("Expected ERROR.String() to be 'ERROR', got %s", ERROR.String())
	}

	// Test unknown level
	unknownLevel := LogLevel(999)
	if unknownLevel.String() != "UNKNOWN" {
		t.Errorf("Expected unknown level to return 'UNKNOWN', got %s", unknownLevel.String())
	}
}

func TestFormatMessage(t *testing.T) {
	app := pocketbase.New()
	logger := GetLogger(app).(*loggerImpl)

	// Test formatting with no key-value pairs
	result := logger.formatMessage("Simple message")
	if result != "Simple message" {
		t.Errorf("Expected 'Simple message', got %s", result)
	}

	// Test formatting with key-value pairs
	result = logger.formatMessage("Message", "key1", "value1", "key2", "value2")
	if result != "Message key1=value1 key2=value2" {
		t.Errorf("Expected 'Message key1=value1 key2=value2', got %s", result)
	}

	// Test formatting with odd number of key-value pairs
	result = logger.formatMessage("Message", "key1", "value1", "key2")
	if result != "Message key1=value1" {
		t.Errorf("Expected 'Message key1=value1', got %s", result)
	}
}
