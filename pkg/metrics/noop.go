package metrics

import (
	"context"
	"net/http"
	"time"
)

// NoOpProvider implements MetricsProvider with no-op operations
// This provider is used when metrics are disabled or when there's a configuration error
type NoOpProvider struct{}

// NoOpTimer implements Timer with no-op operations
type NoOpTimer struct{}

// NewNoOpProvider creates a new no-op metrics provider
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{}
}

// IncrementCounter does nothing in no-op implementation
func (p *NoOpProvider) IncrementCounter(name string, labels map[string]string) {
	// No-op: intentionally empty
}

// IncrementCounterBy does nothing in no-op implementation
func (p *NoOpProvider) IncrementCounterBy(name string, value float64, labels map[string]string) {
	// No-op: intentionally empty
}

// RecordHistogram does nothing in no-op implementation
func (p *NoOpProvider) RecordHistogram(name string, value float64, labels map[string]string) {
	// No-op: intentionally empty
}

// SetGauge does nothing in no-op implementation
func (p *NoOpProvider) SetGauge(name string, value float64, labels map[string]string) {
	// No-op: intentionally empty
}

// StartTimer returns a no-op timer
func (p *NoOpProvider) StartTimer(name string, labels map[string]string) Timer {
	return &NoOpTimer{}
}

// RecordDuration does nothing in no-op implementation
func (p *NoOpProvider) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	// No-op: intentionally empty
}

// GetHandler returns a simple HTTP handler that returns 404
// This ensures the metrics endpoint doesn't break when metrics are disabled
func (p *NoOpProvider) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Metrics are disabled\n"))
	})
}

// Shutdown does nothing in no-op implementation
func (p *NoOpProvider) Shutdown(ctx context.Context) error {
	// No-op: always succeeds immediately
	return nil
}

// Stop does nothing in no-op timer implementation
func (t *NoOpTimer) Stop() {
	// No-op: intentionally empty
}

// StopWithLabels does nothing in no-op timer implementation
func (t *NoOpTimer) StopWithLabels(labels map[string]string) {
	// No-op: intentionally empty
}
