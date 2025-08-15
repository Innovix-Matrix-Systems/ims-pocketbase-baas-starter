# Metrics Collection Guide

This document explains how to use the `pkg/metrics` package to collect metrics and instrument your application.

## Configuration

Enable metrics in your `.env` file:

```bash
METRICS_ENABLED=true
METRICS_PROVIDER=prometheus          # or opentelemetry, disabled
METRICS_NAMESPACE=ims_pocketbase
```

## Basic Usage

### Get Metrics Provider

```go
import "ims-pocketbase-baas-starter/pkg/metrics"

metricsProvider := metrics.GetInstance()
```

### Metric Types

```go
// Counter - count events
metricsProvider.IncrementCounter("user_registrations_total", map[string]string{
    "method": "email",
})

// Histogram - measure distributions (response times, sizes)
metricsProvider.RecordHistogram("request_duration_seconds", 0.125, map[string]string{
    "endpoint": "/api/users",
})

// Gauge - current values (connections, queue size)
metricsProvider.SetGauge("active_connections", 42, map[string]string{
    "server": "web-01",
})

// Timer - measure operation duration
timer := metricsProvider.StartTimer("database_query_duration_seconds", map[string]string{
    "table": "users",
})
// ... perform operation ...
timer.Stop()
```

## Instrumentation Helpers

The package includes helper functions that automatically handle timing, counting, and error tracking:

### Instrument Functions

```go
// Using helper function (recommended)
func processUser(userID string) error {
    metricsProvider := metrics.GetInstance()

    return metrics.InstrumentRecordOperation(metricsProvider, "users", "process", func() error {
        return doProcessing(userID)
    })
}
```

### Instrument HTTP Handlers

```go
func HandleUserStats(e *core.RequestEvent) error {
    metricsProvider := metrics.GetInstance()

    return metrics.InstrumentHTTPHandler(metricsProvider, "GET", "/api/user-stats", func() error {
        stats, err := computeUserStats(e.App)
        if err != nil {
            return err
        }
        return e.JSON(200, stats)
    })
}
```

### Instrument Jobs

```go
func (h *EmailJobHandler) Handle(ctx *cronutils.CronExecutionContext, job *jobutils.JobData) error {
    metricsProvider := metrics.GetInstance()

    return metrics.InstrumentJobHandler(metricsProvider, "email", func() error {
        return h.processEmailJob(job)
    })
}
```

### Instrument Hooks

```go
func HandleUserUpdate(e *core.RecordEvent) error {
    metricsProvider := metrics.GetInstance()

    return metrics.InstrumentHook(metricsProvider, "user_update", func() error {
        return updateUserSettings(e.Record)
    })
}
```

## Other Helpers

```go
// Measure execution time
err := metrics.MeasureExecutionTime(metricsProvider, "operation_duration_seconds",
    map[string]string{"type": "data_processing"}, func() error {
        return processData()
    })

// Record cache operations
metrics.InstrumentCacheOperation(metricsProvider, true)  // cache hit
metrics.InstrumentCacheOperation(metricsProvider, false) // cache miss

// Record queue size
metrics.RecordQueueSize(metricsProvider, "email_queue", 25)
```

## Accessing Metrics

- **Prometheus**: http://localhost:8090/metrics
- **Grafana**: http://localhost:3000 (admin/admin)

## Best Practices

```go
// Good - consistent labels
labels := map[string]string{
    "operation": "create",
    "method": "api",
}

// Bad - avoid high cardinality (unique user IDs)
// Don't use user_id as label

// Good naming conventions
"http_requests_total"           // Use underscores
"request_duration_seconds"      // Include units
"job_queue_size"               // Descriptive names
```

## Safety & Resilience

**Your application is completely safe when Prometheus/Grafana are down!**

The metrics package is designed with multiple safety layers:

### 1. **Graceful Degradation**

- If metrics are disabled (`METRICS_ENABLED=false`), a no-op provider is used
- If Prometheus/Grafana are down, metrics calls simply do nothing
- **Zero impact on application performance or functionality**

### 2. **Panic Recovery**

All metrics operations use `Safe*` functions that recover from panics:

```go
// Even if Prometheus client panics, your app continues running
SafeIncrementCounter(provider, "user_actions", labels)
SafeRecordHistogram(provider, "response_time", 0.5, labels)
```

### 3. **Null Provider Handling**

```go
// Always safe - checks for nil provider
if provider != nil {
    provider.IncrementCounter(name, labels)
}
```

### 4. **Automatic Fallback**

- Unknown provider → No-op provider
- Configuration errors → No-op provider
- Network issues → Operations silently succeed

### 5. **No External Dependencies**

- Metrics are collected in-memory
- Only the `/metrics` endpoint requires Prometheus to scrape
- If scraping fails, metrics continue collecting locally

**Bottom line: Your core application logic is never affected by metrics failures.**

## Troubleshooting

### Metrics Not Showing in Grafana

If you can see metrics in Prometheus (`http://localhost:8090/metrics`) but not in Grafana:

1. **Check Prometheus is scraping your app**:

   - Visit `http://localhost:9090/targets`
   - Ensure `pocketbase-app` target shows "UP" status

2. **Verify metrics in Prometheus**:

   - Go to `http://localhost:9090/graph`
   - Try queries like: `ims_pocketbase_hook_execution_total` or `ims_pocketbase_job_execution_total`

3. **Check Grafana dashboard queries**:

   - Your metrics use the namespace `ims_pocketbase_`
   - Hook metrics: `ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}`
   - Job metrics: `ims_pocketbase_job_execution_total{job_type="email_job"}`

4. **Restart services if needed**:
   ```bash
   docker-compose -f docker-compose.dev.yml restart grafana prometheus
   ```

### Common Metric Names

Your instrumented code creates these metrics:

- `ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}` (from user hooks)
- `ims_pocketbase_job_execution_total{job_type="email_job"}` (from email jobs)
- `ims_pocketbase_http_requests_total{method="GET",path="/api/users"}` (from HTTP handlers)

### Useful Grafana Queries

```promql
# Rate of hook executions
rate(ims_pocketbase_hook_execution_total[5m])

# Job execution rate
rate(ims_pocketbase_job_execution_total[5m])

# 95th percentile response time
histogram_quantile(0.95, rate(ims_pocketbase_http_request_duration_seconds_bucket[5m]))
```

That's it! Simple metrics collection for better observability.
