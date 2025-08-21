package middlewares

import (
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"

	"ims-pocketbase-baas-starter/pkg/common"
)

// AuthMiddleware provides authentication middleware functionality
// It wraps PocketBase's built-in authentication system and provides
// a clean interface for applying authentication to routes
type AuthMiddleware struct {
	app core.App
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
// Returns true if the rule contains @request.auth.id != ” or is not public
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

	// Find the collection
	collection, err := m.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		// Collection not found, let PocketBase handle the error
		return e.Next()
	}

	// Check if the specific operation requires authentication based on the rule
	requiresAuth := false

	switch operation {
	case "list":
		requiresAuth = m.requiresAuthentication(collection.ListRule)
	case "view":
		requiresAuth = m.requiresAuthentication(collection.ViewRule)
	case "create":
		requiresAuth = m.requiresAuthentication(collection.CreateRule)
	case "update":
		requiresAuth = m.requiresAuthentication(collection.UpdateRule)
	case "delete":
		requiresAuth = m.requiresAuthentication(collection.DeleteRule)
	default:
		// For unknown operations, check if any rule requires auth
		requiresAuth = m.requiresAuthentication(collection.ListRule) ||
			m.requiresAuthentication(collection.ViewRule) ||
			m.requiresAuthentication(collection.CreateRule) ||
			m.requiresAuthentication(collection.UpdateRule) ||
			m.requiresAuthentication(collection.DeleteRule)
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
