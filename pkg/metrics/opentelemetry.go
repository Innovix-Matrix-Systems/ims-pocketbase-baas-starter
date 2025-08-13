package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// OpenTelemetryProvider implements MetricsProvider using OpenTelemetry
// This is a simplified implementation that uses the global meter provider
type OpenTelemetryProvider struct {
	config      Config
	meter       metric.Meter
	counters    map[string]metric.Int64Counter
	histograms  map[string]metric.Float64Histogram
	gauges      map[string]metric.Float64ObservableGauge
	gaugeValues map[string]float64
	mu          sync.RWMutex
}

// OpenTelemetryTimer implements Timer for OpenTelemetry metrics
type OpenTelemetryTimer struct {
	histogram metric.Float64Histogram
	labels    []attribute.KeyValue
	startTime time.Time
}

// NewOpenTelemetryProvider creates a new OpenTelemetry metrics provider
func NewOpenTelemetryProvider(config Config) *OpenTelemetryProvider {
	// Use the global meter provider (user should configure it externally)
	meter := otel.Meter(config.Namespace)

	return &OpenTelemetryProvider{
		config:      config,
		meter:       meter,
		counters:    make(map[string]metric.Int64Counter),
		histograms:  make(map[string]metric.Float64Histogram),
		gauges:      make(map[string]metric.Float64ObservableGauge),
		gaugeValues: make(map[string]float64),
	}
}

// IncrementCounter increments a counter metric
func (p *OpenTelemetryProvider) IncrementCounter(name string, labels map[string]string) {
	p.IncrementCounterBy(name, 1.0, labels)
}

// IncrementCounterBy increments a counter metric by a specific value
func (p *OpenTelemetryProvider) IncrementCounterBy(name string, value float64, labels map[string]string) {
	counter := p.getOrCreateCounter(name)
	if counter != nil {
		attrs := p.convertLabels(labels)
		counter.Add(context.Background(), int64(value), metric.WithAttributes(attrs...))
	}
}

// RecordHistogram records a value in a histogram metric
func (p *OpenTelemetryProvider) RecordHistogram(name string, value float64, labels map[string]string) {
	histogram := p.getOrCreateHistogram(name)
	if histogram != nil {
		attrs := p.convertLabels(labels)
		histogram.Record(context.Background(), value, metric.WithAttributes(attrs...))
	}
}

// SetGauge sets a gauge metric value
func (p *OpenTelemetryProvider) SetGauge(name string, value float64, labels map[string]string) {
	// First ensure gauge exists (without holding lock)
	p.getOrCreateGauge(name)

	// Then store the value
	p.mu.Lock()
	defer p.mu.Unlock()

	// Store the value for the gauge (we'll use the first label as a key for simplicity)
	key := name
	if len(labels) > 0 {
		for k, v := range labels {
			key = name + "_" + k + "_" + v
			break // Use first label for key
		}
	}

	p.gaugeValues[key] = value
}

// StartTimer starts a timer for measuring durations
func (p *OpenTelemetryProvider) StartTimer(name string, labels map[string]string) Timer {
	histogram := p.getOrCreateHistogram(name)
	if histogram == nil {
		// Return a no-op timer if histogram creation failed
		return &NoOpTimer{}
	}

	return &OpenTelemetryTimer{
		histogram: histogram,
		labels:    p.convertLabels(labels),
		startTime: time.Now(),
	}
}

// RecordDuration records a duration metric
func (p *OpenTelemetryProvider) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	p.RecordHistogram(name, duration.Seconds(), labels)
}

// GetHandler returns a no-op HTTP handler since OpenTelemetry uses push-based exports
func (p *OpenTelemetryProvider) GetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("OpenTelemetry metrics are exported via OTLP, not HTTP scraping\n"))
	})
}

// Shutdown gracefully shuts down the OpenTelemetry provider
func (p *OpenTelemetryProvider) Shutdown(ctx context.Context) error {
	// For this simplified implementation, we don't manage the meter provider
	// The user should handle shutdown of the global meter provider externally
	return nil
}

// Stop records the duration since timer creation
func (t *OpenTelemetryTimer) Stop() {
	duration := time.Since(t.startTime)
	t.histogram.Record(context.Background(), duration.Seconds(), metric.WithAttributes(t.labels...))
}

// StopWithLabels records the duration with additional labels
func (t *OpenTelemetryTimer) StopWithLabels(additionalLabels map[string]string) {
	duration := time.Since(t.startTime)

	// Merge original labels with additional labels
	allLabels := make([]attribute.KeyValue, len(t.labels))
	copy(allLabels, t.labels)

	for key, value := range additionalLabels {
		if key != "" && value != "" {
			allLabels = append(allLabels, attribute.String(key, value))
		}
	}

	t.histogram.Record(context.Background(), duration.Seconds(), metric.WithAttributes(allLabels...))
}

// getOrCreateCounter gets or creates a counter metric
func (p *OpenTelemetryProvider) getOrCreateCounter(name string) metric.Int64Counter {
	p.mu.RLock()
	counter, exists := p.counters[name]
	p.mu.RUnlock()

	if exists {
		return counter
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if counter, exists := p.counters[name]; exists {
		return counter
	}

	counter, err := p.meter.Int64Counter(name,
		metric.WithDescription("Counter metric: "+name),
	)
	if err != nil {
		// Return nil if counter creation fails
		return nil
	}

	p.counters[name] = counter
	return counter
}

// getOrCreateHistogram gets or creates a histogram metric
func (p *OpenTelemetryProvider) getOrCreateHistogram(name string) metric.Float64Histogram {
	p.mu.RLock()
	histogram, exists := p.histograms[name]
	p.mu.RUnlock()

	if exists {
		return histogram
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if histogram, exists := p.histograms[name]; exists {
		return histogram
	}

	histogram, err := p.meter.Float64Histogram(name,
		metric.WithDescription("Histogram metric: "+name),
	)
	if err != nil {
		// Return nil if histogram creation fails
		return nil
	}

	p.histograms[name] = histogram
	return histogram
}

// getOrCreateGauge gets or creates a gauge metric
func (p *OpenTelemetryProvider) getOrCreateGauge(name string) metric.Float64ObservableGauge {
	p.mu.RLock()
	gauge, exists := p.gauges[name]
	p.mu.RUnlock()

	if exists {
		return gauge
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if gauge, exists := p.gauges[name]; exists {
		return gauge
	}

	// For testing purposes, create a simple gauge without callback
	// In production, users would set up proper callbacks externally
	gauge, err := p.meter.Float64ObservableGauge(name,
		metric.WithDescription("Gauge metric: "+name),
	)
	if err != nil {
		// Return nil if gauge creation fails
		return nil
	}

	p.gauges[name] = gauge
	return gauge
}

// convertLabels converts map[string]string to OpenTelemetry attributes
func (p *OpenTelemetryProvider) convertLabels(labels map[string]string) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels)+len(p.config.Labels))

	// Add global labels first
	for key, value := range p.config.Labels {
		if key != "" && value != "" {
			attrs = append(attrs, attribute.String(key, value))
		}
	}

	// Add provided labels
	for key, value := range labels {
		if key != "" && value != "" {
			attrs = append(attrs, attribute.String(key, value))
		}
	}

	return attrs
}
