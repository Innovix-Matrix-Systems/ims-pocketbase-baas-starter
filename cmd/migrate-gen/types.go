package main

import (
	"errors"
	"fmt"
)

// CLIConfig holds the configuration for the CLI command
type CLIConfig struct {
	MigrationName string
	OutputDir     string
	Verbose       bool
}

// MigrationTemplate holds the data needed to generate a migration file
type MigrationTemplate struct {
	Number     string
	Name       string
	SchemaFile string
}

// MigrationError represents different types of errors that can occur
type MigrationError struct {
	Type    string
	Message string
	Cause   error
}

func (e *MigrationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error type constants
const (
	ErrorTypeValidation = "validation"
	ErrorTypeFileSystem = "filesystem"
	ErrorTypeTemplate   = "template"
)

// Common errors
var (
	ErrMissingMigrationName = errors.New("migration name is required")
	ErrInvalidMigrationName = errors.New("migration name contains invalid characters")
	ErrMigrationsDir        = errors.New("migrations directory not found")
	ErrFileExists           = errors.New("migration file already exists")
)
