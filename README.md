# IMS PocketBase BaaS Starter

A Backend-as-a-Service (BaaS) starter kit built with PocketBase Go framework, enabling custom API routes, business logic, and middleware alongside PocketBase's built-in features. Includes Role-Based Access Control (RBAC), environment-based configuration, and development tools.

## Features

- 🚀 **PocketBase Go Framework** - Full PocketBase functionality with Go extensibility
- 🔐 **RBAC System** - Role-based access control with permissions and roles
- 🛠️ **Custom API Routes** - Add your own REST endpoints and business logic
- 🔧 **Custom Middleware** - Implement Custom Middleware according to your needs
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
   git clone https://github.com/Innovix-Matrix-Systems/ims-pocketbase-baas-starter.git

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

### Migration CLI Generator

The project includes a CLI tool to generate migration files automatically:

```bash
# Generate a new migration
make migrate-gen name=add_user_profiles

# Build the CLI tool
make migrate-gen-build
```

**Features:**

- Automatic sequential numbering
- Name sanitization to kebab-case
- Input validation and error handling
- Prevents duplicate migrations
- Provides helpful next-step guidance

**Example output:**

```
✓ Generated migration file: internal/database/migrations/0003_add-user-profiles.go
✓ Schema file expected at: internal/database/schema/0003_pb_schema.json

Next steps:
1. Design your collections in PocketBase Admin UI
2. Export collections to internal/database/schema/0003_pb_schema.json
3. Update the rollback function with collection names to delete
4. Test the migration in development environment
```

### Default Super Admin

- Email: `superadmin@ims.com`
- Password: `superadmin123456`

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

Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to contribute to this project.

## License

This project is licensed under the [MIT License](LICENSE.md).
