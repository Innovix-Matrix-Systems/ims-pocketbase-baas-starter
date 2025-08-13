package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewNoOpProvider(t *testing.T) {
	provider := NewNoOpProvider()

	if provider == nil {
		t.Fatal("NewNoOpProvider() returned nil")
	}

	// Verify it implements MetricsProvider interface
	var _ MetricsProvider = provider
}

func TestNoOpProviderMethods(t *testing.T) {
	provider := NewNoOpProvider()
	labels := map[string]string{"test": "value"}

	// Test that all methods can be called without panicking
	t.Run("IncrementCounter", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IncrementCounter panicked: %v", r)
			}
		}()
		provider.IncrementCounter("test_counter", labels)
		provider.IncrementCounter("test_counter", nil) // Test with nil labels
	})

	t.Run("IncrementCounterBy", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("IncrementCounterBy panicked: %v", r)
			}
		}()
		provider.IncrementCounterBy("test_counter", 5.0, labels)
		provider.IncrementCounterBy("test_counter", 0.0, nil)     // Test with zero value and nil labels
		provider.IncrementCounterBy("test_counter", -1.0, labels) // Test with negative value
	})

	t.Run("RecordHistogram", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RecordHistogram panicked: %v", r)
			}
		}()
		provider.RecordHistogram("test_histogram", 1.5, labels)
		provider.RecordHistogram("test_histogram", 0.0, nil) // Test with zero value and nil labels
	})

	t.Run("SetGauge", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetGauge panicked: %v", r)
			}
		}()
		provider.SetGauge("test_gauge", 42.0, labels)
		provider.SetGauge("test_gauge", -10.0, nil) // Test with negative value and nil labels
	})

	t.Run("RecordDuration", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("RecordDuration panicked: %v", r)
			}
		}()
		provider.RecordDuration("test_duration", time.Second, labels)
		provider.RecordDuration("test_duration", 0, nil) // Test with zero duration and nil labels
	})
}

func TestNoOpTimer(t *testing.T) {
	provider := NewNoOpProvider()

	t.Run("StartTimer", func(t *testing.T) {
		timer := provider.StartTimer("test_timer", map[string]string{"test": "value"})

		if timer == nil {
			t.Fatal("StartTimer() returned nil")
		}

		// Verify it implements Timer interface
		var _ Timer = timer
	})

	t.Run("TimerMethods", func(t *testing.T) {
		timer := provider.StartTimer("test_timer", nil)

		// Test that timer methods don't panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Timer method panicked: %v", r)
			}
		}()

		timer.Stop()
		timer.StopWithLabels(map[string]string{"additional": "label"})
		timer.StopWithLabels(nil) // Test with nil labels
	})

	t.Run("MultipleTimerOperations", func(t *testing.T) {
		// Test creating multiple timers and calling methods multiple times
		timer1 := provider.StartTimer("timer1", nil)
		timer2 := provider.StartTimer("timer2", map[string]string{"test": "value"})

		timer1.Stop()
		timer2.StopWithLabels(map[string]string{"result": "success"})

		// Call methods again to ensure they're idempotent
		timer1.Stop()
		timer2.Stop()
	})
}

func TestNoOpProviderHTTPHandler(t *testing.T) {
	provider := NewNoOpProvider()
	handler := provider.GetHandler()

	if handler == nil {
		t.Fatal("GetHandler() returned nil")
	}

	// Test HTTP handler behavior
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return 404 status
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	// Should return appropriate message
	expectedBody := "Metrics are disabled\n"
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
	}

	// Test with different HTTP methods
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/metrics", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("Expected status %d for %s method, got %d", http.StatusNotFound, method, w.Code)
			}
		})
	}
}

func TestNoOpProviderShutdown(t *testing.T) {
	provider := NewNoOpProvider()

	// Test shutdown with different contexts
	t.Run("NormalShutdown", func(t *testing.T) {
		ctx := context.Background()
		err := provider.Shutdown(ctx)

		if err != nil {
			t.Errorf("Shutdown() returned error: %v", err)
		}
	})

	t.Run("CancelledContext", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := provider.Shutdown(ctx)

		// No-op should succeed even with cancelled context
		if err != nil {
			t.Errorf("Shutdown() with cancelled context returned error: %v", err)
		}
	})

	t.Run("TimeoutContext", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Wait for timeout
		time.Sleep(1 * time.Millisecond)

		err := provider.Shutdown(ctx)

		// No-op should succeed even with timed out context
		if err != nil {
			t.Errorf("Shutdown() with timeout context returned error: %v", err)
		}
	})
}

// Test performance characteristics of no-op provider
func TestNoOpProviderPerformance(t *testing.T) {
	provider := NewNoOpProvider()
	labels := map[string]string{"test": "value", "env": "test"}

	// Ensure no-op operations are fast
	start := time.Now()

	for i := 0; i < 10000; i++ {
		provider.IncrementCounter("test", labels)
		provider.RecordHistogram("test", float64(i), labels)
		provider.SetGauge("test", float64(i), labels)

		timer := provider.StartTimer("test", labels)
		timer.Stop()
	}

	duration := time.Since(start)

	// No-op operations should be very fast (less than 10ms for 10k operations)
	if duration > 10*time.Millisecond {
		t.Errorf("No-op operations took too long: %v", duration)
	}
}

// Benchmark no-op operations
func BenchmarkNoOpProvider(b *testing.B) {
	provider := NewNoOpProvider()
	labels := map[string]string{"test": "value"}

	b.Run("IncrementCounter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.IncrementCounter("test", labels)
		}
	})

	b.Run("RecordHistogram", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.RecordHistogram("test", 1.0, labels)
		}
	})

	b.Run("SetGauge", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			provider.SetGauge("test", 1.0, labels)
		}
	})

	b.Run("StartTimer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timer := provider.StartTimer("test", labels)
			timer.Stop()
		}
	})
}

// Test concurrent access to no-op provider
func TestNoOpProviderConcurrency(t *testing.T) {
	provider := NewNoOpProvider()
	labels := map[string]string{"test": "value"}

	const numGoroutines = 100
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

			for j := 0; j < operationsPerGoroutine; j++ {
				provider.IncrementCounter("test", labels)
				provider.RecordHistogram("test", float64(j), labels)
				provider.SetGauge("test", float64(j), labels)

				timer := provider.StartTimer("test", labels)
				timer.Stop()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}
