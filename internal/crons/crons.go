package crons

import (
	"os"

	"ims-pocketbase-baas-starter/internal/handlers/cron"
	"ims-pocketbase-baas-starter/pkg/cronutils"
	"ims-pocketbase-baas-starter/pkg/logger"

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

	log := logger.GetLogger(app)
	log.Info("Starting cron job registration process")

	// Define all cron jobs
	crons := []Cron{
		{
			ID:          "system_queue",
			CronExpr:    "* * * * *", // every minutes
			Handler:     cronutils.WithRecovery(app, "system_queue", func() { cron.HandleSystemQueue(app) }),
			Enabled:     os.Getenv("ENABLE_SYSTEM_QUEUE_CRON") != "false", // Enabled by default
			Description: "Process the system queue ",
		},
		{
			ID:          "clean_exported_files",
			CronExpr:    "0 2 * * *", // every day at 2:00 AM
			Handler:     cronutils.WithRecovery(app, "clean_exported_files", func() { cron.HandleClearExportFiles(app) }),
			Enabled:     os.Getenv("ENABLE_CLEAR_EXPORT_FILES_CRON") != "false", // Enabled by default
			Description: "Delete the expired job generated export files",
		},
	}

	log.Info("Registering cron jobs", "total_cron_jobs", len(crons))

	// Register enabled cron jobs with PocketBase cron scheduler
	for _, cronJob := range crons {
		if !cronJob.Enabled {
			log.Info("Skipped disabled cron job", "cron_id", cronJob.ID, "description", cronJob.Description)
			continue
		}

		// Validate cron expression before registering
		if err := cronutils.ValidateCronExpression(cronJob.CronExpr); err != nil {
			log.Error("Invalid cron expression for cron job", "cron_id", cronJob.ID, "cron", cronJob.CronExpr, "error", err)
			continue
		}

		// Register the cron job with PocketBase scheduler
		app.Cron().MustAdd(cronJob.ID, cronJob.CronExpr, cronJob.Handler)

		log.Info("Registered cron job",
			"cron_id", cronJob.ID,
			"cron_expr", cronJob.CronExpr,
			"description", cronJob.Description,
		)
	}

	log.Info("Cron job registration completed", "enabled_cron_jobs", len(crons))
}
