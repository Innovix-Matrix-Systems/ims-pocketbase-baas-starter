package migration

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirectoryExists(t *testing.T) {
	// Test case 1: Existing directory
	t.Run("ExistingDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		err := EnsureDirectoryExists(tempDir)
		if err != nil {
			t.Errorf("Expected no error for existing directory, got: %v", err)
		}
	})

	// Test case 2: Non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentDir := filepath.Join(tempDir, "nonexistent")

		err := EnsureDirectoryExists(nonExistentDir)
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}

		if !errors.Is(err, ErrMigrationsDir) {
			t.Errorf("Expected error to wrap ErrMigrationsDir, got: %v", err)
		}
	})
}

func TestCheckFileExists(t *testing.T) {
	tempDir := t.TempDir()

	// Test case 1: Existing file
	t.Run("ExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing.go")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		exists := CheckFileExists(testFile)
		if !exists {
			t.Error("Expected file to exist")
		}
	})

	// Test case 2: Non-existent file
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "nonexistent.go")
		exists := CheckFileExists(nonExistentFile)
		if exists {
			t.Error("Expected file to not exist")
		}
	})
}

func TestWriteMigrationFile(t *testing.T) {
	tempDir := t.TempDir()

	// Test case 1: Write to new file
	t.Run("WriteNewFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "new_migration.go")
		content := "package migrations\n\n// Test migration"

		err := WriteMigrationFile(testFile, content)
		if err != nil {
			t.Errorf("Expected no error writing new file, got: %v", err)
		}

		// Verify file was created with correct content
		writtenContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(writtenContent) != content {
			t.Errorf("Expected content %q, got %q", content, string(writtenContent))
		}
	})

	// Test case 2: Attempt to overwrite existing file
	t.Run("OverwriteExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "existing_migration.go")

		// Create existing file
		if err := os.WriteFile(testFile, []byte("existing content"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Attempt to write to existing file
		err := WriteMigrationFile(testFile, "new content")
		if err == nil {
			t.Error("Expected error when trying to overwrite existing file")
		}

		if !errors.Is(err, ErrFileExists) {
			t.Errorf("Expected error to wrap ErrFileExists, got: %v", err)
		}
	})

	// Test case 3: Write to non-existent directory
	t.Run("NonExistentDirectory", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		testFile := filepath.Join(nonExistentDir, "migration.go")

		err := WriteMigrationFile(testFile, "content")
		if err == nil {
			t.Error("Expected error when writing to non-existent directory")
		}

		if !errors.Is(err, ErrMigrationsDir) {
			t.Errorf("Expected error to wrap ErrMigrationsDir, got: %v", err)
		}
	})
}

func TestGenerateMigrationFilePath(t *testing.T) {
	testCases := []struct {
		number   int
		name     string
		expected string
	}{
		{1, "init", filepath.Join(MigrationsDir, "0001_init.go")},
		{10, "add-users", filepath.Join(MigrationsDir, "0010_add-users.go")},
		{100, "complex_migration", filepath.Join(MigrationsDir, "0100_complex_migration.go")},
		{9999, "last", filepath.Join(MigrationsDir, "9999_last.go")},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := GenerateMigrationFilePath(tc.number, tc.name)
			if result != tc.expected {
				t.Errorf("GenerateMigrationFilePath(%d, %s) = %s, expected %s",
					tc.number, tc.name, result, tc.expected)
			}
		})
	}
}

func TestGenerateSchemaFilePath(t *testing.T) {
	testCases := []struct {
		number   int
		expected string
	}{
		{1, filepath.Join(SchemaDir, "0001_pb_schema.json")},
		{10, filepath.Join(SchemaDir, "0010_pb_schema.json")},
		{100, filepath.Join(SchemaDir, "0100_pb_schema.json")},
		{9999, filepath.Join(SchemaDir, "9999_pb_schema.json")},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := GenerateSchemaFilePath(tc.number)
			if result != tc.expected {
				t.Errorf("GenerateSchemaFilePath(%d) = %s, expected %s",
					tc.number, result, tc.expected)
			}
		})
	}
}
