package hook

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// HandleRecordCreate handles record creation events
func HandleRecordCreate(e *core.RecordEvent) error {
	// Log the record creation
	e.App.Logger().Info("Record created",
		"collection", e.Record.Collection().Name,
		"id", e.Record.Id,
		"created", e.Record.GetDateTime("created"),
	)

	// Add your custom logic here
	// For example: send notifications, update related records, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordUpdate handles record update events
func HandleRecordUpdate(e *core.RecordEvent) error {
	// Log the record update
	e.App.Logger().Info("Record updated",
		"collection", e.Record.Collection().Name,
		"id", e.Record.Id,
		"updated", e.Record.GetDateTime("updated"),
	)

	// Add your custom logic here
	// For example: track changes, send notifications, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordDelete handles record deletion events
func HandleRecordDelete(e *core.RecordEvent) error {
	// Log the record deletion
	e.App.Logger().Info("Record deleted",
		"collection", e.Record.Collection().Name,
		"id", e.Record.Id,
	)

	// Add your custom logic here
	// For example: cleanup related data, send notifications, etc.

	// Continue with the execution chain
	return e.Next()
}

// HandleRecordAfterCreateSuccess handles successful record creation
func HandleRecordAfterCreateSuccess(e *core.RecordEvent) error {
	// This hook is triggered after the record is successfully persisted
	e.App.Logger().Debug("Record successfully persisted",
		"collection", e.Record.Collection().Name,
		"id", e.Record.Id,
	)

	// Add post-creation logic here
	// For example: send confirmation emails, trigger webhooks, etc.

	return e.Next()
}

// HandleRecordAfterCreateError handles failed record creation
func HandleRecordAfterCreateError(e *core.RecordEvent) error {
	// This hook is triggered when record creation fails
	e.App.Logger().Error("Record creation failed",
		"collection", e.Record.Collection().Name,
		"error", fmt.Sprintf("%v", e),
	)

	// Add error handling logic here
	// For example: cleanup, notifications, etc.

	return e.Next()
}

// HandleUserCreate handles user-specific record creation
func HandleUserCreate(e *core.RecordEvent) error {
	// This is an example of collection-specific hook
	e.App.Logger().Info("New user created",
		"user_id", e.Record.Id,
		"email", e.Record.GetString("email"),
	)

	// Add user-specific logic here
	// For example: send welcome email, create user profile, etc.

	return e.Next()
}

// HandleUserCreateSettings generate default user settings
func HandleUserCreateSettings(e *core.RecordEvent) error {
	e.App.Logger().Info("Creating default settings for new user",
		"user_id", e.Record.Id,
		"email", e.Record.GetString("email"),
	)

	// Find the user_settings collection
	userSettingsCollection, err := e.App.FindCollectionByNameOrId("user_settings")
	if err != nil {
		e.App.Logger().Error("user_settings collection not found", "error", err)
		// Continue without failing if settings collection doesn't exist
		return e.Next()
	}

	// Define default user settings with their values
	defaultUserSettings := []struct {
		SettingSlug string
		Value       string
	}{
		{"theme", "light"},
		{"notifications", "true"},
	}

	// Create user settings for each default setting
	for _, defaultSetting := range defaultUserSettings {
		// Find the setting record by slug
		settingRecord, err := e.App.FindFirstRecordByFilter("settings", "slug = {:slug}", map[string]any{
			"slug": defaultSetting.SettingSlug,
		})
		if err != nil {
			e.App.Logger().Warn("Setting not found, skipping",
				"slug", defaultSetting.SettingSlug,
				"error", err)
			continue
		}

		// Create new user setting record
		userSettingRecord := core.NewRecord(userSettingsCollection)
		userSettingData := map[string]any{
			"user":     e.Record.Id,
			"settings": settingRecord.Id,
			"value":    defaultSetting.Value,
		}

		// Load the data into the record
		userSettingRecord.Load(userSettingData)

		// Save the user setting record
		if err := e.App.Save(userSettingRecord); err != nil {
			e.App.Logger().Error("Failed to create user setting",
				"user_id", e.Record.Id,
				"setting_slug", defaultSetting.SettingSlug,
				"error", err)
			continue
		}

		e.App.Logger().Debug("User setting created",
			"user_id", e.Record.Id,
			"setting_slug", defaultSetting.SettingSlug,
			"value", defaultSetting.Value,
			"user_setting_id", userSettingRecord.Id)
	}

	e.App.Logger().Info("Default user settings creation completed",
		"user_id", e.Record.Id)

	return e.Next()
}
