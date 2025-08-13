package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOpenTelemetryProvider(t *testing.T) {
	config := Config{
		Provider:  ProviderOpenTelemetry,
		Enabled:   true,
		Namespace: "test_app",
		Labels: map[string]string{
			"env": "test",
		},
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint:       "http://localhost:4317",
			Headers:        map[string]string{"api-key": "test"},
			Insecure:       true,
			ExportInterval: 30 * time.Second,
		},
	}

	provider := NewOpenTelemetryProvider(config)

	if provider == nil {
		t.Fatal("NewOpenTelemetryProvider() returned nil")
	}

	if provider.config.Namespace != "test_app" {
		t.Errorf("Expected namespace 'test_app', got '%s'", provider.config.Namespace)
	}

	if provider.meter == nil {
		t.Error("Meter should not be nil")
	}

	// Verify it implements MetricsProvider interface
	var _ MetricsProvider = provider
}

func TestOpenTelemetryProviderCounters(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
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

func TestOpenTelemetryProviderHistograms(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
	})

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

func TestOpenTelemetryProviderGauges(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
	})

	labels := map[string]string{
		"queue": "default",
	}

	// Test SetGauge
	provider.SetGauge("queue_size", 10.0, labels)
	provider.SetGauge("queue_size", 15.0, labels) // Update same gauge
	provider.SetGauge("active_connections", 5.0, nil)

	// Verify gauge values are stored
	if len(provider.gaugeValues) == 0 {
		t.Error("Expected gauge values to be stored")
	}

	// Verify gauges are created
	if len(provider.gauges) == 0 {
		t.Error("Expected gauges to be created")
	}
}

func TestOpenTelemetryTimer(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
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

	// Test StopWithLabels
	timer2 := provider.StartTimer("operation_duration_seconds", labels)
	time.Sleep(5 * time.Millisecond)
	timer2.StopWithLabels(map[string]string{
		"result": "success",
	})

	// Test with nil additional labels
	timer3 := provider.StartTimer("operation_duration_seconds", labels)
	timer3.StopWithLabels(nil)

	// Verify histogram was created
	if len(provider.histograms) != 1 {
		t.Errorf("Expected 1 histogram, got %d", len(provider.histograms))
	}
}

func TestOpenTelemetryProviderHTTPHandler(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
	})

	handler := provider.GetHandler()
	if handler == nil {
		t.Fatal("GetHandler() returned nil")
	}

	// Test HTTP handler (should return 404 for OpenTelemetry)
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "OpenTelemetry") {
		t.Error("Response should mention OpenTelemetry")
	}
}

func TestOpenTelemetryProviderShutdown(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
	})

	// Add some metrics
	provider.IncrementCounter("test_counter", nil)
	provider.SetGauge("test_gauge", 1.0, nil)

	// Test shutdown
	ctx := context.Background()
	err := provider.Shutdown(ctx)

	// Should not return error (even if meterProvider is nil)
	if err != nil {
		t.Errorf("Shutdown() returned error: %v", err)
	}

	// Test shutdown with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() with timeout returned error: %v", err)
	}
}

func TestOpenTelemetryProviderConcurrency(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
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

	if len(provider.gaugeValues) == 0 {
		t.Error("Expected gauge values to be stored")
	}
}

func TestOpenTelemetryProviderGlobalLabels(t *testing.T) {
	config := Config{
		Namespace: "test",
		Labels: map[string]string{
			"env":     "test",
			"service": "api",
		},
	}
	provider := NewOpenTelemetryProvider(config)

	// Add metric with local labels
	localLabels := map[string]string{
		"method": "GET",
	}

	provider.IncrementCounter("requests_total", localLabels)

	// Verify global labels are included in conversion
	attrs := provider.convertLabels(localLabels)

	// Should have both global and local labels
	expectedLabels := 3 // env, service, method
	if len(attrs) != expectedLabels {
		t.Errorf("Expected %d attributes, got %d", expectedLabels, len(attrs))
	}
}

func TestOpenTelemetryProviderErrorHandling(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "test",
	})

	// Test with empty metric name (should not panic)
	provider.IncrementCounter("", nil)
	provider.RecordHistogram("", 1.0, nil)
	provider.SetGauge("", 1.0, nil)

	// Test timer with potential histogram creation failure
	timer := provider.StartTimer("test_timer", nil)
	if timer == nil {
		t.Error("StartTimer should return a timer")
	}

	// Should not panic
	timer.Stop()
	timer.StopWithLabels(map[string]string{"test": "value"})
}

func TestOpenTelemetryProviderLabelHandling(t *testing.T) {
	provider := NewOpenTelemetryProvider(Config{
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

func TestOpenTelemetryProviderConvertLabels(t *testing.T) {
	config := Config{
		Namespace: "test",
		Labels: map[string]string{
			"global1": "value1",
			"global2": "value2",
		},
	}
	provider := NewOpenTelemetryProvider(config)

	labels := map[string]string{
		"local1": "value3",
		"local2": "value4",
		"":       "ignored", // Should be filtered out
		"empty":  "",        // Should be filtered out
	}

	attrs := provider.convertLabels(labels)

	// Should have global1, global2, local1, local2 (4 attributes)
	expectedCount := 4
	if len(attrs) != expectedCount {
		t.Errorf("Expected %d attributes, got %d", expectedCount, len(attrs))
	}

	// Verify specific attributes exist
	attrMap := make(map[string]string)
	for _, attr := range attrs {
		attrMap[string(attr.Key)] = attr.Value.AsString()
	}

	if attrMap["global1"] != "value1" {
		t.Error("Expected global1=value1")
	}
	if attrMap["local1"] != "value3" {
		t.Error("Expected local1=value3")
	}
}

func TestOpenTelemetryProviderFailedExporter(t *testing.T) {
	// Test with invalid endpoint to simulate exporter creation failure
	config := Config{
		Namespace: "test",
		OpenTelemetry: OpenTelemetryConfig{
			Endpoint: "invalid://endpoint",
		},
	}

	provider := NewOpenTelemetryProvider(config)

	// Should still create a provider (with fallback meter)
	if provider == nil {
		t.Fatal("NewOpenTelemetryProvider() should not return nil even with invalid config")
	}

	// Should still be able to use metrics (they just won't be exported)
	provider.IncrementCounter("test", nil)
	provider.RecordHistogram("test", 1.0, nil)
	provider.SetGauge("test", 1.0, nil)
}

// Benchmark OpenTelemetry operations
func BenchmarkOpenTelemetryProvider(b *testing.B) {
	provider := NewOpenTelemetryProvider(Config{
		Namespace: "bench",
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
