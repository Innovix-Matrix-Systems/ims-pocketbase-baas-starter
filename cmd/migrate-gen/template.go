package main

import (
	"fmt"
	"ims-pocketbase-baas-starter/pkg/migration"
	"strings"
	"text/template"
)

// migrationTemplate is the template for generating migration files
const migrationTemplate = `package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Forward migration
		schemaPath := filepath.Join("internal", "database", "schema", "{{.SchemaFile}}")
		schemaData, err := os.ReadFile(schemaPath)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		var collections []interface{}
		if err := json.Unmarshal(schemaData, &collections); err != nil {
			return fmt.Errorf("failed to parse schema JSON: %w", err)
		}

		collectionsData, err := json.Marshal(collections)
		if err != nil {
			return fmt.Errorf("failed to marshal collections: %w", err)
		}

		if err := app.ImportCollectionsByMarshaledJSON(collectionsData, false); err != nil {
			return fmt.Errorf("failed to import collections: %w", err)
		}

		// TODO: Add any data seeding specific to these collections

		return nil
	}, func(app core.App) error {
		// Rollback migration
		collectionsToDelete := []string{
			// TODO: Add collection names to delete during rollback
		}

		for _, collectionName := range collectionsToDelete {
			collection, err := app.FindCollectionByNameOrId(collectionName)
			if err != nil {
				continue // Collection might not exist
			}

			if err := app.Delete(collection); err != nil {
				return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
			}
		}

		return nil
	})
}
`

// GenerateMigrationContent generates the content for a migration file
func GenerateMigrationContent(data MigrationTemplate) (string, error) {
	tmpl, err := template.New("migration").Parse(migrationTemplate)
	if err != nil {
		return "", &MigrationError{
			Type:    ErrorTypeTemplate,
			Message: "failed to parse migration template",
			Cause:   err,
		}
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", &MigrationError{
			Type:    ErrorTypeTemplate,
			Message: "failed to execute migration template",
			Cause:   err,
		}
	}

	return buf.String(), nil
}

// CreateMigrationTemplate creates a MigrationTemplate with the given parameters
func CreateMigrationTemplate(number int, name string) MigrationTemplate {
	numberStr := migration.FormatMigrationNumber(number)
	schemaFile := fmt.Sprintf(migration.SchemaFileFormat, number)

	return MigrationTemplate{
		Number:     numberStr,
		Name:       name,
		SchemaFile: schemaFile,
	}
}
