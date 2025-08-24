# IMS PocketBase BaaS Starter

A production-ready Backend-as-a-Service (BaaS) starter kit that extends PocketBase's Go framework with enterprise-grade features. Combines PocketBase's real-time database, authentication, and file storage with custom business logic, advanced middleware, and comprehensive observability. Designed for rapid development with built-in RBAC, job processing, caching, and monitoring capabilities.

## Features

- 🚀 **PocketBase Go Framework** - Full PocketBase functionality with Go extensibility
- 🔐 **RBAC System** - Role-based access control with permissions and roles
- 🛠️ **Custom API Routes** - Add your own REST endpoints and business logic
- 🔧 **Custom Middleware** - Implement Custom Middleware according to your needs
- 🪝 **Event Hooks System** - Comprehensive event hook management with organized handlers for records, collections, requests, mailer, and realtime events
- ⚡ **Go Cache with TTL** - High-performance in-memory caching with Time-To-Live support for improved application performance
- ⏰ **Cron Jobs & Job Queue** - Scheduled tasks and dynamic job processing with concurrent workers
- 💻 **CLI Command Support** - Command-line interface support for custom scripts and tasks
- 📈 **Metrics & Observability** - Comprehensive monitoring with Prometheus metrics and OpenTelemetry support for performance tracking and system insights
- 📧 **Email Integration** - SMTP configuration with MailHog for development
- 📚 **Auto API Documentation** - Interactive auto generated API Docs: Scalar, Swagger UI, ReDoc, OpenAPI JSON with Postman compatibility
- 🐳 **Docker Support** - Production and development environments
- 🔄 **Hot Reload** - Development environment with automatic code reloading
- ⚙️ **Environment Configuration** - Flexible configuration via environment variables
- 📊 **Future-Proof Migrations** - Automated database setup, seeding, and schema evolution

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.24+ (for local development)

### Development Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/Innovix-Matrix-Systems/ims-pocketbase-baas-starter.git
   # you can also use the use template button and create your onw project from this template
   cd ims-pocketbase-baas-starter
   ```

2. **Setup environment**

   ```bash
   make setup-env
   # Edit .env file with your configuration
   ```

3. **Generate encryption key**

   ```bash
   make generate-key
   # Copy the generated key to .env file
   ```

4. **Start development environment**
   ```bash
   make dev
   ```

### Production Setup

1. **Build and start production containers**
   ```bash
   make build
   make start
   ```

### Default Super Admin

- Email: `superadmin@ims.com`
- Password: `superadmin123456`

## Configuration

The application uses environment variables for configuration. Copy `env.example` to `.env` and update the values:

```bash
make setup-env
make generate-key  # Generate encryption key
# Edit .env file with your configuration
```

Key configuration areas include app settings, SMTP/email, S3 storage, job processing, rate limiting, and security. For complete configuration details, see the [Environment Configuration Guide](docs/environment-configuration.md).

## Available Commands

The project includes comprehensive Makefile commands for development and production:

```bash
make dev          # Start development environment
make dev-logs     # View development logs
make build        # Build production image
make start        # Start production containers
make test         # Run tests
make help         # Show all commands
```

For a complete list of commands and usage examples, see the [Makefile Commands Guide](docs/makefile-commands.md).

## Development Workflow

1. **Start development environment**

   ```bash
   make dev
   ```

2. **Make code changes** - Files are automatically watched and reloaded

3. **View logs**

   ```bash
   make dev-logs
   ```

4. **Access services**
   - PocketBase Admin: http://localhost:8090/\_/
   - API Documentation (API Docs): http://localhost:8090/api-docs
   - API Documentation (ReDoc): http://localhost:8090/api-docs/redoc
   - OpenAPI JSON: http://localhost:8090/api-docs/openapi.json
   - MailHog Web UI: http://localhost:8025
   - Grafana Dashboard: http://localhost:3000 (admin/admin)
   - Prometheus Metrics: http://localhost:9090

## Key Features

- **Database** - Migrations, seeders, and RBAC collections. See [Database Guide](docs/migrations.md)
- **API Documentation** - Auto-generated API Documentation, ReDoc, and OpenAPI JSON. See [API Docs Guide](docs/apidoc.md)
- **Background Jobs** - Cron jobs and job queue system. See [Jobs Guide](docs/cron-jobs.md)
- **Event Hooks** - Comprehensive hook system for extending functionality. See [Hooks Guide](docs/hooks.md)
- **Caching** - High-performance TTL cache system. See [Caching Guide](docs/caching.md)
- **Metrics & Observability** - Prometheus metrics and OpenTelemetry support. See [Metrics Guide](docs/metrics.md)
- **CLI Commands** - Command-line interface for administrative tasks including permission sync and health checks. See [CLI Commands Guide](docs/cli-commands.md)
- **Migration CLI** - Generate migrations with `make migrate-gen name=your_migration`

## Project Structure

This project follows Go project layout standards with clean architecture and modular design. The codebase is organized into:

- **`cmd/`** - Application entry points (server, CLI tools)
- **`internal/`** - Private application code (handlers, hooks, middleware)
- **`pkg/`** - Reusable packages (cache, logger, metrics, utilities)
- **`docs/`** - Comprehensive project documentation
- **`monitoring/`** - Development and production monitoring configurations

For detailed information about the complete project structure, package organization, architectural decisions, and navigation guidance, see the **[📁 docs/](docs/)** folder which contains comprehensive guides for all aspects of the project.

## Contributing

Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to contribute to this project.

## License

This project is licensed under the [MIT License](LICENSE.md).
