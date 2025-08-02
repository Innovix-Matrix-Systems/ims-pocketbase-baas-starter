package jobutils

import (
	"ims-pocketbase-baas-starter/pkg/cronutils"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase"
)

// JobRegistry manages registered job handlers with thread-safe operations
type JobRegistry struct {
	handlers map[string]JobHandler
	mu       sync.RWMutex
}

// JobProcessor coordinates job execution and queue management
type JobProcessor struct {
	app      *pocketbase.PocketBase
	registry *JobRegistry
}

// JobHandler defines the interface that all job handlers must implement
type JobHandler interface {
	// Handle processes a job with the given payload and returns an error if processing fails
	Handle(ctx *cronutils.CronExecutionContext, job *JobData) error

	// GetJobType returns the job type this handler processes
	GetJobType() string
}

// JobData represents standardized job data extracted from queue records
type JobData struct {
	ID          string         // Job ID from queues table
	Name        string         // Job name
	Description string         // Job description
	Type        string         // Job type extracted from payload
	Payload     map[string]any // Parsed JSON payload
	Attempts    int            // Current attempt count
	ReservedAt  *time.Time     // When job was reserved
	CreatedAt   time.Time      // When job was created
	UpdatedAt   time.Time      // When job was updated
}

// JobResult represents the result of job execution
type JobResult struct {
	Success   bool   // Whether the job completed successfully
	Error     error  // Error if job failed
	Retryable bool   // Whether the job should be retried on failure
	Message   string // Additional message about the result
}

// BaseJobPayload represents the common structure for all job types
type BaseJobPayload struct {
	Type    string         `json:"type"`
	Data    any            `json:"data"`
	Options map[string]any `json:"options"`
}

// UserExportJobData represents the data section for user export jobs
type UserExportJobData struct {
	Format string   `json:"format"`
	Fields []string `json:"fields"`
	UserID string   `json:"user_id"`
}

// UserExportJobOptions represents the options section for user export jobs
type UserExportJobOptions struct {
	FilenamePrefix string `json:"filename_prefix"`
	StoreResult    bool   `json:"store_result"`
	ResultExpiry   string `json:"result_expiry"`
}

// UserExportJobPayload represents the complete payload for user export jobs
type UserExportJobPayload struct {
	Type    string               `json:"type"`
	Data    UserExportJobData    `json:"data"`
	Options UserExportJobOptions `json:"options"`
}

// EmailJobData represents the data section for email jobs
type EmailJobData struct {
	To        string         `json:"to"`
	Subject   string         `json:"subject"`
	Template  string         `json:"template"`
	Variables map[string]any `json:"variables"`
}

// EmailJobOptions represents the options section for email jobs
type EmailJobOptions struct {
	RetryCount int `json:"retry_count"`
	Timeout    int `json:"timeout"`
}

// EmailJobPayload represents the complete payload for email jobs
type EmailJobPayload struct {
	Type    string          `json:"type"`
	Data    EmailJobData    `json:"data"`
	Options EmailJobOptions `json:"options"`
}

// DataProcessingJobData represents the data section for data processing jobs
type DataProcessingJobData struct {
	Operation string `json:"operation"`
	Source    string `json:"source"`
	Target    string `json:"target"`
}

// DataProcessingJobOptions represents the options section for data processing jobs
type DataProcessingJobOptions struct {
	Timeout int `json:"timeout,omitempty"`
}

// DataProcessingJobPayload represents the complete payload for data processing jobs
type DataProcessingJobPayload struct {
	Type    string                   `json:"type"`
	Data    DataProcessingJobData    `json:"data"`
	Options DataProcessingJobOptions `json:"options"`
}

// BaseJobResultData represents common result data structure
type BaseJobResultData struct {
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// FileExportResult represents the result data for file export jobs
type FileExportResult struct {
	BaseJobResultData
	ExportRecordId string `json:"export_record_id"` // ID of export_files record
	FileName       string `json:"file_name"`
	FileSize       int64  `json:"file_size"`
	RecordCount    int    `json:"record_count"`
	ContentType    string `json:"content_type"`
}

// EmailResult represents the result data for email jobs
type EmailResult struct {
	BaseJobResultData
	MessageId   string     `json:"message_id,omitempty"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	Recipients  []string   `json:"recipients,omitempty"`
}

// DataProcessingResult represents the result data for data processing jobs
type DataProcessingResult struct {
	BaseJobResultData
	ProcessedRecords int    `json:"processed_records"`
	OutputLocation   string `json:"output_location,omitempty"`
}

// Job status constants
const (
	JobStatusQueued     = "queued"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

// Job type constants
const (
	JobTypeDataProcessing = "data_processing"
	JobTypeEmail          = "email"
)

// Data processing operation constants
const (
	DataProcessingOperationExport    = "export"
	DataProcessingOperationImport    = "import"
	DataProcessingOperationAggregate = "aggregate"
	DataProcessingOperationTransform = "transform"
)

const (
	DataProcessingFileCSV  = "csv"
	DataProcessingFileXLSX = "xlsx"
	DataProcessingFileJSON = "json"
	DataProcessingFilePDF  = "pdf"
)

const (
	DataProcessingCollectionUsers = "users"
)
