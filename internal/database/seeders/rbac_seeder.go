package seeders

import (
	"fmt"
	"ims-pocketbase-baas-starter/pkg/permission"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// SeedRBAC seeds permissions, roles, and creates a super admin user
func SeedRBAC(app core.App) error {
	// 1. Seed Permissions
	if err := seedPermissions(app); err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// 2. Seed Roles
	if err := seedRoles(app); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// 3. Create Super Admin User
	if err := createSuperAdminUser(app); err != nil {
		return fmt.Errorf("failed to create super admin user: %w", err)
	}

	return nil
}

// seedPermissions creates all required permissions
func seedPermissions(app core.App) error {
	permissionDefs := permission.GetAllPermissions()
	requiredPerms := make([]Permission, len(permissionDefs))

	for i, def := range permissionDefs {
		requiredPerms[i] = Permission{
			Slug:        def.Slug,
			Name:        def.Name,
			Description: def.Description,
		}
	}

	permCollection, err := app.FindCollectionByNameOrId("permissions")
	if err != nil {
		return fmt.Errorf("permissions collection not found: %w", err)
	}

	for _, perm := range requiredPerms {
		exists, err := app.FindFirstRecordByFilter("permissions", "slug = {:slug}", dbx.Params{"slug": perm.Slug})
		if err == nil && exists != nil {
			continue // already seeded
		}

		rec := core.NewRecord(permCollection)
		rec.Set("name", perm.Name)
		rec.Set("slug", perm.Slug)
		rec.Set("description", perm.Description)
		if err := app.Save(rec); err != nil {
			return fmt.Errorf("seed permission %s: %w", perm.Slug, err)
		}
		fmt.Printf("✅ Created permission: %s\n", perm.Slug)
	}

	return nil
}

// seedRoles creates the required roles with their permissions
func seedRoles(app core.App) error {
	// Get all permissions first
	allPermissions, err := app.FindAllRecords("permissions")
	if err != nil {
		return fmt.Errorf("failed to get permissions: %w", err)
	}

	// Create permission maps for easy lookup
	permissionMap := make(map[string]string)
	for _, perm := range allPermissions {
		permissionMap[perm.GetString("slug")] = perm.Id
	}

	// Define roles with their permissions
	roles := []Role{
		{
			Name:        "Super Admin",
			Description: "Full system access with all permissions",
			Permissions: []string{
				permission.CacheClear, permission.UserCreate, permission.UserView, permission.UserViewAll, permission.UserUpdate, permission.UserDelete,
				permission.UserRoleAssign, permission.UserPermissionAssign, permission.UserExport,
				permission.RoleCreate, permission.RoleView, permission.RoleViewAll, permission.RoleUpdate, permission.RoleDelete,
			},
		},
		{
			Name:        "Admin",
			Description: "User management and role viewing permissions",
			Permissions: []string{
				permission.UserCreate, permission.UserView, permission.UserViewAll, permission.UserUpdate, permission.UserDelete,
				permission.RoleView, permission.RoleViewAll,
			},
		},
		{
			Name:        "User",
			Description: "Basic user permissions",
			Permissions: []string{
				permission.UserView,
			},
		},
	}

	roleCollection, err := app.FindCollectionByNameOrId("roles")
	if err != nil {
		return fmt.Errorf("roles collection not found: %w", err)
	}

	for _, role := range roles {
		exists, err := app.FindFirstRecordByFilter("roles", "name = {:name}", dbx.Params{"name": role.Name})
		if err == nil && exists != nil {
			continue // already seeded
		}

		// Get permission IDs for this role
		var permissionIds []string
		for _, permSlug := range role.Permissions {
			if permId, ok := permissionMap[permSlug]; ok {
				permissionIds = append(permissionIds, permId)
			}
		}

		rec := core.NewRecord(roleCollection)
		rec.Set("name", role.Name)
		rec.Set("description", role.Description)
		rec.Set("permissions", permissionIds)

		if err := app.Save(rec); err != nil {
			return fmt.Errorf("seed role %s: %w", role.Name, err)
		}
		fmt.Printf("✅ Created role: %s with %d permissions\n", role.Name, len(permissionIds))
	}

	return nil
}

// createSuperAdminUser creates a super admin user with the Super Admin role
func createSuperAdminUser(app core.App) error {
	usersColl, err := app.FindCollectionByNameOrId("users")
	if err != nil {
		return fmt.Errorf("users collection not found: %w", err)
	}

	// Check if super admin already exists
	admin, _ := app.FindFirstRecordByFilter("users", "email = {:email}", dbx.Params{"email": "admin@example.com"})
	if admin != nil {
		fmt.Println("✅ Super admin user already exists, skipping creation")
		return nil
	}

	// Get Super Admin role
	superAdminRole, err := app.FindFirstRecordByFilter("roles", "name = {:name}", dbx.Params{"name": "Super Admin"})
	if err != nil {
		return fmt.Errorf("super admin role not found: %w", err)
	}

	// Create super admin user
	admin = core.NewRecord(usersColl)
	admin.Set("email", "superadminuser@example.com")
	admin.Set("name", "Super Administrator")
	admin.Set("verified", true)
	admin.Set("is_active", true)
	admin.Set("roles", []string{superAdminRole.Id})
	admin.SetPassword("superadmin123")

	if err := app.Save(admin); err != nil {
		return fmt.Errorf("create super admin user: %w", err)
	}

	fmt.Println("✅ Super admin user created successfully")

	return nil
}

// Permission represents a permission structure
type Permission struct {
	Slug        string
	Name        string
	Description string
}

// Role represents a role structure with permissions
type Role struct {
	Name        string
	Description string
	Permissions []string
}
