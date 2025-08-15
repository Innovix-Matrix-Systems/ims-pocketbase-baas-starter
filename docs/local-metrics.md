# Local Development Setup with Metrics

This guide covers setting up the development environment with integrated metrics monitoring using Docker Compose.

## Services Overview

The development environment includes:

- **PocketBase Application** (Port 8090) - Main application with `/metrics` endpoint
- **Prometheus** (Port 9090) - Metrics collection and storage
- **Grafana** (Port 3000) - Metrics visualization and dashboards
- **MailHog** (Port 1025/8025) - Email testing service

## Quick Start

Use the Makefile commands for easy development:

```bash
# Start the complete development environment
make dev

# View logs from all services
make dev-logs

# Stop and clean up
make dev-clean
```

## Service Access

| Service          | URL                           | Credentials |
| ---------------- | ----------------------------- | ----------- |
| PocketBase       | http://localhost:8090         | -           |
| Metrics Endpoint | http://localhost:8090/metrics | -           |
| Prometheus       | http://localhost:9090         | -           |
| Grafana          | http://localhost:3000         | admin/admin |
| MailHog          | http://localhost:8025         | -           |

## Available Metrics

The PocketBase application exposes comprehensive metrics:

### HTTP Metrics

- `ims_pocketbase_http_requests_total` - Total HTTP requests by method/path
- `ims_pocketbase_http_request_duration_seconds` - Request duration histogram
- `ims_pocketbase_http_errors_total` - HTTP errors by status code

### Hook Metrics

- `ims_pocketbase_hook_execution_duration_seconds` - Hook execution time
- `ims_pocketbase_hook_execution_total` - Total hook executions
- `ims_pocketbase_hook_errors_total` - Hook execution errors

### Job Metrics

- `ims_pocketbase_job_execution_duration_seconds` - Job processing time
- `ims_pocketbase_job_execution_total` - Total jobs processed
- `ims_pocketbase_job_errors_total` - Job processing errors
- `ims_pocketbase_job_queue_size` - Current job queue size

### Business Metrics

- `ims_pocketbase_record_operations_total` - Record CRUD operations
- `ims_pocketbase_emails_sent_total` - Emails sent successfully
- `ims_pocketbase_cache_hits_total` - Cache hit count
- `ims_pocketbase_cache_misses_total` - Cache miss count

## Grafana Dashboards

Pre-built dashboards are automatically provisioned:

### PocketBase Overview Dashboard

- HTTP request rates and response times
- Error rates and status code distribution
- Hook execution metrics
- Job processing statistics
- Cache performance metrics

### System Metrics Dashboard

- Application performance metrics
- Resource utilization
- Queue sizes and processing rates

## Configuration

### Prometheus Configuration

The Prometheus configuration (`monitoring/local/prometheus/prometheus.yml`) includes:

```yaml
scrape_configs:
  - job_name: 'pocketbase-app'
    static_configs:
      - targets: ['pocketbase:8090']
    metrics_path: '/metrics'
    scrape_interval: 5s
```

### Grafana Configuration

Grafana is automatically configured with:

- Prometheus as the default data source
- Pre-built dashboards for PocketBase metrics
- Admin credentials: `admin/admin`

## Development Workflow

1. **Start Environment**
   ```bash
   make dev
   ```

2. **Access Services**
   - Develop your application normally
   - View metrics in Grafana: http://localhost:3000
   - Query metrics directly in Prometheus: http://localhost:9090

3. **Monitor Performance**
   - Watch request rates and response times
   - Monitor error rates and investigate issues
   - Track business metrics and user behavior

4. **Debug Issues**
   - Use metrics to identify performance bottlenecks
   - Correlate errors with specific operations
   - Monitor resource usage and scaling needs

## Troubleshooting

### Common Issues

1. **Metrics Not Appearing**
   - Check if `METRICS_ENABLED=true` in `.env`
   - Verify `METRICS_PROVIDER=prometheus`
   - Ensure `/metrics` endpoint is accessible

2. **Grafana Connection Issues**
   - Wait for all services to start completely
   - Check container logs: `make dev-logs`
   - Verify Prometheus is scraping metrics

3. **Performance Issues**
   - Adjust scrape intervals in `prometheus.yml` if needed
   - Monitor resource usage with `make dev-status`

## File Structure

```
monitoring/
├── local/
│   ├── prometheus/
│   │   └── prometheus.yml      # Local Prometheus configuration
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/        # Auto-configured data sources
│   │   └── dashboards/         # Dashboard provisioning
│   └── dashboards/
│       └── pocketbase-metrics.json  # Pre-built dashboard
└── README.md                   # This file (moved to docs/)
```

## Production Notes

This setup is for development only. For production:

- Use external Prometheus/Grafana instances
- Configure authentication and security
- Set up alerting rules
- Implement data retention policies
- Consider Prometheus Operator for Kubernetes

## Related Documentation

- [Main README](../README.md) - Project overview and setup
- [Makefile Commands](../README.md#makefile-commands) - Available development commands
- [Environment Configuration](../README.md#environment-configuration) - Configuration options
