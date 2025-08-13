package metrics

import (
	"strconv"
	"strings"
	"time"

	"ims-pocketbase-baas-starter/pkg/common"
)

// LoadConfig loads metrics configuration from environment variables
func LoadConfig() Config {
	config := Config{
		Provider:  common.GetEnv("METRICS_PROVIDER", ProviderDisabled),
		Enabled:   common.GetEnvBool("METRICS_ENABLED", false),
		Namespace: common.GetEnv("METRICS_NAMESPACE", DefaultNamespace),
		Labels:    parseLabels(common.GetEnv("METRICS_LABELS", "")),
		Prometheus: PrometheusConfig{
			MetricsPath: common.GetEnv("METRICS_PATH", DefaultMetricsPath),
			Buckets:     parseHistogramBuckets(common.GetEnv("METRICS_HISTOGRAM_BUCKETS", "")),
		},
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint:       common.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", DefaultOTLPEndpoint),
			Headers:        parseHeaders(common.GetEnv("OTEL_EXPORTER_OTLP_HEADERS", "")),
			Insecure:       common.GetEnvBool("OTEL_EXPORTER_OTLP_INSECURE", true),
			ExportInterval: parseExportInterval(common.GetEnv("OTEL_METRIC_EXPORT_INTERVAL", "30s")),
		},
	}

	// If provider is set but enabled is false, disable the provider
	if !config.Enabled {
		config.Provider = ProviderDisabled
	}

	// Validate provider type
	if !isValidProvider(config.Provider) {
		config.Provider = ProviderDisabled
	}

	return config
}

// NewConfig creates a new configuration with default values
func NewConfig() Config {
	return Config{
		Provider:  ProviderDisabled,
		Enabled:   false,
		Namespace: DefaultNamespace,
		Labels:    make(map[string]string),
		Prometheus: PrometheusConfig{
			MetricsPath: DefaultMetricsPath,
			Buckets:     DefaultHistogramBuckets,
		},
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint:       DefaultOTLPEndpoint,
			Headers:        make(map[string]string),
			Insecure:       true,
			ExportInterval: DefaultExportInterval,
		},
	}
}

// IsEnabled returns true if metrics collection is enabled
func (c Config) IsEnabled() bool {
	return c.Enabled && c.Provider != ProviderDisabled
}

// IsPrometheus returns true if Prometheus provider is configured
func (c Config) IsPrometheus() bool {
	return c.IsEnabled() && c.Provider == ProviderPrometheus
}

// IsOpenTelemetry returns true if OpenTelemetry provider is configured
func (c Config) IsOpenTelemetry() bool {
	return c.IsEnabled() && c.Provider == ProviderOpenTelemetry
}

// GetMetricsPath returns the metrics endpoint path for Prometheus
func (c Config) GetMetricsPath() string {
	if c.Prometheus.MetricsPath == "" {
		return DefaultMetricsPath
	}
	return c.Prometheus.MetricsPath
}

// GetHistogramBuckets returns histogram buckets with defaults if empty
func (c Config) GetHistogramBuckets() []float64 {
	if len(c.Prometheus.Buckets) == 0 {
		return DefaultHistogramBuckets
	}
	return c.Prometheus.Buckets
}

// GetExportInterval returns export interval with default if zero
func (c Config) GetExportInterval() time.Duration {
	if c.OpenTelemetry.ExportInterval == 0 {
		return DefaultExportInterval
	}
	return c.OpenTelemetry.ExportInterval
}

// parseLabels parses comma-separated key=value pairs into a map
// Format: "key1=value1,key2=value2"
func parseLabels(labelsStr string) map[string]string {
	labels := make(map[string]string)
	if labelsStr == "" {
		return labels
	}

	pairs := strings.Split(labelsStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			if key != "" && value != "" {
				labels[key] = value
			}
		}
	}

	return labels
}

// parseHeaders parses comma-separated key=value pairs into a map for OTLP headers
// Format: "api-key=secret,x-custom=value"
func parseHeaders(headersStr string) map[string]string {
	return parseLabels(headersStr) // Same format as labels
}

// parseHistogramBuckets parses comma-separated float values into a slice
// Format: "0.001,0.005,0.01,0.025,0.05,0.1,0.25,0.5,1.0,2.5,5.0,10.0"
func parseHistogramBuckets(bucketsStr string) []float64 {
	if bucketsStr == "" {
		return DefaultHistogramBuckets
	}

	bucketStrs := strings.Split(bucketsStr, ",")
	buckets := make([]float64, 0, len(bucketStrs))

	for _, bucketStr := range bucketStrs {
		if bucket, err := strconv.ParseFloat(strings.TrimSpace(bucketStr), 64); err == nil {
			buckets = append(buckets, bucket)
		}
	}

	// Return default if no valid buckets were parsed
	if len(buckets) == 0 {
		return DefaultHistogramBuckets
	}

	return buckets
}

// parseExportInterval parses duration string with default fallback
func parseExportInterval(intervalStr string) time.Duration {
	if intervalStr == "" {
		return DefaultExportInterval
	}

	if duration, err := time.ParseDuration(intervalStr); err == nil {
		return duration
	}

	return DefaultExportInterval
}

// isValidProvider checks if the provider type is valid
func isValidProvider(provider string) bool {
	switch provider {
	case ProviderPrometheus, ProviderOpenTelemetry, ProviderDisabled:
		return true
	default:
		return false
	}
}
