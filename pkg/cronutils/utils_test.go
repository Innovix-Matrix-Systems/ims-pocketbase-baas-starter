package cronutils

import (
	"testing"
)

func TestValidateCronExpression(t *testing.T) {
	tests := []struct {
		name        string
		cronExpr    string
		expectError bool
	}{
		// Valid expressions
		{"valid basic", "0 0 * * *", false},
		{"valid with ranges", "0-30 8-17 * * 1-5", false},
		{"valid with lists", "0,15,30,45 * * * *", false},
		{"valid with steps", "*/5 * * * *", false},
		{"valid complex", "0,15,30,45 8-17 * * 1-5", false},
		{"valid 6-field with seconds", "0 0 0 * * *", false},
		{"valid wildcard", "* * * * *", false},

		// Invalid expressions
		{"empty expression", "", true},
		{"too few fields", "0 0 *", true},
		{"too many fields", "0 0 * * * * *", true},
		{"invalid minute", "60 * * * *", true},
		{"invalid hour", "0 24 * * *", true},
		{"invalid day", "0 0 32 * *", true},
		{"invalid month", "0 0 * 13 *", true},
		{"invalid weekday", "0 0 * * 8", true},
		{"invalid range", "0 0 5-3 * *", true},
		{"invalid character", "0 0 * * X", true},
		{"negative value", "0 0 -1 * *", true},
		{"invalid range format", "0 0 1-2-3 * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCronExpression(tt.cronExpr)
			if tt.expectError && err == nil {
				t.Errorf("expected error for %q, but got none", tt.cronExpr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.cronExpr, err)
			}
		})
	}
}
