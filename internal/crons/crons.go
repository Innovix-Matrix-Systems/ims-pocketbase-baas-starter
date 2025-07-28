package crons

import (
	"os"

	"ims-pocketbase-baas-starter/internal/handlers/cron"
	"ims-pocketbase-baas-starter/pkg/cronutils"

	"github.com/pocketbase/pocketbase"
)

// Cron represents a scheduled cron job with its configuration
type Cron struct {
	ID          string // Unique identifier for the cron
	CronExpr    string // Cron expression for scheduling (e.g., "0 2 * * *")
	Handler     func() // Function to execute when cron job runs
	Enabled     bool   // Whether the cron job should be registered and executed
	Description string // Human-readable description of what the cron job does
}

// RegisterCrons registers all scheduled crons with the PocketBase application
// This function should be called during app initialization, before OnServe setup
func RegisterCrons(app *pocketbase.PocketBase) {
	if app == nil {
		panic("RegisterCrons: app cannot be nil")
	}

	app.Logger().Info("Starting cron job registration process")

	// Define all cron jobs
	crons := []Cron{
		{
			ID:          "system_queue",
			CronExpr:    "* * * * *", // every minutes
			Handler:     cronutils.WithRecovery(app, "system_queue", func() { cron.HandleSystemQueue(app) }),
			Enabled:     os.Getenv("ENABLE_SYSTEM_QUEUE_CRON") != "false", // Enabled by default
			Description: "Process the system queue ",
		},
	}

	app.Logger().Info("Registering cron jobs", "total_cron_jobs", len(crons))

	// Register enabled cron jobs with PocketBase cron scheduler
	registeredCount := 0
	for _, cron := range crons {
		if cron.Enabled {
			// Validate cron expression before registration
			if err := cronutils.ValidateCronExpression(cron.CronExpr); err != nil {
				app.Logger().Error("Invalid cron expression for cron job", "cron_id", cron.ID, "cron", cron.CronExpr, "error", err)
				continue
			}

			app.Cron().MustAdd(cron.ID, cron.CronExpr, cron.Handler)
			app.Logger().Info("Registered cron job",
				"id", cron.ID,
				"schedule", cron.CronExpr,
				"description", cron.Description)
			registeredCount++
		} else {
			app.Logger().Info("Skipped disabled cron job",
				"id", cron.ID,
				"description", cron.Description)
		}
	}

	app.Logger().Info("Cron job registration completed",
		"registered", registeredCount,
		"total", len(crons))
}
