package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	"ims-pocketbase-baas-starter/internal/database/seeders"
	"ims-pocketbase-baas-starter/pkg/common"
)

func init() {
	m.Register(func(app core.App) error {
		// 1. Read and parse the schema JSON
		schemaPath := filepath.Join("internal", "database", "schema", "0001_pb_schema.json")
		schemaData, err := os.ReadFile(schemaPath)
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}

		// 2. Parse the collections array directly
		var collections []any
		if err := json.Unmarshal(schemaData, &collections); err != nil {
			return fmt.Errorf("failed to parse schema JSON: %w", err)
		}

		// 3. Convert back to JSON for import
		collectionsData, err := json.Marshal(collections)
		if err != nil {
			return fmt.Errorf("failed to marshal collections: %w", err)
		}

		// 4. Import collections
		if err := app.ImportCollectionsByMarshaledJSON(collectionsData, false); err != nil {
			return fmt.Errorf("failed to import collections: %w", err)
		}

		// 5. Configure app settings using environment variables
		settings := app.Settings()

		// Set basic meta information from environment variables
		settings.Meta.AppName = common.GetEnv("APP_NAME", "IMS_PocketBase_App")
		settings.Meta.AppURL = common.GetEnv("APP_URL", "http://localhost:8090")

		// Configure logs retention
		settings.Logs.MaxDays = common.GetEnvInt("LOGS_MAX_DAYS", 7)

		// Configure SMTP settings from environment variables
		settings.SMTP.Enabled = common.GetEnvBool("SMTP_ENABLED", false)
		if settings.SMTP.Enabled {
			settings.SMTP.Host = common.GetEnv("SMTP_HOST", "")
			settings.SMTP.Port = common.GetEnvInt("SMTP_PORT", 587)
			settings.SMTP.Username = common.GetEnv("SMTP_USERNAME", "")
			settings.SMTP.Password = common.GetEnv("SMTP_PASSWORD", "")
			settings.SMTP.AuthMethod = common.GetEnv("SMTP_AUTH_METHOD", "PLAIN")
			settings.SMTP.TLS = common.GetEnvBool("SMTP_TLS", true)
		}

		// Configure S3 storage from environment variables
		settings.S3.Enabled = common.GetEnvBool("S3_ENABLED", false)
		if settings.S3.Enabled {
			settings.S3.Bucket = common.GetEnv("S3_BUCKET", "")
			settings.S3.Region = common.GetEnv("S3_REGION", "")
			settings.S3.Endpoint = common.GetEnv("S3_ENDPOINT", "")
			settings.S3.AccessKey = common.GetEnv("S3_ACCESS_KEY", "")
			settings.S3.Secret = common.GetEnv("S3_SECRET", "")
		}

		// Configure batch settings from environment variables
		settings.Batch.Enabled = common.GetEnvBool("BATCH_ENABLED", true)
		settings.Batch.MaxRequests = common.GetEnvInt("BATCH_MAX_REQUESTS", 100)

		// Configure rate limiting from environment variables
		settings.RateLimits.Enabled = common.GetEnvBool("RATE_LIMITS_ENABLED", true)
		if settings.RateLimits.Enabled {
			maxHits := common.GetEnvInt("RATE_LIMITS_MAX_HITS", 120)
			duration := int64(common.GetEnvInt("RATE_LIMITS_DURATION", 60))
			settings.RateLimits.Rules = []core.RateLimitRule{
				{Label: "default", MaxRequests: maxHits, Duration: duration}, // e.g., 120 requests per minute
			}
		}

		if err := app.Save(settings); err != nil {
			return fmt.Errorf("failed to save settings: %w", err)
		}

		// 6. Create superuser if none exists
		if err := seeders.CreateSuperUserIfNotExists(app); err != nil {
			return fmt.Errorf("failed to create superuser: %w", err)
		}

		// 7. Seed RBAC data (permissions, roles, and super admin user)
		return seeders.SeedRBAC(app)
	}, func(app core.App) error {
		// Revert operation - drop all non-system collections
		collections, err := app.FindAllCollections()
		if err != nil {
			return fmt.Errorf("failed to fetch collections: %w", err)
		}

		for _, collection := range collections {
			// Skip system collections
			if collection.System {
				continue
			}

			if err := app.Delete(collection); err != nil {
				return fmt.Errorf("failed to delete collection %s: %w", collection.Name, err)
			}
		}

		return nil
	})
}
