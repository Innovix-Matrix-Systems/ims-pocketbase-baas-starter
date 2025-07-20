# Database Migrations Guide

This document outlines the database migration strategy and best practices for the IMS PocketBase BaaS Starter project.

## Overview

Our migration system uses PocketBase's built-in migration framework with an incremental approach for adding new collections and schema changes. This ensures safe, reversible database changes in both development and production environments.

## Migration Strategy

### Initial Migration (0001_init.go)

The initial migration imports the complete database schema from `pb_schema.json` and sets up:

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
│   ├── 0002_add_user_profiles.go # Example: User profile collections
│   ├── 0003_add_audit_logs.go    # Example: Audit logging
│   └── utils.go                  # Migration helper functions
├── schema/
│   ├── pb_schema.json            # Complete initial schema
│   ├── 0002_user_profiles.json   # New collections for migration 0002
│   ├── 0003_audit_logs.json      # New collections for migration 0003
│   └── README.md                 # Schema documentation
└── seeders/
    ├── rbac_seeder.go            # Role-based access control seeding
    └── superuser_seeder.go       # Superuser creation
```

## Creating New Migrations

### Step 1: Design Collections

1. Use PocketBase Admin UI to design your new collections
2. Test the collections thoroughly in development
3. Document the purpose and relationships

### Step 2: Export Schema

1. Export only the new collections from PocketBase Admin UI
2. Save as `internal/database/schema/XXXX_description.json`
3. Use sequential numbering (0002, 0003, etc.)

### Step 3: Create Migration File

Create a new migration file following this template:

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
        schemaPath := filepath.Join("internal", "database", "schema", "0002_user_profiles.json")
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
        collectionsToDelete := []string{"user_profiles", "user_preferences"}

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

### Step 4: Test Migration

```bash
# Test the migration
make dev

# Check logs to ensure migration runs successfully
make dev-logs

# Test rollback if needed (in development only)
```

## Migration Best Practices

### Naming Conventions

- **Migration files**: `XXXX_descriptive_name.go` (e.g., `0002_add_user_profiles.go`)
- **Schema files**: `XXXX_descriptive_name.json` (e.g., `0002_user_profiles.json`)
- **Collection names**: Use snake_case (e.g., `user_profiles`, `audit_logs`)

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

## Examples

See the `examples/` directory for complete migration examples:

- Adding user profile collections
- Implementing audit logging
- Setting up notification system
- Creating content management collections

## Support

For migration-related issues:

1. Check the troubleshooting section above
2. Review PocketBase migration documentation
3. Test in development environment first
4. Create an issue with detailed error logs
