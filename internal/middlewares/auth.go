package middlewares

import (
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"

	"ims-pocketbase-baas-starter/pkg/cache"
	"ims-pocketbase-baas-starter/pkg/common"
)

// AuthMiddleware provides authentication middleware functionality
// It wraps PocketBase's built-in authentication system and provides
// a clean interface for applying authentication to routes
type AuthMiddleware struct {
	app core.App
}

// CollectionAuthInfo holds cached information about collection authentication requirements
type CollectionAuthInfo struct {
	ListRequiresAuth   bool
	ViewRequiresAuth   bool
	CreateRequiresAuth bool
	UpdateRequiresAuth bool
	DeleteRequiresAuth bool
	LastChecked        time.Time
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// WithApp sets the PocketBase app instance for the middleware
func (m *AuthMiddleware) WithApp(app core.App) *AuthMiddleware {
	m.app = app
	return m
}

// RequireAuth returns a hook handler that requires authentication
// This wraps PocketBase's built-in apis.RequireAuth() middleware
func (m *AuthMiddleware) RequireAuth(optCollectionNames ...string) *hook.Handler[*core.RequestEvent] {
	return apis.RequireAuth(optCollectionNames...)
}

// RequireAuthFunc returns a middleware function that requires authentication
// This provides a function-based interface for route middleware by extracting
// the function from PocketBase's hook handler
func (m *AuthMiddleware) RequireAuthFunc(optCollectionNames ...string) func(*core.RequestEvent) error {
	handler := apis.RequireAuth(optCollectionNames...)
	return handler.Func
}

// requiresAuthentication checks if a rule requires authentication
// Returns true if the rule contains @request.auth.id != "" or is not public
func (m *AuthMiddleware) requiresAuthentication(rule *string) bool {
	// If rule is nil, it's locked and requires auth
	if rule == nil {
		return true
	}

	// If rule is empty string, it's public and doesn't require auth
	if *rule == "" {
		return false
	}

	// Check if the rule contains the auth pattern with single or double quotes
	return strings.Contains(*rule, "@request.auth.id != ''") ||
		strings.Contains(*rule, "@request.auth.id != \"\"")
}

// getCachedCollectionAuthInfo retrieves or computes collection authentication info with caching
func (m *AuthMiddleware) getCachedCollectionAuthInfo(collectionName string) (*CollectionAuthInfo, error) {
	// Get cache instance
	cacheService := cache.GetInstance()
	cacheKey := fmt.Sprintf("collection_auth_info_%s", collectionName)

	// Try to get from cache first
	if cachedData, found := cacheService.Get(cacheKey); found {
		if authInfo, ok := cachedData.(*CollectionAuthInfo); ok {
			// Check if cache is still valid (within 1 minute)
			if time.Since(authInfo.LastChecked) < 1*time.Minute {
				return authInfo, nil
			}
		}
	}

	// Cache miss or expired - fetch collection data
	collection, err := m.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, err
	}

	// Create new auth info
	authInfo := &CollectionAuthInfo{
		ListRequiresAuth:   m.requiresAuthentication(collection.ListRule),
		ViewRequiresAuth:   m.requiresAuthentication(collection.ViewRule),
		CreateRequiresAuth: m.requiresAuthentication(collection.CreateRule),
		UpdateRequiresAuth: m.requiresAuthentication(collection.UpdateRule),
		DeleteRequiresAuth: m.requiresAuthentication(collection.DeleteRule),
		LastChecked:        time.Now(),
	}

	// Cache for 1 minute
	cacheService.SetWithExpiration(cacheKey, authInfo, 1*time.Minute)

	return authInfo, nil
}

// getOperationFromPath determines the operation type based on path and HTTP method
func (m *AuthMiddleware) getOperationFromPath(path, method string) (collectionName, operation string, ok bool) {
	if !strings.HasPrefix(path, "/api/collections/") {
		return "", "", false
	}

	// Extract parts from path: /api/collections/{collection}/records[/{id}][/{action}]
	parts := strings.Split(strings.TrimPrefix(path, "/api/collections/"), "/")
	if len(parts) < 1 {
		return "", "", false
	}

	collectionName = parts[0]

	// Skip auth checks for auth-related endpoints as PocketBase handles them
	if len(parts) >= 3 && parts[1] != "records" {
		return "", "", false
	}

	// Skip superuser collection entirely as it's handled by PocketBase
	if collectionName == "_superusers" {
		return "", "", false
	}

	// Determine the operation type based on path and HTTP method
	if len(parts) >= 2 && parts[1] == "records" {
		if len(parts) == 2 || (len(parts) == 3 && parts[2] == "") {
			// /api/collections/{collection}/records
			switch method {
			case "GET":
				operation = "list"
			case "POST":
				operation = "create"
			}
		} else if len(parts) >= 3 {
			// /api/collections/{collection}/records/{id}[/{action}]
			recordId := parts[2]
			if recordId != "" {
				if len(parts) == 3 || (len(parts) == 4 && parts[3] == "") {
					// /api/collections/{collection}/records/{id}
					switch method {
					case "GET":
						operation = "view"
					case "POST", "PATCH":
						operation = "update"
					case "DELETE":
						operation = "delete"
					}
				} else if len(parts) >= 4 {
					// /api/collections/{collection}/records/{id}/{action}
					// Special actions are handled by PocketBase
					return "", "", false
				}
			}
		}
	}

	return collectionName, operation, true
}

// RequireAuthWithExclusionsFunc returns a middleware function that requires authentication
// but respects excluded paths and checks collection rules based on the operation type
func (m *AuthMiddleware) RequireAuthWithExclusionsFunc(e *core.RequestEvent) error {
	// If we don't have access to the app, skip auth checks
	if m.app == nil {
		return e.Next()
	}

	path := e.Request.URL.Path
	method := e.Request.Method

	// Check if path should be excluded
	for _, excludedPath := range common.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return e.Next() // Skip auth for excluded paths
		}
	}

	// Determine operation type
	collectionName, operation, ok := m.getOperationFromPath(path, method)
	if !ok {
		return e.Next()
	}

	// Get cached collection authentication info
	authInfo, err := m.getCachedCollectionAuthInfo(collectionName)
	if err != nil {
		// Collection not found or error, let PocketBase handle the error
		return e.Next()
	}

	// Check if the specific operation requires authentication based on the cached info
	requiresAuth := false

	switch operation {
	case "list":
		requiresAuth = authInfo.ListRequiresAuth
	case "view":
		requiresAuth = authInfo.ViewRequiresAuth
	case "create":
		requiresAuth = authInfo.CreateRequiresAuth
	case "update":
		requiresAuth = authInfo.UpdateRequiresAuth
	case "delete":
		requiresAuth = authInfo.DeleteRequiresAuth
	default:
		// For unknown operations, check if any rule requires auth
		requiresAuth = authInfo.ListRequiresAuth ||
			authInfo.ViewRequiresAuth ||
			authInfo.CreateRequiresAuth ||
			authInfo.UpdateRequiresAuth ||
			authInfo.DeleteRequiresAuth
	}

	// If the operation requires authentication, enforce it
	if requiresAuth {
		authFunc := m.RequireAuthFunc()
		if err := authFunc(e); err != nil {
			return err
		}
	}

	return e.Next()
}
