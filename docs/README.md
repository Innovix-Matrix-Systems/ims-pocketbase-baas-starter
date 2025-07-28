# Documentation

This directory contains comprehensive documentation for the IMS PocketBase BaaS Starter project.

## Available Documentation

### [Database Migrations Guide](migrations.md)

Complete guide for managing database schema changes and migrations:

- Migration strategy and best practices
- Step-by-step instructions for creating new migrations
- File structure and naming conventions
- Troubleshooting and recovery procedures
- Common migration patterns and examples

### [Custom Middleware Setup Guide](middleware.md)

Comprehensive guide for implementing and using authentication middleware:

- Middleware architecture and structure
- Protecting custom API routes
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
