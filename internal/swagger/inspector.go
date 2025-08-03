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

// RouteInspector handles PocketBase router interaction and route extraction
type RouteInspector struct {
	router             interface{}
	app                interface{} // PocketBase app for accessing collections
	allowedCollections []string    // Optional list of collections to include in docs
	mutex              sync.RWMutex
	extractedOnce      bool
	cachedRoutes       []InspectedRoute
	lastError          error
}

// NewRouteInspector creates a new route inspector
func NewRouteInspector(router interface{}) *RouteInspector {
	return &RouteInspector{
		router:             router,
		app:                nil,
		allowedCollections: nil,
		extractedOnce:      false,
		cachedRoutes:       nil,
		lastError:          nil,
	}
}

// NewRouteInspectorWithApp creates a new route inspector with app access
func NewRouteInspectorWithApp(router interface{}, app interface{}) *RouteInspector {
	return &RouteInspector{
		router:             router,
		app:                app,
		allowedCollections: nil,
		extractedOnce:      false,
		cachedRoutes:       nil,
		lastError:          nil,
	}
}

// NewRouteInspectorWithConfig creates a new route inspector with full configuration
func NewRouteInspectorWithConfig(router interface{}, app interface{}, allowedCollections []string) *RouteInspector {
	return &RouteInspector{
		router:             router,
		app:                app,
		allowedCollections: allowedCollections,
		extractedOnce:      false,
		cachedRoutes:       nil,
		lastError:          nil,
	}
}

// InspectedRoute represents a route extracted by the inspector
type InspectedRoute struct {
	Method      string
	Path        string
	Handler     interface{}
	HandlerName string
	Middleware  []string
}

// ExtractRoutes extracts routes from the PocketBase router with caching and error handling
func (ri *RouteInspector) ExtractRoutes() ([]InspectedRoute, error) {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()

	// Return cached results if already extracted
	if ri.extractedOnce {
		if ri.lastError != nil {
			log.Printf("Returning cached fallback routes due to previous error: %v", ri.lastError)
		} else {
			log.Printf("Returning cached extracted routes (%d routes)", len(ri.cachedRoutes))
		}
		return ri.cachedRoutes, ri.lastError
	}

	// Mark as extracted to avoid repeated attempts
	ri.extractedOnce = true

	if ri.router == nil {
		ri.lastError = fmt.Errorf("router is nil")
		log.Printf("Router is nil, using fallback route definitions")
		ri.cachedRoutes = ri.getFallbackRoutes()
		return ri.cachedRoutes, nil
	}

	// Try different extraction methods based on PocketBase version
	routes, err := ri.tryExtractRoutes()
	if err != nil {
		ri.lastError = err
		log.Printf("Route extraction failed: %v, using fallback definitions", err)
		ri.cachedRoutes = ri.getFallbackRoutes()
		return ri.cachedRoutes, nil
	}

	if len(routes) == 0 {
		ri.lastError = fmt.Errorf("no routes extracted from router")
		log.Printf("No routes extracted, using fallback definitions")
		ri.cachedRoutes = ri.getFallbackRoutes()
		return ri.cachedRoutes, nil
	}

	ri.cachedRoutes = routes
	ri.lastError = nil
	log.Printf("Successfully extracted %d routes from router", len(routes))
	return routes, nil
}

// tryExtractRoutes attempts to extract routes using various methods with improved error handling
func (ri *RouteInspector) tryExtractRoutes() ([]InspectedRoute, error) {
	var routes []InspectedRoute
	var errors []error

	// Get router information for version-specific handling
	routerInfo := ri.GetRouterInfo()
	log.Printf("Attempting route extraction from router type: %s", routerInfo["type"])

	// Method 1: Try to extract from Echo router (most common in PocketBase)
	if echoRoutes, err := ri.extractFromEcho(); err == nil && len(echoRoutes) > 0 {
		log.Printf("Successfully extracted %d routes using Echo method", len(echoRoutes))
		routes = append(routes, echoRoutes...)
	} else {
		log.Printf("Echo extraction failed: %v", err)
		errors = append(errors, fmt.Errorf("echo extraction: %w", err))
	}

	// Method 2: Try to extract from PocketBase-specific router methods
	if pbRoutes, err := ri.extractFromPocketBaseRouter(); err == nil && len(pbRoutes) > 0 {
		log.Printf("Successfully extracted %d routes using PocketBase method", len(pbRoutes))
		routes = append(routes, pbRoutes...)
	} else {
		log.Printf("PocketBase router extraction failed: %v", err)
		errors = append(errors, fmt.Errorf("pocketbase router extraction: %w", err))
	}

	// Method 3: Try to extract from ServeEvent router
	if serveRoutes, err := ri.extractFromServeEvent(); err == nil && len(serveRoutes) > 0 {
		log.Printf("Successfully extracted %d routes using ServeEvent method", len(serveRoutes))
		routes = append(routes, serveRoutes...)
	} else {
		log.Printf("ServeEvent extraction failed: %v", err)
		errors = append(errors, fmt.Errorf("serve event extraction: %w", err))
	}

	// Method 4: Try generic reflection-based extraction
	if reflectionRoutes, err := ri.extractFromReflection(); err == nil && len(reflectionRoutes) > 0 {
		log.Printf("Successfully extracted %d routes using reflection method", len(reflectionRoutes))
		routes = append(routes, reflectionRoutes...)
	} else {
		log.Printf("Reflection extraction failed: %v", err)
		errors = append(errors, fmt.Errorf("reflection extraction: %w", err))
	}

	if len(routes) == 0 {
		// Combine all errors for comprehensive error reporting
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return nil, fmt.Errorf("all extraction methods failed: %s", strings.Join(errorMessages, "; "))
	}

	// Remove duplicates and return
	uniqueRoutes := ri.removeDuplicateRoutes(routes)
	log.Printf("After deduplication: %d unique routes", len(uniqueRoutes))
	return uniqueRoutes, nil
}

// extractFromEcho attempts to extract routes from Echo router with improved error handling
func (ri *RouteInspector) extractFromEcho() ([]InspectedRoute, error) {
	var routes []InspectedRoute

	routerValue := reflect.ValueOf(ri.router)
	if routerValue.Kind() == reflect.Ptr {
		if routerValue.IsNil() {
			return nil, fmt.Errorf("router pointer is nil")
		}
		routerValue = routerValue.Elem()
	}

	if routerValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("router is not a struct, got %s", routerValue.Kind())
	}

	// Look for Echo router field with more comprehensive search
	var echoField reflect.Value
	echoFieldNames := []string{"Echo", "echo", "Router", "router", "Engine", "engine", "Mux", "mux"}

	for _, fieldName := range echoFieldNames {
		field := routerValue.FieldByName(fieldName)
		if field.IsValid() && field.CanInterface() {
			log.Printf("Found potential Echo field: %s", fieldName)
			echoField = field
			break
		}
	}

	if !echoField.IsValid() {
		return nil, fmt.Errorf("echo router field not found, tried: %v", echoFieldNames)
	}

	// Extract routes from Echo with better error handling
	echoValue := echoField
	if echoValue.Kind() == reflect.Ptr {
		if echoValue.IsNil() {
			return nil, fmt.Errorf("echo router pointer is nil")
		}
		echoValue = echoValue.Elem()
	}

	if echoValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("echo router is not a struct, got %s", echoValue.Kind())
	}

	// Look for routes field in Echo with multiple possible names
	var routesField reflect.Value
	routeFieldNames := []string{"routes", "Routes", "router", "Router", "handlers", "Handlers"}

	for _, fieldName := range routeFieldNames {
		field := echoValue.FieldByName(fieldName)
		if field.IsValid() {
			log.Printf("Found potential routes field: %s", fieldName)
			routesField = field
			break
		}
	}

	if !routesField.IsValid() {
		return nil, fmt.Errorf("routes field not found in echo router, tried: %v", routeFieldNames)
	}

	// Process routes based on field type with better error handling
	switch routesField.Kind() {
	case reflect.Slice:
		log.Printf("Processing %d routes from slice", routesField.Len())
		for i := 0; i < routesField.Len(); i++ {
			route := routesField.Index(i)
			if inspectedRoute := ri.parseRouteReflection(route); inspectedRoute != nil {
				routes = append(routes, *inspectedRoute)
			}
		}
	case reflect.Map:
		log.Printf("Processing %d routes from map", routesField.Len())
		for _, key := range routesField.MapKeys() {
			route := routesField.MapIndex(key)
			if inspectedRoute := ri.parseRouteReflection(route); inspectedRoute != nil {
				routes = append(routes, *inspectedRoute)
			}
		}
	default:
		return nil, fmt.Errorf("routes field is not a slice or map, got %s", routesField.Kind())
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no valid routes found in echo router")
	}

	return routes, nil
}

// extractFromPocketBaseRouter attempts to extract routes using PocketBase-specific methods
func (ri *RouteInspector) extractFromPocketBaseRouter() ([]InspectedRoute, error) {
	var routes []InspectedRoute

	routerValue := reflect.ValueOf(ri.router)
	if routerValue.Kind() == reflect.Ptr {
		if routerValue.IsNil() {
			return nil, fmt.Errorf("router pointer is nil")
		}
		routerValue = routerValue.Elem()
	}

	if routerValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("router is not a struct, got %s", routerValue.Kind())
	}

	// Try to find PocketBase-specific route information
	// Look for common PocketBase router patterns
	routerType := routerValue.Type()

	// Check for methods that might give us route information
	for i := 0; i < routerType.NumMethod(); i++ {
		method := routerType.Method(i)
		methodName := strings.ToLower(method.Name)

		// Look for route-related methods
		if strings.Contains(methodName, "route") || strings.Contains(methodName, "handler") {
			log.Printf("Found potential route method: %s", method.Name)

			// Try to call methods that might return route information
			if method.Type.NumIn() == 1 && method.Type.NumOut() > 0 {
				// Method takes only receiver, might return route info
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Method %s panicked: %v", method.Name, r)
					}
				}()

				results := routerValue.Method(i).Call([]reflect.Value{})
				if len(results) > 0 {
					// Try to parse the result as route information
					if parsedRoutes := ri.parseMethodResult(results[0], method.Name); len(parsedRoutes) > 0 {
						routes = append(routes, parsedRoutes...)
					}
				}
			}
		}
	}

	// Look for fields that might contain route information
	for i := 0; i < routerValue.NumField(); i++ {
		field := routerValue.Field(i)
		fieldType := routerType.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		fieldName := strings.ToLower(fieldType.Name)
		if strings.Contains(fieldName, "route") || strings.Contains(fieldName, "handler") {
			log.Printf("Found potential route field: %s", fieldType.Name)

			// Special handling for RouterGroup field
			if fieldType.Name == "RouterGroup" {
				if groupRoutes := ri.extractFromRouterGroup(field); len(groupRoutes) > 0 {
					routes = append(routes, groupRoutes...)
				}
			} else {
				if parsedRoutes := ri.parseFieldValue(field, fieldType.Name); len(parsedRoutes) > 0 {
					routes = append(routes, parsedRoutes...)
				}
			}
		}
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found using PocketBase-specific methods")
	}

	return routes, nil
}

// extractFromRouterGroup attempts to extract routes from a RouterGroup field
func (ri *RouteInspector) extractFromRouterGroup(groupField reflect.Value) []InspectedRoute {
	var routes []InspectedRoute

	if !groupField.IsValid() || !groupField.CanInterface() {
		return routes
	}

	// Handle pointer to RouterGroup
	if groupField.Kind() == reflect.Ptr {
		if groupField.IsNil() {
			return routes
		}
		groupField = groupField.Elem()
	}

	if groupField.Kind() != reflect.Struct {
		return routes
	}

	log.Printf("Examining RouterGroup structure...")

	// Look for route-related fields in RouterGroup
	groupType := groupField.Type()
	for i := 0; i < groupField.NumField(); i++ {
		field := groupField.Field(i)
		fieldType := groupType.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		fieldName := strings.ToLower(fieldType.Name)
		log.Printf("RouterGroup field: %s (type: %s)", fieldType.Name, fieldType.Type)

		if strings.Contains(fieldName, "route") || strings.Contains(fieldName, "handler") ||
			strings.Contains(fieldName, "tree") || strings.Contains(fieldName, "node") ||
			strings.Contains(fieldName, "path") || strings.Contains(fieldName, "method") {
			log.Printf("Found potential route field in RouterGroup: %s", fieldType.Name)

			if parsedRoutes := ri.parseFieldValue(field, fieldType.Name); len(parsedRoutes) > 0 {
				routes = append(routes, parsedRoutes...)
			}
		}
	}

	// Also try to call methods on RouterGroup that might return route information
	groupPtr := groupField.Addr()
	if groupPtr.CanInterface() {
		groupType := groupPtr.Type()
		for i := 0; i < groupType.NumMethod(); i++ {
			method := groupType.Method(i)
			methodName := strings.ToLower(method.Name)

			if strings.Contains(methodName, "route") || strings.Contains(methodName, "handler") {
				log.Printf("Found RouterGroup method: %s", method.Name)

				// Try to call methods that don't require parameters
				if method.Type.NumIn() == 1 && method.Type.NumOut() > 0 {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("RouterGroup method %s panicked: %v", method.Name, r)
						}
					}()

					results := groupPtr.Method(i).Call([]reflect.Value{})
					if len(results) > 0 {
						if parsedRoutes := ri.parseMethodResult(results[0], method.Name); len(parsedRoutes) > 0 {
							routes = append(routes, parsedRoutes...)
						}
					}
				}
			}
		}
	}

	return routes
}

// extractFromReflection attempts generic reflection-based route extraction
func (ri *RouteInspector) extractFromReflection() ([]InspectedRoute, error) {
	routerValue := reflect.ValueOf(ri.router)
	if routerValue.Kind() == reflect.Ptr {
		if routerValue.IsNil() {
			return nil, fmt.Errorf("router pointer is nil")
		}
		routerValue = routerValue.Elem()
	}

	if routerValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("router is not a struct for reflection")
	}

	// Recursively search for route-like structures
	foundRoutes := ri.searchForRoutes(routerValue, "router", 0, 3) // Max depth of 3

	if len(foundRoutes) == 0 {
		return nil, fmt.Errorf("no routes found using reflection")
	}

	return foundRoutes, nil
}

// parseMethodResult attempts to parse method call results as route information
func (ri *RouteInspector) parseMethodResult(result reflect.Value, methodName string) []InspectedRoute {
	var routes []InspectedRoute

	if !result.IsValid() {
		return routes
	}

	// Handle different result types
	switch result.Kind() {
	case reflect.Slice:
		for i := 0; i < result.Len(); i++ {
			if route := ri.parseRouteReflection(result.Index(i)); route != nil {
				routes = append(routes, *route)
			}
		}
	case reflect.Map:
		for _, key := range result.MapKeys() {
			if route := ri.parseRouteReflection(result.MapIndex(key)); route != nil {
				routes = append(routes, *route)
			}
		}
	case reflect.Struct:
		if route := ri.parseRouteReflection(result); route != nil {
			routes = append(routes, *route)
		}
	}

	return routes
}

// parseFieldValue attempts to parse field values as route information
func (ri *RouteInspector) parseFieldValue(field reflect.Value, fieldName string) []InspectedRoute {
	var routes []InspectedRoute

	if !field.IsValid() || !field.CanInterface() {
		return routes
	}

	switch field.Kind() {
	case reflect.Slice:
		for i := 0; i < field.Len(); i++ {
			if route := ri.parseRouteReflection(field.Index(i)); route != nil {
				routes = append(routes, *route)
			}
		}
	case reflect.Map:
		for _, key := range field.MapKeys() {
			if route := ri.parseRouteReflection(field.MapIndex(key)); route != nil {
				routes = append(routes, *route)
			}
		}
	case reflect.Struct:
		if route := ri.parseRouteReflection(field); route != nil {
			routes = append(routes, *route)
		}
	}

	return routes
}

// searchForRoutes recursively searches for route-like structures
func (ri *RouteInspector) searchForRoutes(value reflect.Value, path string, depth, maxDepth int) []InspectedRoute {
	var routes []InspectedRoute

	if depth > maxDepth || !value.IsValid() {
		return routes
	}

	switch value.Kind() {
	case reflect.Ptr:
		if !value.IsNil() {
			routes = append(routes, ri.searchForRoutes(value.Elem(), path+"->", depth+1, maxDepth)...)
		}
	case reflect.Struct:
		// Try to parse as a route first
		if route := ri.parseRouteReflection(value); route != nil {
			routes = append(routes, *route)
		}

		// Search struct fields
		valueType := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := value.Field(i)
			fieldType := valueType.Field(i)

			if !fieldType.IsExported() {
				continue
			}

			fieldPath := path + "." + fieldType.Name
			fieldRoutes := ri.searchForRoutes(field, fieldPath, depth+1, maxDepth)
			routes = append(routes, fieldRoutes...)
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			indexPath := fmt.Sprintf("%s[%d]", path, i)
			indexRoutes := ri.searchForRoutes(value.Index(i), indexPath, depth+1, maxDepth)
			routes = append(routes, indexRoutes...)
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			keyPath := fmt.Sprintf("%s[%v]", path, key.Interface())
			mapRoutes := ri.searchForRoutes(value.MapIndex(key), keyPath, depth+1, maxDepth)
			routes = append(routes, mapRoutes...)
		}
	}

	return routes
}

// extractFromServeEvent attempts to extract routes from ServeEvent router
func (ri *RouteInspector) extractFromServeEvent() ([]InspectedRoute, error) {
	var routes []InspectedRoute

	routerValue := reflect.ValueOf(ri.router)
	routerType := reflect.TypeOf(ri.router)

	if routerType.Kind() == reflect.Ptr {
		if routerValue.IsNil() {
			return nil, fmt.Errorf("serve event router pointer is nil")
		}
		routerValue = routerValue.Elem()
		routerType = routerType.Elem()
	}

	// Try to find route-related methods that we can call
	for i := 0; i < routerType.NumMethod(); i++ {
		method := routerType.Method(i)
		methodName := strings.ToLower(method.Name)

		// Look for methods that might contain route information
		if strings.Contains(methodName, "route") || strings.Contains(methodName, "handler") {
			log.Printf("Found potential route method: %s", method.Name)

			// Try to call methods that might return route information
			methodType := method.Type
			if methodType.NumIn() == 1 { // Only receiver
				// Try to call the method safely
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Method %s panicked: %v", method.Name, r)
					}
				}()

				results := routerValue.Method(i).Call([]reflect.Value{})
				if len(results) > 0 {
					if parsedRoutes := ri.parseMethodResult(results[0], method.Name); len(parsedRoutes) > 0 {
						routes = append(routes, parsedRoutes...)
					}
				}
			}
		}
	}

	// Also try to access fields that might contain route information
	if routerValue.Kind() == reflect.Struct {
		for i := 0; i < routerValue.NumField(); i++ {
			field := routerValue.Field(i)
			fieldType := routerType.Field(i)

			if !fieldType.IsExported() {
				continue
			}

			fieldName := strings.ToLower(fieldType.Name)
			if strings.Contains(fieldName, "route") || strings.Contains(fieldName, "handler") ||
				strings.Contains(fieldName, "mux") || strings.Contains(fieldName, "echo") {
				log.Printf("Found potential route field in ServeEvent: %s", fieldType.Name)

				if parsedRoutes := ri.parseFieldValue(field, fieldType.Name); len(parsedRoutes) > 0 {
					routes = append(routes, parsedRoutes...)
				}
			}
		}
	}

	if len(routes) == 0 {
		return nil, fmt.Errorf("no routes found in serve event router")
	}

	return routes, nil
}

// parseRouteReflection parses a route using reflection with improved field detection
func (ri *RouteInspector) parseRouteReflection(routeValue reflect.Value) *InspectedRoute {
	if !routeValue.IsValid() {
		return nil
	}

	if routeValue.Kind() == reflect.Ptr {
		if routeValue.IsNil() {
			return nil
		}
		routeValue = routeValue.Elem()
	}

	if routeValue.Kind() != reflect.Struct {
		return nil
	}

	var method, path string
	var handler interface{}
	var handlerName string

	routeType := routeValue.Type()

	// Try multiple field names for method
	methodFieldNames := []string{"Method", "method", "HTTPMethod", "Verb", "verb"}
	for _, fieldName := range methodFieldNames {
		if methodField := routeValue.FieldByName(fieldName); methodField.IsValid() && methodField.Kind() == reflect.String {
			method = methodField.String()
			break
		}
	}

	// Try multiple field names for path
	pathFieldNames := []string{"Path", "path", "Pattern", "pattern", "Route", "route", "URL", "url"}
	for _, fieldName := range pathFieldNames {
		if pathField := routeValue.FieldByName(fieldName); pathField.IsValid() && pathField.Kind() == reflect.String {
			path = pathField.String()
			break
		}
	}

	// Try multiple field names for handler
	handlerFieldNames := []string{"Handler", "handler", "HandlerFunc", "Func", "func", "Action", "action"}
	for _, fieldName := range handlerFieldNames {
		if handlerField := routeValue.FieldByName(fieldName); handlerField.IsValid() && handlerField.CanInterface() {
			handler = handlerField.Interface()
			handlerName = ri.getHandlerName(handler)
			break
		}
	}

	// Try multiple field names for name/identifier
	nameFieldNames := []string{"Name", "name", "ID", "id", "Identifier", "identifier"}
	for _, fieldName := range nameFieldNames {
		if nameField := routeValue.FieldByName(fieldName); nameField.IsValid() && nameField.Kind() == reflect.String {
			if handlerName == "" {
				handlerName = nameField.String()
			}
			break
		}
	}

	// If we still don't have a handler name, try to generate one from the struct
	if handlerName == "" {
		handlerName = fmt.Sprintf("%s.%s", routeType.PkgPath(), routeType.Name())
	}

	// Validate that we have the minimum required fields
	if method == "" || path == "" {
		// Log what we found for debugging
		log.Printf("Incomplete route found - Method: '%s', Path: '%s', Handler: '%s'", method, path, handlerName)
		return nil
	}

	// Clean up the method and path
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)

	return &InspectedRoute{
		Method:      method,
		Path:        path,
		Handler:     handler,
		HandlerName: handlerName,
		Middleware:  ri.extractMiddlewareFromRoute(routeValue),
	}
}

// getHandlerName extracts a readable handler name
func (ri *RouteInspector) getHandlerName(handler interface{}) string {
	if handler == nil {
		return "unknown"
	}

	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() == reflect.Func {
		// Try to get function name
		handlerValue := reflect.ValueOf(handler)
		if handlerValue.IsValid() {
			return handlerType.String()
		}
	}

	return fmt.Sprintf("%T", handler)
}

// extractMiddlewareFromRoute extracts middleware information from route reflection
func (ri *RouteInspector) extractMiddlewareFromRoute(routeValue reflect.Value) []string {
	var middleware []string

	// Look for middleware field
	if middlewareField := routeValue.FieldByName("Middleware"); middlewareField.IsValid() {
		if middlewareField.Kind() == reflect.Slice {
			for i := 0; i < middlewareField.Len(); i++ {
				mw := middlewareField.Index(i)
				if mw.CanInterface() {
					middleware = append(middleware, fmt.Sprintf("%T", mw.Interface()))
				}
			}
		}
	}

	return middleware
}

// removeDuplicateRoutes removes duplicate routes from the slice
func (ri *RouteInspector) removeDuplicateRoutes(routes []InspectedRoute) []InspectedRoute {
	seen := make(map[string]bool)
	var unique []InspectedRoute

	for _, route := range routes {
		key := fmt.Sprintf("%s:%s", route.Method, route.Path)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, route)
		}
	}

	return unique
}

// getFallbackRoutes returns dynamic routes and known custom routes when extraction fails
func (ri *RouteInspector) getFallbackRoutes() []InspectedRoute {
	var routes []InspectedRoute

	// Add dynamic collection routes based on actual database collections
	if dynamicRoutes := ri.generateDynamicCollectionRoutes(); len(dynamicRoutes) > 0 {
		routes = append(routes, dynamicRoutes...)
		log.Printf("Added %d dynamic collection routes", len(dynamicRoutes))
	}

	// Only add known custom routes if we have an app (indicating we're in a real application context)
	if ri.app != nil {
		customRoutes := ri.getKnownCustomRoutes()
		if len(customRoutes) > 0 {
			routes = append(routes, customRoutes...)
			log.Printf("Added %d known custom routes", len(customRoutes))
		}
	}

	if len(routes) == 0 {
		log.Printf("No routes available for fallback")
	} else {
		log.Printf("Using %d total routes as fallback", len(routes))
	}

	return routes
}

// getKnownCustomRoutes returns the custom routes defined in the application
func (ri *RouteInspector) getKnownCustomRoutes() []InspectedRoute {
	// These are the custom routes defined in internal/routes/routes.go
	// This is a temporary solution until we can extract routes directly from the router
	return []InspectedRoute{
		{
			Method:      "GET",
			Path:        "/api/v1/hello",
			HandlerName: "custom.hello",
			Middleware:  []string{},
		},
		{
			Method:      "GET",
			Path:        "/api/v1/protected",
			HandlerName: "custom.protected",
			Middleware:  []string{"auth"},
		},
		{
			Method:      "GET",
			Path:        "/api/v1/permission-test",
			HandlerName: "custom.permissionTest",
			Middleware:  []string{"auth", "permission"},
		},
		{
			Method:      "POST",
			Path:        "/api/v1/users/export",
			HandlerName: "custom.userExport",
			Middleware:  []string{"auth", "permission"},
		},
		{
			Method:      "GET",
			Path:        "/api/v1/jobs/{id}/status",
			HandlerName: "custom.jobStatus",
			Middleware:  []string{"auth"},
		},
		{
			Method:      "POST",
			Path:        "/api/v1/jobs/{id}/download",
			HandlerName: "custom.jobDownload",
			Middleware:  []string{"auth"},
		},
	}
}

// generateDynamicCollectionRoutes generates routes for all collections in the database
func (ri *RouteInspector) generateDynamicCollectionRoutes() []InspectedRoute {
	var routes []InspectedRoute

	if ri.app == nil {
		log.Printf("No app available for dynamic collection route generation")
		return routes
	}

	// Try to cast app to PocketBase
	app, ok := ri.app.(*pocketbase.PocketBase)
	if !ok {
		log.Printf("App is not a PocketBase instance, cannot generate dynamic routes")
		return routes
	}

	// For now, use a simple approach to detect common collections
	// In a real implementation, you would query the database for actual collections
	commonCollections := ri.detectCommonCollections(app)

	log.Printf("Generating dynamic routes for %d detected collections", len(commonCollections))

	// Generate routes for each collection
	for _, collectionInfo := range commonCollections {
		collectionName := collectionInfo.Name
		collectionType := collectionInfo.Type
		log.Printf("Generating routes for collection: %s (type: %s)", collectionName, collectionType)

		// Generate standard CRUD routes for each collection
		collectionRoutes := []InspectedRoute{
			{
				Method:      "GET",
				Path:        fmt.Sprintf("/api/collections/%s/records", collectionName),
				HandlerName: "pocketbase.listRecords",
				Middleware:  ri.getCollectionMiddleware(collectionName),
			},
			{
				Method:      "POST",
				Path:        fmt.Sprintf("/api/collections/%s/records", collectionName),
				HandlerName: "pocketbase.createRecord",
				Middleware:  ri.getCollectionMiddleware(collectionName),
			},
			{
				Method:      "GET",
				Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collectionName),
				HandlerName: "pocketbase.viewRecord",
				Middleware:  ri.getCollectionMiddleware(collectionName),
			},
			{
				Method:      "PUT",
				Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collectionName),
				HandlerName: "pocketbase.updateRecord",
				Middleware:  ri.getCollectionMiddleware(collectionName),
			},
			{
				Method:      "DELETE",
				Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collectionName),
				HandlerName: "pocketbase.deleteRecord",
				Middleware:  ri.getCollectionMiddleware(collectionName),
			},
		}

		// Add auth routes for auth collections
		if collectionType == "auth" {
			authRoutes := []InspectedRoute{
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/auth-with-password", collectionName),
					HandlerName: "pocketbase.authWithPassword",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/auth-refresh", collectionName),
					HandlerName: "pocketbase.authRefresh",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/request-password-reset", collectionName),
					HandlerName: "pocketbase.requestPasswordReset",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/confirm-password-reset", collectionName),
					HandlerName: "pocketbase.confirmPasswordReset",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/request-verification", collectionName),
					HandlerName: "pocketbase.requestVerification",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/confirm-verification", collectionName),
					HandlerName: "pocketbase.confirmVerification",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/request-email-change", collectionName),
					HandlerName: "pocketbase.requestEmailChange",
					Middleware:  []string{},
				},
				{
					Method:      "POST",
					Path:        fmt.Sprintf("/api/collections/%s/confirm-email-change", collectionName),
					HandlerName: "pocketbase.confirmEmailChange",
					Middleware:  []string{},
				},
			}
			collectionRoutes = append(collectionRoutes, authRoutes...)
		}

		routes = append(routes, collectionRoutes...)
	}

	log.Printf("Generated %d dynamic collection routes", len(routes))
	return routes
}

// CollectionInfo represents basic collection information
type CollectionInfo struct {
	Name string
	Type string
}

// detectCommonCollections queries the database to get actual collections
func (ri *RouteInspector) detectCommonCollections(app *pocketbase.PocketBase) []CollectionInfo {
	var collections []CollectionInfo

	// Query the _collections table to get actual collections
	type dbCollection struct {
		Name string `db:"name"`
		Type string `db:"type"`
	}

	var dbCollections []dbCollection
	err := app.DB().NewQuery("SELECT name, type FROM _collections WHERE type IN ('base', 'auth')").All(&dbCollections)
	if err != nil {
		log.Printf("Failed to query collections from database: %v", err)
		// Return empty slice instead of fake data
		return collections
	}

	// Convert database results to CollectionInfo, filtering by allowed collections if specified
	for _, dbCol := range dbCollections {
		// If allowedCollections is specified, only include collections in that list
		if ri.allowedCollections != nil && len(ri.allowedCollections) > 0 {
			allowed := false
			for _, allowedName := range ri.allowedCollections {
				if dbCol.Name == allowedName {
					allowed = true
					break
				}
			}
			if !allowed {
				log.Printf("Skipping collection '%s' - not in allowed list", dbCol.Name)
				continue
			}
		}

		collections = append(collections, CollectionInfo{
			Name: dbCol.Name,
			Type: dbCol.Type,
		})
	}

	if ri.allowedCollections != nil && len(ri.allowedCollections) > 0 {
		log.Printf("Found %d collections (filtered by allowed list %v): %v", len(collections), ri.allowedCollections, collections)
	} else {
		log.Printf("Found %d actual collections in database: %v", len(collections), collections)
	}
	return collections
}

// getCollectionMiddleware determines middleware for a specific collection
func (ri *RouteInspector) getCollectionMiddleware(collectionName string) []string {
	// Check if this is a protected collection
	protectedCollections := []string{"users", "roles", "permissions"}
	for _, protected := range protectedCollections {
		if collectionName == protected {
			return []string{"auth"}
		}
	}
	return []string{}
}

// IsRouterAccessible checks if the router can be accessed for inspection
func (ri *RouteInspector) IsRouterAccessible() bool {
	if ri.router == nil {
		return false
	}

	// Try basic reflection access
	routerValue := reflect.ValueOf(ri.router)
	if !routerValue.IsValid() {
		return false
	}

	// Check if we can access the router structure
	if routerValue.Kind() == reflect.Ptr {
		if routerValue.IsNil() {
			return false
		}
		routerValue = routerValue.Elem()
	}

	return routerValue.Kind() == reflect.Struct
}

// GetRouterInfo returns information about the router for debugging
func (ri *RouteInspector) GetRouterInfo() map[string]interface{} {
	info := make(map[string]interface{})

	if ri.router == nil {
		info["status"] = "nil"
		return info
	}

	routerValue := reflect.ValueOf(ri.router)
	routerType := reflect.TypeOf(ri.router)

	info["type"] = routerType.String()
	info["kind"] = routerValue.Kind().String()
	info["accessible"] = ri.IsRouterAccessible()

	if routerValue.Kind() == reflect.Ptr {
		info["is_nil"] = routerValue.IsNil()
		if !routerValue.IsNil() {
			routerValue = routerValue.Elem()
			routerType = routerType.Elem()
		}
	}

	if routerValue.Kind() == reflect.Struct {
		var fields []string
		for i := 0; i < routerValue.NumField(); i++ {
			field := routerType.Field(i)
			if field.IsExported() {
				fields = append(fields, field.Name)
			}
		}
		info["exported_fields"] = fields

		var methods []string
		for i := 0; i < routerType.NumMethod(); i++ {
			method := routerType.Method(i)
			if method.IsExported() {
				methods = append(methods, method.Name)
			}
		}
		info["exported_methods"] = methods
	}

	return info
}

// GetLastError returns the last error encountered during route extraction
func (ri *RouteInspector) GetLastError() error {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	return ri.lastError
}

// ClearCache clears the cached routes and allows re-extraction
func (ri *RouteInspector) ClearCache() {
	ri.mutex.Lock()
	defer ri.mutex.Unlock()
	ri.extractedOnce = false
	ri.cachedRoutes = nil
	ri.lastError = nil
}

// GetCachedRoutes returns cached routes if available
func (ri *RouteInspector) GetCachedRoutes() ([]InspectedRoute, bool) {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()
	return ri.cachedRoutes, ri.extractedOnce
}

// GetExtractionStatus returns information about the extraction status
func (ri *RouteInspector) GetExtractionStatus() map[string]interface{} {
	ri.mutex.RLock()
	defer ri.mutex.RUnlock()

	status := map[string]interface{}{
		"extracted_once": ri.extractedOnce,
		"cached_routes":  len(ri.cachedRoutes),
		"has_error":      ri.lastError != nil,
	}

	if ri.lastError != nil {
		status["last_error"] = ri.lastError.Error()
	}

	return status
}

// CreateFromServeEvent creates a RouteInspector from a ServeEvent
func CreateFromServeEvent(se *core.ServeEvent) *RouteInspector {
	return NewRouteInspector(se.Router)
}

// CreateFromServeEventWithApp creates a RouteInspector from a ServeEvent with app access
func CreateFromServeEventWithApp(se *core.ServeEvent, app interface{}) *RouteInspector {
	return NewRouteInspectorWithApp(se.Router, app)
}
