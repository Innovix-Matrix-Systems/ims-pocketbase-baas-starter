package metrics

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"
)

// MockMetricsProvider is a test implementation of MetricsProvider
type MockMetricsProvider struct {
	counters   map[string]float64
	histograms map[string][]float64
	gauges     map[string]float64
	timers     []string
	mu         sync.RWMutex
}

// NewMockMetricsProvider creates a new mock metrics provider
func NewMockMetricsProvider() *MockMetricsProvider {
	return &MockMetricsProvider{
		counters:   make(map[string]float64),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
		timers:     make([]string, 0),
	}
}

func (m *MockMetricsProvider) IncrementCounter(name string, labels map[string]string) {
	m.IncrementCounterBy(name, 1.0, labels)
}

func (m *MockMetricsProvider) IncrementCounterBy(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name
	if len(labels) > 0 {
		key += "_labeled"
	}
	m.counters[key] += value
}

func (m *MockMetricsProvider) RecordHistogram(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name
	if len(labels) > 0 {
		key += "_labeled"
	}
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *MockMetricsProvider) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name
	if len(labels) > 0 {
		key += "_labeled"
	}
	m.gauges[key] = value
}

func (m *MockMetricsProvider) StartTimer(name string, labels map[string]string) Timer {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := name
	if len(labels) > 0 {
		key += "_labeled"
	}
	m.timers = append(m.timers, key)
	return &MockTimer{provider: m, name: key}
}

func (m *MockMetricsProvider) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	m.RecordHistogram(name, duration.Seconds(), labels)
}

func (m *MockMetricsProvider) GetHandler() http.Handler {
	return nil
}

func (m *MockMetricsProvider) Shutdown(ctx context.Context) error {
	return nil
}

// MockTimer implements Timer for testing
type MockTimer struct {
	provider *MockMetricsProvider
	name     string
}

func (t *MockTimer) Stop() {
	// Record a small duration for testing
	t.provider.RecordHistogram(t.name, 0.001, nil)
}

func (t *MockTimer) StopWithLabels(labels map[string]string) {
	// Record a small duration for testing
	t.provider.RecordHistogram(t.name, 0.001, labels)
}

// Helper methods for testing
func (m *MockMetricsProvider) GetCounterValue(name string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.counters[name]
}

func (m *MockMetricsProvider) GetHistogramValues(name string) []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.histograms[name]
}

func (m *MockMetricsProvider) GetGaugeValue(name string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gauges[name]
}

func (m *MockMetricsProvider) GetTimerCount(name string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, timer := range m.timers {
		if timer == name {
			count++
		}
	}
	return count
}

func TestInstrumentHook(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("successful hook execution", func(t *testing.T) {
		err := InstrumentHook(mock, "test_hook", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check metrics were recorded
		if mock.GetCounterValue(MetricHookExecutionTotal+"_labeled") != 1 {
			t.Error("Expected hook execution total to be incremented")
		}

		if mock.GetCounterValue(MetricHookErrorsTotal+"_labeled") != 0 {
			t.Error("Expected no hook errors")
		}

		if mock.GetTimerCount(MetricHookExecutionDuration+"_labeled") != 1 {
			t.Error("Expected timer to be started")
		}
	})

	t.Run("failed hook execution", func(t *testing.T) {
		testError := errors.New("test error")
		err := InstrumentHook(mock, "test_hook_error", func() error {
			return testError
		})

		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Check error metrics were recorded
		if mock.GetCounterValue(MetricHookErrorsTotal+"_labeled") == 0 {
			t.Error("Expected hook errors to be incremented")
		}
	})

	t.Run("nil provider", func(t *testing.T) {
		err := InstrumentHook(nil, "test_hook", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error with nil provider, got %v", err)
		}
	})
}

func TestInstrumentJobHandler(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("successful job execution", func(t *testing.T) {
		err := InstrumentJobHandler(mock, "test_job", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check metrics were recorded
		if mock.GetCounterValue(MetricJobExecutionTotal+"_labeled") != 1 {
			t.Error("Expected job execution total to be incremented")
		}

		if mock.GetCounterValue(MetricJobErrorsTotal+"_labeled") != 0 {
			t.Error("Expected no job errors")
		}

		if mock.GetTimerCount(MetricJobExecutionDuration+"_labeled") != 1 {
			t.Error("Expected timer to be started")
		}
	})

	t.Run("failed job execution", func(t *testing.T) {
		testError := errors.New("job failed")
		err := InstrumentJobHandler(mock, "test_job_error", func() error {
			return testError
		})

		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Check error metrics were recorded
		if mock.GetCounterValue(MetricJobErrorsTotal+"_labeled") == 0 {
			t.Error("Expected job errors to be incremented")
		}
	})
}

func TestInstrumentHTTPHandler(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("successful HTTP request", func(t *testing.T) {
		err := InstrumentHTTPHandler(mock, "GET", "/api/test", func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check metrics were recorded
		if mock.GetCounterValue(MetricHTTPRequestsTotal+"_labeled") != 1 {
			t.Error("Expected HTTP requests total to be incremented")
		}

		if mock.GetTimerCount(MetricHTTPRequestDuration+"_labeled") != 1 {
			t.Error("Expected timer to be started")
		}
	})

	t.Run("failed HTTP request", func(t *testing.T) {
		testError := errors.New("HTTP error")
		err := InstrumentHTTPHandler(mock, "POST", "/api/error", func() error {
			return testError
		})

		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Check error metrics were recorded
		if mock.GetCounterValue(MetricHandlerErrorsTotal+"_labeled") == 0 {
			t.Error("Expected handler errors to be incremented")
		}
	})
}

func TestInstrumentRecordOperation(t *testing.T) {
	mock := NewMockMetricsProvider()

	err := InstrumentRecordOperation(mock, "users", "create", func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check metrics were recorded
	if mock.GetCounterValue(MetricRecordOperationsTotal+"_labeled") != 1 {
		t.Error("Expected record operations total to be incremented")
	}

	if mock.GetTimerCount(MetricHookExecutionDuration+"_labeled") != 1 {
		t.Error("Expected timer to be started")
	}
}

func TestInstrumentEmailOperation(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("successful email", func(t *testing.T) {
		err := InstrumentEmailOperation(mock, func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Check metrics were recorded
		if mock.GetCounterValue(MetricEmailsSentTotal) != 1 {
			t.Error("Expected emails sent total to be incremented")
		}
	})

	t.Run("failed email", func(t *testing.T) {
		testError := errors.New("email failed")
		err := InstrumentEmailOperation(mock, func() error {
			return testError
		})

		if err != testError {
			t.Errorf("Expected test error, got %v", err)
		}

		// Should not increment success counter on error
		if mock.GetCounterValue(MetricEmailsSentTotal) != 1 { // Still 1 from previous test
			t.Error("Expected emails sent total to not be incremented on error")
		}
	})
}

func TestInstrumentCacheOperation(t *testing.T) {
	mock := NewMockMetricsProvider()

	// Test cache hit
	InstrumentCacheOperation(mock, true)
	if mock.GetCounterValue(MetricCacheHitsTotal) != 1 {
		t.Error("Expected cache hits to be incremented")
	}

	// Test cache miss
	InstrumentCacheOperation(mock, false)
	if mock.GetCounterValue(MetricCacheMissesTotal) != 1 {
		t.Error("Expected cache misses to be incremented")
	}

	// Test with nil provider
	InstrumentCacheOperation(nil, true) // Should not panic
}

func TestSafeExecute(t *testing.T) {
	t.Run("normal execution", func(t *testing.T) {
		err := SafeExecute(func() error {
			return errors.New("test error")
		})

		if err == nil || err.Error() != "test error" {
			t.Errorf("Expected test error, got %v", err)
		}
	})

	t.Run("panic recovery", func(t *testing.T) {
		err := SafeExecute(func() error {
			panic("test panic")
		})

		if err == nil || err.Error() != "panic recovered: test panic" {
			t.Errorf("Expected panic recovery error, got %v", err)
		}
	})

	t.Run("successful execution", func(t *testing.T) {
		err := SafeExecute(func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestSafeFunctions(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("SafeIncrementCounter", func(t *testing.T) {
		SafeIncrementCounter(mock, "test_counter", nil)
		if mock.GetCounterValue("test_counter") != 1 {
			t.Error("Expected counter to be incremented")
		}

		// Test with nil provider
		SafeIncrementCounter(nil, "test_counter", nil) // Should not panic
	})

	t.Run("SafeIncrementCounterBy", func(t *testing.T) {
		SafeIncrementCounterBy(mock, "test_counter_by", 5.0, nil)
		if mock.GetCounterValue("test_counter_by") != 5.0 {
			t.Error("Expected counter to be incremented by 5")
		}
	})

	t.Run("SafeRecordHistogram", func(t *testing.T) {
		SafeRecordHistogram(mock, "test_histogram", 1.5, nil)
		values := mock.GetHistogramValues("test_histogram")
		if len(values) != 1 || values[0] != 1.5 {
			t.Error("Expected histogram value to be recorded")
		}
	})

	t.Run("SafeSetGauge", func(t *testing.T) {
		SafeSetGauge(mock, "test_gauge", 42.0, nil)
		if mock.GetGaugeValue("test_gauge") != 42.0 {
			t.Error("Expected gauge value to be set")
		}
	})

	t.Run("SafeStartTimer and SafeStopTimer", func(t *testing.T) {
		timer := SafeStartTimer(mock, "test_timer", nil)
		if timer == nil {
			t.Error("Expected timer to be returned")
		}

		SafeStopTimer(timer) // Should not panic

		// Test with nil provider
		timer = SafeStartTimer(nil, "test_timer", nil)
		if timer == nil {
			t.Error("Expected no-op timer to be returned")
		}

		SafeStopTimer(nil) // Should not panic
	})

	t.Run("SafeRecordDuration", func(t *testing.T) {
		SafeRecordDuration(mock, "test_duration", 100*time.Millisecond, nil)
		values := mock.GetHistogramValues("test_duration")
		if len(values) == 0 {
			t.Error("Expected duration to be recorded")
		}
	})
}

func TestMeasureExecutionTime(t *testing.T) {
	mock := NewMockMetricsProvider()

	err := MeasureExecutionTime(mock, "test_execution", nil, func() error {
		time.Sleep(1 * time.Millisecond) // Small delay for measurement
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	values := mock.GetHistogramValues("test_execution")
	if len(values) != 1 {
		t.Error("Expected execution time to be recorded")
	}

	// Test with nil provider
	err = MeasureExecutionTime(nil, "test_execution", nil, func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error with nil provider, got %v", err)
	}
}

func TestWithMetrics(t *testing.T) {
	mock := NewMockMetricsProvider()

	t.Run("with provider", func(t *testing.T) {
		err := WithMetrics(mock, func(provider MetricsProvider) error {
			if provider != mock {
				t.Error("Expected same provider to be passed")
			}
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("with nil provider", func(t *testing.T) {
		err := WithMetrics(nil, func(provider MetricsProvider) error {
			if provider == nil {
				t.Error("Expected no-op provider to be provided")
			}
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestBatchIncrementCounters(t *testing.T) {
	mock := NewMockMetricsProvider()

	counters := map[string]map[string]string{
		"counter1": {"label1": "value1"},
		"counter2": {"label2": "value2"},
		"counter3": nil,
	}

	BatchIncrementCounters(mock, counters)

	// Check that all counters were incremented
	if mock.GetCounterValue("counter1_labeled") != 1 {
		t.Error("Expected counter1 to be incremented")
	}
	if mock.GetCounterValue("counter2_labeled") != 1 {
		t.Error("Expected counter2 to be incremented")
	}
	if mock.GetCounterValue("counter3") != 1 {
		t.Error("Expected counter3 to be incremented")
	}

	// Test with nil provider
	BatchIncrementCounters(nil, counters) // Should not panic
}

func TestRecordQueueSize(t *testing.T) {
	mock := NewMockMetricsProvider()

	RecordQueueSize(mock, "default", 10)

	if mock.GetGaugeValue(MetricJobQueueSize+"_labeled") != 10 {
		t.Error("Expected queue size to be recorded")
	}

	// Test with nil provider
	RecordQueueSize(nil, "default", 5) // Should not panic
}

func TestConcurrentInstrumentation(t *testing.T) {
	mock := NewMockMetricsProvider()
	const numGoroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Test concurrent hook instrumentation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				_ = InstrumentHook(mock, "concurrent_hook", func() error {
					return nil
				})
			}
		}()
	}

	wg.Wait()

	// Verify metrics were recorded correctly
	expectedTotal := float64(numGoroutines * operationsPerGoroutine)
	if mock.GetCounterValue(MetricHookExecutionTotal+"_labeled") != expectedTotal {
		t.Errorf("Expected %f hook executions, got %f", expectedTotal, mock.GetCounterValue(MetricHookExecutionTotal+"_labeled"))
	}
}

// Benchmark instrumentation functions
func BenchmarkInstrumentHook(b *testing.B) {
	mock := NewMockMetricsProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InstrumentHook(mock, "benchmark_hook", func() error {
			return nil
		})
	}
}

func BenchmarkSafeIncrementCounter(b *testing.B) {
	mock := NewMockMetricsProvider()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SafeIncrementCounter(mock, "benchmark_counter", nil)
	}
}
