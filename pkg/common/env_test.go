package common

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		expected     string
		setEnv       bool
	}{
		{
			name:         "environment variable exists with value",
			envKey:       "TEST_STRING_VALID",
			envValue:     "hello_world",
			defaultValue: "default",
			expected:     "hello_world",
			setEnv:       true,
		},
		{
			name:         "environment variable not set",
			envKey:       "TEST_STRING_NOT_SET",
			envValue:     "",
			defaultValue: "fallback_value",
			expected:     "fallback_value",
			setEnv:       false,
		},
		{
			name:         "environment variable set to empty string",
			envKey:       "TEST_STRING_EMPTY",
			envValue:     "",
			defaultValue: "default_empty",
			expected:     "default_empty",
			setEnv:       true,
		},
		{
			name:         "environment variable with spaces",
			envKey:       "TEST_STRING_SPACES",
			envValue:     "  value with spaces  ",
			defaultValue: "default",
			expected:     "  value with spaces  ",
			setEnv:       true,
		},
		{
			name:         "environment variable with special characters",
			envKey:       "TEST_STRING_SPECIAL",
			envValue:     "value@#$%^&*()",
			defaultValue: "default",
			expected:     "value@#$%^&*()",
			setEnv:       true,
		},
		{
			name:         "empty default value",
			envKey:       "TEST_STRING_EMPTY_DEFAULT",
			envValue:     "actual_value",
			defaultValue: "",
			expected:     "actual_value",
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before test
			os.Unsetenv(tt.envKey)

			// Set environment variable if needed
			if tt.setEnv && tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey) // Clean up after test
			}

			result := GetEnv(tt.envKey, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetEnv(%q, %q) = %q, want %q",
					tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue int
		expected     int
		setEnv       bool
	}{
		{
			name:         "valid integer environment variable",
			envKey:       "TEST_INT_VALID",
			envValue:     "42",
			defaultValue: 10,
			expected:     42,
			setEnv:       true,
		},
		{
			name:         "environment variable not set",
			envKey:       "TEST_INT_NOT_SET",
			envValue:     "",
			defaultValue: 100,
			expected:     100,
			setEnv:       false,
		},
		{
			name:         "invalid integer environment variable",
			envKey:       "TEST_INT_INVALID",
			envValue:     "not_a_number",
			defaultValue: 50,
			expected:     50,
			setEnv:       true,
		},
		{
			name:         "empty string environment variable",
			envKey:       "TEST_INT_EMPTY",
			envValue:     "",
			defaultValue: 25,
			expected:     25,
			setEnv:       true,
		},
		{
			name:         "negative integer environment variable",
			envKey:       "TEST_INT_NEGATIVE",
			envValue:     "-15",
			defaultValue: 5,
			expected:     -15,
			setEnv:       true,
		},
		{
			name:         "zero value environment variable",
			envKey:       "TEST_INT_ZERO",
			envValue:     "0",
			defaultValue: 99,
			expected:     0,
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before test
			os.Unsetenv(tt.envKey)

			// Set environment variable if needed
			if tt.setEnv && tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey) // Clean up after test
			}

			result := GetEnvInt(tt.envKey, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetEnvInt(%q, %d) = %d, want %d",
					tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue bool
		expected     bool
		setEnv       bool
	}{
		{
			name:         "true string value",
			envKey:       "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "false string value",
			envKey:       "TEST_BOOL_FALSE",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "1 numeric value",
			envKey:       "TEST_BOOL_ONE",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "0 numeric value",
			envKey:       "TEST_BOOL_ZERO",
			envValue:     "0",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "uppercase TRUE",
			envKey:       "TEST_BOOL_UPPER_TRUE",
			envValue:     "TRUE",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "uppercase FALSE",
			envKey:       "TEST_BOOL_UPPER_FALSE",
			envValue:     "FALSE",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "single character t",
			envKey:       "TEST_BOOL_T",
			envValue:     "t",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "single character f",
			envKey:       "TEST_BOOL_F",
			envValue:     "f",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "environment variable not set",
			envKey:       "TEST_BOOL_NOT_SET",
			envValue:     "",
			defaultValue: true,
			expected:     true,
			setEnv:       false,
		},
		{
			name:         "invalid boolean value",
			envKey:       "TEST_BOOL_INVALID",
			envValue:     "maybe",
			defaultValue: false,
			expected:     false,
			setEnv:       true,
		},
		{
			name:         "empty string environment variable",
			envKey:       "TEST_BOOL_EMPTY",
			envValue:     "",
			defaultValue: true,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "mixed case True",
			envKey:       "TEST_BOOL_MIXED_TRUE",
			envValue:     "True",
			defaultValue: false,
			expected:     true,
			setEnv:       true,
		},
		{
			name:         "mixed case False",
			envKey:       "TEST_BOOL_MIXED_FALSE",
			envValue:     "False",
			defaultValue: true,
			expected:     false,
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment variable before test
			os.Unsetenv(tt.envKey)

			// Set environment variable if needed
			if tt.setEnv && tt.envValue != "" {
				os.Setenv(tt.envKey, tt.envValue)
				defer os.Unsetenv(tt.envKey) // Clean up after test
			}

			result := GetEnvBool(tt.envKey, tt.defaultValue)

			if result != tt.expected {
				t.Errorf("GetEnvBool(%q, %t) = %t, want %t",
					tt.envKey, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
