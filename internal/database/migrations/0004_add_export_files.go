package migrations

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
		schemaPath := filepath.Join("internal", "database", "schema", "0004_pb_schema.json")
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
		collectionsToDelete := []string{"export_files"}

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
