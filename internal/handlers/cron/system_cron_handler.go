package cron

import (
	"fmt"
	"time"

	"ims-pocketbase-baas-starter/internal/jobs"
	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/cronutils"
	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"

	"github.com/pocketbase/dbx"
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
	maxWorkers := common.GetEnvInt("JOB_MAX_WORKERS", 5)                 // Default 5 concurrent workers
	batchSize := common.GetEnvInt("JOB_BATCH_SIZE", 50)                  // Process up to 50 jobs per run
	reservationTimeout := common.GetEnvInt("JOB_RESERVATION_TIMEOUT", 5) // Default 5 minutes reservation timeout

	// Fetch pending jobs (not reserved or reservation expired)
	expiredTime := time.Now().Add(-time.Duration(reservationTimeout) * time.Minute)

	queues, err := app.FindRecordsByFilter(
		"queues",
		"reserved_at = '' || reserved_at < {:expired}",
		"-created", // Order by created descending (FIFO)
		batchSize,
		0,
		dbx.Params{"expired": expiredTime.Format(time.RFC3339)},
	)

	if err != nil {
		ctx.LogError(err, "Error fetching queues data")
		return
	}

	// Record queue size metrics
	metricsProvider := metrics.GetInstance()
	if metricsProvider != nil {
		metrics.RecordQueueSize(metricsProvider, "system", len(queues))
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

		logger := logger.GetLogger(app)
		logger.Info("Job processing batch completed",
			"total_jobs", len(queues),
			"successful", successCount,
			"failed", failureCount,
			"workers", maxWorkers)
	} else {
		ctx.LogDebug("No pending jobs in queue", "No jobs found to process")
	}

	ctx.LogEnd("System queue process operations completed successfully")
}
