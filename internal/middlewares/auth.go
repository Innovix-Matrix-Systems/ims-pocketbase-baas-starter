package middlewares

import (
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// AuthMiddleware provides authentication middleware functionality
// It wraps PocketBase's built-in authentication system and provides
// a clean interface for applying authentication to routes
type AuthMiddleware struct{}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// RequireAuth returns a hook handler that requires authentication
// This wraps PocketBase's built-in apis.RequireAuth() middleware
//
// Parameters:
//   - optCollectionNames: Optional collection names to filter allowed auth collections
//     If empty, any auth collection is allowed
//
// Returns:
//   - *hook.Handler[*core.RequestEvent]: PocketBase hook handler for authentication
//
// Example:
//
//	handler := middleware.RequireAuth()                    // any auth collection
//	handler := middleware.RequireAuth("users", "_superusers") // only specified collections
func (m *AuthMiddleware) RequireAuth(optCollectionNames ...string) *hook.Handler[*core.RequestEvent] {
	return apis.RequireAuth(optCollectionNames...)
}

// RequireAuthFunc returns a middleware function that requires authentication
// This provides a function-based interface for route middleware by extracting
// the function from PocketBase's hook handler
//
// Parameters:
//   - optCollectionNames: Optional collection names to filter allowed auth collections
//
// Returns:
//   - func(*core.RequestEvent) error: Middleware function for direct route use
//
// Example:
//
//	authFunc := middleware.RequireAuthFunc("users")
//	router.GET("/protected", authFunc, handlerFunc)
func (m *AuthMiddleware) RequireAuthFunc(optCollectionNames ...string) func(*core.RequestEvent) error {
	handler := apis.RequireAuth(optCollectionNames...)
	return handler.Func
}

// BindToRouter binds authentication middleware to the router
// This method provides a hook for binding middleware to the router
// and can be extended for specific routing requirements
//
// Parameters:
//   - se: ServeEvent containing the router and other server context
//
// Returns:
//   - error: nil on success, error on failure
func (m *AuthMiddleware) BindToRouter(se *core.ServeEvent) error {
	// This method provides a hook for binding middleware to the router
	// Implementation will be expanded as needed for specific routing requirements
	return nil
}
