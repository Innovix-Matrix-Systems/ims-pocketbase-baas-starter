package middlewares

import (
	"log"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// PermissionMiddleware provides permission-based middleware functionality
// It extends the authentication system to check for specific permissions
type PermissionMiddleware struct{}

// NewPermissionMiddleware creates a new instance of PermissionMiddleware
func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{}
}

// getUserPermissions extracts and processes all permissions for a user
// This includes direct permissions and those inherited from roles
//
// Parameters:
//   - app: The PocketBase app instance
//   - user: The authenticated user record
//
// Returns:
//   - []string: Array of permission slugs the user has access to
func (m *PermissionMiddleware) getUserPermissions(app core.App, user *core.Record) []string {
	// Extract and log permissions and roles arrays
	userPermissions, _ := user.Get("permissions").([]string)
	roles, _ := user.Get("roles").([]string)

	// fetch roles from collection
	roleRecords, err := app.FindRecordsByIds("roles", roles)
	if err != nil {
		log.Printf("Error fetching roles: %v", err)
	}
	for _, role := range roleRecords {
		perms := role.GetStringSlice("permissions")
		userPermissions = append(userPermissions, perms...)
	}

	uniquePerms := make(map[string]struct{})
	for _, p := range userPermissions {
		uniquePerms[p] = struct{}{}
	}
	userPermissions = userPermissions[:0] // reset slice
	for p := range uniquePerms {
		userPermissions = append(userPermissions, p)
	}

	//fetch permissions form collection
	permissionsRecords, err := app.FindRecordsByIds("permissions", userPermissions)
	if err != nil {
		log.Printf("Error fetching permissions: %v", err)
	}

	userPermissions = userPermissions[:0] // reset slice
	//now from permission records we only need to get the array of permission slugs
	for _, permission := range permissionsRecords {
		permSlug := permission.GetString("slug")
		userPermissions = append(userPermissions, permSlug)
	}

	return userPermissions
}

// HasPermission checks if a user has any of the specified permissions
// This checks both direct permissions assigned to the user and permissions from roles
//
// Parameters:
//   - userPermissions: The authenticated user's permissions array slugs
//   - permissions: String array of permission slugs to check
//
// Returns:
//   - bool: True if the user has any of the specified permissions, false otherwise
func (m *PermissionMiddleware) HasPermission(userPermissions []string, permissions []string) bool {
	if len(permissions) == 0 || len(userPermissions) == 0 {
		return false
	}

	// Create a map of user permissions for O(1) lookups
	userPermMap := make(map[string]struct{}, len(userPermissions))
	for _, perm := range userPermissions {
		userPermMap[perm] = struct{}{}
	}

	// Check if any required permission exists in the user's permission map
	for _, requiredPerm := range permissions {
		if _, exists := userPermMap[requiredPerm]; exists {
			return true
		}
	}

	return false
}

// RequirePermission returns a middleware function that requires specific permissions
// This middleware checks if the authenticated user has any of the specified permissions
//
// Parameters:
//   - permissions: String array of permission slugs to check
//
// Returns:
//   - func(*core.RequestEvent) error: Middleware function for direct route use
//
// Example:
//
//	permFunc := middleware.RequirePermission("resource.view")
//	router.GET("/protected", authFunc, permFunc, handlerFunc)
func (m *PermissionMiddleware) RequirePermission(permissions ...string) func(*core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// If no permissions are required, allow the request
		if len(permissions) == 0 {
			return nil
		}

		// Extract the authenticated user from the request event
		// The Auth field contains the authenticated user record
		user := e.Auth

		// Get user permissions from the user record and roles
		userPermissions := m.getUserPermissions(e.App, user)

		if user == nil {
			// User not found - this should not happen if used after auth middleware
			// but we handle it gracefully
			return apis.NewForbiddenError("Authentication required", nil)
		}

		// Check if the user is a superuser of pocketbase they will bypass this check
		if user.IsSuperuser() {
			return nil
		}

		// Check if the user has any of the required permissions
		if m.HasPermission(userPermissions, permissions) {
			// User has permission, allow the request to proceed
			return nil
		}

		// User doesn't have the required permissions, return 403 Forbidden
		return apis.NewForbiddenError("You don't have permission to access this resource", nil)
	}
}
