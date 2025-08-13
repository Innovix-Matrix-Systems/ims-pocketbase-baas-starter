package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusProvider implements MetricsProvider using Prometheus client library
type PrometheusProvider struct {
	config     Config
	registry   *prometheus.Registry
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges     map[string]*prometheus.GaugeVec
	mu         sync.RWMutex
}

// PrometheusTimer implements Timer for Prometheus metrics
type PrometheusTimer struct {
	histogram *prometheus.HistogramVec
	labels    prometheus.Labels
	startTime time.Time
}

// NewPrometheusProvider creates a new Prometheus metrics provider
func NewPrometheusProvider(config Config) *PrometheusProvider {
	registry := prometheus.NewRegistry()

	provider := &PrometheusProvider{
		config:     config,
		registry:   registry,
		counters:   make(map[string]*prometheus.CounterVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
	}

	return provider
}

// IncrementCounter increments a counter metric
func (p *PrometheusProvider) IncrementCounter(name string, labels map[string]string) {
	p.IncrementCounterBy(name, 1.0, labels)
}

// IncrementCounterBy increments a counter metric by a specific value
func (p *PrometheusProvider) IncrementCounterBy(name string, value float64, labels map[string]string) {
	counter := p.getOrCreateCounter(name, labels)
	if counter != nil {
		labelKeys := p.getLabelKeys(labels)
		promLabels := p.convertLabelsWithKeys(labels, labelKeys)
		counter.With(promLabels).Add(value)
	}
}

// RecordHistogram records a value in a histogram metric
func (p *PrometheusProvider) RecordHistogram(name string, value float64, labels map[string]string) {
	histogram := p.getOrCreateHistogram(name, labels)
	if histogram != nil {
		labelKeys := p.getLabelKeys(labels)
		promLabels := p.convertLabelsWithKeys(labels, labelKeys)
		histogram.With(promLabels).Observe(value)
	}
}

// SetGauge sets a gauge metric value
func (p *PrometheusProvider) SetGauge(name string, value float64, labels map[string]string) {
	gauge := p.getOrCreateGauge(name, labels)
	if gauge != nil {
		labelKeys := p.getLabelKeys(labels)
		promLabels := p.convertLabelsWithKeys(labels, labelKeys)
		gauge.With(promLabels).Set(value)
	}
}

// StartTimer starts a timer for measuring durations
func (p *PrometheusProvider) StartTimer(name string, labels map[string]string) Timer {
	histogram := p.getOrCreateHistogram(name, labels)
	if histogram == nil {
		// Return a no-op timer if histogram creation failed
		return &NoOpTimer{}
	}

	labelKeys := p.getLabelKeys(labels)
	promLabels := p.convertLabelsWithKeys(labels, labelKeys)

	return &PrometheusTimer{
		histogram: histogram,
		labels:    promLabels,
		startTime: time.Now(),
	}
}

// RecordDuration records a duration metric
func (p *PrometheusProvider) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	p.RecordHistogram(name, duration.Seconds(), labels)
}

// GetHandler returns the Prometheus HTTP handler for the /metrics endpoint
func (p *PrometheusProvider) GetHandler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Registry:          p.registry,
	})
}

// Shutdown gracefully shuts down the Prometheus provider
func (p *PrometheusProvider) Shutdown(ctx context.Context) error {
	// Prometheus client doesn't require explicit shutdown
	// Just clear internal state
	p.mu.Lock()
	defer p.mu.Unlock()

	p.counters = make(map[string]*prometheus.CounterVec)
	p.histograms = make(map[string]*prometheus.HistogramVec)
	p.gauges = make(map[string]*prometheus.GaugeVec)

	return nil
}

// Stop records the duration since timer creation
func (t *PrometheusTimer) Stop() {
	duration := time.Since(t.startTime)
	t.histogram.With(t.labels).Observe(duration.Seconds())
}

// StopWithLabels records the duration with additional labels
// Note: Additional labels must have been included when the histogram was created
func (t *PrometheusTimer) StopWithLabels(additionalLabels map[string]string) {
	duration := time.Since(t.startTime)

	// For Prometheus, we can only use labels that were defined when the metric was created
	// So we'll just use the original labels and ignore additional ones
	// This is a limitation of the Prometheus client library
	t.histogram.With(t.labels).Observe(duration.Seconds())
}

// getOrCreateCounter gets or creates a counter metric
func (p *PrometheusProvider) getOrCreateCounter(name string, labels map[string]string) *prometheus.CounterVec {
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

	labelKeys := p.getLabelKeys(labels)
	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.config.Namespace,
			Name:      name,
			Help:      "Counter metric: " + name,
		},
		labelKeys,
	)

	if err := p.registry.Register(counter); err != nil {
		// If registration fails, return nil to avoid panics
		return nil
	}

	p.counters[name] = counter
	return counter
}

// getOrCreateHistogram gets or creates a histogram metric
func (p *PrometheusProvider) getOrCreateHistogram(name string, labels map[string]string) *prometheus.HistogramVec {
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

	labelKeys := p.getLabelKeys(labels)
	histogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.config.Namespace,
			Name:      name,
			Help:      "Histogram metric: " + name,
			Buckets:   p.config.GetHistogramBuckets(),
		},
		labelKeys,
	)

	if err := p.registry.Register(histogram); err != nil {
		// If registration fails, return nil to avoid panics
		return nil
	}

	p.histograms[name] = histogram
	return histogram
}

// getOrCreateGauge gets or creates a gauge metric
func (p *PrometheusProvider) getOrCreateGauge(name string, labels map[string]string) *prometheus.GaugeVec {
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

	labelKeys := p.getLabelKeys(labels)
	gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.config.Namespace,
			Name:      name,
			Help:      "Gauge metric: " + name,
		},
		labelKeys,
	)

	if err := p.registry.Register(gauge); err != nil {
		// If registration fails, return nil to avoid panics
		return nil
	}

	p.gauges[name] = gauge
	return gauge
}

// getLabelKeys extracts label keys from a labels map and adds global labels
func (p *PrometheusProvider) getLabelKeys(labels map[string]string) []string {
	keySet := make(map[string]bool)

	// Add global label keys
	for key := range p.config.Labels {
		keySet[key] = true
	}

	// Add provided label keys
	for key := range labels {
		if key != "" {
			keySet[key] = true
		}
	}

	// Convert to slice
	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}

	return keys
}

// convertLabelsWithKeys converts map[string]string to prometheus.Labels ensuring all keys are present
func (p *PrometheusProvider) convertLabelsWithKeys(labels map[string]string, expectedKeys []string) prometheus.Labels {
	promLabels := make(prometheus.Labels)

	// Initialize all expected keys with empty values
	for _, key := range expectedKeys {
		promLabels[key] = ""
	}

	// Add global labels first
	for key, value := range p.config.Labels {
		if key != "" {
			promLabels[key] = value
		}
	}

	// Add provided labels (can override global labels)
	for key, value := range labels {
		if key != "" {
			promLabels[key] = value
		}
	}

	return promLabels
}
