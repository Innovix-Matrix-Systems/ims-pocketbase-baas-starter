package seeders

import (
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// SeedSettings seeds default application settings
func SeedSettings(app core.App) error {
	fmt.Println("🌱 Seeding application settings...")

	if err := seedDefaultSettings(app); err != nil {
		return fmt.Errorf("failed to seed default settings: %w", err)
	}

	fmt.Println("✅ Settings seeding completed successfully")
	return nil
}

// seedDefaultSettings creates default application settings
func seedDefaultSettings(app core.App) error {
	// Define default settings based on the actual schema (name, slug, description)
	defaultSettings := []Setting{
		{
			Name:        "Theme",
			Slug:        "theme",
			Description: "Default application theme (light, dark, auto)",
		},
		{
			Name:        "Notifications",
			Slug:        "notifications",
			Description: "Enable/disable notifications system-wide",
		},
		{
			Name:        "Language",
			Slug:        "language",
			Description: "Default application language (en, es, fr, de, it)",
		},
		{
			Name:        "Timezone",
			Slug:        "timezone",
			Description: "Default application timezone (UTC, America/New_York, Europe/London, Asia/Tokyo)",
		},
		{
			Name:        "Maintenance Mode",
			Slug:        "maintenance_mode",
			Description: "Enable/disable maintenance mode",
		},
	}

	// Get settings collection
	settingsCollection, err := app.FindCollectionByNameOrId("settings")
	if err != nil {
		return fmt.Errorf("settings collection not found: %w", err)
	}

	// Seed each setting
	for _, setting := range defaultSettings {
		// Check if setting already exists
		exists, err := app.FindFirstRecordByFilter("settings", "slug = {:slug}", dbx.Params{"slug": setting.Slug})
		if err == nil && exists != nil {
			fmt.Printf("⏭️  Setting '%s' already exists, skipping\n", setting.Slug)
			continue
		}

		// Create new setting record
		record := core.NewRecord(settingsCollection)
		record.Set("name", setting.Name)
		record.Set("slug", setting.Slug)
		record.Set("description", setting.Description)

		// Save the setting
		if err := app.Save(record); err != nil {
			return fmt.Errorf("failed to create setting '%s': %w", setting.Slug, err)
		}

		fmt.Printf("✅ Created setting: %s (%s)\n", setting.Name, setting.Slug)
	}

	return nil
}

// GetSettingBySlug retrieves a setting by slug
func GetSettingBySlug(app core.App, slug string) (*core.Record, error) {
	setting, err := app.FindFirstRecordByFilter("settings", "slug = {:slug}", dbx.Params{"slug": slug})
	if err != nil {
		return nil, fmt.Errorf("setting '%s' not found: %w", slug, err)
	}

	return setting, nil
}

// GetAllSettings retrieves all settings
func GetAllSettings(app core.App) ([]*core.Record, error) {
	settings, err := app.FindAllRecords("settings")
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return settings, nil
}

// UpdateSettingDescription updates a setting description by slug
func UpdateSettingDescription(app core.App, slug, description string) error {
	setting, err := app.FindFirstRecordByFilter("settings", "slug = {:slug}", dbx.Params{"slug": slug})
	if err != nil {
		return fmt.Errorf("setting '%s' not found: %w", slug, err)
	}

	setting.Set("description", description)
	if err := app.Save(setting); err != nil {
		return fmt.Errorf("failed to update setting '%s': %w", slug, err)
	}

	return nil
}

// Setting represents a setting structure matching the schema
type Setting struct {
	Name        string
	Slug        string
	Description string
}
