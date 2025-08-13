package cronutils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase"
)

// CronExecutionContext provides common utilities for cron execution
type CronExecutionContext struct {
	App       *pocketbase.PocketBase
	CronID    string
	StartTime time.Time
}

// NewCronExecutionContext creates a new job execution context
func NewCronExecutionContext(app *pocketbase.PocketBase, CronID string) *CronExecutionContext {
	return &CronExecutionContext{
		App:       app,
		CronID:    CronID,
		StartTime: time.Now(),
	}
}

// LogStart logs the start of a job execution
func (ctx *CronExecutionContext) LogStart(message string) {
	log := logger.GetLogger(ctx.App)
	log.Info(fmt.Sprintf("Job %s started", ctx.CronID), "message", message, "start_time", ctx.StartTime)
}

// LogEnd logs the end of a job execution with duration
func (ctx *CronExecutionContext) LogEnd(message string) {
	duration := time.Since(ctx.StartTime)
	log := logger.GetLogger(ctx.App)
	log.Info(fmt.Sprintf("Job %s completed", ctx.CronID), "message", message, "duration", duration)
}

// LogError logs an error during job execution
func (ctx *CronExecutionContext) LogError(err error, message string) {
	duration := time.Since(ctx.StartTime)
	log := logger.GetLogger(ctx.App)
	log.Error(fmt.Sprintf("Job %s failed", ctx.CronID), "error", err, "message", message, "duration", duration)
}

// LogDebug logs for dev and debugging
func (ctx *CronExecutionContext) LogDebug(data any, message string) {
	log := logger.GetLogger(ctx.App)
	log.Debug(fmt.Sprintf("message: %s, data: %v", message, data))
}

// WithRecovery wraps a job function with panic recovery
func WithRecovery(app *pocketbase.PocketBase, CronID string, jobFunc func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log := logger.GetLogger(app)
				log.Error(fmt.Sprintf("Job %s panicked", CronID), "panic", r)
			}
		}()
		jobFunc()
	}
}

// ValidateCronExpression validates a cron expression format
func ValidateCronExpression(cronExpr string) error {
	if cronExpr == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}

	// Split the expression into fields
	fields := strings.Fields(cronExpr)

	// Standard cron has 5 fields: minute hour day month weekday
	// Some systems support 6 fields with seconds as the first field
	if len(fields) != 5 && len(fields) != 6 {
		return fmt.Errorf("invalid cron expression: expected 5 or 6 fields, got %d", len(fields))
	}

	// Regular expression for validating each field
	// Supports: numbers, ranges (1-5), lists (1,3,5), steps (*/5, 1-10/2), wildcards (*)
	fieldPattern := `^(\*(/\d+)?|(\d+(-\d+)?(,\d+(-\d+)?)*)(\/\d+)?|\?)$`

	fieldRegex := regexp.MustCompile(fieldPattern)

	// Field ranges for validation (min-max values)
	var fieldRanges [][]int
	if len(fields) == 6 {
		// With seconds field: second, minute, hour, day, month, weekday
		fieldRanges = [][]int{{0, 59}, {0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	} else {
		// Standard 5-field: minute, hour, day, month, weekday
		fieldRanges = [][]int{{0, 59}, {0, 23}, {1, 31}, {1, 12}, {0, 7}}
	}

	// Validate each field
	for i, field := range fields {
		// Check basic pattern
		if !fieldRegex.MatchString(field) {
			return fmt.Errorf("invalid cron field %d: %s", i+1, field)
		}

		// Skip wildcard and question mark
		if field == "*" || field == "?" {
			continue
		}

		// Validate numeric ranges
		if err := validateFieldRange(field, fieldRanges[i][0], fieldRanges[i][1]); err != nil {
			return fmt.Errorf("invalid cron field %d: %v", i+1, err)
		}
	}

	return nil
}

// validateFieldRange validates that numeric values in a cron field are within acceptable ranges
func validateFieldRange(field string, min, max int) error {
	// Handle step notation (e.g., "*/5" or "1-10/2")
	parts := strings.Split(field, "/")
	baseField := parts[0]

	// If it starts with *, it's a step pattern like */5
	if baseField == "*" {
		return nil // */N is always valid for step patterns
	}

	// Handle comma-separated values
	values := strings.Split(baseField, ",")
	for _, value := range values {
		// Handle range notation (e.g., "1-5")
		if strings.Contains(value, "-") {
			rangeParts := strings.Split(value, "-")
			if len(rangeParts) != 2 {
				return fmt.Errorf("invalid range format: %s", value)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return fmt.Errorf("invalid range start: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return fmt.Errorf("invalid range end: %s", rangeParts[1])
			}

			if start < min || start > max || end < min || end > max {
				return fmt.Errorf("range %s outside valid bounds [%d-%d]", value, min, max)
			}

			if start > end {
				return fmt.Errorf("invalid range: start %d greater than end %d", start, end)
			}
		} else {
			// Handle single numeric value
			num, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid numeric value: %s", value)
			}

			if num < min || num > max {
				return fmt.Errorf("value %d outside valid bounds [%d-%d]", num, min, max)
			}
		}
	}

	return nil
}
