package swagger

import (
	"fmt"
	"ims-pocketbase-baas-starter/pkg/common"
	"log"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CombinedOpenAPISpec represents the unified OpenAPI specification
type CombinedOpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Servers    []Server            `json:"servers,omitempty"`
	Paths      map[string]PathItem `json:"paths"`
	Components *Components         `json:"components,omitempty"`
	Tags       []Tag               `json:"tags,omitempty"`
}

// Server represents an OpenAPI server object
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Tag represents an OpenAPI tag object
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Info represents the OpenAPI info object
type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// PathItem represents a path item in OpenAPI
type PathItem map[string]Operation

// Operation represents an operation in OpenAPI
type Operation struct {
	Summary     string                `json:"summary,omitempty"`
	Description string                `json:"description,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Tags        []string              `json:"tags,omitempty"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
}

// Components represents OpenAPI components
type Components struct {
	Schemas         map[string]interface{}    `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme in OpenAPI
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

// UnifiedConfig holds the unified generator configuration
type UnifiedConfig struct {
	Title                     string
	Version                   string
	Description               string
	ServerURL                 string
	ExcludedCollections       []string
	IncludeSystem             bool
	CustomRoutes              []CustomRoute
	EnableAuth                bool
	IncludeExamples           bool
	EnableDiscovery           bool // Enable automatic route discovery
	EnableDynamicContentTypes bool // Enable dynamic content type detection for file fields
}

// RouteInfo contains metadata about a discovered route
type RouteInfo struct {
	Method      string           `json:"method"`
	Path        string           `json:"path"`
	Handler     string           `json:"handler"`
	Summary     string           `json:"summary"`
	Description string           `json:"description"`
	Tags        []string         `json:"tags"`
	IsProtected bool             `json:"is_protected"`
	Middleware  []string         `json:"middleware"`
	Parameters  []RouteParameter `json:"parameters,omitempty"`
}

// RouteParameter represents a route parameter
type RouteParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // path, query, header
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// RouteRegistry provides thread-safe storage for discovered routes
type RouteRegistry struct {
	routes    map[string]map[string]RouteInfo // path -> method -> RouteInfo
	mutex     sync.RWMutex
	collected bool
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]map[string]RouteInfo),
	}
}

// AddRoute adds a route to the registry
func (r *RouteRegistry) AddRoute(info RouteInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.routes[info.Path] == nil {
		r.routes[info.Path] = make(map[string]RouteInfo)
	}
	r.routes[info.Path][strings.ToLower(info.Method)] = info
}

// GetRoute retrieves a specific route
func (r *RouteRegistry) GetRoute(path, method string) (RouteInfo, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if pathRoutes, exists := r.routes[path]; exists {
		if route, exists := pathRoutes[strings.ToLower(method)]; exists {
			return route, true
		}
	}
	return RouteInfo{}, false
}

// GetAllRoutes returns all registered routes
func (r *RouteRegistry) GetAllRoutes() map[string]map[string]RouteInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Create a deep copy to avoid race conditions
	result := make(map[string]map[string]RouteInfo)
	for path, methods := range r.routes {
		result[path] = make(map[string]RouteInfo)
		for method, info := range methods {
			result[path][method] = info
		}
	}
	return result
}

// IsCollected returns whether routes have been collected
func (r *RouteRegistry) IsCollected() bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.collected
}

// SetCollected marks the registry as having collected routes
func (r *RouteRegistry) SetCollected(collected bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.collected = collected
}

// Count returns the total number of routes
func (r *RouteRegistry) Count() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	count := 0
	for _, methods := range r.routes {
		count += len(methods)
	}
	return count
}

// Generator handles OpenAPI specification generation
type Generator struct {
	app       *pocketbase.PocketBase
	config    UnifiedConfig
	discovery Discovery
	schemaGen SchemaGen
	routeGen  RouteGen
	registry  *RouteRegistry
	collector *SimpleRouteCollector
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(app *pocketbase.PocketBase, config UnifiedConfig) *Generator {
	// Initialize discovery
	discovery := NewCollectionDiscoveryWithConfig(
		app,
		config.ExcludedCollections,
		config.IncludeSystem,
	)

	// Initialize schema generator
	schemaGen := NewSchemaGeneratorWithConfig(
		config.IncludeExamples,
		config.IncludeSystem,
	)

	// Initialize route generator
	routeGen := NewRouteGeneratorWithConfig(schemaGen, config.EnableDynamicContentTypes)

	// Register custom routes
	for _, customRoute := range config.CustomRoutes {
		routeGen.RegisterCustomRoute(customRoute)
	}

	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollector(app, registry)

	return &Generator{
		app:       app,
		config:    config,
		discovery: discovery,
		schemaGen: schemaGen,
		routeGen:  routeGen,
		registry:  registry,
		collector: collector,
	}
}

// CollectRoutes triggers route collection using the collector
func (g *Generator) CollectRoutes(se *core.ServeEvent) error {
	if g.config.EnableDiscovery {
		return g.collector.CollectRoutes(se)
	}
	return nil
}

// GenerateSpec generates the unified OpenAPI specification
func (g *Generator) GenerateSpec() (*CombinedOpenAPISpec, error) {
	log.Printf("Starting OpenAPI specification generation...")

	// Step 1: Discover collections
	collections, err := g.discovery.DiscoverCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to discover collections: %w", err)
	}
	log.Printf("Discovered %d collections", len(collections))

	// Step 2: Generate schemas for all collections
	schemas, err := g.generateAllSchemas(collections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schemas: %w", err)
	}
	log.Printf("Generated %d schemas", len(schemas))

	// Step 3: Generate all routes
	routes, err := g.routeGen.GetAllRoutes(collections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes: %w", err)
	}
	log.Printf("Generated %d routes", len(routes))

	// Build specification
	spec := &CombinedOpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       g.config.Title,
			Version:     g.config.Version,
			Description: g.config.Description,
		},
		Servers: []Server{
			{
				URL:         g.config.ServerURL,
				Description: "PocketBase server",
			},
		},
		Paths:      g.buildPaths(routes),
		Components: g.buildComponents(schemas),
		Tags:       g.buildTags(collections, routes),
	}

	log.Printf("OpenAPI specification generated successfully")
	return spec, nil
}

// Global generator instance for sharing between OnServe hook and endpoints
var globalGenerator *Generator

// InitializeGenerator creates and stores a global generator instance
func InitializeGenerator(app *pocketbase.PocketBase) *Generator {
	config := DefaultUnifiedConfig()
	globalGenerator = NewGenerator(app, config)
	return globalGenerator
}

// GetGlobalGenerator returns the global generator instance
func GetGlobalGenerator() *Generator {
	return globalGenerator
}

// GenerateOpenAPI generates OpenAPI specification from PocketBase app
func GenerateOpenAPI(app *pocketbase.PocketBase) (*CombinedOpenAPISpec, error) {
	// Use global generator if available, otherwise create a new one
	var generator *Generator
	if globalGenerator != nil {
		generator = globalGenerator
	} else {
		generator = InitializeGenerator(app)
	}

	return generator.GenerateSpec()
}

// generateAllSchemas generates schemas for all collections
func (g *Generator) generateAllSchemas(collections []EnhancedCollectionInfo) (map[string]interface{}, error) {
	allSchemas := make(map[string]interface{})

	for _, collection := range collections {
		// Generate main collection schema
		schema, err := g.schemaGen.GenerateCollectionSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate schema for collection %s: %v", collection.Name, err)
			continue
		}
		allSchemas[g.schemaGen.GetSchemaName(collection)] = schema

		// Generate create schema
		createSchema, err := g.schemaGen.GenerateCreateSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate create schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[g.schemaGen.GetCreateSchemaName(collection)] = createSchema
		}

		// Generate update schema
		updateSchema, err := g.schemaGen.GenerateUpdateSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate update schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[g.schemaGen.GetUpdateSchemaName(collection)] = updateSchema
		}

		// Generate list response schema
		listResponseSchema, err := g.schemaGen.GenerateListResponseSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate list response schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[g.schemaGen.GetListResponseSchemaName(collection)] = listResponseSchema
		}
	}

	return allSchemas, nil
}

// buildPaths builds the paths section of the OpenAPI spec
func (g *Generator) buildPaths(routes []GeneratedRoute) map[string]PathItem {
	paths := make(map[string]PathItem)

	for _, route := range routes {
		if paths[route.Path] == nil {
			paths[route.Path] = make(PathItem)
		}

		operation := Operation{
			Summary:     route.Summary,
			Description: route.Description,
			Tags:        route.Tags,
			Responses:   make(map[string]Response),
			Security:    route.Security,
			Parameters:  route.Parameters,
			RequestBody: route.RequestBody,
			OperationID: route.OperationID,
		}

		// Copy responses from GeneratedRoute to Operation
		for code, resp := range route.Responses {
			operation.Responses[code] = Response{
				Description: resp.Description,
				Content:     resp.Content,
			}
		}

		paths[route.Path][strings.ToLower(route.Method)] = operation
	}

	return paths
}

// buildComponents builds the components section of the OpenAPI spec
func (g *Generator) buildComponents(schemas map[string]interface{}) *Components {
	components := &Components{
		Schemas: schemas,
	}

	// Add security schemes if auth is enabled
	if g.config.EnableAuth {
		components.SecuritySchemes = map[string]SecurityScheme{
			"BearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
				Description:  "JWT token authentication",
			},
		}
	}

	return components
}

// buildTags builds the tags section of the OpenAPI spec with controlled ordering
func (g *Generator) buildTags(collections []EnhancedCollectionInfo, routes []GeneratedRoute) []Tag {
	tagMap := make(map[string]string)
	caser := cases.Title(language.English)

	// Add collection-based tags
	for _, collection := range collections {
		tagName := caser.String(collection.Name)
		tagMap[tagName] = fmt.Sprintf("Operations for %s collection", collection.Name)
	}

	// Add standard tags
	tagMap["System"] = "System health and monitoring endpoints"
	tagMap["Custom"] = "Custom API endpoints"
	tagMap["Jobs"] = "Job management endpoints"
	tagMap["Users"] = "User management endpoints"

	if g.config.EnableAuth {
		tagMap["Authentication"] = "Authentication operations"
	}

	// Add custom tags from routes
	for _, route := range routes {
		for _, tag := range route.Tags {
			if _, exists := tagMap[tag]; !exists {
				tagMap[tag] = fmt.Sprintf("Operations for %s", tag)
			}
		}
	}

	// Define tag order - system tags first, then auth collections, then other collections
	tagOrder := []string{"System", "Authentication"}

	// Add auth collections next
	for _, collection := range collections {
		if collection.Type == "auth" {
			tagOrder = append(tagOrder, caser.String(collection.Name))
		}
	}

	// Add remaining collections
	for _, collection := range collections {
		if collection.Type != "auth" {
			tagOrder = append(tagOrder, caser.String(collection.Name))
		}
	}

	// Add any custom tags that aren't in our order
	for tag := range tagMap {
		found := false
		for _, orderedTag := range tagOrder {
			if tag == orderedTag {
				found = true
				break
			}
		}
		if !found {
			tagOrder = append(tagOrder, tag)
		}
	}

	// Convert to slice in the defined order
	var tags []Tag
	for _, tagName := range tagOrder {
		if description, exists := tagMap[tagName]; exists {
			tags = append(tags, Tag{
				Name:        tagName,
				Description: description,
			})
		}
	}

	return tags
}

// RefreshCollections refreshes the collection cache
func (g *Generator) RefreshCollections() error {
	if g.discovery != nil {
		g.discovery.RefreshCollectionCache()
	}
	return nil
}

// AddCustomRoute adds a custom route to the generator
func (g *Generator) AddCustomRoute(route CustomRoute) {
	g.config.CustomRoutes = append(g.config.CustomRoutes, route)
	if g.routeGen != nil {
		g.routeGen.RegisterCustomRoute(route)
	}
}

// GetCollectionStats returns statistics about discovered collections
func (g *Generator) GetCollectionStats() (map[string]int, error) {
	if g.discovery == nil {
		return nil, fmt.Errorf("discovery service not initialized")
	}
	return g.discovery.GetCollectionStats()
}

// ValidateConfiguration validates the generator configuration
func (g *Generator) ValidateConfiguration() error {
	if g.config.Title == "" {
		return fmt.Errorf("title is required")
	}
	if g.config.Version == "" {
		return fmt.Errorf("version is required")
	}
	if g.config.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}
	return nil
}

// GetConfiguration returns the current configuration
func (g *Generator) GetConfiguration() UnifiedConfig {
	return g.config
}

// UpdateConfiguration updates the generator configuration
func (g *Generator) UpdateConfiguration(config UnifiedConfig) error {
	if err := g.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	g.config = config

	// Reinitialize components with new config
	g.discovery = NewCollectionDiscoveryWithConfig(
		g.app,
		config.ExcludedCollections,
		config.IncludeSystem,
	)

	g.schemaGen = NewSchemaGeneratorWithConfig(
		config.IncludeExamples,
		config.IncludeSystem,
	)

	g.routeGen = NewRouteGeneratorWithConfig(g.schemaGen, config.EnableDynamicContentTypes)

	// Re-register custom routes
	for _, customRoute := range config.CustomRoutes {
		g.routeGen.RegisterCustomRoute(customRoute)
	}

	return nil
}

// validateConfig validates a configuration
func (g *Generator) validateConfig(config UnifiedConfig) error {
	if config.Title == "" {
		return fmt.Errorf("title is required")
	}
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}
	if config.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}
	return nil
}

// GetHealthStatus returns the health status of the generator
func (g *Generator) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status": "healthy",
		"components": map[string]interface{}{
			"discovery": g.discovery != nil,
			"schemaGen": g.schemaGen != nil,
			"routeGen":  g.routeGen != nil,
			"app":       g.app != nil,
		},
	}

	// Test collection access
	if g.discovery != nil {
		if err := g.discovery.ValidateCollectionAccess(); err != nil {
			status["status"] = "unhealthy"
			status["error"] = err.Error()
		}
	}

	return status
}

// DefaultUnifiedConfig returns a default configuration
func DefaultUnifiedConfig() UnifiedConfig {
	return UnifiedConfig{
		Title:                     common.GetEnv("APP_NAME", "IMS Pocketbase") + " API",
		Version:                   "1.0.0",
		Description:               "Auto-generated API documentation for PocketBase collections",
		ServerURL:                 common.GetEnv("APP_URL", "http://localhost:8090"),
		ExcludedCollections:       []string{},
		IncludeSystem:             false,
		CustomRoutes:              []CustomRoute{},
		EnableAuth:                true,
		IncludeExamples:           true,
		EnableDiscovery:           true,
		EnableDynamicContentTypes: true, // Enable dynamic content types by default
	}
}
