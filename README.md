# IMS PocketBase BaaS Starter

A Backend-as-a-Service (BaaS) starter kit built with PocketBase Go framework, enabling custom API routes, business logic, and middleware alongside PocketBase's built-in features. Includes Role-Based Access Control (RBAC), environment-based configuration, and development tools.

## Features

- 🚀 **PocketBase Go Framework** - Full PocketBase functionality with Go extensibility
- 🔐 **RBAC System** - Role-based access control with permissions and roles
- 🛠️ **Custom API Routes** - Add your own REST endpoints and business logic
- 🔧 **Custom Middleware** - Implement Custom Middleware according to your needs
- ⏰ **Cron Jobs & Job Queue** - Scheduled tasks and dynamic job processing with concurrent workers
- 📧 **Email Integration** - SMTP configuration with MailHog for development
- 📚 **Auto API Documentation** - Swagger UI, ReDoc, OpenAPI JSON with Postman compatibility
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

## Makefile Commands

| Development                              | Production                                    | Utility                                  |
| ---------------------------------------- | --------------------------------------------- | ---------------------------------------- |
| `dev` - Start dev environment            | `build` - Build production image              | `help` - Show all commands               |
| `dev-build` - Build dev image            | `start` - Start containers                    | `generate-key` - Generate encryption key |
| `dev-logs` - Show dev logs               | `stop` - Stop containers                      | `setup-env` - Setup environment file     |
| `dev-clean` - Clean dev env              | `restart` - Restart containers                | `test` - Run tests                       |
| `dev-data-clean` - Clean dev data        | `down` - Stop and remove containers           | `lint` - Run linter                      |
| `dev-start` - Alias for dev              | `logs` - Show container logs                  | `format` - Format Go code                |
| `dev-status` - Show dev container status | `clean` - Remove containers, networks, images | `status` - Show container status         |
|                                          | `clean-data` - Remove only volumes            | `prod-start` - Alias for start           |

## Environment Configuration

Copy `env.example` to `.env` and configure the following:

### App Configuration

- `APP_NAME` - Application name
- `APP_URL` - Application URL

### Job processing settings

- `JOB_MAX_WORKERS` - Concurrent workers (default: 5)
- `JOB_BATCH_SIZE` - Jobs per cron run (default: 50)
- `JOB_MAX_RETRIES` - Maximum retry attempts (default: 3)
- `ENABLE_SYSTEM_QUEUE_CRON` - Enable queue processing (default: true)

### SMTP Configuration (for email)

- `SMTP_ENABLED` - Enable/disable SMTP
- `SMTP_HOST` - SMTP server host
- `SMTP_PORT` - SMTP server port
- `SMTP_USERNAME` - SMTP username
- `SMTP_PASSWORD` - SMTP password
- `SMTP_AUTH_METHOD` - Authentication method
- `SMTP_TLS` - Enable TLS

### S3 Configuration (for file storage)

- `S3_ENABLED` - Enable/disable S3
- `S3_BUCKET` - S3 bucket name
- `S3_REGION` - S3 region
- `S3_ENDPOINT` - S3 endpoint
- `S3_ACCESS_KEY` - S3 access key
- `S3_SECRET` - S3 secret key

### Security

- `PB_ENCRYPTION_KEY` - PocketBase encryption key (32 characters)

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
   - API Documentation (Swagger): http://localhost:8090/api-docs
   - API Documentation (ReDoc): http://localhost:8090/api-docs/redoc
   - OpenAPI JSON: http://localhost:8090/api-docs/openapi.json
   - MailHog Web UI: http://localhost:8025

## Database

The application includes:

- **Migrations** - Database schema setup
- **Seeders** - Initial data seeding (RBAC, super admin)
- **Collections** - User management, roles, permissions

For detailed information about database migrations and schema management, see the [Database Migrations Guide](docs/migrations.md).

## API Documentation

The application automatically generates comprehensive API documentation for all collections and custom routes:

- **Swagger UI** - Interactive API explorer at `http://localhost:8090/api-docs`
- **ReDoc** - Clean documentation interface at `http://localhost:8090/api-docs/redoc`
- **OpenAPI JSON** - Machine-readable spec at `http://localhost:8090/api-docs/openapi.json`
- **Postman Compatible** - Import OpenAPI JSON directly into Postman/Insomnia

**Features:**
- Auto-discovery of all PocketBase collections
- Complete CRUD operation documentation
- Authentication flow documentation
- File upload support with multipart forms
- Custom route integration
- Example data generation

For detailed information about the Swagger system, configuration, and usage, see the [Swagger Documentation Guide](docs/swagger.md).

## Cron Jobs & Job Queue System

The application includes a comprehensive background task processing system with:

- **Cron Jobs** - Scheduled tasks with environment-based control
- **Job Queue** - Dynamic job processing with concurrent workers
- **Built-in Handlers** - Email jobs, data processing jobs
- **Extensible Architecture** - Easy to add custom job types

For detailed information about cron jobs, job queue system, and creating custom handlers, see the [Cron Jobs & Job Queue Guide](docs/cron-jobs.md).

### Migration CLI Generator

The project includes a CLI tool to generate migration files automatically:

```bash
make migrate-gen name=add_user_profiles
```

**Features:** Automatic sequential numbering, name sanitization, input validation, and helpful next-step guidance.

## Project Structure

```
├── cmd/
│   ├── migrate-gen/     # Migration CLI generator
│   └── server/          # Application entry point
├── docs/                # Project documentation
│   ├── README.md       # Documentation index
│   ├── cron-jobs.md    # Cron jobs & job queue guide
│   ├── middleware.md   # Custom middleware guide
│   └── migrations.md   # Database migration guide
├── internal/
│   ├── app/            # Application setup and configuration
│   ├── crons/          # Cron job definitions and registration
│   ├── database/
│   │   ├── migrations/ # Database migrations
│   │   ├── schema/     # PocketBase schema files
│   │   └── seeders/    # Data seeders (RBAC, admin)
│   ├── handlers/
│   │   ├── cron/       # Cron job handlers
│   │   └── jobs/       # Job queue handlers
│   ├── jobs/           # Job processor management
│   ├── middlewares/    # HTTP middlewares (auth, permissions)
│   └── routes/         # Custom API route definitions
├── pkg/
│   ├── cronutils/      # Cron execution utilities
│   ├── jobutils/       # Job processing utilities
│   └── migration/      # Migration utilities and scanner
├── pb_data/            # PocketBase data directory
├── pb_public/          # PocketBase public assets
├── .github/            # GitHub workflows and templates
├── Dockerfile          # Production Dockerfile
├── Dockerfile.dev      # Development Dockerfile
├── docker-compose.yml  # Production compose
├── docker-compose.dev.yml # Development compose
├── Makefile           # Development commands
└── .env.example       # Environment template
```

## Contributing

Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to contribute to this project.

## License

This project is licensed under the [MIT License](LICENSE.md).
