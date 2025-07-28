package main

import (
	"testing"

	"ims-pocketbase-baas-starter/pkg/migration"
)

func TestParseArgs(t *testing.T) {
	// Test case 1: Valid arguments
	t.Run("ValidArguments", func(t *testing.T) {
		args := []string{"migrate-gen", "add_user_profiles"}
		config, err := ParseArgs(args)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if config.MigrationName != "add-user-profiles" {
			t.Errorf("Expected migration name 'add-user-profiles', got: %s", config.MigrationName)
		}

		if config.OutputDir != migration.MigrationsDir {
			t.Errorf("Expected output dir %s, got: %s", migration.MigrationsDir, config.OutputDir)
		}

		if config.Verbose != false {
			t.Errorf("Expected verbose false, got: %t", config.Verbose)
		}
	})

	// Test case 2: Missing migration name
	t.Run("MissingMigrationName", func(t *testing.T) {
		args := []string{"migrate-gen"}
		_, err := ParseArgs(args)
		if err == nil {
			t.Fatal("Expected error for missing migration name")
		}

		migrationErr, ok := err.(*MigrationError)
		if !ok {
			t.Fatalf("Expected MigrationError, got: %T", err)
		}

		if migrationErr.Type != ErrorTypeValidation {
			t.Errorf("Expected error type %s, got: %s", ErrorTypeValidation, migrationErr.Type)
		}
	})

	// Test case 3: Empty arguments
	t.Run("EmptyArguments", func(t *testing.T) {
		args := []string{}
		_, err := ParseArgs(args)
		if err == nil {
			t.Fatal("Expected error for empty arguments")
		}
	})

	// Test case 4: Invalid migration name
	t.Run("InvalidMigrationName", func(t *testing.T) {
		args := []string{"migrate-gen", "invalid@name!"}
		_, err := ParseArgs(args)
		if err == nil {
			t.Fatal("Expected error for invalid migration name")
		}

		migrationErr, ok := err.(*MigrationError)
		if !ok {
			t.Fatalf("Expected MigrationError, got: %T", err)
		}

		if migrationErr.Type != ErrorTypeValidation {
			t.Errorf("Expected error type %s, got: %s", ErrorTypeValidation, migrationErr.Type)
		}
	})
}

func TestValidateMigrationName(t *testing.T) {
	// Test case 1: Valid names
	validNames := []string{
		"add_users",
		"add-users",
		"AddUsers",
		"add123",
		"migration_with_numbers_123",
		"simple",
		"a",
		"123",
	}

	for _, name := range validNames {
		t.Run("Valid_"+name, func(t *testing.T) {
			err := ValidateMigrationName(name)
			if err != nil {
				t.Errorf("Expected no error for valid name %s, got: %v", name, err)
			}
		})
	}

	// Test case 2: Invalid names
	invalidNames := []string{
		"",             // Empty
		"add@users",    // Special character @
		"add users",    // Space
		"add.users",    // Dot
		"add/users",    // Slash
		"add\\users",   // Backslash
		"add!users",    // Exclamation
		"add#users",    // Hash
		"add$users",    // Dollar
		"add%users",    // Percent
		"add^users",    // Caret
		"add&users",    // Ampersand
		"add*users",    // Asterisk
		"add(users)",   // Parentheses
		"add[users]",   // Brackets
		"add{users}",   // Braces
		"add|users",    // Pipe
		"add;users",    // Semicolon
		"add:users",    // Colon
		"add'users",    // Single quote
		"add\"users\"", // Double quote
		"add<users>",   // Angle brackets
		"add,users",    // Comma
		"add?users",    // Question mark
		"add~users",    // Tilde
		"add`users",    // Backtick
	}

	for _, name := range invalidNames {
		t.Run("Invalid_"+name, func(t *testing.T) {
			err := ValidateMigrationName(name)
			if err == nil {
				t.Errorf("Expected error for invalid name %s", name)
			}

			migrationErr, ok := err.(*MigrationError)
			if !ok {
				t.Fatalf("Expected MigrationError, got: %T", err)
			}

			if migrationErr.Type != ErrorTypeValidation {
				t.Errorf("Expected error type %s, got: %s", ErrorTypeValidation, migrationErr.Type)
			}
		})
	}
}

func TestSanitizeMigrationName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"AddUsers", "addusers"},
		{"add_users", "add-users"},
		{"ADD_USERS", "add-users"},
		{"add__users", "add-users"},
		{"add___users", "add-users"},
		{"_add_users_", "add-users"},
		{"Add_User_Profiles", "add-user-profiles"},
		{"create-audit-logs", "create-audit-logs"},
		{"COMPLEX_Migration_Name", "complex-migration-name"},
		{"simple", "simple"},
		{"123", "123"},
		{"add123users", "add123users"},
		{"", ""},
		{"---", ""},
		{"_-_-_", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := SanitizeMigrationName(tc.input)
			if result != tc.expected {
				t.Errorf("SanitizeMigrationName(%s) = %s, expected %s", tc.input, result, tc.expected)
			}
		})
	}
}
