package metrics

import (
	"sync"
)

var (
	instance MetricsProvider
	once     sync.Once
)

// GetInstance returns the singleton metrics provider
func GetInstance() MetricsProvider {
	once.Do(func() {
		config := LoadConfig()
		instance = NewProvider(config)
	})
	return instance
}

// InitializeProvider initializes the singleton with custom config
// This should be called during application startup
func InitializeProvider(config Config) MetricsProvider {
	once.Do(func() {
		instance = NewProvider(config)
	})
	return instance
}

// NewProvider creates a new metrics provider based on configuration
func NewProvider(config Config) MetricsProvider {
	if !config.IsEnabled() {
		return NewNoOpProvider()
	}

	switch config.Provider {
	case ProviderPrometheus:
		return NewPrometheusProvider(config)
	case ProviderOpenTelemetry:
		return NewOpenTelemetryProvider(config)
	default:
		// Fallback to no-op for unknown providers
		return NewNoOpProvider()
	}
}

// Reset resets the singleton instance (used for testing)
func Reset() {
	once = sync.Once{}
	instance = nil
}

// IsInitialized returns true if the singleton instance has been initialized
func IsInitialized() bool {
	return instance != nil
}
