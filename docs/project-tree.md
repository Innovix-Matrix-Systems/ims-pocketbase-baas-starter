# Project Structure Guide

This document provides a comprehensive overview of the IMS PocketBase BaaS Starter project structure, explaining the purpose and organization of each directory and key files.

## Directory Structure Overview

```
ims-pocketbase-baas-starter/
├── 📁 cmd/                     # Application entry points
├── 📁 monitoring/              # Monitoring configurations (Prometheus, Grafana)
├── 📁 docs/                    # Project documentation
├── 📁 internal/                # Private application code
├── 📁 pb_data/                 # PocketBase data directory
├── 📁 pb_public/               # PocketBase public assets
├── 📁 pkg/                     # Reusable packages
├── 📁 scripts/                 # Build and setup scripts
├── 🐳 Dockerfile               # Production container definition
├── 🐳 docker-compose.yml       # Production container orchestration
├── 🐹 go.mod                   # Go module definition
├── ⚙️ makefile                 # Development commands
└── 📄 README.md                # Main project documentation
```

## Detailed Directory Breakdown

### 📁 `cmd/` - Application Entry Points

Contains the main application executables following Go project layout standards.

```
cmd/
├── migrate-gen/          # Migration CLI generator
│   ├── main.go          # CLI entry point
│   ├── cli.go           # Command-line interface logic
│   ├── template.go      # Migration template generation
│   └── types.go         # CLI-specific types
└── server/              # Main application server
    └── main.go          # Server entry point
```

**Purpose:** Separates different executable commands, making the project modular and following Go conventions.

### 📁 `internal/` - Private Application Code

Contains application-specific code that should not be imported by other projects.

```
internal/
├── app/                 # Application setup and configuration
│   ├── app.go          # Main app initialization and DI orchestration
│   └── app_test.go     # Application setup tests
├── crons/              # Cron job definitions
│   └── crons.go        # Cron job registration and configuration
├── database/           # Database-related code
│   ├── migrations/     # Database schema migrations
│   ├── schema/         # PocketBase schema JSON files
│   └── seeders/        # Data seeding utilities
├── handlers/           # Business logic handlers
│   ├── cron/          # Cron job handlers
│   ├── export/        # Data export handlers
│   ├── hook/          # Event hook handlers
│   ├── jobs/          # Job queue handlers
│   └── route/         # Custom route handlers
├── hooks/             # Event hook registration
│   ├── hooks.go       # Hook registration orchestration
│   └── hooks_test.go  # Hook system tests
├── jobs/              # Job management
│   ├── jobs.go        # Job handler registration (new pattern)
│   └── manager.go     # Job manager singleton
├── middlewares/       # HTTP middlewares
│   ├── middlewares.go # Middleware registration (new pattern)
│   ├── auth.go        # Authentication middleware
│   ├── metrics.go     # Metrics collection middleware
│   └── permission.go  # Permission-based access control
├── routes/            # Custom API routes
│   └── routes.go      # Route registration (new pattern)
└── apidoc/           # API documentation generation
    ├── generator.go   # OpenAPI spec generation
    ├── discovery.go   # Collection discovery
    ├── schema.go      # Schema generation
    └── endpoints.go   # API docs endpoints
```

### 📁 `pkg/` - Reusable Packages

Contains reusable packages that could potentially be imported by other projects.

```
pkg/
├── cache/             # Caching system
│   ├── cache.go      # Cache service with TTL support
│   └── cache_test.go # Cache system tests
├── common/            # Common utilities
│   ├── env.go        # Environment variable utilities
│   ├── response.go   # HTTP response utilities
│   └── route.go      # Route configuration
├── cronutils/         # Cron execution utilities
│   ├── utils.go      # Cron validation and execution context
│   └── utils_test.go # Cron utilities tests
├── jobutils/          # Job processing utilities
│   ├── processor.go  # Job processor implementation
│   ├── types.go      # Job-related types and interfaces
│   ├── payload.go    # Job payload parsing utilities
│   ├── file.go       # File handling for jobs
│   └── worker_pool.go # Concurrent job processing
├── logger/            # Centralized logging system
│   ├── logger.go     # Logger singleton implementation
│   ├── utils.go      # Logger utilities
│   └── logger_test.go # Logger tests
├── metrics/           # Metrics and observability
│   ├── metrics.go    # Main metrics interface and factory
│   ├── config.go     # Configuration management
│   ├── prometheus.go # Prometheus implementation
│   ├── opentelemetry.go # OpenTelemetry implementation
│   ├── noop.go       # No-op implementation
│   ├── instrumentation.go # Helper functions
│   ├── types.go      # Metric types and constants
│   └── *_test.go     # Comprehensive test suite
├── migration/         # Migration utilities
│   ├── scanner.go    # Migration file scanning
│   ├── filesystem.go # File system operations
│   └── *_test.go     # Migration tests
└── permission/        # Permission system
    ├── permissions.go # Permission constants and definitions
    └── permissions_test.go # Permission tests
```

### 📊 `monitoring/` - Monitoring Configurations

Contains monitoring infrastructure configurations for both development and production environments.

```
monitoring/
├── local/             # Development monitoring setup
│   ├── grafana/       # Grafana configuration
│   │   ├── dashboards/    # Pre-built dashboards
│   │   └── provisioning/ # Grafana provisioning config
│   └── prometheus/    # Prometheus configuration
│       └── prometheus.yml # Local metrics scraping
└── production/        # Production monitoring setup
    ├── grafana/       # Production Grafana config
    ├── prometheus/    # Production Prometheus config
    └── alertmanager/  # Alert management configuration
```

### 📁 `docs/` - Project Documentation

Comprehensive documentation for all aspects of the project.

```
docs/
├── README.md              # Documentation index
├── dependency-injection.md # DI patterns and strategies
├── cron-jobs.md          # Job queue and cron system
├── hooks.md              # Event hooks system
├── logger.md             # Logging system
├── middleware.md         # Custom middleware
├── migrations.md         # Database migrations
├── apidoc.md            # API documentation
├── docker-metrics.md     # Metrics monitoring setup
├── git-hooks.md          # Git hooks setup
└── project-tree.md       # This file - project structure
```

## Key Design Principles

### 1. **Clean Architecture**
- Clear separation between `internal/` (private) and `pkg/` (public) code
- Layered architecture with proper dependency flow
- Domain-driven design with focused packages

### 2. **Dependency Injection Patterns**
- **Singleton Pattern**: Shared services (`pkg/metrics`, `pkg/logger`, `pkg/cache`)
- **Constructor Injection**: Business logic components (`internal/handlers`)
- **Factory Pattern**: Configurable implementations (`pkg/metrics/config.go`)
- **Interface-Based**: Loose coupling throughout the application

### 3. **Testing Strategy**
- Co-located test files following Go conventions (`*_test.go`)
- Comprehensive test coverage for all packages
- Mock-friendly architecture with interface-based design
- Singleton reset functionality for isolated testing

### 4. **Configuration Management**
- Environment-driven configuration (`.env` files)
- Centralized environment utilities (`pkg/common/env.go`)
- Sensible defaults with override capabilities
- Production and development configurations

## File Naming Conventions

### Go Files
- `*.go` - Implementation files
- `*_test.go` - Test files (co-located with implementation)
- `types.go` - Type definitions and constants
- `config.go` - Configuration structures and loading

### Documentation Files
- `README.md` - Main documentation or directory index
- `*.md` - Specific topic documentation
- `CONTRIBUTING.md` - Contribution guidelines
- `LICENSE.md` - License information

### Configuration Files
- `.env` - Environment variables (not in version control)
- `.env.example` - Environment template
- `.env.production` - Production environment template
- `docker-compose.yml` - Production container configuration
- `docker-compose.dev.yml` - Development container configuration
- `Dockerfile` - Production container definition
- `makefile` - Development commands

### Monitoring Files
- `monitoring/local/` - Development monitoring setup
- `monitoring/production/` - Production monitoring deployment
- `monitoring/*/prometheus/` - Prometheus configurations
- `monitoring/*/grafana/` - Grafana dashboards and provisioning

## Package Dependencies

### Dependency Flow
```
internal/app/ (orchestrates everything)
    ↓
internal/middlewares/ (HTTP layer)
    ↓
internal/handlers/ (business logic)
    ↓
pkg/ (utilities and services)
```

### Key Relationships
- `internal/app/app.go` orchestrates all dependency injection
- `pkg/metrics/` provides observability across all layers
- `pkg/logger/` provides logging across all components
- `pkg/cache/` provides caching for performance optimization
- `internal/middlewares/` provides cross-cutting concerns
- `internal/handlers/` implements business logic

## Development Workflow Integration

### Hot Reload Support
- `.air.toml` - Air configuration for hot reload
- `docker-compose.dev.yml` - Development environment with volume mounts
- `Dockerfile.dev` - Development container with Air

### Code Quality
- `makefile` - Standardized development commands
- `.githooks/` - Pre-commit and pre-push validation
- `scripts/` - Setup and utility scripts

### Testing Infrastructure
- Comprehensive test coverage across all packages
- Benchmark tests for performance-critical components
- Integration tests for complex workflows
- Mock-friendly architecture for isolated testing

## Production Deployment

### Container Strategy
- Multi-stage Docker builds for optimized production images
- Separate development and production configurations
- Volume mounts for persistent data (`pb_data/`)
- Health checks and monitoring endpoints

### Monitoring and Observability
- **Local Development**: `monitoring/local/` - Prometheus + Grafana for development
- **Production Deployment**: `monitoring/production/` - Scalable monitoring infrastructure
- **Metrics Collection**: Prometheus scraping with configurable intervals
- **Visualization**: Grafana dashboards with pre-built PocketBase metrics
- **Alerting**: Alertmanager integration for production notifications
- **Structured Logging**: Multiple log levels with centralized collection

## Best Practices Demonstrated

1. **Go Project Layout**: Follows standard Go project structure
2. **Separation of Concerns**: Clear boundaries between layers
3. **Dependency Injection**: Multiple strategies for different use cases
4. **Testing**: Comprehensive test coverage with proper isolation
5. **Documentation**: Extensive documentation for all components
6. **Configuration**: Environment-driven with sensible defaults
7. **Containerization**: Production-ready Docker setup
8. **Code Quality**: Linting, formatting, and Git hooks

## Navigation Tips

- **Start with** `README.md` for project overview
- **Explore** `internal/app/app.go` to understand application initialization
- **Review** `pkg/` packages to understand core utilities
- **Check** `docs/` for detailed guides on specific topics
- **Use** `makefile` commands for development workflow
- **Refer to** `.env.example` for configuration options

This structure provides a solid foundation for building scalable, maintainable Backend-as-a-Service applications with PocketBase and Go.