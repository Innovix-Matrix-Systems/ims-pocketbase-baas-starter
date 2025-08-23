package apidoc

// GetCustomRoutes returns all predefined custom routes
// These routes are automatically registered when creating a new generator
// through the DefaultConfig function
func GetCustomRoutes() []CustomRoute {
	return defineCustomRoutes()
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
		{
			Method:      "GET",
			Path:        "/api/v1/cache-status",
			Summary:     "Cache Status",
			Description: "Get the current status of the global cache store including collection hash and cache statistics",
			Tags:        []string{"System"},
			Protected:   false,
		},
		{
			Method:      "DELETE",
			Path:        "/api/v1/cache",
			Summary:     "Clear Cache",
			Description: "Clear all cache entries in the system",
			Tags:        []string{"System"},
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
			Parameters: []Parameter{
				{
					Name:        "id",
					In:          "path",
					Required:    true,
					Schema:      map[string]any{"type": "string"},
					Description: "The unique identifier of the job",
				},
			},
		},
		{
			Method:      "POST",
			Path:        "/api/v1/jobs/{id}/download",
			Summary:     "Download Job File",
			Description: "Download the file associated with a job",
			Tags:        []string{"Jobs"},
			Protected:   true,
			Parameters: []Parameter{
				{
					Name:        "id",
					In:          "path",
					Required:    true,
					Schema:      map[string]any{"type": "string"},
					Description: "The unique identifier of the job",
				},
			},
		},
	}
}
