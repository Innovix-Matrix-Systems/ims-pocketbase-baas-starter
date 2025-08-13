# Docker Development Setup with Metrics

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

### Hook Metrics

- `ims_pocketbase_hook_execution_total` - Hook executions by type
- `ims_pocketbase_hook_execution_duration_seconds` - Hook execution time
- `ims_pocketbase_hook_errors_total` - Hook execution errors

### Job Metrics

- `ims_pocketbase_job_execution_total` - Job executions by type
- `ims_pocketbase_job_execution_duration_seconds` - Job execution time
- `ims_pocketbase_job_errors_total` - Job execution errors
- `ims_pocketbase_job_queue_size` - Current job queue size

### Business Metrics

- `ims_pocketbase_record_operations_total` - Database record operations
- `ims_pocketbase_emails_sent_total` - Email delivery tracking
- `ims_pocketbase_cache_hits_total` / `ims_pocketbase_cache_misses_total` - Cache performance

## Configuration

### Metrics Configuration

The application loads metrics configuration from environment variables. Add to your `.env` file:

```bash
METRICS_PROVIDER=prometheus
METRICS_ENABLED=true
METRICS_NAMESPACE=ims_pocketbase
```

### Prometheus Configuration

Edit `docker/prometheus/prometheus.yml` to customize:

- Scrape intervals (default: 5s for PocketBase)
- Additional targets
- Alerting rules

### Grafana Dashboards

- Pre-configured dashboard: http://localhost:3000/d/pocketbase-metrics
- Custom dashboards: Add JSON files to `docker/grafana/dashboards/`
- Data sources: Automatically configured via `docker/grafana/provisioning/`

## Development Workflow

1. **Start environment**: `make dev`
2. **Generate traffic**: Make API calls, trigger hooks, process jobs
3. **View metrics**:
   - Raw: http://localhost:8090/metrics
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000/d/pocketbase-metrics
4. **Monitor and optimize** based on metrics data

## Troubleshooting

### Metrics Not Appearing

1. **Check metrics endpoint**:

   ```bash
   curl http://localhost:8090/metrics
   ```

2. **Verify Prometheus targets**:

   - Go to http://localhost:9090/targets
   - Ensure `pocketbase-app` shows as UP

3. **Check service logs**:
   ```bash
   make dev-logs
   # or specific service
   docker-compose -f docker-compose.dev.yml logs prometheus
   ```

### Grafana Issues

1. **Verify data source**: http://localhost:3000/datasources
2. **Test queries**: http://localhost:3000/explore
3. **Check dashboard**: http://localhost:3000/d/pocketbase-metrics

### Performance Considerations

- Metrics collection has minimal overhead
- Adjust scrape intervals in `prometheus.yml` if needed
- Monitor resource usage with `make dev-status`

## File Structure

```
docker/
├── prometheus/
│   └── prometheus.yml          # Prometheus configuration
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
