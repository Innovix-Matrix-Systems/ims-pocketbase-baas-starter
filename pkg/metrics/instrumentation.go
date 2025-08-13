package metrics

import (
	"fmt"
	"time"
)

// InstrumentHook wraps hook execution with metrics collection
// It measures execution time and tracks success/failure rates
func InstrumentHook(provider MetricsProvider, hookType string, fn func() error) error {
	if provider == nil {
		return fn()
	}

	labels := map[string]string{
		LabelHookType: hookType,
	}

	// Start timing
	timer := SafeStartTimer(provider, MetricHookExecutionDuration, labels)
	defer SafeStopTimer(timer)

	// Increment total counter
	SafeIncrementCounter(provider, MetricHookExecutionTotal, labels)

	// Execute the function
	err := SafeExecute(fn)

	// Record success/failure
	statusLabels := make(map[string]string)
	for k, v := range labels {
		statusLabels[k] = v
	}
	if err != nil {
		statusLabels[LabelStatus] = LabelError
		SafeIncrementCounter(provider, MetricHookErrorsTotal, statusLabels)
	} else {
		statusLabels[LabelStatus] = LabelSuccess
	}

	return err
}

// InstrumentJobHandler wraps job handler execution with metrics collection
// It measures execution time and tracks job processing metrics
func InstrumentJobHandler(provider MetricsProvider, jobType string, fn func() error) error {
	if provider == nil {
		return fn()
	}

	labels := map[string]string{
		LabelJobType: jobType,
	}

	// Start timing
	timer := SafeStartTimer(provider, MetricJobExecutionDuration, labels)
	defer SafeStopTimer(timer)

	// Increment total counter
	SafeIncrementCounter(provider, MetricJobExecutionTotal, labels)

	// Execute the function
	err := SafeExecute(fn)

	// Record success/failure
	statusLabels := make(map[string]string)
	for k, v := range labels {
		statusLabels[k] = v
	}
	if err != nil {
		statusLabels[LabelStatus] = LabelError
		SafeIncrementCounter(provider, MetricJobErrorsTotal, statusLabels)
	} else {
		statusLabels[LabelStatus] = LabelSuccess
	}

	return err
}

// InstrumentHTTPHandler wraps HTTP handler execution with metrics collection
// It measures request duration and tracks HTTP metrics
func InstrumentHTTPHandler(provider MetricsProvider, method, path string, fn func() error) error {
	if provider == nil {
		return fn()
	}

	labels := map[string]string{
		LabelMethod: method,
		LabelPath:   path,
	}

	// Start timing
	timer := SafeStartTimer(provider, MetricHTTPRequestDuration, labels)
	defer SafeStopTimer(timer)

	// Increment total counter
	SafeIncrementCounter(provider, MetricHTTPRequestsTotal, labels)

	// Execute the function
	err := SafeExecute(fn)

	// Record success/failure with status code
	statusLabels := make(map[string]string)
	for k, v := range labels {
		statusLabels[k] = v
	}
	if err != nil {
		statusLabels[LabelStatusCode] = "500" // Internal server error
		statusLabels[LabelStatus] = LabelError
		SafeIncrementCounter(provider, MetricHandlerErrorsTotal, statusLabels)
	} else {
		statusLabels[LabelStatusCode] = "200" // Success
		statusLabels[LabelStatus] = LabelSuccess
	}

	return err
}

// InstrumentRecordOperation wraps record operations with metrics collection
func InstrumentRecordOperation(provider MetricsProvider, collection, operation string, fn func() error) error {
	if provider == nil {
		return fn()
	}

	labels := map[string]string{
		LabelCollection: collection,
		LabelOperation:  operation,
	}

	// Start timing
	timer := SafeStartTimer(provider, MetricHookExecutionDuration, labels)
	defer SafeStopTimer(timer)

	// Increment total counter
	SafeIncrementCounter(provider, MetricRecordOperationsTotal, labels)

	// Execute the function
	err := SafeExecute(fn)

	// Record success/failure
	statusLabels := make(map[string]string)
	for k, v := range labels {
		statusLabels[k] = v
	}
	if err != nil {
		statusLabels[LabelStatus] = LabelError
	} else {
		statusLabels[LabelStatus] = LabelSuccess
	}

	return err
}

// InstrumentEmailOperation wraps email operations with metrics collection
func InstrumentEmailOperation(provider MetricsProvider, fn func() error) error {
	if provider == nil {
		return fn()
	}

	// Start timing
	timer := SafeStartTimer(provider, MetricHookExecutionDuration, nil)
	defer SafeStopTimer(timer)

	// Execute the function
	err := SafeExecute(fn)

	// Record email metrics
	if err != nil {
		// Don't increment success counter on error
	} else {
		SafeIncrementCounter(provider, MetricEmailsSentTotal, nil)
	}

	return err
}

// InstrumentCacheOperation wraps cache operations with metrics collection
func InstrumentCacheOperation(provider MetricsProvider, hit bool) {
	if provider == nil {
		return
	}

	if hit {
		SafeIncrementCounter(provider, MetricCacheHitsTotal, nil)
	} else {
		SafeIncrementCounter(provider, MetricCacheMissesTotal, nil)
	}
}

// SafeExecute executes a function with panic recovery
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	return fn()
}

// SafeIncrementCounter safely increments a counter with error recovery
func SafeIncrementCounter(provider MetricsProvider, name string, labels map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			// In production, you might want to use a proper logger here
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		provider.IncrementCounter(name, labels)
	}
}

// SafeIncrementCounterBy safely increments a counter by value with error recovery
func SafeIncrementCounterBy(provider MetricsProvider, name string, value float64, labels map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		provider.IncrementCounterBy(name, value, labels)
	}
}

// SafeRecordHistogram safely records a histogram value with error recovery
func SafeRecordHistogram(provider MetricsProvider, name string, value float64, labels map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		provider.RecordHistogram(name, value, labels)
	}
}

// SafeSetGauge safely sets a gauge value with error recovery
func SafeSetGauge(provider MetricsProvider, name string, value float64, labels map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		provider.SetGauge(name, value, labels)
	}
}

// SafeStartTimer safely starts a timer with error recovery
func SafeStartTimer(provider MetricsProvider, name string, labels map[string]string) Timer {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		return provider.StartTimer(name, labels)
	}
	return &NoOpTimer{}
}

// SafeStopTimer safely stops a timer with error recovery
func SafeStopTimer(timer Timer) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if timer != nil {
		timer.Stop()
	}
}

// SafeRecordDuration safely records a duration with error recovery
func SafeRecordDuration(provider MetricsProvider, name string, duration time.Duration, labels map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			// Log panic but don't propagate it
			_ = r // Acknowledge the recovered panic
		}
	}()

	if provider != nil {
		provider.RecordDuration(name, duration, labels)
	}
}

// MeasureExecutionTime measures the execution time of a function and records it as a histogram
func MeasureExecutionTime(provider MetricsProvider, metricName string, labels map[string]string, fn func() error) error {
	if provider == nil {
		return fn()
	}

	start := time.Now()
	err := SafeExecute(fn)
	duration := time.Since(start)

	SafeRecordDuration(provider, metricName, duration, labels)
	return err
}

// WithMetrics is a helper function that provides a metrics provider to a function
// This is useful for dependency injection patterns
func WithMetrics(provider MetricsProvider, fn func(MetricsProvider) error) error {
	if provider == nil {
		provider = NewNoOpProvider()
	}
	return fn(provider)
}

// BatchIncrementCounters increments multiple counters atomically (as much as possible)
func BatchIncrementCounters(provider MetricsProvider, counters map[string]map[string]string) {
	if provider == nil {
		return
	}

	for name, labels := range counters {
		SafeIncrementCounter(provider, name, labels)
	}
}

// RecordQueueSize records the current size of a queue as a gauge
func RecordQueueSize(provider MetricsProvider, queueName string, size int) {
	if provider == nil {
		return
	}

	labels := map[string]string{
		"queue": queueName,
	}

	SafeSetGauge(provider, MetricJobQueueSize, float64(size), labels)
}
