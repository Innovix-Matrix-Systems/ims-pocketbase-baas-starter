package jobutils

import (
	"context"
	"fmt"
	"ims-pocketbase-baas-starter/pkg/cronutils"
	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// WorkerPool manages a pool of persistent workers for job processing
type WorkerPool struct {
	workers     []*Worker
	jobQueue    chan *core.Record
	resultQueue chan WorkerJobResult
	quit        chan bool
	wg          sync.WaitGroup
	maxWorkers  int
	app         *pocketbase.PocketBase
	registry    *JobRegistry
}

// Worker represents a single worker in the pool
type Worker struct {
	id          int
	jobQueue    chan *core.Record
	resultQueue chan WorkerJobResult
	quit        chan bool
	app         *pocketbase.PocketBase
	registry    *JobRegistry
}

// WorkerJobResult represents the result of job processing
type WorkerJobResult struct {
	JobID string
	Error error
}

// NewWorkerPool creates a new persistent worker pool
func NewWorkerPool(app *pocketbase.PocketBase, registry *JobRegistry, maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 5 // Default to 5 workers
	}

	pool := &WorkerPool{
		workers:     make([]*Worker, 0, maxWorkers),
		jobQueue:    make(chan *core.Record, maxWorkers*2),    // Buffer for jobs
		resultQueue: make(chan WorkerJobResult, maxWorkers*2), // Buffer for results
		quit:        make(chan bool),
		maxWorkers:  maxWorkers,
		app:         app,
		registry:    registry,
	}

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		worker := &Worker{
			id:          i,
			jobQueue:    pool.jobQueue,
			resultQueue: pool.resultQueue,
			quit:        make(chan bool),
			app:         app,
			registry:    registry,
		}
		pool.workers = append(pool.workers, worker)
		pool.wg.Add(1)
		go worker.start(&pool.wg)
	}

	log := logger.GetLogger(app)
	log.Info("Worker pool started", "workers", maxWorkers)
	return pool
}

// ProcessJobs processes a batch of jobs using the worker pool
func (wp *WorkerPool) ProcessJobs(jobs []*core.Record) []error {
	if len(jobs) == 0 {
		return nil
	}

	// Record queue size metrics before processing
	metricsProvider := metrics.GetInstance()
	if metricsProvider != nil {
		metrics.RecordQueueSize(metricsProvider, "worker_pool", len(wp.jobQueue))
	}

	// Send jobs to workers
	for _, job := range jobs {
		select {
		case wp.jobQueue <- job:
			// Job queued successfully
		case <-time.After(30 * time.Second):
			// Timeout - this shouldn't happen with proper buffer sizing
			log := logger.GetLogger(wp.app)
			log.Error("Job queue timeout", "job_id", job.Id)
		}
	}

	// Collect results
	results := make([]error, len(jobs))
	for i := range len(jobs) {
		select {
		case result := <-wp.resultQueue:
			// Find the job index and store the result
			for j, job := range jobs {
				if job.Id == result.JobID {
					results[j] = result.Error
					break
				}
			}
		case <-time.After(5 * time.Minute):
			// Timeout for job processing
			log := logger.GetLogger(wp.app)
			log.Error("Job processing timeout")
			results[i] = fmt.Errorf("job processing timeout")
		}
	}

	return results
}

// ProcessJobsConcurrently is an alias for ProcessJobs for compatibility
func (wp *WorkerPool) ProcessJobsConcurrently(jobs []*core.Record, maxWorkers int) []error {
	// Ignore maxWorkers parameter as we use the pool's configured workers
	return wp.ProcessJobs(jobs)
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(ctx context.Context) error {
	log := logger.GetLogger(wp.app)
	log.Info("Shutting down worker pool")

	// Close job queue to signal no more jobs
	close(wp.jobQueue)

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log := logger.GetLogger(wp.app)
		log.Info("Worker pool shutdown completed")
		return nil
	case <-ctx.Done():
		// Force shutdown
		close(wp.quit)
		log := logger.GetLogger(wp.app)
		log.Warn("Worker pool force shutdown due to timeout")
		return ctx.Err()
	}
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() map[string]any {
	return map[string]any{
		"worker_count":      len(wp.workers),
		"max_workers":       wp.maxWorkers,
		"job_queue_size":    len(wp.jobQueue),
		"job_queue_cap":     cap(wp.jobQueue),
		"result_queue_size": len(wp.resultQueue),
		"result_queue_cap":  cap(wp.resultQueue),
	}
}

// Worker methods

// start begins the worker's job processing loop
func (w *Worker) start(wg *sync.WaitGroup) {
	defer wg.Done()

	log := logger.GetLogger(w.app)
	log.Debug("Worker started", "worker_id", w.id)

	for {
		select {
		case job, ok := <-w.jobQueue:
			if !ok {
				// Job queue closed, worker should exit
				log.Debug("Worker shutting down", "worker_id", w.id)
				return
			}

			// Process the job
			err := w.processJob(job)

			// Send result
			w.resultQueue <- WorkerJobResult{
				JobID: job.Id,
				Error: err,
			}

		case <-w.quit:
			// Forced shutdown
			log := logger.GetLogger(w.app)
			log.Debug("Worker force shutdown", "worker_id", w.id)
			return
		}
	}
}

// processJob processes a single job (similar to JobProcessor.ProcessJob but optimized for worker pool)
func (w *Worker) processJob(record *core.Record) error {
	startTime := time.Now()

	// Validate job record
	if err := w.validateJobRecord(record); err != nil {
		return fmt.Errorf("job validation failed: %w", err)
	}

	// Check if job is already reserved by another process
	if w.isJobReserved(record) {
		return fmt.Errorf("job %s is already reserved", record.Id)
	}

	// Reserve the job
	if err := w.reserveJob(record); err != nil {
		return err
	}

	// Parse job data
	jobData, err := ParseJobDataFromRecord(record)
	if err != nil {
		return w.failJob(record, fmt.Errorf("failed to parse job data: %w", err))
	}

	// Get handler
	handler, err := w.registry.GetHandler(jobData.Type)
	if err != nil {
		return w.failJob(record, fmt.Errorf("no handler for job type '%s': %w", jobData.Type, err))
	}

	// Execute job with panic recovery
	var jobErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				jobErr = fmt.Errorf("job handler panicked: %v", r)
				log := logger.GetLogger(w.app)
				log.Error("Job handler panic", "job_id", record.Id, "panic", r)
			}
		}()

		// Create execution context using cronutils
		ctx := cronutils.NewCronExecutionContext(w.app, record.Id)

		jobErr = handler.Handle(ctx, jobData)
	}()

	// Handle result
	if jobErr != nil {
		log := logger.GetLogger(w.app)
		log.Error("Job failed", "job_id", record.Id, "worker_id", w.id, "error", jobErr)
		return w.failJob(record, jobErr)
	}

	// Complete job
	log := logger.GetLogger(w.app)
	log.Debug("Job completed", "job_id", record.Id, "worker_id", w.id, "duration", time.Since(startTime))
	return w.completeJob(record)
}

// Helper methods (similar to JobProcessor but optimized)

func (w *Worker) validateJobRecord(record *core.Record) error {
	if record == nil {
		return fmt.Errorf("job record cannot be nil")
	}
	if record.Id == "" {
		return fmt.Errorf("job record must have a valid ID")
	}
	if record.Collection().Name != QueuesCollection {
		return fmt.Errorf("job record must be from the '%s' collection", QueuesCollection)
	}
	return nil
}

func (w *Worker) isJobReserved(record *core.Record) bool {
	reservedAtStr := record.GetString("reserved_at")
	if reservedAtStr == "" {
		return false
	}

	reservedAt, err := time.Parse(time.RFC3339, reservedAtStr)
	if err != nil {
		return false
	}

	return time.Since(reservedAt) < 5*time.Minute
}

func (w *Worker) reserveJob(record *core.Record) error {
	now := time.Now()
	record.Set("reserved_at", now.Format(time.RFC3339))
	return w.app.Save(record)
}

func (w *Worker) completeJob(record *core.Record) error {
	return w.app.Delete(record)
}

func (w *Worker) failJob(record *core.Record, jobErr error) error {
	currentAttempts := int(record.GetFloat("attempts"))
	record.Set("attempts", currentAttempts+1)
	record.Set("reserved_at", "")

	if err := w.app.Save(record); err != nil {
		log := logger.GetLogger(w.app)
		log.Error("Failed to update failed job", "job_id", record.Id, "error", err)
		return fmt.Errorf("failed to update failed job: %w", err)
	}

	return jobErr
}
