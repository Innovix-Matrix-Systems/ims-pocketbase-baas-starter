package seeders

import (
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase/core"
)

// CreateSuperUserIfNotExists creates a superuser if none exists
func CreateSuperUserIfNotExists(app core.App) error {
	// Check if any superusers exist
	superusers, err := app.FindAllRecords("_superusers")
	if err != nil {
		return fmt.Errorf("failed to check existing superusers: %w", err)
	}

	// If superusers already exist, skip creation
	if len(superusers) > 0 {
		fmt.Println("✅ Superuser already exists, skipping creation")
		return nil
	}

	// Get superuser collection
	collection, err := app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		return fmt.Errorf("failed to find superusers collection: %w", err)
	}

	// Create new superuser record
	record := core.NewRecord(collection)

	// Configuration from environment variables with defaults
	config := getSuperUserConfig()

	// Set the fields
	record.Set("email", config.Email)
	record.Set("password", config.Password)

	// Save the superuser
	if err := app.Save(record); err != nil {
		return fmt.Errorf("failed to create superuser: %w", err)
	}

	fmt.Printf("✅ Superuser created successfully with email: %s\n", config.Email)

	if config.IsDefaultPassword {
		fmt.Println("⚠️  WARNING: Using default password! Please change it after first login!")
		fmt.Println("   Default credentials:")
		fmt.Printf("   Email: %s\n", config.Email)
		fmt.Printf("   Password: %s\n", config.Password)
	}

	return nil
}

// SuperUserConfig holds the configuration for creating a superuser
type SuperUserConfig struct {
	Email             string
	Password          string
	IsDefaultPassword bool
}

// getSuperUserConfig returns the superuser configuration from environment variables or defaults
func getSuperUserConfig() SuperUserConfig {
	config := SuperUserConfig{
		Email:             "superadmin@ims.com",
		Password:          "superadmin123456",
		IsDefaultPassword: true,
	}

	// Override with environment variables if provided
	if email := os.Getenv("SUPERUSER_EMAIL"); email != "" {
		config.Email = email
	}

	if password := os.Getenv("SUPERUSER_PASSWORD"); password != "" {
		config.Password = password
		config.IsDefaultPassword = false
	}

	return config
}

// CreateMultipleSuperUsers creates multiple superusers from a predefined list
// Useful for team setups
func CreateMultipleSuperUsers(app core.App, users []SuperUserConfig) error {
	collection, err := app.FindCollectionByNameOrId("_superusers")
	if err != nil {
		return fmt.Errorf("failed to find superusers collection: %w", err)
	}

	for _, user := range users {
		// Check if user already exists
		existing, _ := app.FindFirstRecordByFilter("_superusers", "email = {:email}", map[string]any{
			"email": user.Email,
		})

		if existing != nil {
			fmt.Printf("⏭️  Superuser %s already exists, skipping\n", user.Email)
			continue
		}

		// Create new superuser record
		record := core.NewRecord(collection)
		record.Set("email", user.Email)
		record.Set("password", user.Password)

		// Save the superuser
		if err := app.Save(record); err != nil {
			fmt.Printf("❌ Failed to create superuser %s: %v\n", user.Email, err)
			continue
		}

		fmt.Printf("✅ Superuser created: %s\n", user.Email)
	}

	return nil
}
