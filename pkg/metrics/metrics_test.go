package metrics

import (
	"os"
	"sync"
	"testing"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		config       Config
		expectedType string
		shouldBeNoOp bool
	}{
		{
			name: "disabled provider",
			config: Config{
				Provider: ProviderDisabled,
				Enabled:  false,
			},
			expectedType: "*metrics.NoOpProvider",
			shouldBeNoOp: true,
		},
		{
			name: "enabled false with prometheus",
			config: Config{
				Provider: ProviderPrometheus,
				Enabled:  false,
			},
			expectedType: "*metrics.NoOpProvider",
			shouldBeNoOp: true,
		},
		{
			name: "unknown provider",
			config: Config{
				Provider: "unknown",
				Enabled:  true,
			},
			expectedType: "*metrics.NoOpProvider",
			shouldBeNoOp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.config)

			if provider == nil {
				t.Fatal("NewProvider() returned nil")
			}

			// For now, we can only test that we get a provider back
			// We'll test specific provider types in their respective test files
			if tt.shouldBeNoOp {
				// Test that it behaves like a no-op (doesn't panic)
				provider.IncrementCounter("test", nil)
				provider.RecordHistogram("test", 1.0, nil)
				provider.SetGauge("test", 1.0, nil)

				timer := provider.StartTimer("test", nil)
				timer.Stop()

				handler := provider.GetHandler()
				if handler == nil {
					t.Error("GetHandler() should not return nil even for no-op provider")
				}
			}
		})
	}
}

func TestSingletonPattern(t *testing.T) {
	// Reset singleton before test
	Reset()

	// Verify not initialized
	if IsInitialized() {
		t.Error("Expected singleton to not be initialized")
	}

	// Set environment for test
	os.Setenv("METRICS_PROVIDER", "disabled")
	os.Setenv("METRICS_ENABLED", "false")
	defer func() {
		os.Unsetenv("METRICS_PROVIDER")
		os.Unsetenv("METRICS_ENABLED")
	}()

	// Get instance multiple times
	instance1 := GetInstance()
	instance2 := GetInstance()

	// Should be the same instance
	if instance1 != instance2 {
		t.Error("GetInstance() should return the same instance")
	}

	// Should be initialized now
	if !IsInitialized() {
		t.Error("Expected singleton to be initialized after GetInstance()")
	}
}

func TestInitializeProvider(t *testing.T) {
	// Reset singleton before test
	Reset()

	config := Config{
		Provider: ProviderDisabled,
		Enabled:  false,
	}

	// Initialize with custom config
	instance1 := InitializeProvider(config)
	instance2 := GetInstance()

	// Should be the same instance
	if instance1 != instance2 {
		t.Error("InitializeProvider() and GetInstance() should return the same instance")
	}

	// Should be initialized
	if !IsInitialized() {
		t.Error("Expected singleton to be initialized after InitializeProvider()")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Reset singleton before test
	Reset()

	// Set environment for test
	os.Setenv("METRICS_PROVIDER", "disabled")
	os.Setenv("METRICS_ENABLED", "false")
	defer func() {
		os.Unsetenv("METRICS_PROVIDER")
		os.Unsetenv("METRICS_ENABLED")
	}()

	const numGoroutines = 100
	instances := make([]MetricsProvider, numGoroutines)
	var wg sync.WaitGroup

	// Launch multiple goroutines to get instance
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			instances[index] = GetInstance()
		}(i)
	}

	wg.Wait()

	// All instances should be the same
	firstInstance := instances[0]
	for i := 1; i < numGoroutines; i++ {
		if instances[i] != firstInstance {
			t.Errorf("Instance %d is different from first instance", i)
		}
	}
}

func TestReset(t *testing.T) {
	// Initialize singleton
	_ = GetInstance()

	if !IsInitialized() {
		t.Error("Expected singleton to be initialized")
	}

	// Reset
	Reset()

	if IsInitialized() {
		t.Error("Expected singleton to not be initialized after Reset()")
	}

	// Should be able to initialize again
	_ = GetInstance()

	if !IsInitialized() {
		t.Error("Expected singleton to be initialized after second GetInstance()")
	}
}

// Benchmark singleton access
func BenchmarkGetInstance(b *testing.B) {
	// Reset and initialize once
	Reset()
	os.Setenv("METRICS_PROVIDER", "disabled")
	defer os.Unsetenv("METRICS_PROVIDER")

	_ = GetInstance() // Initialize

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetInstance()
		}
	})
}
