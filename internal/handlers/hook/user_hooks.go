package hook

import (
	"encoding/json"
	"fmt"
	"time"

	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/jobutils"
	"ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase/core"
)

// HandleUserWelcomeEmail handles sending a welcome email to new users
func HandleUserWelcomeEmail(e *core.RecordEvent) error {
	// Log the user creation
	if log := logger.FromApp(e.App); log != nil {
		log.Info("New user created - preparing welcome email",
			"user_id", e.Record.Id,
			"email", e.Record.GetString("email"))
	}

	// Get app name from settings or use default
	appName := common.GetEnv("APP_NAME", "N/A")
	settings := e.App.Settings()
	if settings.Meta.AppName != "" {
		appName = settings.Meta.AppName
	}

	// Get user details
	email := e.Record.GetString("email")
	name := e.Record.GetString("name")
	appUrl := common.GetEnv("APP_URL", "N/A")
	if settings.Meta.AppURL != "" {
		appUrl = settings.Meta.AppURL
	}
	if name == "" {
		name = email // Use email as name if name is not provided
	}

	// Create email job payload
	payload := jobutils.EmailJobPayload{
		Type: jobutils.JobTypeEmail,
		Data: jobutils.EmailJobData{
			To:       email,
			Subject:  fmt.Sprintf("Welcome to %s!", appName),
			Template: "welcome",
			Variables: map[string]any{
				"AppName": appName,
				"Name":    name,
				"Email":   email,
				"AppURL":  appUrl,
				"Year":    time.Now().Year(),
			},
		},
		Options: jobutils.EmailJobOptions{
			RetryCount: 3,
			Timeout:    30,
		},
	}

	// Create job record in queues collection
	collection, err := e.App.FindCollectionByNameOrId("queues")
	if err != nil {
		if log := logger.FromApp(e.App); log != nil {
			log.Error("Failed to find queues collection", "error", err)
		}
		return err
	}

	jobRecord := core.NewRecord(collection)
	jobRecord.Set("name", fmt.Sprintf("Welcome email for %s", email))
	jobRecord.Set("description", fmt.Sprintf("Send welcome email to new user %s", email))

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		if log := logger.FromApp(e.App); log != nil {
			log.Error("Failed to marshal email payload", "error", err)
		}
		return err
	}
	jobRecord.Set("payload", string(payloadBytes))
	jobRecord.Set("attempts", 0)

	if err := e.App.Save(jobRecord); err != nil {
		if log := logger.FromApp(e.App); log != nil {
			log.Error("Failed to queue welcome email job", "error", err)
		}
		return err
	}

	if log := logger.FromApp(e.App); log != nil {
		log.Info("Welcome email job queued successfully",
			"user_id", e.Record.Id,
			"email", email,
			"job_id", jobRecord.Id)
	}

	return e.Next()
}

// HandleUserCreateSettings generate default user settings
func HandleUserCreateSettings(e *core.RecordEvent) error {
	if log := logger.FromApp(e.App); log != nil {
		log.Info("Creating default settings for new user",
			"user_id", e.Record.Id,
			"email", e.Record.GetString("email"),
		)
	}

	// Find the user_settings collection
	userSettingsCollection, err := e.App.FindCollectionByNameOrId("user_settings")
	if err != nil {
		if log := logger.FromApp(e.App); log != nil {
			log.Error("user_settings collection not found", "error", err)
		}
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
			if log := logger.FromApp(e.App); log != nil {
				log.Warn("Setting not found, skipping",
					"slug", defaultSetting.SettingSlug,
					"error", err)
			}
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
			if log := logger.FromApp(e.App); log != nil {
				log.Error("Failed to create user setting",
					"user_id", e.Record.Id,
					"setting_slug", defaultSetting.SettingSlug,
					"error", err)
			}
			continue
		}

		if log := logger.FromApp(e.App); log != nil {
			log.Debug("User setting created",
				"user_id", e.Record.Id,
				"setting_slug", defaultSetting.SettingSlug,
				"value", defaultSetting.Value,
				"user_setting_id", userSettingRecord.Id)
		}
	}

	if log := logger.FromApp(e.App); log != nil {
		log.Info("Default user settings creation completed",
			"user_id", e.Record.Id)
	}

	return e.Next()
}
