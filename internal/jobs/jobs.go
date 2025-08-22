package jobs

import (
	"ims-pocketbase-baas-starter/internal/handlers/jobs"
	"ims-pocketbase-baas-starter/pkg/jobutils"
	"ims-pocketbase-baas-starter/pkg/logger"

	"github.com/pocketbase/pocketbase"
)

// Job represents a background job with its configuration
type Job struct {
	Type        string              // Job type identifier
	Handler     jobutils.JobHandler // Handler function to process this job type
	Enabled     bool                // Whether the job handler should be registered
	Description string              // Human-readable description of what the job does
}

// RegisterJobs registers all background job handlers with the PocketBase application
// This function should be called during app initialization, before OnServe setup
func RegisterJobs(app *pocketbase.PocketBase) error {
	if app == nil {
		panic("RegisterJobs: app cannot be nil")
	}

	log := logger.GetLogger(app)
	log.Info("Starting job handler registration process")

	// Get the job manager and processor
	jobManager := GetJobManager()
	processor := jobManager.GetProcessor()

	if processor == nil {
		err := jobManager.Initialize(app)
		if err != nil {
			log.Error("Failed to initialize job manager", "error", err)
			return err
		}
		processor = jobManager.GetProcessor()
	}

	// Define all job handlers
	jobs := []Job{
		{
			Type:        jobutils.JobTypeEmail,
			Handler:     jobs.NewEmailJobHandler(app),
			Enabled:     true, // Always enabled
			Description: "Process email jobs",
		},
		{
			Type:        jobutils.JobTypeDataProcessing,
			Handler:     jobs.NewDataProcessingJobHandler(app),
			Enabled:     true, // Always enabled
			Description: "Process data processing jobs",
		},
		// Add more job handlers here as needed:
		// {
		//     Type:        "example_job",
		//     Handler:     jobs.NewExampleJobHandler(app),
		//     Enabled:     true,
		//     Description: "Process example jobs",
		// },
	}

	log.Info("Registering job handlers", "total_job_handlers", len(jobs))

	// Register enabled job handlers with the job processor
	registry := processor.GetRegistry()
	registeredCount := 0

	for _, jobHandler := range jobs {
		if !jobHandler.Enabled {
			log.Info("Skipped disabled job handler", "job_type", jobHandler.Type, "description", jobHandler.Description)
			continue
		}

		// Register the job handler with the registry
		if err := registry.Register(jobHandler.Handler); err != nil {
			log.Error("Failed to register job handler",
				"job_type", jobHandler.Type,
				"description", jobHandler.Description,
				"error", err)
			continue
		}

		log.Info("Registered job handler",
			"job_type", jobHandler.Type,
			"description", jobHandler.Description,
		)
		registeredCount++
	}

	log.Info("Job handler registration completed", "registered_handlers", registeredCount)
	return nil
}
