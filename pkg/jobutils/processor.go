package jobutils

import (
	"encoding/json"
	"fmt"
	"ims-pocketbase-baas-starter/pkg/common"
	"ims-pocketbase-baas-starter/pkg/cronutils"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// NewJobRegistry creates a new job registry
func NewJobRegistry() *JobRegistry {
	return &JobRegistry{
		handlers: make(map[string]JobHandler),
	}
}

// Register adds a job handler to the registry
func (r *JobRegistry) Register(handler JobHandler) error {
	if handler == nil {
		return fmt.Errorf("job handler cannot be nil")
	}

	jobType := handler.GetJobType()
	if jobType == "" {
		return fmt.Errorf("job handler must return a non-empty job type")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.handlers[jobType]; exists {
		return fmt.Errorf("job handler for type '%s' is already registered", jobType)
	}

	r.handlers[jobType] = handler
	return nil
}

// GetHandler retrieves a job handler by job type
func (r *JobRegistry) GetHandler(jobType string) (JobHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	handler, exists := r.handlers[jobType]
	if !exists {
		return nil, fmt.Errorf("no handler registered for job type '%s'", jobType)
	}

	return handler, nil
}

// ListHandlers returns a list of all registered job types
func (r *JobRegistry) ListHandlers() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.handlers))
	for jobType := range r.handlers {
		types = append(types, jobType)
	}

	return types
}

// ParseJobDataFromRecord extracts JobData from a PocketBase record
func ParseJobDataFromRecord(record *core.Record) (*JobData, error) {
	if record == nil {
		return nil, fmt.Errorf("record cannot be nil")
	}

	// Parse the JSON payload
	var payload map[string]any
	payloadStr := record.GetString("payload")
	if payloadStr != "" {
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return nil, fmt.Errorf("failed to parse job payload: %w", err)
		}
	}

	// Extract job type from payload
	jobType := ""
	if payload != nil {
		if typeVal, ok := payload["type"]; ok {
			if typeStr, ok := typeVal.(string); ok {
				jobType = typeStr
			}
		}
	}

	if jobType == "" {
		return nil, fmt.Errorf("job payload must contain a 'type' field")
	}

	// Parse reserved_at timestamp
	var reservedAt *time.Time
	if reservedAtStr := record.GetString("reserved_at"); reservedAtStr != "" {
		if parsed, err := time.Parse(time.RFC3339, reservedAtStr); err == nil {
			reservedAt = &parsed
		}
	}

	return &JobData{
		ID:          record.Id,
		Name:        record.GetString("name"),
		Description: record.GetString("description"),
		Type:        jobType,
		Payload:     payload,
		Attempts:    int(record.GetFloat("attempts")),
		ReservedAt:  reservedAt,
		CreatedAt:   record.GetDateTime("created").Time(),
		UpdatedAt:   record.GetDateTime("updated").Time(),
	}, nil
}

// ValidateJobPayload validates that a job payload has the required structure
func ValidateJobPayload(payload map[string]any) error {
	if payload == nil {
		return fmt.Errorf("job payload cannot be nil")
	}

	// Check for required 'type' field
	typeVal, exists := payload["type"]
	if !exists {
		return fmt.Errorf("job payload must contain a 'type' field")
	}

	typeStr, ok := typeVal.(string)
	if !ok || typeStr == "" {
		return fmt.Errorf("job payload 'type' field must be a non-empty string")
	}

	// Validate 'data' field if present
	if dataVal, exists := payload["data"]; exists {
		if _, ok := dataVal.(map[string]any); !ok {
			return fmt.Errorf("job payload 'data' field must be an object")
		}
	}

	// Validate 'options' field if present
	if optionsVal, exists := payload["options"]; exists {
		if _, ok := optionsVal.(map[string]any); !ok {
			return fmt.Errorf("job payload 'options' field must be an object")
		}
	}

	return nil
}

// NewJobProcessor creates a new job processor with initialized registry and worker pool
func NewJobProcessor(app *pocketbase.PocketBase) *JobProcessor {
	if app == nil {
		panic("NewJobProcessor: app cannot be nil")
	}

	registry := NewJobRegistry()

	return &JobProcessor{
		app:        app,
		registry:   registry,
		workerPool: NewWorkerPool(app, registry, common.GetEnvInt("JOB_MAX_WORKERS", 5)), // Default 5 workers
	}
}

// GetRegistry returns the job registry for handler registration
func (p *JobProcessor) GetRegistry() *JobRegistry {
	return p.registry
}

// RegisterHandler is a convenience method to register a job handler
func (p *JobProcessor) RegisterHandler(handler JobHandler) error {
	return p.registry.Register(handler)
}

// parseJobData extracts and validates job data from a PocketBase record
func (p *JobProcessor) parseJobData(record *core.Record) (*JobData, error) {
	jobData, err := ParseJobDataFromRecord(record)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job data: %w", err)
	}

	// Validate the payload structure
	if err := ValidateJobPayload(jobData.Payload); err != nil {
		return nil, fmt.Errorf("invalid job payload: %w", err)
	}

	return jobData, nil
}

// validateJobRecord performs basic validation on a job record before processing
func (p *JobProcessor) validateJobRecord(record *core.Record) error {
	if record == nil {
		return fmt.Errorf("job record cannot be nil")
	}

	if record.Id == "" {
		return fmt.Errorf("job record must have a valid ID")
	}

	if record.Collection().Name != "queues" {
		return fmt.Errorf("job record must be from the 'queues' collection")
	}

	return nil
}

// reserveJob updates the reserved_at timestamp to mark the job as being processed
func (p *JobProcessor) reserveJob(record *core.Record) error {
	now := time.Now()
	record.Set("reserved_at", now.Format(time.RFC3339))

	if err := p.app.Save(record); err != nil {
		return fmt.Errorf("failed to reserve job %s: %w", record.Id, err)
	}

	p.app.Logger().Debug("Job reserved", "job_id", record.Id, "reserved_at", now)
	return nil
}

// completeJob deletes the job from the queue after successful processing
func (p *JobProcessor) completeJob(record *core.Record) error {
	if err := p.app.Delete(record); err != nil {
		return fmt.Errorf("failed to delete completed job %s: %w", record.Id, err)
	}

	p.app.Logger().Info("Job completed and removed from queue",
		"job_id", record.Id,
		"job_name", record.GetString("name"))
	return nil
}

// failJob increments the attempts counter and handles job failure
func (p *JobProcessor) failJob(record *core.Record, jobErr error) error {
	currentAttempts := int(record.GetFloat("attempts"))
	newAttempts := currentAttempts + 1

	// Update attempts counter and clear reservation
	record.Set("attempts", newAttempts)
	record.Set("reserved_at", "")

	if err := p.app.Save(record); err != nil {
		p.app.Logger().Error("Failed to update failed job record",
			"job_id", record.Id,
			"error", err)
		return fmt.Errorf("failed to update failed job %s: %w", record.Id, err)
	}

	p.app.Logger().Error("Job failed",
		"job_id", record.Id,
		"job_name", record.GetString("name"),
		"attempts", newAttempts,
		"error", jobErr)

	return jobErr
}

// isJobReserved checks if a job is currently reserved by another process
func (p *JobProcessor) isJobReserved(record *core.Record) bool {
	reservedAtStr := record.GetString("reserved_at")
	if reservedAtStr == "" {
		return false
	}

	reservedAt, err := time.Parse(time.RFC3339, reservedAtStr)
	if err != nil {
		// If we can't parse the timestamp, consider it not reserved
		return false
	}

	// Consider job reserved if reservation is less than 5 minutes old
	reservationTimeout := 5 * time.Minute
	return time.Since(reservedAt) < reservationTimeout
}

// ProcessJob processes a single job with complete lifecycle management
func (p *JobProcessor) ProcessJob(record *core.Record) error {
	// Step 1: Validate the job record
	if err := p.validateJobRecord(record); err != nil {
		return fmt.Errorf("job validation failed: %w", err)
	}

	// Step 2: Check if job is already reserved
	if p.isJobReserved(record) {
		return fmt.Errorf("job %s is already reserved by another process", record.Id)
	}

	// Step 3: Reserve the job
	if err := p.reserveJob(record); err != nil {
		return err
	}

	// Step 4: Parse job data
	jobData, err := p.parseJobData(record)
	if err != nil {
		return p.failJob(record, fmt.Errorf("failed to parse job data: %w", err))
	}

	// Step 5: Get appropriate handler
	handler, err := p.registry.GetHandler(jobData.Type)
	if err != nil {
		return p.failJob(record, fmt.Errorf("no handler found for job type '%s': %w", jobData.Type, err))
	}

	// Step 6: Create execution context
	ctx := cronutils.NewCronExecutionContext(p.app, record.Id)

	// Step 7: Execute job with panic recovery
	var jobErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				jobErr = fmt.Errorf("job handler panicked: %v", r)
				ctx.LogError(jobErr, "Job handler panic recovered")
			}
		}()

		ctx.LogStart(fmt.Sprintf("Processing %s job: %s", jobData.Type, jobData.Name))
		jobErr = handler.Handle(ctx, jobData)
	}()

	// Step 8: Handle job result
	if jobErr != nil {
		ctx.LogError(jobErr, "Job processing failed")
		return p.failJob(record, jobErr)
	}

	// Step 9: Complete job successfully
	ctx.LogEnd("Job processed successfully")
	return p.completeJob(record)
}

// ProcessJobsConcurrently processes multiple jobs concurrently using the persistent worker pool
func (p *JobProcessor) ProcessJobsConcurrently(records []*core.Record, maxWorkers int) []error {
	if len(records) == 0 {
		return nil
	}

	// Use the persistent worker pool for better performance
	if p.workerPool != nil {
		return p.workerPool.ProcessJobs(records)
	}

	// Fallback to the old method if worker pool is not available
	return p.processJobsConcurrentlyFallback(records, maxWorkers)
}

// processJobsConcurrentlyFallback is the original implementation as fallback
func (p *JobProcessor) processJobsConcurrentlyFallback(records []*core.Record, maxWorkers int) []error {
	if len(records) == 0 {
		return nil
	}

	// Ensure we don't create more workers than jobs
	if maxWorkers > len(records) {
		maxWorkers = len(records)
	}

	// Create channels for job distribution and error collection
	jobChan := make(chan *core.Record, len(records))
	errorChan := make(chan error, len(records))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			p.app.Logger().Debug("Job worker started", "worker_id", workerID)

			for record := range jobChan {
				err := p.ProcessJob(record)
				errorChan <- err // Send error (or nil) to error channel
			}

			p.app.Logger().Debug("Job worker finished", "worker_id", workerID)
		}(i)
	}

	// Send all jobs to workers
	for _, record := range records {
		jobChan <- record
	}
	close(jobChan)

	// Wait for all workers to complete and close error channel
	go func() {
		wg.Wait()
		close(errorChan)
	}()

	// Collect all errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	// Log processing summary
	successCount := 0
	failureCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		} else {
			failureCount++
		}
	}

	p.app.Logger().Info("Concurrent job processing completed",
		"total_jobs", len(records),
		"successful", successCount,
		"failed", failureCount,
		"workers", maxWorkers)

	return errors
}

// ProcessJobs processes multiple jobs sequentially (fallback method)
func (p *JobProcessor) ProcessJobs(records []*core.Record) []error {
	errors := make([]error, len(records))

	for i, record := range records {
		errors[i] = p.ProcessJob(record)
	}

	return errors
}
