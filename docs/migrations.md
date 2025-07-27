# Database Migrations Guide

This document outlines the database migration strategy and best practices for the IMS PocketBase BaaS Starter project.

## Overview

Our migration system uses PocketBase's built-in migration framework with an incremental approach for adding new collections and schema changes. This ensures safe, reversible database changes in both development and production environments.

## Migration Strategy

### Initial Migration (0001_init.go)

The initial migration imports the complete database schema from `0001_pb_schema.json` and sets up:

- All base collections (users, roles, permissions)
- System collections (\_superusers, \_authOrigins, etc.)
- Application settings from environment variables
- Initial data seeding (superuser, RBAC data)

### Future Migrations (Incremental Approach)

For new collections or schema changes, we use targeted migrations that only affect specific collections:

1. **Export only new collections** from PocketBase Admin UI
2. **Create numbered migration files** with specific changes
3. **Import only the new collections** in the migration
4. **Provide precise rollback functionality**

## File Structure

```
internal/database/
├── migrations/
│   ├── 0001_init.go              # Initial schema and setup
│   ├── 0002_add_user_settings.go # User settings collections
│   ├── 0003_add_audit_logs.go    # Example: Audit logging
│   └── utils.go                  # Migration helper functions
├── schema/
│   ├── 0001_pb_schema.json       # Complete initial schema
│   ├── 0002_pb_schema.json       # User settings collections for migration 0002
│   ├── 0003_pb_schema.json       # Example: Future schema additions
│   └── README.md                 # Schema documentation
└── seeders/
    ├── rbac_seeder.go            # Role-based access control seeding
    └── superuser_seeder.go       # Superuser creation
```

## Migration CLI Generator

The project includes a CLI tool to automatically generate migration files with proper structure and naming conventions.

### Quick Start

Generate a new migration using the makefile:

```bash
make migrate-gen name=add_user_profiles
```

This will:
- Scan existing migrations to determine the next sequential number
- Create a properly structured migration file (e.g., `0003_add_user_profiles.go`)
- Display the expected schema file location
- Provide next steps for completing the migration

### Usage Examples

```bash
# Generate migration with underscore naming
make migrate-gen name=add_user_settings

# Generate migration with hyphen naming  
make migrate-gen name=add-audit-logs

# Generate migration with mixed case (will be converted to kebab-case)
make migrate-gen name=AddNotificationSystem
```

### Generated File Structure

The CLI generates migration files with this structure:

```go
package migrations

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        // Forward migration
        schemaPath := filepath.Join("internal", "database", "schema", "0003_pb_schema.json")
        // ... schema import logic
        
        // TODO: Add any data seeding specific to these collections
        return nil
    }, func(app core.App) error {
        // Rollback migration
        collectionsToDelete := []string{
            // TODO: Add collection names to delete during rollback
        }
        // ... rollback logic
        return nil
    })
}
```

### CLI Features

- **Automatic Numbering**: Scans existing migrations and assigns the next sequential number
- **Name Sanitization**: Converts migration names to kebab-case format
- **Input Validation**: Ensures migration names contain only valid characters
- **Duplicate Prevention**: Prevents overwriting existing migration files
- **Helpful Output**: Shows file paths and next steps after generation

### Building the CLI

To build a standalone binary:

```bash
make migrate-gen-build
```

This creates `bin/migrate-gen` which can be used directly:

```bash
./bin/migrate-gen add_user_profiles
```

## Creating New Migrations

### Step 1: Generate Migration File

Use the CLI generator to create the migration file:

```bash
make migrate-gen name=your_migration_name
```

### Step 2: Design Collections

1. Use PocketBase Admin UI to design your new collections
2. Test the collections thoroughly in development
3. Document the purpose and relationships

### Step 3: Export Schema

1. Export only the new collections from PocketBase Admin UI
2. Save as `internal/database/schema/XXXX_pb_schema.json` (the CLI will tell you the exact filename)
3. The schema file number should match your migration number

### Step 4: Update Migration File

The generated migration file includes TODO comments for customization:

```go
// internal/database/migrations/0002_add_user_profiles.go
package migrations

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/pocketbase/pocketbase/core"
    m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
    m.Register(func(app core.App) error {
        // Forward migration
        schemaPath := filepath.Join("internal", "database", "schema", "0002_pb_schema.json")
        schemaData, err := os.ReadFile(schemaPath)
        if err != nil {
            return fmt.Errorf("failed to read schema file: %w", err)
        }

        var collections []interface{}
        if err := json.Unmarshal(schemaData, &collections); err != nil {
            return fmt.Errorf("failed to parse schema JSON: %w", err)
        }

        collectionsData, err := json.Marshal(collections)
        if err != nil {
            return fmt.Errorf("failed to marshal collections: %w", err)
        }

        if err := app.ImportCollectionsByMarshaledJSON(collectionsData, false); err != nil {
            return fmt.Errorf("failed to import collections: %w", err)
        }

        // Optional: Add any data seeding specific to these collections

        return nil
    }, func(app core.App) error {
        // Rollback migration
        collectionsToDelete := []string{"settings", "user_settings"}

        for _, collectionName := range collectionsToDelete {
            collection, err := app.FindCollectionByNameOrId(collectionName)
            if err != nil {
                continue // Collection might not exist
            }

            if err := app.Delete(collection); err != nil {
                return fmt.Errorf("failed to delete collection %s: %w", collectionName, err)
            }
        }

        return nil
    })
}
```

### Step 5: Test Migration

```bash
# Test the migration
make dev

# Check logs to ensure migration runs successfully
make dev-logs

# Test rollback if needed (in development only)
```

## Migration Best Practices

### Naming Conventions

- **Migration files**: `XXXX_descriptive_name.go` (e.g., `0002_add_user_settings.go`)
- **Schema files**: `XXXX_pb_schema.json` (e.g., `0002_pb_schema.json`)
- **Collection names**: Use snake_case (e.g., `settings`, `user_settings`)

### Safety Guidelines

1. **Always test migrations** in development first
2. **Provide rollback functionality** for every migration
3. **Keep migrations focused** - one feature per migration
4. **Document breaking changes** in migration comments
5. **Backup production data** before running migrations

### Data Seeding in Migrations

- **Initial data only**: Use seeders for essential data (roles, permissions)
- **Make seeding idempotent**: Check if data exists before creating
- **Separate concerns**: Keep schema changes and data seeding separate when possible

### Environment Considerations

- **Development**: Migrations run automatically with `automigrate: true`
- **Production**: Run migrations manually with proper backup procedures
- **Testing**: Use separate test databases for migration testing

## Common Migration Patterns

### Adding a New Collection

```go
// Import new collection from JSON schema file
// Add any required initial data
// Provide rollback to delete the collection
```

### Modifying Existing Collection

```go
// Export the modified collection
// Import with updated schema
// Handle data migration if field types changed
// Provide rollback to previous schema
```

### Adding Relationships

```go
// Ensure related collections exist
// Add relation fields
// Update access rules if needed
// Test cascade delete behavior
```

## Troubleshooting

### Common Issues

1. **Schema conflicts**: Ensure collection IDs don't conflict
2. **Missing dependencies**: Check if related collections exist
3. **Permission errors**: Verify access rules are correctly set
4. **Data type conflicts**: Handle field type changes carefully

### Recovery Procedures

1. **Failed migration**: Use rollback function to revert changes
2. **Corrupted data**: Restore from backup and retry
3. **Schema inconsistency**: Export current schema and compare with expected

## Migration Checklist

Before creating a migration:

- [ ] Collections designed and tested in development
- [ ] Schema exported to numbered JSON file
- [ ] Migration file created with proper rollback
- [ ] Migration tested locally
- [ ] Documentation updated
- [ ] Backup procedures planned for production

## Common Migration Examples

### Adding User Profile Collections

```bash
# Generate the migration
make migrate-gen name=add_user_profiles

# Design collections in PocketBase Admin UI:
# - user_profiles (relation to users)
# - profile_settings
# - user_preferences

# Export to internal/database/schema/0003_pb_schema.json

# Update rollback function:
collectionsToDelete := []string{"user_profiles", "profile_settings", "user_preferences"}
```

### Implementing Audit Logging

```bash
# Generate the migration
make migrate-gen name=add_audit_logs

# Design collections:
# - audit_logs (user actions, timestamps, metadata)
# - audit_settings (retention policies)

# Export to internal/database/schema/0004_pb_schema.json

# Update rollback function:
collectionsToDelete := []string{"audit_logs", "audit_settings"}
```

### Setting Up Notification System

```bash
# Generate the migration
make migrate-gen name=add_notification_system

# Design collections:
# - notifications (user notifications)
# - notification_templates (email/push templates)
# - notification_preferences (user preferences)

# Export to internal/database/schema/0005_pb_schema.json

# Update rollback function:
collectionsToDelete := []string{"notifications", "notification_templates", "notification_preferences"}
```

### Content Management Collections

```bash
# Generate the migration
make migrate-gen name=add_cms_collections

# Design collections:
# - articles (blog posts, content)
# - categories (content categorization)
# - tags (content tagging)
# - media (file attachments)

# Export to internal/database/schema/0006_pb_schema.json

# Update rollback function:
collectionsToDelete := []string{"articles", "categories", "tags", "media"}
```

### CLI Usage Patterns

```bash
# Different naming styles (all converted to kebab-case)
make migrate-gen name=add_user_settings     # → 0003_add-user-settings.go
make migrate-gen name=AddUserSettings       # → 0003_adduserasettings.go  
make migrate-gen name=add-user-settings     # → 0003_add-user-settings.go

# Complex migration names
make migrate-gen name=refactor_user_authentication_system
# → 0004_refactor-user-authentication-system.go

# Simple migrations
make migrate-gen name=fix_permissions       # → 0005_fix-permissions.go
make migrate-gen name=update_schema         # → 0006_update-schema.go
```

## Support

For migration-related issues:

1. Check the troubleshooting section above
2. Review PocketBase migration documentation
3. Test in development environment first
4. Create an issue with detailed error logs
