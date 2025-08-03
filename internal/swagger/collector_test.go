package swagger

import (
	"testing"
)

func TestNewSimpleRouteCollector(t *testing.T) {
	registry := NewRouteRegistry()

	collector := NewSimpleRouteCollector(nil, registry)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if collector.registry != registry {
		t.Error("Expected collector registry to match provided registry")
	}
}

func TestShouldIncludeRoute(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		route    DiscoveredRoute
		expected bool
	}{
		{
			name:     "API route should be included",
			route:    DiscoveredRoute{Method: "GET", Path: "/api/collections/users/records"},
			expected: true,
		},
		{
			name:     "Documentation route should be excluded",
			route:    DiscoveredRoute{Method: "GET", Path: "/api-docs/openapi.json"},
			expected: false,
		},
		{
			name:     "Static file route should be excluded",
			route:    DiscoveredRoute{Method: "GET", Path: "/{path...}"},
			expected: false,
		},
		{
			name:     "Internal route should be excluded",
			route:    DiscoveredRoute{Method: "GET", Path: "/_internal/health"},
			expected: false,
		},
		{
			name:     "Custom API route should be included",
			route:    DiscoveredRoute{Method: "GET", Path: "/api/v1/hello"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.shouldIncludeRoute(tt.route)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for route %s", tt.expected, result, tt.route.Path)
			}
		})
	}
}

func TestIsProtectedRoute(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Users collection should be protected",
			path:     "/api/collections/users/records",
			expected: true,
		},
		{
			name:     "Protected custom route should be protected",
			path:     "/api/v1/protected",
			expected: true,
		},
		{
			name:     "Public hello route should not be protected",
			path:     "/api/v1/hello",
			expected: false,
		},
		{
			name:     "Auth route should be protected",
			path:     "/api/collections/users/auth-with-password",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.isProtectedRoute(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestGenerateTags(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "Collections path should get PocketBase Collections tag",
			path:     "/api/collections/users/records",
			expected: []string{"PocketBase Collections"},
		},
		{
			name:     "Admin path should get Admin Authentication tag",
			path:     "/api/admins/auth-with-password",
			expected: []string{"Admin Authentication"},
		},
		{
			name:     "Custom API path should get Custom API tag",
			path:     "/api/v1/hello",
			expected: []string{"Custom API"},
		},
		{
			name:     "Files path should get Files tag",
			path:     "/api/files/collection/id/filename",
			expected: []string{"Files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.generateTags(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d for path %s", len(tt.expected), len(result), tt.path)
				return
			}

			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("Expected tag %s, got %s for path %s", tt.expected[i], tag, tt.path)
				}
			}
		})
	}
}

func TestExtractParameters(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "Path with no parameters",
			path:     "/api/v1/hello",
			expected: 0,
		},
		{
			name:     "Path with one parameter",
			path:     "/api/collections/{collection}/records",
			expected: 1,
		},
		{
			name:     "Path with multiple parameters",
			path:     "/api/collections/{collection}/records/{id}",
			expected: 2,
		},
		{
			name:     "Path with file parameters",
			path:     "/api/files/{collection}/{id}/{filename}",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.extractParameters(tt.path)
			if len(result) != tt.expected {
				t.Errorf("Expected %d parameters, got %d for path %s", tt.expected, len(result), tt.path)
			}

			// Verify all parameters are path parameters and required
			for _, param := range result {
				if param.In != "path" {
					t.Errorf("Expected parameter to be in path, got %s", param.In)
				}
				if !param.Required {
					t.Errorf("Expected path parameter to be required")
				}
				if param.Type != "string" {
					t.Errorf("Expected parameter type to be string, got %s", param.Type)
				}
			}
		})
	}
}

func TestCollectorFallbackIntegration(t *testing.T) {
	registry := NewRouteRegistry()
	_ = NewSimpleRouteCollector(nil, registry) // collector not used in this test

	// Test that inspector returns empty routes when no app is available
	inspector := NewRouteInspector(nil)
	routes := inspector.getFallbackRoutes()

	// With no app, we expect no routes (only dynamic routes from database are returned)
	if len(routes) != 0 {
		t.Errorf("Expected no routes without app, got %d routes", len(routes))
	}
}

func TestIsDocumentationEndpoint(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "OpenAPI JSON endpoint should be documentation",
			path:     "/api-docs/openapi.json",
			expected: true,
		},
		{
			name:     "Swagger UI endpoint should be documentation",
			path:     "/api-docs",
			expected: true,
		},
		{
			name:     "ReDoc endpoint should be documentation",
			path:     "/api-docs/redoc",
			expected: true,
		},
		{
			name:     "Regular API endpoint should not be documentation",
			path:     "/api/v1/hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.isDocumentationEndpoint(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestIsStaticFileRoute(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Catch-all static route should be static",
			path:     "/{path...}",
			expected: true,
		},
		{
			name:     "Static directory should be static",
			path:     "/static/css/style.css",
			expected: true,
		},
		{
			name:     "Assets directory should be static",
			path:     "/assets/js/app.js",
			expected: true,
		},
		{
			name:     "API endpoint should not be static",
			path:     "/api/v1/hello",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.isStaticFileRoute(tt.path)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(nil, registry)

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "GET records should generate List Records",
			method:   "GET",
			path:     "/api/collections/users/records",
			expected: "List Records",
		},
		{
			name:     "POST records should generate Create Record",
			method:   "POST",
			path:     "/api/collections/users/records",
			expected: "Create Record",
		},
		{
			name:     "GET single record should generate View Record",
			method:   "GET",
			path:     "/api/collections/users/records/{id}",
			expected: "View Record",
		},
		{
			name:     "PUT single record should generate Update Record",
			method:   "PUT",
			path:     "/api/collections/users/records/{id}",
			expected: "Update Record",
		},
		{
			name:     "DELETE single record should generate Delete Record",
			method:   "DELETE",
			path:     "/api/collections/users/records/{id}",
			expected: "Delete Record",
		},
		{
			name:     "Hello endpoint should generate Hello World",
			method:   "GET",
			path:     "/api/v1/hello",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.generateSummary(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for %s %s", tt.expected, result, tt.method, tt.path)
			}
		})
	}
}
