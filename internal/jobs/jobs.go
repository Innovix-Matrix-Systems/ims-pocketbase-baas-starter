package jobs

import (
	"os"

	"ims-pocketbase-baas-starter/internal/handlers/cron"
	"ims-pocketbase-baas-starter/pkg/common/jobutils"

	"github.com/pocketbase/pocketbase"
)

// Job represents a scheduled job with its configuration
type Job struct {
	ID          string // Unique identifier for the job
	CronExpr    string // Cron expression for scheduling (e.g., "0 2 * * *")
	Handler     func() // Function to execute when job runs
	Enabled     bool   // Whether the job should be registered and executed
	Description string // Human-readable description of what the job does
}

// RegisterJobs registers all scheduled jobs with the PocketBase application
// This function should be called during app initialization, before OnServe setup
func RegisterJobs(app *pocketbase.PocketBase) {
	if app == nil {
		panic("RegisterJobs: app cannot be nil")
	}

	app.Logger().Info("Starting job registration process")

	// Define all jobs in one place
	jobs := []Job{
		{
			ID:          "system_queue",
			CronExpr:    "* * * * *", // every minutes
			Handler:     jobutils.WithRecovery(app, "system_queue", func() { cron.HandleSystemQueue(app) }),
			Enabled:     os.Getenv("ENABLE_SYSTEM_QUEUE_CRON") != "false", // Enabled by default
			Description: "Process the system queue ",
		},
	}

	app.Logger().Info("Registering jobs", "total_jobs", len(jobs))

	// Register enabled jobs with PocketBase cron scheduler
	registeredCount := 0
	for _, job := range jobs {
		if job.Enabled {
			// Validate cron expression before registration
			if err := jobutils.ValidateCronExpression(job.CronExpr); err != nil {
				app.Logger().Error("Invalid cron expression for job", "job_id", job.ID, "cron", job.CronExpr, "error", err)
				continue
			}

			app.Cron().MustAdd(job.ID, job.CronExpr, job.Handler)
			app.Logger().Info("Registered job",
				"id", job.ID,
				"schedule", job.CronExpr,
				"description", job.Description)
			registeredCount++
		} else {
			app.Logger().Info("Skipped disabled job",
				"id", job.ID,
				"description", job.Description)
		}
	}

	app.Logger().Info("Job registration completed",
		"registered", registeredCount,
		"total", len(jobs))
}
