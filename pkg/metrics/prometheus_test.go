package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewPrometheusProvider(t *testing.T) {
	config := Config{
		Provider:  ProviderPrometheus,
		Enabled:   true,
		Namespace: "test_app",
		Labels: map[string]string{
			"env": "test",
		},
		Prometheus: PrometheusConfig{
			MetricsPath: "/metrics",
			Buckets:     []float64{0.1, 0.5, 1.0, 5.0},
		},
	}

	provider := NewPrometheusProvider(config)

	if provider == nil {
		t.Fatal("NewPrometheusProvider() returned nil")
	}

	if provider.config.Namespace != "test_app" {
		t.Errorf("Expected namespace 'test_app', got '%s'", provider.config.Namespace)
	}

	if provider.registry == nil {
		t.Error("Registry should not be nil")
	}

	// Verify it implements MetricsProvider interface
	var _ MetricsProvider = provider
}

func TestPrometheusProviderCounters(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
		Labels:    map[string]string{"global": "value"},
	})

	labels := map[string]string{
		"method": "GET",
		"status": "200",
	}

	// Test IncrementCounter
	provider.IncrementCounter("http_requests_total", labels)
	provider.IncrementCounter("http_requests_total", labels)

	// Test IncrementCounterBy
	provider.IncrementCounterBy("http_requests_total", 5.0, labels)

	// Verify metrics are created
	if len(provider.counters) != 1 {
		t.Errorf("Expected 1 counter, got %d", len(provider.counters))
	}

	// Test with nil labels
	provider.IncrementCounter("test_counter", nil)

	// Test with empty labels
	provider.IncrementCounter("test_counter", map[string]string{})
}

func TestPrometheusProviderHistograms(t *testing.T) {
	config := Config{
		Namespace: "test",
		Prometheus: PrometheusConfig{
			Buckets: []float64{0.1, 0.5, 1.0, 5.0},
		},
	}
	provider := NewPrometheusProvider(config)

	labels := map[string]string{
		"handler": "api",
	}

	// Test RecordHistogram
	provider.RecordHistogram("request_duration_seconds", 0.25, labels)
	provider.RecordHistogram("request_duration_seconds", 1.5, labels)

	// Test RecordDuration
	provider.RecordDuration("processing_time", 500*time.Millisecond, labels)

	// Verify metrics are created
	if len(provider.histograms) != 2 {
		t.Errorf("Expected 2 histograms, got %d", len(provider.histograms))
	}
}

func TestPrometheusProviderGauges(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	labels := map[string]string{
		"queue": "default",
	}

	// Test SetGauge
	provider.SetGauge("queue_size", 10.0, labels)
	provider.SetGauge("queue_size", 15.0, labels) // Update same gauge
	provider.SetGauge("active_connections", 5.0, nil)

	// Verify metrics are created
	if len(provider.gauges) != 2 {
		t.Errorf("Expected 2 gauges, got %d", len(provider.gauges))
	}
}

func TestPrometheusTimer(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
		Prometheus: PrometheusConfig{
			Buckets: DefaultHistogramBuckets,
		},
	})

	labels := map[string]string{
		"operation": "test",
	}

	// Test StartTimer and Stop
	timer := provider.StartTimer("operation_duration_seconds", labels)
	if timer == nil {
		t.Fatal("StartTimer() returned nil")
	}

	// Verify it implements Timer interface
	var _ Timer = timer

	// Simulate some work
	time.Sleep(10 * time.Millisecond)
	timer.Stop()

	// Test StopWithLabels (additional labels are ignored in Prometheus)
	timer2 := provider.StartTimer("operation_duration_seconds", labels)
	time.Sleep(5 * time.Millisecond)
	timer2.StopWithLabels(map[string]string{
		"result": "success", // This will be ignored due to Prometheus limitations
	})

	// Test with nil additional labels
	timer3 := provider.StartTimer("operation_duration_seconds", labels)
	timer3.StopWithLabels(nil)

	// Verify histogram was created
	if len(provider.histograms) != 1 {
		t.Errorf("Expected 1 histogram, got %d", len(provider.histograms))
	}
}

func TestPrometheusProviderHTTPHandler(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	// Add some metrics
	provider.IncrementCounter("test_counter", map[string]string{"label": "value"})
	provider.SetGauge("test_gauge", 42.0, nil)

	handler := provider.GetHandler()
	if handler == nil {
		t.Fatal("GetHandler() returned nil")
	}

	// Test HTTP handler
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()

	// Check for Prometheus format
	if !strings.Contains(body, "test_test_counter") {
		t.Error("Response should contain counter metric")
	}

	if !strings.Contains(body, "test_test_gauge") {
		t.Error("Response should contain gauge metric")
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") && !strings.Contains(contentType, "application/openmetrics-text") {
		t.Errorf("Unexpected content type: %s", contentType)
	}
}

func TestPrometheusProviderShutdown(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	// Add some metrics
	provider.IncrementCounter("test_counter", nil)
	provider.SetGauge("test_gauge", 1.0, nil)

	// Verify metrics exist
	if len(provider.counters) == 0 || len(provider.gauges) == 0 {
		t.Error("Expected metrics to be created")
	}

	// Test shutdown
	ctx := context.Background()
	err := provider.Shutdown(ctx)

	if err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}

	// Verify internal state is cleared
	if len(provider.counters) != 0 {
		t.Error("Counters should be cleared after shutdown")
	}

	if len(provider.gauges) != 0 {
		t.Error("Gauges should be cleared after shutdown")
	}

	if len(provider.histograms) != 0 {
		t.Error("Histograms should be cleared after shutdown")
	}
}

func TestPrometheusProviderConcurrency(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	const numGoroutines = 50
	const operationsPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	// Launch multiple goroutines performing operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Goroutine %d panicked: %v", id, r)
				}
				done <- true
			}()

			labels := map[string]string{
				"goroutine": string(rune(id)),
			}

			for j := 0; j < operationsPerGoroutine; j++ {
				provider.IncrementCounter("concurrent_counter", labels)
				provider.RecordHistogram("concurrent_histogram", float64(j), labels)
				provider.SetGauge("concurrent_gauge", float64(j), labels)

				timer := provider.StartTimer("concurrent_timer", labels)
				timer.Stop()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify metrics were created
	if len(provider.counters) == 0 {
		t.Error("Expected counters to be created")
	}

	if len(provider.histograms) == 0 {
		t.Error("Expected histograms to be created")
	}

	if len(provider.gauges) == 0 {
		t.Error("Expected gauges to be created")
	}
}

func TestPrometheusProviderGlobalLabels(t *testing.T) {
	config := Config{
		Namespace: "test",
		Labels: map[string]string{
			"env":     "test",
			"service": "api",
		},
	}
	provider := NewPrometheusProvider(config)

	// Add metric with local labels
	localLabels := map[string]string{
		"method": "GET",
		"env":    "override", // This should override global label
	}

	provider.IncrementCounter("requests_total", localLabels)

	// Test that handler includes metrics
	handler := provider.GetHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()

	// Should contain the metric
	if !strings.Contains(body, "test_requests_total") {
		t.Error("Response should contain the metric")
	}
}

func TestPrometheusProviderErrorHandling(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	// Test with empty metric name (should not panic)
	provider.IncrementCounter("", nil)
	provider.RecordHistogram("", 1.0, nil)
	provider.SetGauge("", 1.0, nil)

	// Test timer with failed histogram creation
	timer := provider.StartTimer("", nil)
	if timer == nil {
		t.Error("StartTimer should return a timer even if histogram creation fails")
	}

	// Should not panic
	timer.Stop()
	timer.StopWithLabels(map[string]string{"test": "value"})
}

func TestPrometheusProviderLabelHandling(t *testing.T) {
	provider := NewPrometheusProvider(Config{
		Namespace: "test",
	})

	// Test with various label scenarios
	testCases := []map[string]string{
		nil,                                  // nil labels
		{},                                   // empty labels
		{"key": "value"},                     // normal labels
		{"key": ""},                          // empty value (should be filtered)
		{"": "value"},                        // empty key (should be filtered)
		{"key1": "value1", "key2": "value2"}, // multiple labels
	}

	for i, labels := range testCases {
		metricName := "test_metric_" + string(rune(i))
		provider.IncrementCounter(metricName, labels)
		provider.RecordHistogram(metricName, 1.0, labels)
		provider.SetGauge(metricName, 1.0, labels)
	}

	// Should not panic and should create metrics
	if len(provider.counters) == 0 {
		t.Error("Expected some counters to be created")
	}
}

// Benchmark Prometheus operations
func BenchmarkPrometheusProvider(b *testing.B) {
	provider := NewPrometheusProvider(Config{
		Namespace: "bench",
		Prometheus: PrometheusConfig{
			Buckets: DefaultHistogramBuckets,
		},
	})

	labels := map[string]string{
		"method": "GET",
		"status": "200",
	}

	b.Run("IncrementCounter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.IncrementCounter("requests_total", labels)
		}
	})

	b.Run("RecordHistogram", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.RecordHistogram("request_duration", 0.1, labels)
		}
	})

	b.Run("SetGauge", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.SetGauge("active_connections", float64(i), labels)
		}
	})

	b.Run("StartTimer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timer := provider.StartTimer("operation_duration", labels)
			timer.Stop()
		}
	})
}
