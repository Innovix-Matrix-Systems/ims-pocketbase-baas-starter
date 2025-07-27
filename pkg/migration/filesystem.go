package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Common errors
var (
	ErrFileExists = errors.New("migration file already exists")
)

// EnsureDirectoryExists checks if a directory exists and creates it if it doesn't
func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s (%w)", path, ErrMigrationsDir)
	}
	return nil
}

// CheckFileExists checks if a file exists at the given path
func CheckFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// WriteMigrationFile writes the migration content to the specified file path
func WriteMigrationFile(path, content string) error {
	// Check if file already exists
	if CheckFileExists(path) {
		return fmt.Errorf("migration file already exists: %s (%w)", path, ErrFileExists)
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := EnsureDirectoryExists(dir); err != nil {
		return err
	}

	// Write the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write migration file %s: %w", path, err)
	}

	return nil
}

// GenerateMigrationFilePath generates the full file path for a migration
func GenerateMigrationFilePath(number int, name string) string {
	filename := fmt.Sprintf(MigrationFileFormat, number, name)
	return filepath.Join(MigrationsDir, filename)
}

// GenerateSchemaFilePath generates the expected schema file path for a migration
func GenerateSchemaFilePath(number int) string {
	filename := fmt.Sprintf(SchemaFileFormat, number)
	return filepath.Join(SchemaDir, filename)
}
