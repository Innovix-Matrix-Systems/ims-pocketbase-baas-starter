package permission

import (
	"testing"
)

func TestPermissionConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"UserCreate constant", UserCreate, "user.create"},
		{"UserView constant", UserView, "user.view"},
		{"UserViewAll constant", UserViewAll, "user.view.all"},
		{"UserUpdate constant", UserUpdate, "user.update"},
		{"UserDelete constant", UserDelete, "user.delete"},
		{"UserRoleAssign constant", UserRoleAssign, "user.role.assign"},
		{"UserPermissionAssign constant", UserPermissionAssign, "user.permission.assign"},
		{"UserExport constant", UserExport, "user.export"},
		{"RoleCreate constant", RoleCreate, "role.create"},
		{"RoleView constant", RoleView, "role.view"},
		{"RoleViewAll constant", RoleViewAll, "role.view.all"},
		{"RoleUpdate constant", RoleUpdate, "role.update"},
		{"RoleDelete constant", RoleDelete, "role.delete"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Expected %s to be %q, got %q", tt.name, tt.expected, tt.constant)
			}
		})
	}
}

func TestGetAllPermissions(t *testing.T) {
	permissions := GetAllPermissions()

	// Test that we get the expected number of permissions
	expectedCount := 14 // Updated to include CacheClear permission
	if len(permissions) != expectedCount {
		t.Errorf("Expected %d permissions, got %d", expectedCount, len(permissions))
	}

	// Test that all permissions have required fields
	for i, perm := range permissions {
		if perm.Slug == "" {
			t.Errorf("Permission at index %d has empty Slug", i)
		}
		if perm.Name == "" {
			t.Errorf("Permission at index %d has empty Name", i)
		}
		if perm.Description == "" {
			t.Errorf("Permission at index %d has empty Description", i)
		}
	}

	// Test specific permissions exist
	expectedPermissions := map[string]struct {
		name        string
		description string
	}{
		CacheClear:           {"Clear Cache", "Can clear the system cache"},
		UserCreate:           {"Create User", "Can create new users"},
		UserView:             {"View User", "Can view user details"},
		UserViewAll:          {"View All Users", "Can view all users"},
		UserUpdate:           {"Update User", "Can update user information"},
		UserDelete:           {"Delete User", "Can delete users"},
		UserRoleAssign:       {"Assign Role To User", "Can assign roles to users"},
		UserPermissionAssign: {"Assign Permission To User", "Can assign permissions to users"},
		UserExport:           {"Export Users", "Can export user data as CSV"},
		RoleCreate:           {"Create Role", "Can create new roles"},
		RoleView:             {"View Role", "Can view role details"},
		RoleViewAll:          {"View All Roles", "Can view all roles"},
		RoleUpdate:           {"Update Role", "Can update role information"},
		RoleDelete:           {"Delete Role", "Can delete roles"},
	}

	// Create a map of returned permissions for easy lookup
	returnedPerms := make(map[string]PermissionDefinition)
	for _, perm := range permissions {
		returnedPerms[perm.Slug] = perm
	}

	// Verify each expected permission exists with correct details
	for slug, expected := range expectedPermissions {
		perm, exists := returnedPerms[slug]
		if !exists {
			t.Errorf("Expected permission %q not found", slug)
			continue
		}

		if perm.Name != expected.name {
			t.Errorf("Permission %q: expected name %q, got %q", slug, expected.name, perm.Name)
		}

		if perm.Description != expected.description {
			t.Errorf("Permission %q: expected description %q, got %q", slug, expected.description, perm.Description)
		}
	}
}

func TestPermissionDefinitionStruct(t *testing.T) {
	// Test that PermissionDefinition struct can be created and accessed
	perm := PermissionDefinition{
		Slug:        "test.permission",
		Name:        "Test Permission",
		Description: "A test permission",
	}

	if perm.Slug != "test.permission" {
		t.Errorf("Expected Slug to be 'test.permission', got %q", perm.Slug)
	}

	if perm.Name != "Test Permission" {
		t.Errorf("Expected Name to be 'Test Permission', got %q", perm.Name)
	}

	if perm.Description != "A test permission" {
		t.Errorf("Expected Description to be 'A test permission', got %q", perm.Description)
	}
}

func TestPermissionUniqueness(t *testing.T) {
	permissions := GetAllPermissions()
	slugs := make(map[string]bool)

	for _, perm := range permissions {
		if slugs[perm.Slug] {
			t.Errorf("Duplicate permission slug found: %q", perm.Slug)
		}
		slugs[perm.Slug] = true
	}
}
