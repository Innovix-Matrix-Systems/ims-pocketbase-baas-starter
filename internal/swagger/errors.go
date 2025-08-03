package swagger

import (
	"fmt"
	"log"
	"strings"
)

// ErrorLevel represents the severity of an error
type ErrorLevel int

const (
	ErrorLevelWarning ErrorLevel = iota
	ErrorLevelError
	ErrorLevelCritical
)

// SwaggerError represents an error that occurred during swagger generation
type SwaggerError struct {
	Level       ErrorLevel
	Component   string
	Operation   string
	Message     string
	Context     map[string]interface{}
	Cause       error
	Recoverable bool
}

// Error implements the error interface
func (e SwaggerError) Error() string {
	var sb strings.Builder

	levelStr := "ERROR"
	switch e.Level {
	case ErrorLevelWarning:
		levelStr = "WARNING"
	case ErrorLevelCritical:
		levelStr = "CRITICAL"
	}

	sb.WriteString(fmt.Sprintf("[%s] %s", levelStr, e.Component))
	if e.Operation != "" {
		sb.WriteString(fmt.Sprintf(" - %s", e.Operation))
	}
	sb.WriteString(fmt.Sprintf(": %s", e.Message))

	if e.Cause != nil {
		sb.WriteString(fmt.Sprintf(" (caused by: %v)", e.Cause))
	}

	if len(e.Context) > 0 {
		sb.WriteString(" [context:")
		for k, v := range e.Context {
			sb.WriteString(fmt.Sprintf(" %s=%v", k, v))
		}
		sb.WriteString("]")
	}

	return sb.String()
}

// ErrorHandler manages errors during swagger generation
type ErrorHandler struct {
	errors         []SwaggerError
	warningCount   int
	errorCount     int
	criticalCount  int
	maxErrors      int
	failOnWarnings bool
	logErrors      bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		errors:         []SwaggerError{},
		maxErrors:      100,
		failOnWarnings: false,
		logErrors:      true,
	}
}

// NewErrorHandlerWithConfig creates a new error handler with configuration
func NewErrorHandlerWithConfig(maxErrors int, failOnWarnings bool, logErrors bool) *ErrorHandler {
	return &ErrorHandler{
		errors:         []SwaggerError{},
		maxErrors:      maxErrors,
		failOnWarnings: failOnWarnings,
		logErrors:      logErrors,
	}
}

// AddError adds an error to the handler
func (eh *ErrorHandler) AddError(err SwaggerError) {
	eh.errors = append(eh.errors, err)

	switch err.Level {
	case ErrorLevelWarning:
		eh.warningCount++
		if eh.logErrors {
			log.Printf("WARNING: %s", err.Error())
		}
	case ErrorLevelError:
		eh.errorCount++
		if eh.logErrors {
			log.Printf("ERROR: %s", err.Error())
		}
	case ErrorLevelCritical:
		eh.criticalCount++
		if eh.logErrors {
			log.Printf("CRITICAL: %s", err.Error())
		}
	}

	// Check if we've exceeded max errors
	if len(eh.errors) >= eh.maxErrors {
		panic(fmt.Errorf("maximum error count (%d) exceeded", eh.maxErrors))
	}
}

// AddWarning adds a warning-level error
func (eh *ErrorHandler) AddWarning(component, operation, message string, context map[string]interface{}) {
	eh.AddError(SwaggerError{
		Level:       ErrorLevelWarning,
		Component:   component,
		Operation:   operation,
		Message:     message,
		Context:     context,
		Recoverable: true,
	})
}

// AddErrorWithCause adds an error with a cause
func (eh *ErrorHandler) AddErrorWithCause(component, operation, message string, cause error, context map[string]interface{}) {
	eh.AddError(SwaggerError{
		Level:       ErrorLevelError,
		Component:   component,
		Operation:   operation,
		Message:     message,
		Context:     context,
		Cause:       cause,
		Recoverable: false,
	})
}

// AddCriticalError adds a critical error
func (eh *ErrorHandler) AddCriticalError(component, operation, message string, cause error) {
	eh.AddError(SwaggerError{
		Level:       ErrorLevelCritical,
		Component:   component,
		Operation:   operation,
		Message:     message,
		Cause:       cause,
		Recoverable: false,
	})
}

// HasErrors returns true if there are any errors (not warnings)
func (eh *ErrorHandler) HasErrors() bool {
	return eh.errorCount > 0 || eh.criticalCount > 0
}

// HasCriticalErrors returns true if there are any critical errors
func (eh *ErrorHandler) HasCriticalErrors() bool {
	return eh.criticalCount > 0
}

// ShouldFail returns true if generation should fail based on current errors
func (eh *ErrorHandler) ShouldFail() bool {
	if eh.HasCriticalErrors() {
		return true
	}
	if eh.HasErrors() {
		return true
	}
	if eh.failOnWarnings && eh.warningCount > 0 {
		return true
	}
	return false
}

// GetSummary returns a summary of all errors
func (eh *ErrorHandler) GetSummary() string {
	return fmt.Sprintf("Errors: %d critical, %d errors, %d warnings (total: %d)",
		eh.criticalCount, eh.errorCount, eh.warningCount, len(eh.errors))
}

// GetErrors returns all errors of a specific level
func (eh *ErrorHandler) GetErrors(level ErrorLevel) []SwaggerError {
	var filtered []SwaggerError
	for _, err := range eh.errors {
		if err.Level == level {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetAllErrors returns all errors
func (eh *ErrorHandler) GetAllErrors() []SwaggerError {
	return eh.errors
}

// Clear clears all errors
func (eh *ErrorHandler) Clear() {
	eh.errors = []SwaggerError{}
	eh.warningCount = 0
	eh.errorCount = 0
	eh.criticalCount = 0
}

// RecoveryStrategy represents a strategy for recovering from errors
type RecoveryStrategy interface {
	Recover(err SwaggerError) (interface{}, error)
	CanRecover(err SwaggerError) bool
}

// FieldTypeRecovery implements recovery for unknown field types
type FieldTypeRecovery struct{}

// Recover attempts to recover from an unknown field type error
func (ftr *FieldTypeRecovery) Recover(err SwaggerError) (interface{}, error) {
	if fieldType, ok := err.Context["fieldType"].(string); ok {
		// Return a fallback schema for unknown field type
		return &FieldSchema{
			Type:        "string",
			Description: fmt.Sprintf("Unknown field type: %s (using string fallback)", fieldType),
			Required:    false,
		}, nil
	}
	return nil, fmt.Errorf("cannot recover: missing field type in context")
}

// CanRecover checks if this strategy can recover from the error
func (ftr *FieldTypeRecovery) CanRecover(err SwaggerError) bool {
	return err.Component == "FieldMapper" && strings.Contains(err.Message, "unknown field type")
}

// CollectionAccessRecovery implements recovery for collection access errors
type CollectionAccessRecovery struct{}

// Recover attempts to recover from collection access errors
func (car *CollectionAccessRecovery) Recover(err SwaggerError) (interface{}, error) {
	if collectionName, ok := err.Context["collection"].(string); ok {
		// Return an empty collection info as fallback
		return &EnhancedCollectionInfo{
			Name:   collectionName,
			Type:   "base",
			Fields: []FieldInfo{},
			System: false,
		}, nil
	}
	return nil, fmt.Errorf("cannot recover: missing collection name in context")
}

// CanRecover checks if this strategy can recover from the error
func (car *CollectionAccessRecovery) CanRecover(err SwaggerError) bool {
	return err.Component == "CollectionDiscovery" && err.Recoverable
}

// ErrorRecoveryManager manages error recovery strategies
type ErrorRecoveryManager struct {
	strategies []RecoveryStrategy
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager() *ErrorRecoveryManager {
	return &ErrorRecoveryManager{
		strategies: []RecoveryStrategy{
			&FieldTypeRecovery{},
			&CollectionAccessRecovery{},
		},
	}
}

// TryRecover attempts to recover from an error
func (erm *ErrorRecoveryManager) TryRecover(err SwaggerError) (interface{}, error) {
	for _, strategy := range erm.strategies {
		if strategy.CanRecover(err) {
			return strategy.Recover(err)
		}
	}
	return nil, fmt.Errorf("no recovery strategy available for error: %s", err.Error())
}

// AddStrategy adds a recovery strategy
func (erm *ErrorRecoveryManager) AddStrategy(strategy RecoveryStrategy) {
	erm.strategies = append(erm.strategies, strategy)
}
