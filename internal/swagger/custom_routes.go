package swagger

// CustomRouteRegistry holds all custom route definitions
type CustomRouteRegistry struct {
	routes []CustomRoute
}

// NewCustomRouteRegistry creates a new custom route registry
func NewCustomRouteRegistry() *CustomRouteRegistry {
	return &CustomRouteRegistry{
		routes: defineCustomRoutes(),
	}
}

// GetRoutes returns all custom routes
func (r *CustomRouteRegistry) GetRoutes() []CustomRoute {
	return r.routes
}

// RegisterWithGenerator registers all custom routes with the given generator
func (r *CustomRouteRegistry) RegisterWithGenerator(generator *Generator) {
	for _, route := range r.routes {
		generator.AddCustomRoute(route)
	}
}

// defineCustomRoutes defines all custom routes for the application
func defineCustomRoutes() []CustomRoute {
	return []CustomRoute{
		// System routes
		{
			Method:      "GET",
			Path:        "/api/health",
			Summary:     "Health Check",
			Description: "Check the health status of the API",
			Tags:        []string{"System"},
			Protected:   false,
		},

		// Custom API routes
		{
			Method:      "GET",
			Path:        "/api/v1/hello",
			Summary:     "Hello Endpoint",
			Description: "Returns a hello message from custom route",
			Tags:        []string{"Custom"},
			Protected:   false,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/protected",
			Summary:     "Protected Endpoint",
			Description: "Returns a message for authenticated users",
			Tags:        []string{"Custom"},
			Protected:   true,
		},
		{
			Method:      "GET",
			Path:        "/api/v1/permission-test",
			Summary:     "Permission Test Endpoint",
			Description: "Tests user creation permission",
			Tags:        []string{"Custom"},
			Protected:   true,
		},

		// User management routes
		{
			Method:      "POST",
			Path:        "/api/v1/users/export",
			Summary:     "Export Users",
			Description: "Export users data (requires export permission)",
			Tags:        []string{"Users"},
			Protected:   true,
		},

		// Job management routes
		{
			Method:      "GET",
			Path:        "/api/v1/jobs/{id}/status",
			Summary:     "Get Job Status",
			Description: "Get the status of a specific job",
			Tags:        []string{"Jobs"},
			Protected:   true,
		},
		{
			Method:      "POST",
			Path:        "/api/v1/jobs/{id}/download",
			Summary:     "Download Job File",
			Description: "Download the file associated with a job",
			Tags:        []string{"Jobs"},
			Protected:   true,
		},
	}
}

// RegisterCustomRoutes is a convenience function to register all custom routes with a generator
func RegisterCustomRoutes(generator *Generator) {
	registry := NewCustomRouteRegistry()
	registry.RegisterWithGenerator(generator)
}
