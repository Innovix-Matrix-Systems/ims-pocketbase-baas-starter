# Documentation

This directory contains comprehensive documentation for the IMS PocketBase BaaS Starter project.

## Getting Started (Essential)

### [Project Structure Guide](project-tree.md)

Comprehensive overview of the project structure and organization:

- Complete directory structure breakdown with explanations
- Detailed package organization and responsibilities
- Key design principles and architectural patterns
- File naming conventions and best practices
- Package dependencies and relationships
- Development workflow integration
- Production deployment considerations
- Navigation tips and getting started guidance

### [Environment Configuration Guide](environment-configuration.md)

Comprehensive guide for configuring the application through environment variables:

- Complete configuration reference with examples
- App, SMTP, S3, job processing, and security settings
- Environment-specific configuration examples (dev/prod)
- Security best practices and validation
- Troubleshooting common configuration issues

### [Makefile Commands Guide](makefile-commands.md)

Complete reference for all available development and production commands:

- Development commands (dev, dev-build, dev-logs, dev-clean)
- Production commands (build, start, stop, restart, clean)
- Utility commands (test, lint, format, generate-key, setup-env)
- Common workflows and usage examples
- Tips and troubleshooting

### [CLI Commands Guide](cli-commands.md)

Complete guide for using and extending the command-line interface:

- Built-in commands (hello, version, health)
- Running commands in development and production
- Adding new custom commands
- Advanced features and best practices
- Testing and troubleshooting

### [Database Migrations Guide](migrations.md)

Complete guide for managing database schema changes and migrations:

- Migration strategy and best practices
- Step-by-step instructions for creating new migrations
- File structure and naming conventions
- Troubleshooting and recovery procedures
- Common migration patterns and examples

## Core Features

### [Custom Routes Guide](custom-routes.md)

Comprehensive guide for creating custom API routes and endpoints:

- Creating custom route handlers and registration following the new consistent pattern
- Route organization and file structure with array-based configuration
- Integration with API documentation
- Authentication and middleware integration
- Request/response handling and validation
- Testing custom routes and error handling
- Advanced routing patterns and best practices

### [Event Hooks System Guide](hooks.md)

Complete guide for implementing and managing PocketBase event hooks:

- Hook system architecture and organization
- Record, collection, request, mailer, and realtime hooks
- Creating custom hook handlers and registration
- Collection-specific hooks and execution order
- Best practices and common use cases
- Error handling and testing strategies
- Practical examples for audit logging, validation, and notifications

### [Custom Middleware Setup Guide](middleware.md)

Comprehensive guide for implementing and using authentication middleware:

- Middleware architecture and structure following the new consistent pattern
- Protecting custom API routes with array-based middleware configuration
- Applying middleware to default PocketBase routes
- Collection-specific authentication
- Testing and troubleshooting middleware
- Future extension possibilities

### [Cron Jobs & Job Queue Guide](cron-jobs.md)

Complete guide for background task processing and job queue management:

- Cron job system architecture and configuration
- Job queue processing with concurrent workers
- Built-in job handlers (email, data processing)
- Creating custom job handlers and cron jobs
- Job lifecycle management and error handling
- Performance optimization and monitoring
- Environment configuration and troubleshooting

### [Custom Email System Guide](custom-emails.md)

Complete guide for sending custom emails using the job queue system with template support:

- Email system architecture and SMTP configuration
- Creating HTML and text email templates with variables
- Sending emails via API, programmatically, and in event hooks
- Email job processing and payload structure
- Common email templates (welcome, password reset, verification)
- Testing with MailHog and troubleshooting
- Best practices and integration examples

## Performance & Monitoring

### [Caching System Guide](caching.md)

Complete guide for using the built-in TTL caching system:

- Cache system architecture and singleton pattern
- Basic caching operations (get, set, delete, flush)
- TTL (Time-To-Live) configuration and best practices
- Cache invalidation strategies and patterns
- Performance monitoring and debugging
- Integration with existing application components

### [Metrics Collection Guide](metrics.md)

Complete guide for collecting metrics and instrumenting your application:

- Metrics configuration and setup
- Basic metric types (counters, histograms, gauges, timers)
- Instrumentation patterns for functions, handlers, and jobs
- Business metrics examples and best practices
- Helper functions and common patterns
- Accessing metrics through Prometheus and Grafana

### [Centralized Logger System Guide](logger.md)

Complete guide for the unified logging system with singleton pattern:

- Logger architecture and singleton implementation
- Multiple log levels (DEBUG, INFO, WARN, ERROR)
- Database storage integration with PocketBase logger
- Structured logging with key-value pairs
- Configuration options and usage examples
- Best practices for application logging

## Documentation & Development

### [API Docs Guide](apidoc.md)

Comprehensive guide for the automatic API documentation system:

- Interactive API documentation, ReDoc, and OpenAPI JSON generation
- Collection discovery and schema generation
- Route generation for CRUD and authentication
- Custom route integration and configuration
- File upload documentation and multipart forms
- Postman/Insomnia integration and client SDK generation
- Architecture overview and troubleshooting

### [Local Metrics Setup Guide](local-metrics.md)

Complete guide for development environment with integrated metrics monitoring:

- Local development setup with Prometheus and Grafana
- Metrics collection and visualization configuration
- Pre-built dashboards for HTTP, hooks, jobs, and business metrics
- Development workflow with metrics monitoring
- Troubleshooting metrics collection and visualization
- Performance monitoring and optimization strategies

### [Git Hooks Setup Guide](git-hooks.md)

Guide for setting up Git hooks for code quality and automation:

- Pre-commit hooks for code formatting and linting
- Pre-push hooks for testing and validation
- Automated code quality checks
- Integration with development workflow

## Advanced Topics

### [Dependency Injection Guide](dependency-injection.md)

Comprehensive guide for understanding dependency injection patterns used throughout the project:

- Multiple DI strategies working together harmoniously
- Singleton pattern with lazy initialization for shared services
- Constructor injection for business logic components
- Factory pattern for configurable implementations
- Function injection for event-driven architecture
- Interface-based dependency injection for loose coupling
- Complete application startup DI flow with visual diagrams
- Testing strategies with mock dependencies and singleton reset
- Environment-driven configuration and best practices
- Production-ready architecture patterns and benefits

## Quick Links

- **[Project README](../README.md)** - Main project documentation
- **[Environment Configuration](../env.example)** - Configuration options
- **[Docker Setup](../docker-compose.yml)** - Container configuration
- **[Makefile Commands](../makefile)** - Available development commands

## Contributing to Documentation

Please see our [Contributing Guide](../CONTRIBUTING.md) for details on how to contribute to this project's documentation.

## Documentation Standards

- Use clear, concise language
- Include practical examples
- Provide troubleshooting sections
- Keep formatting consistent
- Update links when files are moved or renamed
