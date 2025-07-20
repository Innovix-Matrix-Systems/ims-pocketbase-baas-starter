# IMS PocketBase BaaS Starter

A Backend-as-a-Service (BaaS) starter kit built with PocketBase and Go, featuring Role-Based Access Control (RBAC), environment-based configuration, and development tools.

## Features

- 🚀 **PocketBase Backend** - Self-hosted backend with real-time subscriptions
- 🔐 **RBAC System** - Role-based access control with permissions and roles
- 📧 **Email Integration** - SMTP configuration with MailHog for development
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
   git clone <repository-url>
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

## Makefile Commands

### Development Commands

- `make dev` - Start development environment with hot reload
- `make dev-build` - Build development Docker image
- `make dev-logs` - Show development container logs
- `make dev-clean` - Clean development environment

### Production Commands

- `make build` - Build production Docker image
- `make start` - Start production containers
- `make stop` - Stop containers
- `make restart` - Restart containers
- `make down` - Stop and remove containers
- `make logs` - Show container logs
- `make clean` - Remove containers, networks, and images
- `make delete-data` - Remove containers, networks, images, and volumes

### Utility Commands

- `make help` - Show all available commands
- `make generate-key` - Generate encryption key
- `make setup-env` - Setup environment file
- `make test` - Run tests
- `make lint` - Run linter
- `make format` - Format Go code

## Environment Configuration

Copy `env.example` to `.env` and configure the following:

### App Configuration

- `APP_NAME` - Application name
- `APP_URL` - Application URL

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
   - MailHog Web UI: http://localhost:8025

## Database

The application includes:

- **Migrations** - Database schema setup
- **Seeders** - Initial data seeding (RBAC, super admin)
- **Collections** - User management, roles, permissions

For detailed information about database migrations and schema management, see the [Database Migrations Guide](docs/migrations.md).

### Default Super Admin

- Email: `admin@example.com`
- Password: `admin123456`

## Project Structure

```
├── cmd/
│   └── server/          # Application entry point
├── docs/                # Project documentation
│   ├── README.md       # Documentation index
│   └── migrations.md   # Database migration guide
├── internal/
│   ├── app/            # Application setup
│   ├── database/
│   │   ├── migrations/ # Database migrations
│   │   ├── schema/     # PocketBase schema
│   │   └── seeders/    # Data seeders
│   ├── handlers/       # HTTP handlers
│   ├── middlewares/    # HTTP middlewares
│   └── routes/         # Route definitions
├── pb_public/          # PocketBase public assets
├── Dockerfile          # Production Dockerfile
├── Dockerfile.dev      # Development Dockerfile
├── docker-compose.yml  # Production compose
├── docker-compose.dev.yml # Development compose
├── Makefile           # Development commands
└── .env.example       # Environment template
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Format code: `make format`
6. Submit a pull request

## License

This project is licensed under the MIT License.
