package migration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// MigrationInfo represents information about an existing migration
type MigrationInfo struct {
	Number int
	Name   string
	Path   string
}

// Configuration constants
const (
	MigrationsDir        = "internal/database/migrations"
	SchemaDir            = "internal/database/schema"
	MigrationFilePattern = `^\d{4}_.*\.go$`
	SchemaFileFormat     = "%04d_pb_schema.json"
	MigrationFileFormat  = "%04d_%s.go"
)

// Common errors
var (
	ErrMigrationsDir = errors.New("migrations directory not found")
)

// ScanExistingMigrations scans the migrations directory for existing migration files
func ScanExistingMigrations(dir string) ([]MigrationInfo, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %s (%w)", dir, ErrMigrationsDir)
	}

	// Read directory contents
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []MigrationInfo
	migrationPattern := regexp.MustCompile(MigrationFilePattern)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !migrationPattern.MatchString(filename) {
			continue
		}

		// Extract migration number and name from filename
		number, name, err := parseMigrationFilename(filename)
		if err != nil {
			// Skip malformed migration files
			continue
		}

		migrations = append(migrations, MigrationInfo{
			Number: number,
			Name:   name,
			Path:   filepath.Join(dir, filename),
		})
	}

	// Sort migrations by number
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Number < migrations[j].Number
	})

	return migrations, nil
}

// GetNextMigrationNumber determines the next migration number based on existing migrations
func GetNextMigrationNumber(migrations []MigrationInfo) int {
	if len(migrations) == 0 {
		return 1
	}

	// Find the highest migration number
	highest := 0
	for _, migration := range migrations {
		if migration.Number > highest {
			highest = migration.Number
		}
	}

	return highest + 1
}

// parseMigrationFilename extracts the migration number and name from a filename
// Expected format: 0001_migration_name.go
func parseMigrationFilename(filename string) (int, string, error) {
	// Check if filename ends with .go
	if !strings.HasSuffix(filename, ".go") {
		return 0, "", fmt.Errorf("migration file must have .go extension: %s", filename)
	}

	// Remove .go extension
	nameWithoutExt := filename[:len(filename)-3]

	// Split by underscore
	parts := regexp.MustCompile(`^(\d{4})_(.+)$`).FindStringSubmatch(nameWithoutExt)
	if len(parts) != 3 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	// Parse number
	number, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, "", fmt.Errorf("invalid migration number in filename %s: %w", filename, err)
	}

	return number, parts[2], nil
}

// FormatMigrationNumber formats a migration number as a 4-digit zero-padded string
func FormatMigrationNumber(number int) string {
	return fmt.Sprintf("%04d", number)
}
