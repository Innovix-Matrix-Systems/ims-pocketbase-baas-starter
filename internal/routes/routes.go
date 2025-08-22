package routes

import (
	"ims-pocketbase-baas-starter/internal/handlers/route"
	"ims-pocketbase-baas-starter/internal/middlewares"
	"ims-pocketbase-baas-starter/pkg/permission"

	"github.com/pocketbase/pocketbase/core"
)

// Route represents a custom application route with its configuration
type Route struct {
	Method      string                           // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path        string                           // Route path
	Handler     func(*core.RequestEvent) error   // Handler function to execute when route is called
	Middlewares []func(*core.RequestEvent) error // Middlewares to apply to this route
	Enabled     bool                             // Whether the route should be registered
	Description string                           // Human-readable description of what the route does
}

// RegisterCustom registers all custom routes with the PocketBase application
// This function follows the same pattern as RegisterCrons and RegisterJobs
func RegisterCustom(e *core.ServeEvent) {
	authMiddleware := middlewares.NewAuthMiddleware()
	permissionMiddleware := middlewares.NewPermissionMiddleware()

	g := e.Router.Group("/api/v1")

	// Define all custom routes
	routes := []Route{
		{
			Method:      "GET",
			Path:        "/cache-status",
			Handler:     route.HandleCacheStatus,
			Middlewares: []func(*core.RequestEvent) error{},
			Enabled:     true,
			Description: "Cache status route (public for monitoring)",
		},
		{
			Method:  "DELETE",
			Path:    "/cache",
			Handler: route.HandleCacheClear,
			Middlewares: []func(*core.RequestEvent) error{
				authMiddleware.RequireAuthFunc(),
				permissionMiddleware.RequirePermission(permission.CacheClear),
			},
			Enabled:     true,
			Description: "Clear all system cache (requires auth and cache.clear permission)",
		},
		{
			Method:  "POST",
			Path:    "/users/export",
			Handler: route.HandleUserExport,
			Middlewares: []func(*core.RequestEvent) error{
				authMiddleware.RequireAuthFunc(),
				permissionMiddleware.RequirePermission(permission.UserExport),
			},
			Enabled:     true,
			Description: "User export route",
		},
		{
			Method:  "GET",
			Path:    "/jobs/{id}/status",
			Handler: route.HandleGetJobStatus,
			Middlewares: []func(*core.RequestEvent) error{
				authMiddleware.RequireAuthFunc(),
			},
			Enabled:     true,
			Description: "Get job status route",
		},
		{
			Method:  "POST",
			Path:    "/jobs/{id}/download",
			Handler: route.HandleDownloadJobFile,
			Middlewares: []func(*core.RequestEvent) error{
				authMiddleware.RequireAuthFunc(),
			},
			Enabled:     true,
			Description: "Download job file route",
		},
		// Add more routes here as needed:
	}

	// Register enabled routes
	for _, route := range routes {
		if !route.Enabled {
			continue
		}

		// Create the final handler with middlewares applied
		finalHandler := route.Handler
		for i := len(route.Middlewares) - 1; i >= 0; i-- {
			middleware := route.Middlewares[i]
			nextHandler := finalHandler
			finalHandler = func(e *core.RequestEvent) error {
				if err := middleware(e); err != nil {
					return err
				}
				return nextHandler(e)
			}
		}

		// Register the route with the appropriate HTTP method
		switch route.Method {
		case "GET":
			g.GET(route.Path, finalHandler)
		case "POST":
			g.POST(route.Path, finalHandler)
		case "PUT":
			g.PUT(route.Path, finalHandler)
		case "DELETE":
			g.DELETE(route.Path, finalHandler)
		case "PATCH":
			g.PATCH(route.Path, finalHandler)
		default:
			// For unsupported methods, skip registration
			continue
		}
	}
}
