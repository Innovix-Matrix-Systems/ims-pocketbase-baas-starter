# Troubleshooting Metrics in Grafana

## Issue: email_job and user_create_setting metrics not showing in Grafana

### Solution Summary
The metrics are being correctly exposed by your application and scraped by Prometheus. The issue was that the Grafana dashboard wasn't configured to display these specific metrics.

## Verification Steps

### 1. Check if metrics are exposed by your application
```bash
curl http://localhost:8090/metrics | grep -E "(email_job|user_create_setting|hook_execution_total|job_execution_total)"
```

Expected output should include:
- `ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}`
- `ims_pocketbase_job_execution_total{job_type="email_job"}`

### 2. Check if Prometheus is scraping the metrics
Visit: http://localhost:9090/targets
- Look for the `pocketbase-app` target
- Status should be "UP"

### 3. Query metrics directly in Prometheus
Visit: http://localhost:9090/graph

Try these queries:
- `ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}`
- `ims_pocketbase_job_execution_total{job_type="email_job"}`
- `rate(ims_pocketbase_hook_execution_total[5m])`
- `rate(ims_pocketbase_job_execution_total[5m])`

### 4. Check Grafana Dashboard
Visit: http://localhost:3000 (admin/admin)

The updated dashboard now includes:
- Combined "Hook & Job Execution Rate" panel
- "Specific Metrics: Email Job & User Create Settings" panel
- "Execution Duration: Email Job & User Create Settings" panel

## Metric Names Reference

Your application exposes these metrics with the namespace `ims_pocketbase_`:

### Hook Metrics
- `ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}` - Total executions
- `ims_pocketbase_hook_execution_duration_seconds{hook_type="user_create_settings"}` - Execution time histogram

### Job Metrics  
- `ims_pocketbase_job_execution_total{job_type="email_job"}` - Total executions
- `ims_pocketbase_job_execution_duration_seconds{job_type="email_job"}` - Execution time histogram

## Useful Grafana Queries

### Rate of executions (per second)
```promql
rate(ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}[5m])
rate(ims_pocketbase_job_execution_total{job_type="email_job"}[5m])
```

### 95th percentile execution time
```promql
histogram_quantile(0.95, rate(ims_pocketbase_hook_execution_duration_seconds_bucket{hook_type="user_create_settings"}[5m]))
histogram_quantile(0.95, rate(ims_pocketbase_job_execution_duration_seconds_bucket{job_type="email_job"}[5m]))
```

### Total count
```promql
ims_pocketbase_hook_execution_total{hook_type="user_create_settings"}
ims_pocketbase_job_execution_total{job_type="email_job"}
```

## Common Issues

### Metrics not appearing
1. **Application not running**: Ensure your PocketBase app is running on port 8090
2. **Prometheus not scraping**: Check prometheus.yml configuration and restart Prometheus
3. **Grafana cache**: Restart Grafana container or refresh the dashboard
4. **No data**: Trigger the hooks/jobs to generate metrics data

### Dashboard not updating
1. Restart Grafana: `docker-compose -f docker-compose.dev.yml restart grafana`
2. Check time range in Grafana (top right)
3. Verify data source connection in Grafana settings

## Triggering Metrics for Testing

To generate test data:

### User Create Settings Hook
Create a new user via the API:
```bash
curl -X POST http://localhost:8090/api/collections/_pb_users_auth_/records \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","passwordConfirm":"password123"}'
```

### Email Job
This gets triggered automatically when certain email events occur in your application.