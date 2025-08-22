# Cron Jobs and Job Queue System

This document covers the cron job scheduling system and job queue processing functionality in the IMS PocketBase BaaS Starter.

## Overview

The application includes two complementary systems for background task processing:

1. **Cron Job System** - Scheduled tasks that run at specific intervals
2. **Job Queue System** - Dynamic job processing with handlers for different job types

## Cron Job System

### Architecture

The cron system is organized in the `internal/crons/` directory and follows a centralized registration pattern:

```
internal/
├── crons/
│   └── crons.go          # Cron registration and configuration
├── handlers/
│   └── cron/
│       └── system.go     # Cron job handlers
└── jobs/
    └── manager.go        # Job processor management
```

### Configuration

Cron jobs are defined in `internal/crons/crons.go` with the following structure:

```go
type Cron struct {
    ID          string // Unique identifier
    CronExpr    string // Cron expression (e.g., "* * * * *")
    Handler     func() // Function to execute
    Enabled     bool   // Whether the job is enabled
    Description string // Human-readable description
}
```

### Built-in Cron Jobs

#### System Queue Processor

- **ID**: `system_queue`
- **Schedule**: Every minute (`* * * * *`)
- **Function**: Processes jobs from the database queue
- **Environment Variable**: `ENABLE_SYSTEM_QUEUE_CRON` (default: enabled)

### Adding New Cron Jobs

1. **Define the cron job** in `internal/crons/crons.go`:

```go
{
    ID:          "my_custom_job",
    CronExpr:    "0 2 * * *", // Daily at 2 AM
    Handler:     cronutils.WithRecovery(app, "my_custom_job", func() { 
        myCustomHandler(app) 
    }),
    Enabled:     os.Getenv("ENABLE_MY_CUSTOM_JOB") != "false",
    Description: "My custom scheduled task",
}
```

2. **Create the handler function** in `internal/handlers/cron/`:

```go
func myCustomHandler(app *pocketbase.PocketBase) {
    ctx := cronutils.NewCronExecutionContext(app, "my_custom_job")
    ctx.LogStart("Starting my custom job")
    
    // Your job logic here
    
    ctx.LogEnd("My custom job completed")
}
```

### Environment Variables

- `ENABLE_SYSTEM_QUEUE_CRON` - Enable/disable system queue processing (default: `true`)

## Job Queue System

### Architecture

The job queue system processes dynamic jobs stored in the database with different handlers for different job types:

```
internal/
├── jobs/
│   └── manager.go        # Job processor singleton
├── handlers/
│   └── jobs/
│       ├── registry.go   # Handler registration
│       ├── email.go      # Email job handler
│       └── data.go       # Data processing handler
└── pkg/
    ├── jobutils/
    │   └── processor.go  # Job processing utilities
    └── cronutils/
        └── utils.go      # Cron execution utilities
```

### Database Schema

Jobs are stored in the `queues` table with the following structure:

```json
{
  "id": "unique_job_id",
  "name": "job_name",
  "description": "Job description",
  "payload": {
    "type": "email",
    "data": {
      "to": "user@example.com",
      "subject": "Welcome {{name}}!",
      "template": "welcome_email",
      "variables": {
        "name": "John Doe",
        "company": "Acme Corp"
      }
    },
    "options": {
      "retry_count": 3,
      "timeout": 30
    }
  },
  "attempts": 0,
  "reserved_at": null,
  "created": "2025-01-01T00:00:00Z",
  "updated": "2025-01-01T00:00:00Z"
}
```

### Job Processing Flow

1. **Cron Trigger** - System queue cron runs every minute
2. **Job Fetching** - Fetches unreserved jobs from database
3. **Job Reservation** - Updates `reserved_at` to prevent duplicate processing
4. **Handler Routing** - Routes job to appropriate handler based on `type`
5. **Job Execution** - Handler processes the job
6. **Completion** - Successful jobs are deleted, failed jobs increment `attempts`

### Built-in Job Handlers

#### Email Job Handler

Processes email jobs with template variable replacement:

```json
{
  "type": "email",
  "data": {
    "to": "user@example.com",
    "subject": "Welcome {{name}}!",
    "template": "welcome_email",
    "variables": {
      "name": "John Doe",
      "company": "Acme Corp"
    }
  }
}
```

#### Data Processing Job Handler

Handles various data processing operations:

```json
{
  "type": "data_processing",
  "data": {
    "operation": "transform|aggregate|export|import",
    "source": "source_identifier",
    "target": "target_identifier"
  }
}
```

### Adding New Job Handlers

1. **Create the handler** in `internal/handlers/jobs/`:

```go
type MyJobHandler struct {
    app *pocketbase.PocketBase
}

func NewMyJobHandler(app *pocketbase.PocketBase) *MyJobHandler {
    return &MyJobHandler{app: app}
}

func (h *MyJobHandler) Handle(ctx *cronutils.CronExecutionContext, job *jobutils.JobData) error {
    ctx.LogStart(fmt.Sprintf("Processing my job: %s", job.ID))
    
    // Extract job data
    jobData, ok := job.Payload["data"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("invalid job payload structure")
    }
    
    // Process the job
    // Your job logic here
    
    ctx.LogEnd("My job processed successfully")
    return nil
}

func (h *MyJobHandler) GetJobType() string {
    return "my_job_type"
}
```

2. **Register the handler** in `internal/jobs/jobs.go`:

```go
// In the jobHandlers slice in RegisterJobs function
{
    Type:        "my_job_type",
    Handler:     jobs.NewMyJobHandler(app),
    Enabled:     true,
    Description: "Process my custom jobs",
},
```

### Job Queue Configuration

Environment variables for job processing:

- `JOB_MAX_WORKERS` - Maximum concurrent workers (default: `5`)
- `JOB_BATCH_SIZE` - Jobs processed per cron run (default: `50`)
- `JOB_MAX_RETRIES` - Maximum retry attempts (default: `3`)
- `JOB_TIMEOUT_SECONDS` - Job timeout in seconds (default: `30`)
- `JOB_RESERVATION_TIMEOUT` - Job reservation timeout in minutes (default: `5`)

### Adding Jobs to Queue

You can add jobs to the queue through the PocketBase API or programmatically:

#### Via API

```bash
curl -X POST http://localhost:8090/api/collections/queues/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "welcome_email",
    "description": "Send welcome email to new user",
    "payload": {
      "type": "email",
      "data": {
        "to": "user@example.com",
        "subject": "Welcome {{name}}!",
        "template": "welcome_email",
        "variables": {
          "name": "John Doe"
        }
      }
    }
  }'
```

#### Programmatically

```go
func addEmailJob(app *pocketbase.PocketBase, to, name string) error {
    collection, err := app.FindCollectionByNameOrId("queues")
    if err != nil {
        return err
    }
    
    record := core.NewRecord(collection)
    record.Set("name", "welcome_email")
    record.Set("description", "Send welcome email")
    record.Set("payload", map[string]interface{}{
        "type": "email",
        "data": map[string]interface{}{
            "to": to,
            "subject": "Welcome {{name}}!",
            "template": "welcome_email",
            "variables": map[string]interface{}{
                "name": name,
            },
        },
    })
    
    return app.Save(record)
}
```

## Monitoring and Debugging

### Logging

Both cron jobs and job queue processing include comprehensive logging:

- **Job Start/End** - Execution timing and status
- **Error Handling** - Detailed error information with context
- **Performance Metrics** - Processing times and success/failure rates
- **Debug Information** - Job data and processing details

### PocketBase Admin UI

- **Cron Jobs** - View and manually trigger cron jobs in Dashboard > Settings > Crons
- **Queue Jobs** - View and manage queue jobs in the `queues` collection
- **Logs** - Monitor job execution through application logs

### Common Issues

#### Jobs Not Processing

1. Check if `ENABLE_SYSTEM_QUEUE_CRON` is enabled
2. Verify cron job is registered and running
3. Check job payload format and required fields
4. Review application logs for errors

#### Job Handler Not Found

1. Ensure handler is registered in `internal/jobs/jobs.go`
2. Verify job `type` matches handler's `GetJobType()`
3. Check for handler registration errors in logs

#### Jobs Stuck in Processing

1. Check `reserved_at` timestamps (jobs auto-recover after 5 minutes)
2. Review job timeout configuration
3. Look for handler panics or infinite loops

## Performance Considerations

### Concurrent Processing

- Jobs are processed concurrently using worker pools
- Default: 5 workers, configurable via `JOB_MAX_WORKERS`
- Workers process jobs within the 1-minute cron interval

### Database Optimization

- Jobs use reservation system to prevent duplicate processing
- Completed jobs are deleted to keep queue table clean
- Failed jobs increment attempt counter for retry logic

### Resource Management

- Job handlers should be stateless and thread-safe
- Long-running jobs should implement timeout handling
- Consider job complexity when setting worker count

## Best Practices

### Job Design

1. **Idempotent Operations** - Jobs should be safe to retry
2. **Timeout Handling** - Implement reasonable timeouts
3. **Error Handling** - Provide clear error messages
4. **Logging** - Include sufficient context for debugging

### Performance

1. **Batch Processing** - Process multiple items per job when possible
2. **Resource Limits** - Avoid memory-intensive operations
3. **Database Connections** - Reuse connections efficiently
4. **Monitoring** - Track job performance and failure rates

### Security

1. **Input Validation** - Validate all job payload data
2. **Access Control** - Ensure proper permissions for job operations
3. **Sensitive Data** - Handle credentials and personal data securely
4. **Rate Limiting** - Prevent job queue flooding

## Testing

### Unit Testing

Test job handlers independently:

```go
func TestEmailJobHandler(t *testing.T) {
    app := testutils.NewTestApp()
    handler := NewEmailJobHandler(app)
    
    jobData := &jobutils.JobData{
        Type: "email",
        Payload: map[string]interface{}{
            "data": map[string]interface{}{
                "to": "test@example.com",
                "subject": "Test",
            },
        },
    }
    
    ctx := cronutils.NewCronExecutionContext(app, "test")
    err := handler.Handle(ctx, jobData)
    
    assert.NoError(t, err)
}
```

### Integration Testing

Test complete job processing flow:

```go
func TestJobProcessing(t *testing.T) {
    app := testutils.NewTestApp()
    processor := jobutils.NewJobProcessor(app)
    
    // Add job to queue
    // Process job
    // Verify results
}
```

## Migration from Legacy Systems

If migrating from other job queue systems:

1. **Map Job Types** - Identify equivalent job types
2. **Payload Format** - Convert to standardized payload structure
3. **Handler Logic** - Implement handlers for existing job types
4. **Configuration** - Update environment variables
5. **Testing** - Thoroughly test job processing

## Troubleshooting

### Debug Mode

Enable debug logging for detailed job processing information:

```bash
# Set log level to debug
export LOG_LEVEL=debug
```

### Manual Job Processing

For debugging, you can manually trigger job processing:

1. Access PocketBase Admin UI
2. Go to Dashboard > Settings > Crons
3. Find "system_queue" and click "Run"

### Queue Inspection

Monitor queue status through the admin UI:

1. Go to Collections > queues
2. Check job status, attempts, and reserved_at timestamps
3. Filter by job type or creation date

For additional support, refer to the main [README](../README.md) or check the application logs.