package cron

import (
	"time"

	"ims-pocketbase-baas-starter/internal/jobs"
	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/cronutils"
	log "ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
)

// HandleSystemQueue processes jobs from the queue table using the job processor
func HandleSystemQueue(app *pocketbase.PocketBase) {
	ctx := cronutils.NewCronExecutionContext(app, "system_queue")
	ctx.LogStart("Starting system queue process operations")

	jobManager := jobs.GetJobManager()
	processor := jobManager.GetProcessor()

	if processor == nil {
		ctx.LogError(nil, "Job processor not initialized")
		return
	}

	maxWorkers := common.GetEnvInt("JOB_MAX_WORKERS", 5)                 // Default 5 concurrent workers
	batchSize := common.GetEnvInt("JOB_BATCH_SIZE", 50)                  // Process up to 50 jobs per run
	reservationTimeout := common.GetEnvInt("JOB_RESERVATION_TIMEOUT", 5) // Default 5 minutes reservation timeout

	// Fetch pending jobs: either not reserved or reservation has expired
	expiredTime := time.Now().Add(-time.Duration(reservationTimeout) * time.Minute)

	// Format for PocketBase: RFC3339 with 'T' replaced by space (e.g., "2025-11-04 19:39:00Z")
	pbExpiredTime := expiredTime.Format("2006-01-02 15:04:05Z")

	queues, err := app.FindRecordsByFilter(
		"queues",
		"reserved_at = '' || reserved_at < {:expired}",
		"-created", // FIFO: oldest first
		batchSize,
		0,
		dbx.Params{"expired": pbExpiredTime},
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

	if len(queues) > 0 {
		errors := processor.ProcessJobsConcurrently(queues, maxWorkers)
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

		log.Info("Job processing batch completed",
			"total_jobs", len(queues),
			"successful", successCount,
			"failed", failureCount,
			"workers", maxWorkers)
	}

	ctx.LogEnd("System queue process operations completed successfully")
}
