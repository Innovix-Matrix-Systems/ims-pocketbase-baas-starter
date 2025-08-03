package swagger

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// SimpleRouteCollector handles route discovery during OnServe hook execution
type SimpleRouteCollector struct {
	registry           *RouteRegistry
	app                *pocketbase.PocketBase
	allowedCollections []string
	mutex              sync.Mutex
}

// NewSimpleRouteCollector creates a new route collector
func NewSimpleRouteCollector(app *pocketbase.PocketBase, registry *RouteRegistry) *SimpleRouteCollector {
	return &SimpleRouteCollector{
		registry:           registry,
		app:                app,
		allowedCollections: nil,
	}
}

// NewSimpleRouteCollectorWithConfig creates a new route collector with configuration
func NewSimpleRouteCollectorWithConfig(app *pocketbase.PocketBase, registry *RouteRegistry, allowedCollections []string) *SimpleRouteCollector {
	return &SimpleRouteCollector{
		registry:           registry,
		app:                app,
		allowedCollections: allowedCollections,
	}
}

// CollectRoutes discovers routes from the PocketBase router during OnServe hook
func (c *SimpleRouteCollector) CollectRoutes(se *core.ServeEvent) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Avoid collecting routes multiple times
	if c.registry.IsCollected() {
		return nil
	}

	log.Printf("Starting route discovery...")

	// Create RouteInspector for router access
	inspector := NewRouteInspectorWithConfig(se.Router, c.app, c.allowedCollections)

	// Log router information for debugging
	routerInfo := inspector.GetRouterInfo()
	log.Printf("Router info: %+v", routerInfo)

	// Extract routes using the inspector
	inspectedRoutes, err := inspector.ExtractRoutes()
	if err != nil {
		log.Printf("Route inspection failed: %v", err)
		// Continue with fallback routes - inspector handles this internally
	}

	// Convert inspected routes to DiscoveredRoute format
	var routes []DiscoveredRoute
	for _, inspectedRoute := range inspectedRoutes {
		route := DiscoveredRoute{
			Method:  inspectedRoute.Method,
			Path:    inspectedRoute.Path,
			Handler: inspectedRoute.Handler,
		}
		routes = append(routes, route)
	}

	// Process and filter discovered routes
	for _, route := range routes {
		if c.shouldIncludeRoute(route) {
			// Extract metadata and add to registry
			routeInfo := c.extractRouteMetadata(route)
			c.registry.AddRoute(routeInfo)
		}
	}

	// Mark registry as collected
	c.registry.SetCollected(true)

	log.Printf("Route discovery completed. Collected %d routes", c.registry.Count())
	return nil
}

// DiscoveredRoute represents a route discovered from the router
type DiscoveredRoute struct {
	Method  string
	Path    string
	Handler interface{}
}

// shouldIncludeRoute determines if a route should be included in the registry
func (c *SimpleRouteCollector) shouldIncludeRoute(route DiscoveredRoute) bool {
	// Filter out documentation endpoints
	if c.isDocumentationEndpoint(route.Path) {
		return false
	}

	// Filter out static file routes
	if c.isStaticFileRoute(route.Path) {
		return false
	}

	// Filter out internal/system routes
	if c.isInternalRoute(route.Path) {
		return false
	}

	return true
}

// isDocumentationEndpoint checks if the route is a documentation endpoint
func (c *SimpleRouteCollector) isDocumentationEndpoint(path string) bool {
	docPaths := []string{
		"/api-docs",
		"/api-docs/",
		"/api-docs/openapi.json",
		"/api-docs/redoc",
		"/swagger",
		"/swagger/",
		"/docs",
		"/docs/",
	}

	for _, docPath := range docPaths {
		if strings.HasPrefix(path, docPath) {
			return true
		}
	}

	return false
}

// isStaticFileRoute checks if the route serves static files
func (c *SimpleRouteCollector) isStaticFileRoute(path string) bool {
	staticPatterns := []string{
		"/{path...}",
		"/static/",
		"/assets/",
		"/public/",
		"/*filepath",
	}

	for _, pattern := range staticPatterns {
		if strings.Contains(path, pattern) || strings.HasPrefix(path, pattern) {
			return true
		}
	}

	return false
}

// isInternalRoute checks if the route is an internal system route
func (c *SimpleRouteCollector) isInternalRoute(path string) bool {
	internalPrefixes := []string{
		"/_",
		"/internal/",
		"/system/",
		"/health",
		"/metrics",
	}

	for _, prefix := range internalPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// extractRouteMetadata extracts metadata from a discovered route
func (c *SimpleRouteCollector) extractRouteMetadata(route DiscoveredRoute) RouteInfo {
	routeInfo := RouteInfo{
		Method:      strings.ToUpper(route.Method),
		Path:        route.Path,
		Handler:     c.getHandlerName(route.Handler),
		IsProtected: c.isProtectedRoute(route.Path),
		Middleware:  c.extractMiddleware(route),
		Tags:        c.generateTags(route.Path),
		Summary:     c.generateSummary(route.Method, route.Path),
		Description: c.generateDescription(route.Method, route.Path),
		Parameters:  c.extractParameters(route.Path),
	}

	return routeInfo
}

// getHandlerName extracts a readable handler name
func (c *SimpleRouteCollector) getHandlerName(handler interface{}) string {
	if handler == nil {
		return "unknown"
	}

	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() == reflect.Func {
		return handlerType.String()
	}

	return fmt.Sprintf("%T", handler)
}

// isProtectedRoute determines if a route requires authentication
func (c *SimpleRouteCollector) isProtectedRoute(path string) bool {
	// Check against protected collections from common package
	protectedPrefixes := []string{
		"/api/collections/users",
		"/api/collections/roles",
		"/api/collections/permissions",
		"/api/v1/protected",
		"/api/v1/permission-test",
		"/api/v1/users/export",
		"/api/v1/jobs",
	}

	for _, prefix := range protectedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	// Check if it's an admin route
	if strings.HasPrefix(path, "/api/admins") && !strings.Contains(path, "auth-with-password") {
		return true
	}

	return false
}

// extractMiddleware extracts middleware information from route
func (c *SimpleRouteCollector) extractMiddleware(route DiscoveredRoute) []string {
	var middleware []string

	if c.isProtectedRoute(route.Path) {
		middleware = append(middleware, "auth")
	}

	// Add permission middleware for specific routes
	if strings.Contains(route.Path, "permission-test") {
		middleware = append(middleware, "permission")
	}

	return middleware
}

// generateTags generates appropriate tags for the route
func (c *SimpleRouteCollector) generateTags(path string) []string {
	if strings.HasPrefix(path, "/api/collections") {
		return []string{"PocketBase Collections"}
	}

	if strings.HasPrefix(path, "/api/admins") {
		return []string{"Admin Authentication"}
	}

	if strings.HasPrefix(path, "/api/users") {
		return []string{"User Authentication"}
	}

	if strings.HasPrefix(path, "/api/files") {
		return []string{"Files"}
	}

	if strings.HasPrefix(path, "/api/v1") {
		return []string{"Custom API"}
	}

	return []string{"API"}
}

// generateSummary generates a summary for the route
func (c *SimpleRouteCollector) generateSummary(method, path string) string {
	method = strings.ToUpper(method)

	// Handle specific known routes
	if strings.Contains(path, "/records") && !strings.Contains(path, "/{id}") {
		switch method {
		case "GET":
			return "List Records"
		case "POST":
			return "Create Record"
		}
	}

	if strings.Contains(path, "/records/{id}") {
		switch method {
		case "GET":
			return "View Record"
		case "PUT", "PATCH":
			return "Update Record"
		case "DELETE":
			return "Delete Record"
		}
	}

	if strings.Contains(path, "auth-with-password") {
		return "Authenticate with Password"
	}

	if strings.Contains(path, "/hello") {
		return "Hello World"
	}

	if strings.Contains(path, "/protected") {
		return "Protected Endpoint"
	}

	// Generate generic summary
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) > 0 {
		resource := pathParts[len(pathParts)-1]
		// Remove parameter brackets
		resource = strings.ReplaceAll(resource, "{", "")
		resource = strings.ReplaceAll(resource, "}", "")

		switch method {
		case "GET":
			return fmt.Sprintf("Get %s", strings.Title(resource))
		case "POST":
			return fmt.Sprintf("Create %s", strings.Title(resource))
		case "PUT", "PATCH":
			return fmt.Sprintf("Update %s", strings.Title(resource))
		case "DELETE":
			return fmt.Sprintf("Delete %s", strings.Title(resource))
		}
	}

	return fmt.Sprintf("%s %s", method, path)
}

// generateDescription generates a description for the route
func (c *SimpleRouteCollector) generateDescription(method, path string) string {
	summary := c.generateSummary(method, path)

	if c.isProtectedRoute(path) {
		return summary + " (requires authentication)"
	}

	return summary
}

// extractParameters extracts path and query parameters from the route path
func (c *SimpleRouteCollector) extractParameters(path string) []RouteParameter {
	var parameters []RouteParameter

	// Extract path parameters (e.g., {id}, {collection})
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := strings.Trim(part, "{}")

			param := RouteParameter{
				Name:        paramName,
				In:          "path",
				Required:    true,
				Type:        "string",
				Description: fmt.Sprintf("The %s identifier", paramName),
			}

			// Add specific descriptions for known parameters
			switch paramName {
			case "id":
				param.Description = "The record ID"
			case "collection":
				param.Description = "The collection name"
			case "filename":
				param.Description = "The filename"
			}

			parameters = append(parameters, param)
		}
	}

	return parameters
}
