package main

import (
	"strings"
	"testing"
)

func TestCreateMigrationTemplate(t *testing.T) {
	testCases := []struct {
		number       int
		name         string
		expectedNum  string
		expectedFile string
	}{
		{1, "init", "0001", "0001_pb_schema.json"},
		{10, "add-users", "0010", "0010_pb_schema.json"},
		{100, "complex_migration", "0100", "0100_pb_schema.json"},
		{9999, "last", "9999", "9999_pb_schema.json"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			template := CreateMigrationTemplate(tc.number, tc.name)

			if template.Number != tc.expectedNum {
				t.Errorf("Expected number %s, got: %s", tc.expectedNum, template.Number)
			}

			if template.Name != tc.name {
				t.Errorf("Expected name %s, got: %s", tc.name, template.Name)
			}

			if template.SchemaFile != tc.expectedFile {
				t.Errorf("Expected schema file %s, got: %s", tc.expectedFile, template.SchemaFile)
			}
		})
	}
}

func TestGenerateMigrationContent(t *testing.T) {
	// Test case 1: Valid template data
	t.Run("ValidTemplate", func(t *testing.T) {
		template := MigrationTemplate{
			Number:     "0001",
			Name:       "test_migration",
			SchemaFile: "0001_pb_schema.json",
		}

		content, err := GenerateMigrationContent(template)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that content contains expected elements
		expectedElements := []string{
			"package migrations",
			"import (",
			"\"encoding/json\"",
			"\"fmt\"",
			"\"os\"",
			"\"path/filepath\"",
			"\"github.com/pocketbase/pocketbase/core\"",
			"m \"github.com/pocketbase/pocketbase/migrations\"",
			"func init() {",
			"m.Register(",
			"schemaPath := filepath.Join(\"internal\", \"database\", \"schema\", \"0001_pb_schema.json\")",
			"os.ReadFile(schemaPath)",
			"json.Unmarshal(schemaData, &collections)",
			"json.Marshal(collections)",
			"app.ImportCollectionsByMarshaledJSON(collectionsData, false)",
			"TODO: Add any data seeding specific to these collections",
			"collectionsToDelete := []string{",
			"TODO: Add collection names to delete during rollback",
			"app.FindCollectionByNameOrId(collectionName)",
			"app.Delete(collection)",
		}

		for _, element := range expectedElements {
			if !strings.Contains(content, element) {
				t.Errorf("Expected content to contain: %s", element)
			}
		}

		// Check that the schema file path is correctly templated
		if !strings.Contains(content, "0001_pb_schema.json") {
			t.Error("Expected content to contain templated schema file path")
		}
	})

	// Test case 2: Different template values
	t.Run("DifferentTemplateValues", func(t *testing.T) {
		template := MigrationTemplate{
			Number:     "0042",
			Name:       "add_user_profiles",
			SchemaFile: "0042_pb_schema.json",
		}

		content, err := GenerateMigrationContent(template)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Check that the correct schema file is referenced
		if !strings.Contains(content, "0042_pb_schema.json") {
			t.Error("Expected content to contain correct schema file path")
		}

		// Ensure it's a valid Go file structure
		if !strings.HasPrefix(content, "package migrations") {
			t.Error("Expected content to start with package declaration")
		}

		if !strings.Contains(content, "func init() {") {
			t.Error("Expected content to contain init function")
		}
	})

	// Test case 3: Edge cases
	t.Run("EdgeCases", func(t *testing.T) {
		edgeCases := []MigrationTemplate{
			{Number: "0001", Name: "", SchemaFile: "0001_pb_schema.json"},
			{Number: "9999", Name: "very_long_migration_name_with_many_underscores", SchemaFile: "9999_pb_schema.json"},
			{Number: "0001", Name: "123", SchemaFile: "0001_pb_schema.json"},
		}

		for i, template := range edgeCases {
			t.Run("", func(t *testing.T) {
				content, err := GenerateMigrationContent(template)
				if err != nil {
					t.Fatalf("Case %d: Expected no error, got: %v", i, err)
				}

				// Basic validation that it's still a valid template
				if !strings.Contains(content, "package migrations") {
					t.Errorf("Case %d: Expected valid package declaration", i)
				}

				if !strings.Contains(content, template.SchemaFile) {
					t.Errorf("Case %d: Expected schema file %s in content", i, template.SchemaFile)
				}
			})
		}
	})
}

func TestMigrationTemplateStructure(t *testing.T) {
	// Test that the generated content has the correct structure
	template := MigrationTemplate{
		Number:     "0001",
		Name:       "test",
		SchemaFile: "0001_pb_schema.json",
	}

	content, err := GenerateMigrationContent(template)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check for proper Go syntax elements
	lines := strings.Split(content, "\n")

	// Should start with package declaration
	if !strings.HasPrefix(lines[0], "package migrations") {
		t.Error("Expected first line to be package declaration")
	}

	// Should have import block
	hasImport := false
	for _, line := range lines {
		if strings.Contains(line, "import (") {
			hasImport = true
			break
		}
	}
	if !hasImport {
		t.Error("Expected import block")
	}

	// Should have init function
	hasInit := false
	for _, line := range lines {
		if strings.Contains(line, "func init() {") {
			hasInit = true
			break
		}
	}
	if !hasInit {
		t.Error("Expected init function")
	}

	// Should have m.Register call
	hasRegister := false
	for _, line := range lines {
		if strings.Contains(line, "m.Register(") {
			hasRegister = true
			break
		}
	}
	if !hasRegister {
		t.Error("Expected m.Register call")
	}

	// Should have both forward and rollback functions
	hasForward := strings.Contains(content, "// Forward migration")
	hasRollback := strings.Contains(content, "// Rollback migration")

	if !hasForward {
		t.Error("Expected forward migration comment")
	}
	if !hasRollback {
		t.Error("Expected rollback migration comment")
	}
}
