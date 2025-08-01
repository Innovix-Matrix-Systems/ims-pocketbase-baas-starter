package migrations

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"

	"ims-pocketbase-baas-starter/internal/database/seeders"
)

// getEnvWithDefault returns environment variable value or default if not set
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBoolWithDefault returns boolean environment variable value or default if not set
func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvIntWithDefault returns integer environment variable value or default if not set
func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

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
		settings.Meta.AppName = getEnvWithDefault("APP_NAME", "IMS_PocketBase_App")
		settings.Meta.AppURL = getEnvWithDefault("APP_URL", "http://localhost:8090")

		// Configure logs retention
		settings.Logs.MaxDays = getEnvIntWithDefault("LOGS_MAX_DAYS", 7)

		// Configure SMTP settings from environment variables
		settings.SMTP.Enabled = getEnvBoolWithDefault("SMTP_ENABLED", false)
		if settings.SMTP.Enabled {
			settings.SMTP.Host = getEnvWithDefault("SMTP_HOST", "")
			settings.SMTP.Port = getEnvIntWithDefault("SMTP_PORT", 587)
			settings.SMTP.Username = getEnvWithDefault("SMTP_USERNAME", "")
			settings.SMTP.Password = getEnvWithDefault("SMTP_PASSWORD", "")
			settings.SMTP.AuthMethod = getEnvWithDefault("SMTP_AUTH_METHOD", "PLAIN")
			settings.SMTP.TLS = getEnvBoolWithDefault("SMTP_TLS", true)
		}

		// Configure S3 storage from environment variables
		settings.S3.Enabled = getEnvBoolWithDefault("S3_ENABLED", false)
		if settings.S3.Enabled {
			settings.S3.Bucket = getEnvWithDefault("S3_BUCKET", "")
			settings.S3.Region = getEnvWithDefault("S3_REGION", "")
			settings.S3.Endpoint = getEnvWithDefault("S3_ENDPOINT", "")
			settings.S3.AccessKey = getEnvWithDefault("S3_ACCESS_KEY", "")
			settings.S3.Secret = getEnvWithDefault("S3_SECRET", "")
		}

		// Configure batch settings from environment variables
		settings.Batch.Enabled = getEnvBoolWithDefault("BATCH_ENABLED", true)
		settings.Batch.MaxRequests = getEnvIntWithDefault("BATCH_MAX_REQUESTS", 100)

		// Configure rate limiting from environment variables
		settings.RateLimits.Enabled = getEnvBoolWithDefault("RATE_LIMITS_ENABLED", true)
		if settings.RateLimits.Enabled {
			maxHits := getEnvIntWithDefault("RATE_LIMITS_MAX_HITS", 120)
			duration := int64(getEnvIntWithDefault("RATE_LIMITS_DURATION", 60))
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
