package middlewares

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"

	"ims-pocketbase-baas-starter/pkg/metrics"
)

// Middleware represents an application middleware with its configuration
type Middleware struct {
	ID          string                         // Unique identifier for the middleware
	Handler     func(*core.RequestEvent) error // Handler function to execute
	Enabled     bool                           // Whether the middleware should be registered
	Description string                         // Human-readable description of what the middleware does
	Order       int                            // Order of execution (lower numbers execute first)
}

// RegisterMiddlewares registers all application middlewares with the PocketBase router
func RegisterMiddlewares(e *core.ServeEvent) {
	// Define all middlewares
	middlewares := []Middleware{
		{
			ID:          "metricsCollection",
			Handler:     getMetricsMiddlewareHandler(),
			Enabled:     true,
			Description: "Collect HTTP request metrics",
			Order:       1,
		},
		{
			ID:          "jwtAuth",
			Handler:     getAuthMiddlewareHandler(e),
			Enabled:     true,
			Description: "JWT authentication with exclusions",
			Order:       2,
		},
		// Add more middlewares here as needed:
		// {
		//     ID:          "exampleMiddleware",
		//     Handler:     getExampleMiddlewareHandler(),
		//     Enabled:     true,
		//     Description: "Example middleware description",
		//     Order:       3,
		// },
	}

	// Register enabled middlewares
	for _, middleware := range middlewares {
		if !middleware.Enabled {
			continue
		}

		e.Router.Bind(&hook.Handler[*core.RequestEvent]{
			Id:   middleware.ID,
			Func: middleware.Handler,
		})
	}
}

// getMetricsMiddlewareHandler creates the metrics middleware handler
func getMetricsMiddlewareHandler() func(*core.RequestEvent) error {
	metricsProvider := metrics.GetInstance()
	metricsMiddleware := NewMetricsMiddleware(metricsProvider)
	return metricsMiddleware.RequireMetricsFunc()
}

// getAuthMiddlewareHandler creates the auth middleware handler
func getAuthMiddlewareHandler(e *core.ServeEvent) func(*core.RequestEvent) error {
	authMiddleware := NewAuthMiddleware().WithApp(e.App)
	return authMiddleware.RequireAuthWithExclusionsFunc
}
