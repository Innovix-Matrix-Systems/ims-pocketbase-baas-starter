package middlewares

import (
	"ims-pocketbase-baas-starter/pkg/cache"
	"ims-pocketbase-baas-starter/pkg/logger"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	RolesCollection       = "roles"
	PermissionsCollection = "permissions"
	PermissionCacheTime   = 1 * time.Minute
)

// PermissionMiddleware provides permission-based middleware functionality
// It extends the authentication system to check for specific permissions
type PermissionMiddleware struct {
	cache    *cache.CacheService
	cacheKey cache.CacheKey
}

// NewPermissionMiddleware creates a new instance of PermissionMiddleware
func NewPermissionMiddleware() *PermissionMiddleware {
	return &PermissionMiddleware{
		cache:    cache.GetInstance(),
		cacheKey: cache.CacheKey{},
	}
}

// getUserPermissions extracts and processes all permissions for a user with caching
// This includes direct permissions and those inherited from roles
// Uses centralized caching to avoid N+1 query problems
//
// Parameters:
//   - app: The PocketBase app instance
//   - user: The authenticated user record
//
// Returns:
//   - []string: Array of permission slugs the user has access to
func (m *PermissionMiddleware) getUserPermissions(app core.App, user *core.Record) []string {
	// Check cache first
	cacheKey := m.cacheKey.UserPermissions(user.Id)
	if cachedPerms, found := m.cache.GetStringSlice(cacheKey); found {
		return cachedPerms
	}

	// Cache miss - fetch and compute permissions
	permissions := m.fetchUserPermissions(app, user)

	// Cache for 5 minutes
	m.cache.SetWithExpiration(cacheKey, permissions, PermissionCacheTime)

	return permissions
}

// fetchUserPermissions fetches user permissions from database (optimized to avoid N+1 queries)
func (m *PermissionMiddleware) fetchUserPermissions(app core.App, user *core.Record) []string {
	// Extract user's direct permissions and roles
	userPermissions := user.GetStringSlice("permissions")
	roles := user.GetStringSlice("roles")

	// Batch fetch all roles to get their permissions
	if len(roles) > 0 {
		roleRecords, err := app.FindRecordsByIds(RolesCollection, roles)
		if err != nil {
			logger.FromAppOrDefault(app).Error("Error fetching roles", "error", err)
		} else {
			// Collect permissions from all roles
			for _, role := range roleRecords {
				rolePerms := role.GetStringSlice("permissions")
				userPermissions = append(userPermissions, rolePerms...)
			}
		}
	}

	// Remove duplicates
	uniquePerms := make(map[string]struct{})
	for _, p := range userPermissions {
		if p != "" {
			uniquePerms[p] = struct{}{}
		}
	}

	// Convert back to slice
	permissionIDs := make([]string, 0, len(uniquePerms))
	for p := range uniquePerms {
		permissionIDs = append(permissionIDs, p)
	}

	// Batch fetch all permission records to get slugs
	if len(permissionIDs) == 0 {
		return []string{}
	}

	permissionsRecords, err := app.FindRecordsByIds(PermissionsCollection, permissionIDs)
	if err != nil {
		logger.FromAppOrDefault(app).Error("Error fetching permissions", "error", err)
		return []string{}
	}

	// Extract permission slugs
	permissionSlugs := make([]string, 0, len(permissionsRecords))
	for _, permission := range permissionsRecords {
		slug := permission.GetString("slug")
		if slug != "" {
			permissionSlugs = append(permissionSlugs, slug)
		}
	}

	return permissionSlugs
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

// InvalidateUserPermissions invalidates cached permissions for a specific user
func (m *PermissionMiddleware) InvalidateUserPermissions(userID string) {
	cacheKey := m.cacheKey.UserPermissions(userID)
	m.cache.Delete(cacheKey)
}

// InvalidateAllUserPermissions invalidates all cached user permissions
func (m *PermissionMiddleware) InvalidateAllUserPermissions() int {
	return m.cache.InvalidateUserPermissions()
}

// GetCacheStats returns cache statistics for debugging
func (m *PermissionMiddleware) GetCacheStats() map[string]any {
	return m.cache.GetStats()
}
