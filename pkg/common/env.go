package common

import (
	"os"
	"strconv"
)

// GetEnv retrieves an environment variable as a string.
// If the environment variable exists and is not empty, it returns the variable's value.
// Otherwise, it returns the provided defaultValue.
//
// Parameters:
//   - key: the environment variable name to look up
//   - defaultValue: the value to return if the environment variable is not set or empty
//
// Returns the string value of the environment variable or the default value.
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an environment variable as an integer.
// If the environment variable exists and can be converted to an integer,
// it returns the converted value. Otherwise, it returns the provided defaultValue.
//
// Parameters:
//   - key: the environment variable name to look up
//   - defaultValue: the value to return if the environment variable is not set or invalid
//
// Returns the integer value of the environment variable or the default value.
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool retrieves an environment variable as a boolean.
// If the environment variable exists and can be converted to a boolean,
// it returns the converted value. Otherwise, it returns the provided defaultValue.
//
// Valid boolean values are: "1", "t", "T", "TRUE", "true", "True", "0", "f", "F", "FALSE", "false", "False"
// as defined by Go's strconv.ParseBool function.
//
// Parameters:
//   - key: the environment variable name to look up
//   - defaultValue: the value to return if the environment variable is not set or invalid
//
// Returns the boolean value of the environment variable or the default value.
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
