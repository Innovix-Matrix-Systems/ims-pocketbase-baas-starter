package metrics

import (
	"context"
	"net/http"
	"time"
)

// MetricsProvider defines the interface for all metrics implementations
type MetricsProvider interface {
	// Counter operations
	IncrementCounter(name string, labels map[string]string)
	IncrementCounterBy(name string, value float64, labels map[string]string)

	// Histogram operations
	RecordHistogram(name string, value float64, labels map[string]string)

	// Gauge operations
	SetGauge(name string, value float64, labels map[string]string)

	// Timing operations
	StartTimer(name string, labels map[string]string) Timer
	RecordDuration(name string, duration time.Duration, labels map[string]string)

	// Provider-specific operations
	GetHandler() http.Handler // For Prometheus /metrics endpoint
	Shutdown(ctx context.Context) error
}

// Timer interface for measuring durations
type Timer interface {
	Stop()                                   // Records the duration since timer creation
	StopWithLabels(labels map[string]string) // Records duration with additional labels
}

// Config holds metrics configuration
type Config struct {
	Provider  string            `json:"provider"`  // "prometheus", "opentelemetry", "disabled"
	Enabled   bool              `json:"enabled"`   // Master enable/disable switch
	Namespace string            `json:"namespace"` // Metrics namespace prefix
	Labels    map[string]string `json:"labels"`    // Global labels applied to all metrics

	// Prometheus-specific config
	Prometheus PrometheusConfig `json:"prometheus"`

	// OpenTelemetry-specific config
	OpenTelemetry OpenTelemetryConfig `json:"opentelemetry"`
}

// PrometheusConfig holds Prometheus-specific settings
type PrometheusConfig struct {
	MetricsPath string    `json:"metrics_path"` // Default: "/metrics"
	Buckets     []float64 `json:"buckets"`      // Histogram buckets
}

// OpenTelemetryConfig holds OpenTelemetry-specific settings
type OpenTelemetryConfig struct {
	Endpoint       string            `json:"endpoint"`        // OTLP endpoint
	Headers        map[string]string `json:"headers"`         // Additional headers
	Insecure       bool              `json:"insecure"`        // Use insecure connection
	ExportInterval time.Duration     `json:"export_interval"` // Export interval
}

// Provider type constants
const (
	ProviderPrometheus    = "prometheus"
	ProviderOpenTelemetry = "opentelemetry"
	ProviderDisabled      = "disabled"
)

// Metric name constants following Prometheus conventions
const (
	// Hook metrics
	MetricHookExecutionDuration = "hook_execution_duration_seconds"
	MetricHookExecutionTotal    = "hook_execution_total"
	MetricHookErrorsTotal       = "hook_errors_total"

	// Handler metrics
	MetricHandlerDuration      = "handler_duration_seconds"
	MetricHandlerRequestsTotal = "handler_requests_total"
	MetricHandlerErrorsTotal   = "handler_errors_total"

	// Job metrics
	MetricJobExecutionDuration = "job_execution_duration_seconds"
	MetricJobExecutionTotal    = "job_execution_total"
	MetricJobErrorsTotal       = "job_errors_total"
	MetricJobQueueSize         = "job_queue_size"

	// Business metrics
	MetricRecordOperationsTotal = "record_operations_total"
	MetricEmailsSentTotal       = "emails_sent_total"
	MetricCacheHitsTotal        = "cache_hits_total"
	MetricCacheMissesTotal      = "cache_misses_total"

	// HTTP metrics
	MetricHTTPRequestDuration = "http_request_duration_seconds"
	MetricHTTPRequestsTotal   = "http_requests_total"
)

// Standard label keys
const (
	LabelHookType    = "hook_type"
	LabelCollection  = "collection"
	LabelOperation   = "operation"
	LabelStatus      = "status"
	LabelJobType     = "job_type"
	LabelHandlerName = "handler"
	LabelMethod      = "method"
	LabelPath        = "path"
	LabelStatusCode  = "status_code"
	LabelError       = "error"
	LabelSuccess     = "success"
)

// Default configuration values
const (
	DefaultNamespace      = "ims_pocketbase"
	DefaultMetricsPath    = "/metrics"
	DefaultExportInterval = 30 * time.Second
	DefaultOTLPEndpoint   = "http://localhost:4317"
)

// Default histogram buckets for duration metrics (in seconds)
var DefaultHistogramBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
}
