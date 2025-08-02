package migration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestScanExistingMigrations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test case 1: Empty directory
	t.Run("EmptyDirectory", func(t *testing.T) {
		migrations, err := ScanExistingMigrations(tempDir)
		if err != nil {
			t.Fatalf("Expected no error for empty directory, got: %v", err)
		}
		if len(migrations) != 0 {
			t.Fatalf("Expected 0 migrations, got: %d", len(migrations))
		}
	})

	// Test case 2: Directory with valid migration files
	t.Run("ValidMigrationFiles", func(t *testing.T) {
		// Create test migration files
		testFiles := []string{
			"0001_init.go",
			"0002_add_users.go",
			"0005_add_settings.go",
		}

		for _, filename := range testFiles {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte("package migrations"), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		migrations, err := ScanExistingMigrations(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(migrations) != 3 {
			t.Fatalf("Expected 3 migrations, got: %d", len(migrations))
		}

		// Check if migrations are sorted by number
		expectedNumbers := []int{1, 2, 5}
		for i, migration := range migrations {
			if migration.Number != expectedNumbers[i] {
				t.Errorf("Expected migration number %d, got: %d", expectedNumbers[i], migration.Number)
			}
		}

		// Check specific migration details
		if migrations[0].Name != "init" {
			t.Errorf("Expected migration name 'init', got: %s", migrations[0].Name)
		}
		if migrations[1].Name != "add_users" {
			t.Errorf("Expected migration name 'add_users', got: %s", migrations[1].Name)
		}
	})

	// Test case 3: Directory with invalid files (should be ignored)
	t.Run("InvalidFiles", func(t *testing.T) {
		// Clean the temp directory
		os.RemoveAll(tempDir)
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}

		// Create invalid files that should be ignored
		invalidFiles := []string{
			"README.md",
			"migration.txt",
			"001_invalid.go", // Wrong number format
			"0001_test",      // Missing .go extension
		}

		for _, filename := range invalidFiles {
			filePath := filepath.Join(tempDir, filename)
			if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file %s: %v", filename, err)
			}
		}

		// Add one valid file
		validFile := filepath.Join(tempDir, "0001_valid.go")
		if err := os.WriteFile(validFile, []byte("package migrations"), 0644); err != nil {
			t.Fatalf("Failed to create valid test file: %v", err)
		}

		migrations, err := ScanExistingMigrations(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(migrations) != 1 {
			t.Fatalf("Expected 1 migration, got: %d", len(migrations))
		}

		if migrations[0].Name != "valid" {
			t.Errorf("Expected migration name 'valid', got: %s", migrations[0].Name)
		}
	})

	// Test case 4: Non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		_, err := ScanExistingMigrations(nonExistentDir)
		if err == nil {
			t.Fatal("Expected error for non-existent directory")
		}

		if !errors.Is(err, ErrMigrationsDir) {
			t.Errorf("Expected error to wrap ErrMigrationsDir, got: %v", err)
		}
	})
}

func TestGetNextMigrationNumber(t *testing.T) {
	// Test case 1: Empty migrations list
	t.Run("EmptyMigrations", func(t *testing.T) {
		migrations := []MigrationInfo{}
		nextNumber := GetNextMigrationNumber(migrations)
		if nextNumber != 1 {
			t.Errorf("Expected next number 1, got: %d", nextNumber)
		}
	})

	// Test case 2: Sequential migrations
	t.Run("SequentialMigrations", func(t *testing.T) {
		migrations := []MigrationInfo{
			{Number: 1, Name: "init"},
			{Number: 2, Name: "add_users"},
			{Number: 3, Name: "add_settings"},
		}
		nextNumber := GetNextMigrationNumber(migrations)
		if nextNumber != 4 {
			t.Errorf("Expected next number 4, got: %d", nextNumber)
		}
	})

	// Test case 3: Non-sequential migrations (gaps)
	t.Run("NonSequentialMigrations", func(t *testing.T) {
		migrations := []MigrationInfo{
			{Number: 1, Name: "init"},
			{Number: 3, Name: "add_users"},
			{Number: 7, Name: "add_settings"},
		}
		nextNumber := GetNextMigrationNumber(migrations)
		if nextNumber != 8 {
			t.Errorf("Expected next number 8, got: %d", nextNumber)
		}
	})

	// Test case 4: Unordered migrations
	t.Run("UnorderedMigrations", func(t *testing.T) {
		migrations := []MigrationInfo{
			{Number: 5, Name: "add_settings"},
			{Number: 1, Name: "init"},
			{Number: 3, Name: "add_users"},
		}
		nextNumber := GetNextMigrationNumber(migrations)
		if nextNumber != 6 {
			t.Errorf("Expected next number 6, got: %d", nextNumber)
		}
	})
}

func TestFormatMigrationNumber(t *testing.T) {
	testCases := []struct {
		input    int
		expected string
	}{
		{1, "0001"},
		{10, "0010"},
		{100, "0100"},
		{1000, "1000"},
		{9999, "9999"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := FormatMigrationNumber(tc.input)
			if result != tc.expected {
				t.Errorf("FormatMigrationNumber(%d) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}

func TestParseMigrationFilename(t *testing.T) {
	testCases := []struct {
		filename     string
		expectedNum  int
		expectedName string
		shouldError  bool
	}{
		{"0001_init.go", 1, "init", false},
		{"0002_add_users.go", 2, "add_users", false},
		{"0010_complex_migration_name.go", 10, "complex_migration_name", false},
		{"9999_last_migration.go", 9999, "last_migration", false},
		{"001_invalid.go", 0, "", true}, // Wrong number format
		{"0001_test", 0, "", true},      // Missing .go extension
		{"invalid.go", 0, "", true},     // No number
		{"0001_.go", 0, "", true},       // Empty name after underscore (invalid)
		{"0001_a.go", 1, "a", false},    // Single character name (valid)
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			num, name, err := parseMigrationFilename(tc.filename)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for filename %s, but got none", tc.filename)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for filename %s: %v", tc.filename, err)
				return
			}

			if num != tc.expectedNum {
				t.Errorf("Expected number %d, got %d", tc.expectedNum, num)
			}

			if name != tc.expectedName {
				t.Errorf("Expected name %s, got %s", tc.expectedName, name)
			}
		})
	}
}
