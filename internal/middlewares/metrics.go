package middlewares

import (
	"ims-pocketbase-baas-starter/pkg/metrics"

	"github.com/pocketbase/pocketbase/core"
)

// MetricsMiddleware provides HTTP request metrics collection
type MetricsMiddleware struct {
	provider metrics.MetricsProvider
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(provider metrics.MetricsProvider) *MetricsMiddleware {
	return &MetricsMiddleware{
		provider: provider,
	}
}

// RequireMetricsFunc returns a middleware function that collects HTTP metrics
func (m *MetricsMiddleware) RequireMetricsFunc() func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if m.provider == nil {
			return e.Next()
		}

		// Use the instrumentation helper to wrap the request
		return metrics.InstrumentHTTPHandler(m.provider, e.Request.Method, e.Request.URL.Path, func() error {
			return e.Next()
		})
	}
}

// RequireMetrics returns a middleware function that collects HTTP metrics (convenience method)
func RequireMetrics(provider metrics.MetricsProvider) func(*core.RequestEvent) error {
	middleware := NewMetricsMiddleware(provider)
	return middleware.RequireMetricsFunc()
}

// normalizePath normalizes URL paths for consistent metrics
// This helps reduce cardinality by grouping similar paths
func normalizePath(path string) string {
	// For now, return the path as-is
	// In production, you might want to:
	// - Replace IDs with placeholders (e.g., /api/users/123 -> /api/users/{id})
	// - Group similar paths
	// - Limit path length

	// Simple length limiting to prevent high cardinality
	if len(path) > 100 {
		return path[:100] + "..."
	}

	return path
}

// InstrumentHandler wraps a handler function with metrics collection
func InstrumentHandler(provider metrics.MetricsProvider, handlerName string, handler func(*core.RequestEvent) error) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		if provider == nil {
			return handler(e)
		}

		return metrics.InstrumentHTTPHandler(provider, e.Request.Method, e.Request.URL.Path, func() error {
			return handler(e)
		})
	}
}

// RecordCustomMetric is a helper function to record custom metrics from handlers
func RecordCustomMetric(provider metrics.MetricsProvider, metricName string, value float64, labels map[string]string) {
	if provider == nil {
		return
	}

	metrics.SafeRecordHistogram(provider, metricName, value, labels)
}

// IncrementCustomCounter is a helper function to increment custom counters from handlers
func IncrementCustomCounter(provider metrics.MetricsProvider, metricName string, labels map[string]string) {
	if provider == nil {
		return
	}

	metrics.SafeIncrementCounter(provider, metricName, labels)
}

// SetCustomGauge is a helper function to set custom gauge values from handlers
func SetCustomGauge(provider metrics.MetricsProvider, metricName string, value float64, labels map[string]string) {
	if provider == nil {
		return
	}

	metrics.SafeSetGauge(provider, metricName, value, labels)
}
