package metrics

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"METRICS_PROVIDER", "METRICS_ENABLED", "METRICS_NAMESPACE", "METRICS_LABELS",
		"METRICS_PATH", "METRICS_HISTOGRAM_BUCKETS", "OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_HEADERS", "OTEL_EXPORTER_OTLP_INSECURE", "OTEL_METRIC_EXPORT_INTERVAL",
	}

	for _, env := range envVars {
		originalEnv[env] = os.Getenv(env)
		os.Unsetenv(env)
	}

	// Restore environment after test
	defer func() {
		for _, env := range envVars {
			if val, exists := originalEnv[env]; exists {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: Config{
				Provider:  ProviderDisabled,
				Enabled:   false,
				Namespace: DefaultNamespace,
				Labels:    map[string]string{},
				Prometheus: PrometheusConfig{
					MetricsPath: DefaultMetricsPath,
					Buckets:     DefaultHistogramBuckets,
				},
				OpenTelemetry: OpenTelemetryConfig{
					Endpoint:       DefaultOTLPEndpoint,
					Headers:        map[string]string{},
					Insecure:       true,
					ExportInterval: DefaultExportInterval,
				},
			},
		},
		{
			name: "prometheus configuration",
			envVars: map[string]string{
				"METRICS_PROVIDER":  "prometheus",
				"METRICS_ENABLED":   "true",
				"METRICS_NAMESPACE": "test_app",
				"METRICS_PATH":      "/custom-metrics",
			},
			expected: Config{
				Provider:  ProviderPrometheus,
				Enabled:   true,
				Namespace: "test_app",
				Labels:    map[string]string{},
				Prometheus: PrometheusConfig{
					MetricsPath: "/custom-metrics",
					Buckets:     DefaultHistogramBuckets,
				},
				OpenTelemetry: OpenTelemetryConfig{
					Endpoint:       DefaultOTLPEndpoint,
					Headers:        map[string]string{},
					Insecure:       true,
					ExportInterval: DefaultExportInterval,
				},
			},
		},
		{
			name: "opentelemetry configuration",
			envVars: map[string]string{
				"METRICS_PROVIDER":            "opentelemetry",
				"METRICS_ENABLED":             "true",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://otel-collector:4317",
				"OTEL_EXPORTER_OTLP_HEADERS":  "api-key=secret,x-tenant=test",
				"OTEL_EXPORTER_OTLP_INSECURE": "false",
				"OTEL_METRIC_EXPORT_INTERVAL": "60s",
			},
			expected: Config{
				Provider:  ProviderOpenTelemetry,
				Enabled:   true,
				Namespace: DefaultNamespace,
				Labels:    map[string]string{},
				Prometheus: PrometheusConfig{
					MetricsPath: DefaultMetricsPath,
					Buckets:     DefaultHistogramBuckets,
				},
				OpenTelemetry: OpenTelemetryConfig{
					Endpoint: "http://otel-collector:4317",
					Headers: map[string]string{
						"api-key":  "secret",
						"x-tenant": "test",
					},
					Insecure:       false,
					ExportInterval: 60 * time.Second,
				},
			},
		},
		{
			name: "disabled when enabled is false",
			envVars: map[string]string{
				"METRICS_PROVIDER": "prometheus",
				"METRICS_ENABLED":  "false",
			},
			expected: Config{
				Provider:  ProviderDisabled,
				Enabled:   false,
				Namespace: DefaultNamespace,
				Labels:    map[string]string{},
				Prometheus: PrometheusConfig{
					MetricsPath: DefaultMetricsPath,
					Buckets:     DefaultHistogramBuckets,
				},
				OpenTelemetry: OpenTelemetryConfig{
					Endpoint:       DefaultOTLPEndpoint,
					Headers:        map[string]string{},
					Insecure:       true,
					ExportInterval: DefaultExportInterval,
				},
			},
		},
		{
			name: "invalid provider defaults to disabled",
			envVars: map[string]string{
				"METRICS_PROVIDER": "invalid",
				"METRICS_ENABLED":  "true",
			},
			expected: Config{
				Provider:  ProviderDisabled,
				Enabled:   true,
				Namespace: DefaultNamespace,
				Labels:    map[string]string{},
				Prometheus: PrometheusConfig{
					MetricsPath: DefaultMetricsPath,
					Buckets:     DefaultHistogramBuckets,
				},
				OpenTelemetry: OpenTelemetryConfig{
					Endpoint:       DefaultOTLPEndpoint,
					Headers:        map[string]string{},
					Insecure:       true,
					ExportInterval: DefaultExportInterval,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := LoadConfig()

			if !reflect.DeepEqual(config, tt.expected) {
				t.Errorf("LoadConfig() = %+v, expected %+v", config, tt.expected)
			}

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	expected := Config{
		Provider:  ProviderDisabled,
		Enabled:   false,
		Namespace: DefaultNamespace,
		Labels:    map[string]string{},
		Prometheus: PrometheusConfig{
			MetricsPath: DefaultMetricsPath,
			Buckets:     DefaultHistogramBuckets,
		},
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint:       DefaultOTLPEndpoint,
			Headers:        map[string]string{},
			Insecure:       true,
			ExportInterval: DefaultExportInterval,
		},
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("NewConfig() = %+v, expected %+v", config, expected)
	}
}

func TestConfigMethods(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		tests  map[string]bool
	}{
		{
			name: "disabled config",
			config: Config{
				Provider: ProviderDisabled,
				Enabled:  false,
			},
			tests: map[string]bool{
				"IsEnabled":       false,
				"IsPrometheus":    false,
				"IsOpenTelemetry": false,
			},
		},
		{
			name: "prometheus config",
			config: Config{
				Provider: ProviderPrometheus,
				Enabled:  true,
			},
			tests: map[string]bool{
				"IsEnabled":       true,
				"IsPrometheus":    true,
				"IsOpenTelemetry": false,
			},
		},
		{
			name: "opentelemetry config",
			config: Config{
				Provider: ProviderOpenTelemetry,
				Enabled:  true,
			},
			tests: map[string]bool{
				"IsEnabled":       true,
				"IsPrometheus":    false,
				"IsOpenTelemetry": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.config.IsEnabled(); got != tt.tests["IsEnabled"] {
				t.Errorf("Config.IsEnabled() = %v, expected %v", got, tt.tests["IsEnabled"])
			}
			if got := tt.config.IsPrometheus(); got != tt.tests["IsPrometheus"] {
				t.Errorf("Config.IsPrometheus() = %v, expected %v", got, tt.tests["IsPrometheus"])
			}
			if got := tt.config.IsOpenTelemetry(); got != tt.tests["IsOpenTelemetry"] {
				t.Errorf("Config.IsOpenTelemetry() = %v, expected %v", got, tt.tests["IsOpenTelemetry"])
			}
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single label",
			input: "env=production",
			expected: map[string]string{
				"env": "production",
			},
		},
		{
			name:  "multiple labels",
			input: "env=production,service=api,version=1.0",
			expected: map[string]string{
				"env":     "production",
				"service": "api",
				"version": "1.0",
			},
		},
		{
			name:     "invalid format",
			input:    "invalid,format=value",
			expected: map[string]string{"format": "value"},
		},
		{
			name:     "empty values ignored",
			input:    "key1=,=value2,key3=value3",
			expected: map[string]string{"key3": "value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLabels(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseLabels(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseHistogramBuckets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []float64
	}{
		{
			name:     "empty string returns default",
			input:    "",
			expected: DefaultHistogramBuckets,
		},
		{
			name:     "valid buckets",
			input:    "0.1,0.5,1.0,5.0",
			expected: []float64{0.1, 0.5, 1.0, 5.0},
		},
		{
			name:     "invalid values ignored",
			input:    "0.1,invalid,1.0,another",
			expected: []float64{0.1, 1.0},
		},
		{
			name:     "all invalid returns default",
			input:    "invalid,values,only",
			expected: DefaultHistogramBuckets,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseHistogramBuckets(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseHistogramBuckets(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseExportInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "empty string returns default",
			input:    "",
			expected: DefaultExportInterval,
		},
		{
			name:     "valid duration",
			input:    "60s",
			expected: 60 * time.Second,
		},
		{
			name:     "invalid duration returns default",
			input:    "invalid",
			expected: DefaultExportInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExportInterval(tt.input)
			if result != tt.expected {
				t.Errorf("parseExportInterval(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
