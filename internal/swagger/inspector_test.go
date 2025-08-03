package swagger

import (
	"reflect"
	"testing"
)

func TestNewRouteInspector(t *testing.T) {
	router := "mock-router"
	inspector := NewRouteInspector(router)

	if inspector == nil {
		t.Fatal("Expected inspector to be created, got nil")
	}

	if inspector.router != router {
		t.Error("Expected inspector router to match provided router")
	}
}

func TestIsRouterAccessible(t *testing.T) {
	tests := []struct {
		name     string
		router   interface{}
		expected bool
	}{
		{
			name:     "Nil router should not be accessible",
			router:   nil,
			expected: false,
		},
		{
			name:     "String router should not be accessible",
			router:   "mock-router",
			expected: false,
		},
		{
			name:     "Struct router should be accessible",
			router:   struct{ Name string }{Name: "test"},
			expected: true,
		},
		{
			name:     "Pointer to struct should be accessible",
			router:   &struct{ Name string }{Name: "test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewRouteInspector(tt.router)
			result := inspector.IsRouterAccessible()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for router %T", tt.expected, result, tt.router)
			}
		})
	}
}

func TestGetRouterInfo(t *testing.T) {
	tests := []struct {
		name   string
		router interface{}
	}{
		{
			name:   "Nil router",
			router: nil,
		},
		{
			name:   "String router",
			router: "mock-router",
		},
		{
			name:   "Struct router",
			router: struct{ Name string }{Name: "test"},
		},
		{
			name:   "Pointer to struct",
			router: &struct{ Name string }{Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewRouteInspector(tt.router)
			info := inspector.GetRouterInfo()

			if info == nil {
				t.Error("Expected router info to be returned, got nil")
			}

			// Check that status is set for nil router
			if tt.router == nil {
				if status, exists := info["status"]; !exists || status != "nil" {
					t.Error("Expected status 'nil' for nil router")
				}
			} else {
				// Check that type information is present
				if _, exists := info["type"]; !exists {
					t.Error("Expected type information in router info")
				}
				if _, exists := info["kind"]; !exists {
					t.Error("Expected kind information in router info")
				}
				if _, exists := info["accessible"]; !exists {
					t.Error("Expected accessible information in router info")
				}
			}
		})
	}
}

func TestExtractRoutes(t *testing.T) {
	tests := []struct {
		name           string
		router         interface{}
		expectFallback bool
	}{
		{
			name:           "Nil router should use fallback",
			router:         nil,
			expectFallback: true,
		},
		{
			name:           "String router should use fallback",
			router:         "mock-router",
			expectFallback: true,
		},
		{
			name:           "Struct without routes should use fallback",
			router:         struct{ Name string }{Name: "test"},
			expectFallback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inspector := NewRouteInspector(tt.router)
			routes, err := inspector.ExtractRoutes()

			// Should not return error even if extraction fails (uses fallback)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if tt.expectFallback {
				// With no app, we expect empty routes (no hardcoded fallback routes)
				if len(routes) != 0 {
					t.Errorf("Expected no routes without app, got %d routes", len(routes))
				}
			}
		})
	}
}

func TestGetFallbackRoutes(t *testing.T) {
	inspector := NewRouteInspector(nil)
	routes := inspector.getFallbackRoutes()

	// With no app, we expect no fallback routes (only dynamic routes from database)
	if len(routes) != 0 {
		t.Errorf("Expected no fallback routes without app, got %d routes", len(routes))
	}
}

func TestRemoveDuplicateRoutes(t *testing.T) {
	inspector := NewRouteInspector(nil)

	// Create test routes with duplicates
	routes := []InspectedRoute{
		{Method: "GET", Path: "/api/test", HandlerName: "handler1"},
		{Method: "POST", Path: "/api/test", HandlerName: "handler2"},
		{Method: "GET", Path: "/api/test", HandlerName: "handler3"}, // Duplicate
		{Method: "GET", Path: "/api/other", HandlerName: "handler4"},
	}

	unique := inspector.removeDuplicateRoutes(routes)

	// Should have 3 unique routes (GET+POST /api/test, GET /api/other)
	if len(unique) != 3 {
		t.Errorf("Expected 3 unique routes, got %d", len(unique))
	}

	// Check that the right routes are kept
	routeKeys := make(map[string]bool)
	for _, route := range unique {
		key := route.Method + ":" + route.Path
		if routeKeys[key] {
			t.Errorf("Duplicate route found: %s", key)
		}
		routeKeys[key] = true
	}

	// Verify expected routes are present
	expectedKeys := []string{
		"GET:/api/test",
		"POST:/api/test",
		"GET:/api/other",
	}

	for _, expectedKey := range expectedKeys {
		if !routeKeys[expectedKey] {
			t.Errorf("Expected route %s not found in unique routes", expectedKey)
		}
	}
}

func TestGetHandlerName(t *testing.T) {
	inspector := NewRouteInspector(nil)

	tests := []struct {
		name     string
		handler  interface{}
		expected string
	}{
		{
			name:     "Nil handler should return unknown",
			handler:  nil,
			expected: "unknown",
		},
		{
			name:     "String handler should return type",
			handler:  "test-handler",
			expected: "string",
		},
		{
			name:     "Function handler should return function type",
			handler:  func() {},
			expected: "func()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inspector.getHandlerName(tt.handler)
			if tt.handler == nil && result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			} else if tt.handler != nil && result == "" {
				t.Error("Expected non-empty handler name")
			}
		})
	}
}

func TestParseRouteReflection(t *testing.T) {
	inspector := NewRouteInspector(nil)

	// Test with a mock route struct
	type MockRoute struct {
		Method  string
		Path    string
		Handler interface{}
		Name    string
	}

	tests := []struct {
		name     string
		route    interface{}
		expected *InspectedRoute
	}{
		{
			name: "Valid route should be parsed",
			route: MockRoute{
				Method:  "GET",
				Path:    "/api/test",
				Handler: "test-handler",
				Name:    "test-route",
			},
			expected: &InspectedRoute{
				Method:      "GET",
				Path:        "/api/test",
				Handler:     "test-handler",
				HandlerName: "test-route",
			},
		},
		{
			name: "Route with missing method should return nil",
			route: MockRoute{
				Path:    "/api/test",
				Handler: "test-handler",
			},
			expected: nil,
		},
		{
			name: "Route with missing path should return nil",
			route: MockRoute{
				Method:  "GET",
				Handler: "test-handler",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inspector.parseRouteReflection(reflect.ValueOf(tt.route))

			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil result, got %+v", result)
				}
			} else {
				if result == nil {
					t.Fatal("Expected route to be parsed, got nil")
				}

				if result.Method != tt.expected.Method {
					t.Errorf("Expected method %s, got %s", tt.expected.Method, result.Method)
				}
				if result.Path != tt.expected.Path {
					t.Errorf("Expected path %s, got %s", tt.expected.Path, result.Path)
				}
				if result.Handler != tt.expected.Handler {
					t.Errorf("Expected handler %v, got %v", tt.expected.Handler, result.Handler)
				}
			}
		})
	}
}

func TestGetLastError(t *testing.T) {
	inspector := NewRouteInspector(nil)

	// Initially should have no error
	if err := inspector.GetLastError(); err != nil {
		t.Errorf("Expected no initial error, got %v", err)
	}

	// After extraction, should have no error for nil router (uses fallback)
	_, _ = inspector.ExtractRoutes()
	if err := inspector.GetLastError(); err == nil {
		t.Error("Expected error for nil router, got nil")
	}
}

func TestClearCache(t *testing.T) {
	inspector := NewRouteInspector(nil)

	// Extract routes to populate cache (will be empty without app)
	routes1, _ := inspector.ExtractRoutes()
	// With no app, we expect empty routes
	if len(routes1) != 0 {
		t.Errorf("Expected no routes without app, got %d routes", len(routes1))
	}

	// Verify cache is populated (even if empty)
	if _, extracted := inspector.GetCachedRoutes(); !extracted {
		t.Error("Expected extraction to be marked as done")
	}

	// Clear cache
	inspector.ClearCache()

	// Verify cache is cleared
	if cached, extracted := inspector.GetCachedRoutes(); extracted {
		t.Error("Expected extraction to be marked as not done after clear")
	} else if len(cached) != 0 {
		t.Error("Expected cached routes to be empty after clear")
	}

	if err := inspector.GetLastError(); err != nil {
		t.Error("Expected no error after cache clear")
	}
}

func TestGetExtractionStatus(t *testing.T) {
	inspector := NewRouteInspector(nil)

	// Check initial status
	status := inspector.GetExtractionStatus()
	if status["extracted_once"].(bool) {
		t.Error("Expected extracted_once to be false initially")
	}
	if status["cached_routes"].(int) != 0 {
		t.Error("Expected cached_routes to be 0 initially")
	}
	if status["has_error"].(bool) {
		t.Error("Expected has_error to be false initially")
	}

	// Extract routes
	_, _ = inspector.ExtractRoutes()

	// Check status after extraction
	status = inspector.GetExtractionStatus()
	if !status["extracted_once"].(bool) {
		t.Error("Expected extracted_once to be true after extraction")
	}
	// With no app, cached_routes will be 0
	if status["cached_routes"].(int) != 0 {
		t.Errorf("Expected cached_routes to be 0 without app, got %d", status["cached_routes"].(int))
	}
	if !status["has_error"].(bool) {
		t.Error("Expected has_error to be true for nil router")
	}
	if _, exists := status["last_error"]; !exists {
		t.Error("Expected last_error to be present when has_error is true")
	}
}
