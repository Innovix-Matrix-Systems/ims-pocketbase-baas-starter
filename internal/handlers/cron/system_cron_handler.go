package cron

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"ims-pocketbase-baas-starter/internal/jobs"
	"ims-pocketbase-baas-starter/pkg/cronutils"

	"github.com/pocketbase/pocketbase"
)

// HandleSystemQueue processes jobs from the queue table using the job processor
func HandleSystemQueue(app *pocketbase.PocketBase) {
	ctx := cronutils.NewCronExecutionContext(app, "system_queue")
	ctx.LogStart("Starting system queue process operations")

	// Get the pre-initialized job processor
	jobManager := jobs.GetJobManager()
	processor := jobManager.GetProcessor()

	if processor == nil {
		ctx.LogError(nil, "Job processor not initialized")
		return
	}

	// Get configuration for concurrent processing
	maxWorkers := getEnvInt("JOB_MAX_WORKERS", 5) // Default 5 concurrent workers
	batchSize := getEnvInt("JOB_BATCH_SIZE", 50)  // Process up to 50 jobs per run

	// Fetch pending jobs (not reserved or reservation expired)
	expiredTime := time.Now().Add(-5 * time.Minute) // 5 minute reservation timeout

	queues, err := app.FindRecordsByFilter(
		"queues",
		"reserved_at = '' || reserved_at < {:expired}",
		"-created", // Order by created descending (FIFO)
		batchSize,
		0,
		map[string]any{"expired": expiredTime.Format(time.RFC3339)},
	)

	if err != nil {
		ctx.LogError(err, "Error fetching queues data")
		return
	}

	// Process jobs concurrently if any are available
	if len(queues) > 0 {
		ctx.LogDebug(fmt.Sprintf("Processing %d jobs with %d workers", len(queues), maxWorkers), "Starting concurrent job processing")

		errors := processor.ProcessJobsConcurrently(queues, maxWorkers)

		// Count and log results
		successCount := 0
		failureCount := 0
		for _, err := range errors {
			if err == nil {
				successCount++
			} else {
				failureCount++
				ctx.LogError(err, "Job processing error")
			}
		}

		app.Logger().Info("Job processing batch completed",
			"total_jobs", len(queues),
			"successful", successCount,
			"failed", failureCount,
			"workers", maxWorkers)
	} else {
		ctx.LogDebug("No pending jobs in queue", "No jobs found to process")
	}

	ctx.LogEnd("System queue process operations completed successfully")
}

// getEnvInt gets an integer value from environment variable with a default fallback
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
