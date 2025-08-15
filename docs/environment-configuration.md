# Environment Configuration Guide

This document provides comprehensive information about configuring the IMS PocketBase BaaS Starter through environment variables.

## Setup

Copy `env.example` to `.env` and configure the following variables according to your needs:

```bash
cp env.example .env
```

## Configuration Categories

### App Configuration

Basic application settings that define the core behavior.

- **`APP_NAME`** - Application name used in logs and UI
  - Default: `IMS_PocketBase_App`
  - Example: `MyApp_Production`

- **`APP_URL`** - Base URL where the application is accessible
  - Default: `http://localhost:8090`
  - Example: `https://api.myapp.com`

### Logging Configuration

Controls application logging behavior and retention.

- **`LOGS_MAX_DAYS`** - Maximum number of days to retain log files
  - Default: `7`
  - Example: `30` (for production environments)

### Job Processing Settings

Configuration for the background job queue and cron system.

- **`JOB_MAX_WORKERS`** - Number of concurrent workers for job processing
  - Default: `5`
  - Range: `1-20` (adjust based on server capacity)

- **`JOB_BATCH_SIZE`** - Number of jobs processed per cron execution
  - Default: `50`
  - Range: `10-200`

- **`JOB_MAX_RETRIES`** - Maximum retry attempts for failed jobs
  - Default: `3`
  - Range: `1-10`

- **`ENABLE_SYSTEM_QUEUE_CRON`** - Enable/disable automatic job queue processing
  - Default: `true`
  - Values: `true`, `false`

### SMTP Configuration (Email)

Email server configuration for sending notifications and system emails.

- **`SMTP_ENABLED`** - Enable/disable SMTP email functionality
  - Default: `true`
  - Values: `true`, `false`

- **`SMTP_HOST`** - SMTP server hostname
  - Development: `mailhog` (for Docker MailHog)
  - Production: `smtp.gmail.com`, `smtp.sendgrid.net`, etc.

- **`SMTP_PORT`** - SMTP server port
  - Development: `1025` (MailHog)
  - Production: `587` (TLS), `465` (SSL), `25` (plain)

- **`SMTP_USERNAME`** - SMTP authentication username
  - Leave empty for development with MailHog
  - Production: Your email or API key

- **`SMTP_PASSWORD`** - SMTP authentication password
  - Leave empty for development with MailHog
  - Production: Your password or API secret

- **`SMTP_AUTH_METHOD`** - Authentication method
  - Default: `PLAIN`
  - Options: `PLAIN`, `LOGIN`, `CRAM-MD5`

- **`SMTP_TLS`** - Enable TLS encryption
  - Default: `true`
  - Values: `true`, `false`

### S3 Configuration (File Storage)

Amazon S3 or S3-compatible storage configuration for file uploads.

- **`S3_ENABLED`** - Enable/disable S3 file storage
  - Default: `false`
  - Values: `true`, `false`

- **`S3_BUCKET`** - S3 bucket name
  - Example: `my-app-files`

- **`S3_REGION`** - AWS region
  - Default: `us-east-1`
  - Example: `eu-west-1`, `ap-southeast-1`

- **`S3_ENDPOINT`** - S3 endpoint URL
  - AWS: `https://s3.amazonaws.com`
  - MinIO: `http://localhost:9000`
  - DigitalOcean Spaces: `https://nyc3.digitaloceanspaces.com`

- **`S3_ACCESS_KEY`** - S3 access key ID
  - AWS: Your AWS access key
  - MinIO: Your MinIO access key

- **`S3_SECRET`** - S3 secret access key
  - AWS: Your AWS secret key
  - MinIO: Your MinIO secret key

### Batch Processing Configuration

Settings for batch operations and bulk data processing.

- **`BATCH_ENABLED`** - Enable/disable batch processing features
  - Default: `true`
  - Values: `true`, `false`

- **`BATCH_MAX_REQUESTS`** - Maximum requests per batch operation
  - Default: `100`
  - Range: `10-1000`

### Rate Limiting Configuration

API rate limiting settings to prevent abuse and ensure fair usage.

- **`RATE_LIMITS_ENABLED`** - Enable/disable rate limiting
  - Default: `true`
  - Values: `true`, `false`

- **`RATE_LIMITS_MAX_HITS`** - Maximum requests per time window
  - Default: `120`
  - Range: `10-10000`

- **`RATE_LIMITS_DURATION`** - Rate limit time window in seconds
  - Default: `60`
  - Range: `1-3600`

### Security Configuration

Critical security settings for production deployments.

- **`PB_ENCRYPTION_KEY`** - PocketBase encryption key (32 characters)
  - **Required for production**
  - Generate using: `openssl rand -base64 24`
  - Example: `your-32-char-encryption-key-here`

### Metrics Configuration

Observability and monitoring settings (when metrics package is enabled).

- **`METRICS_PROVIDER`** - Metrics provider type
  - Options: `prometheus`, `opentelemetry`, `disabled`
  - Default: `disabled`

- **`METRICS_ENABLED`** - Master switch for metrics collection
  - Default: `false`
  - Values: `true`, `false`

- **`METRICS_NAMESPACE`** - Metrics namespace prefix
  - Default: `ims_pocketbase`
  - Example: `myapp_production`

- **`METRICS_PATH`** - Prometheus metrics endpoint path
  - Default: `/metrics`

### OpenTelemetry Configuration

OpenTelemetry-specific settings for distributed tracing and metrics.

- **`OTEL_EXPORTER_OTLP_ENDPOINT`** - OTLP endpoint URL
  - Example: `http://localhost:4317`

- **`OTEL_EXPORTER_OTLP_HEADERS`** - Additional headers for OTLP export
  - Example: `api-key=secret`

- **`OTEL_EXPORTER_OTLP_INSECURE`** - Use insecure connection
  - Default: `true` (for development)
  - Values: `true`, `false`

- **`OTEL_METRIC_EXPORT_INTERVAL`** - Metric export interval
  - Default: `30s`
  - Format: Go duration string (`30s`, `1m`, `5m`)

## Environment-Specific Examples

### Development Environment (.env)
```bash
APP_NAME=MyApp_Development
APP_URL=http://localhost:8090

# Use MailHog for email testing
SMTP_ENABLED=true
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=

# Disable S3 for local development
S3_ENABLED=false

# Enable metrics for development monitoring
METRICS_ENABLED=true
METRICS_PROVIDER=prometheus

# Development encryption key (generate your own)
PB_ENCRYPTION_KEY=dev-key-change-in-production-32
```

### Production Environment (.env)
```bash
APP_NAME=MyApp_Production
APP_URL=https://api.myapp.com

# Production SMTP settings
SMTP_ENABLED=true
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_TLS=true

# Production S3 settings
S3_ENABLED=true
S3_BUCKET=myapp-production-files
S3_REGION=us-east-1
S3_ACCESS_KEY=your-aws-access-key
S3_SECRET=your-aws-secret-key

# Production job processing
JOB_MAX_WORKERS=10
JOB_BATCH_SIZE=100

# Production rate limiting
RATE_LIMITS_MAX_HITS=1000
RATE_LIMITS_DURATION=60

# Production metrics
METRICS_ENABLED=true
METRICS_PROVIDER=opentelemetry
OTEL_EXPORTER_OTLP_ENDPOINT=https://your-otel-collector:4317

# Secure encryption key (generate using openssl)
PB_ENCRYPTION_KEY=your-secure-32-char-production-key
```

## Security Best Practices

1. **Never commit `.env` files** to version control
2. **Generate unique encryption keys** for each environment
3. **Use strong passwords** for SMTP and S3 credentials
4. **Enable TLS** for all external connections in production
5. **Rotate credentials regularly** in production environments
6. **Use environment-specific configurations** (dev/staging/prod)

## Validation

The application validates configuration on startup and will:
- Log warnings for missing optional configurations
- Fail to start if required configurations are missing
- Use sensible defaults where possible
- Provide clear error messages for invalid values

## Troubleshooting

### Common Issues

1. **SMTP Connection Failed**
   - Verify `SMTP_HOST` and `SMTP_PORT`
   - Check firewall settings
   - Validate credentials

2. **S3 Upload Errors**
   - Verify bucket permissions
   - Check access key and secret
   - Ensure bucket exists in specified region

3. **Job Processing Not Working**
   - Check `ENABLE_SYSTEM_QUEUE_CRON=true`
   - Verify `JOB_MAX_WORKERS > 0`
   - Review application logs

4. **Rate Limiting Too Restrictive**
   - Increase `RATE_LIMITS_MAX_HITS`
   - Adjust `RATE_LIMITS_DURATION`
   - Consider disabling for development

For more troubleshooting help, check the application logs or refer to the specific feature documentation in the [docs/](.) folder.