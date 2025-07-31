package main

import (
	"regexp"
	"strings"

	"ims-pocketbase-baas-starter/pkg/migration"
)

// ParseArgs parses command-line arguments and returns CLI configuration
func ParseArgs(args []string) (*CLIConfig, error) {
	if len(args) < 2 {
		return nil, &MigrationError{
			Type:    ErrorTypeValidation,
			Message: "migration name is required",
			Cause:   ErrMissingMigrationName,
		}
	}

	migrationName := args[1]
	if err := ValidateMigrationName(migrationName); err != nil {
		return nil, err
	}

	// Sanitize migration name to snake_case
	sanitizedName := SanitizeMigrationName(migrationName)

	return &CLIConfig{
		MigrationName: sanitizedName,
		OutputDir:     migration.MigrationsDir,
		Verbose:       false, // TODO: Add verbose flag support
	}, nil
}

// ValidateMigrationName validates that the migration name is acceptable
func ValidateMigrationName(name string) error {
	if name == "" {
		return &MigrationError{
			Type:    ErrorTypeValidation,
			Message: "migration name cannot be empty",
			Cause:   ErrMissingMigrationName,
		}
	}

	// Check for invalid characters (allow letters, numbers, hyphens, underscores)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(name) {
		return &MigrationError{
			Type:    ErrorTypeValidation,
			Message: "migration name can only contain letters, numbers, hyphens, and underscores",
			Cause:   ErrInvalidMigrationName,
		}
	}

	return nil
}

// SanitizeMigrationName converts migration name to snake_case
func SanitizeMigrationName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace hyphens with underscores (normalize to snake_case)
	name = strings.ReplaceAll(name, "-", "_")

	// Replace multiple consecutive underscores with single underscore
	multiUnderscore := regexp.MustCompile(`_+`)
	name = multiUnderscore.ReplaceAllString(name, "_")

	// Remove leading/trailing underscores
	name = strings.Trim(name, "_")

	return name
}

// ShowUsage displays usage information for the CLI
func ShowUsage() {
	usage := `Usage: migrate-gen <migration_name>

Generate a new migration file for the PocketBase project.

Arguments:
  migration_name    Name for the migration (will be sanitized to snake_case)

Examples:
  migrate-gen add_user_profiles
  migrate-gen create_audit_logs
  migrate-gen AddNotificationSystem

The generated migration file will be placed in internal/database/migrations/
with the next sequential number (e.g., 0003_add_user_profiles.go).
`
	println(usage)
}
