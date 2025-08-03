package swagger

import (
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
)

// EnhancedGenerator orchestrates all components to generate complete OpenAPI specifications
type EnhancedGenerator struct {
	app       *pocketbase.PocketBase
	config    EnhancedConfig
	discovery Discovery
	schemaGen SchemaGen
	routeGen  RouteGen
}

// EnhancedConfig holds configuration for the enhanced generator
type EnhancedConfig struct {
	Title               string
	Version             string
	Description         string
	ServerURL           string
	ExcludedCollections []string
	IncludeSystem       bool
	CustomRoutes        []CustomRoute
	EnableAuth          bool
	IncludeExamples     bool
}

// EnhancedOpenAPISpec represents the complete OpenAPI 3.0 specification
type EnhancedOpenAPISpec struct {
	OpenAPI    string                            `json:"openapi"`
	Info       EnhancedInfo                      `json:"info"`
	Servers    []EnhancedServer                  `json:"servers,omitempty"`
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components *EnhancedComponents               `json:"components,omitempty"`
	Tags       []EnhancedTag                     `json:"tags,omitempty"`
}

// EnhancedInfo represents OpenAPI info object
type EnhancedInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// EnhancedServer represents OpenAPI server object
type EnhancedServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// EnhancedComponents represents OpenAPI components
type EnhancedComponents struct {
	Schemas         map[string]interface{}    `json:"schemas,omitempty"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// EnhancedTag represents OpenAPI tag object
type EnhancedTag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// EnhancedOpenAPIGenerator interface for enhanced OpenAPI generation
type EnhancedOpenAPIGenerator interface {
	GenerateSpec() (*EnhancedOpenAPISpec, error)
	RefreshCollections() error
	AddCustomRoute(route CustomRoute)
	GetCollectionStats() (map[string]int, error)
}

// NewEnhancedGenerator creates a new enhanced OpenAPI generator
func NewEnhancedGenerator(app *pocketbase.PocketBase, config EnhancedConfig) *EnhancedGenerator {
	// Initialize discovery service
	discovery := NewCollectionDiscoveryWithConfig(
		app,
		[]string{}, // Empty allowed collections - we'll use excluded instead
		config.ExcludedCollections,
		config.IncludeSystem,
	)

	// Initialize schema generator
	schemaGen := NewSchemaGeneratorWithConfig(
		config.IncludeExamples,
		config.IncludeSystem,
	)

	// Initialize route generator
	routeGen := NewRouteGenerator(schemaGen)

	// Register custom routes
	for _, customRoute := range config.CustomRoutes {
		routeGen.RegisterCustomRoute(customRoute)
	}

	return &EnhancedGenerator{
		app:       app,
		config:    config,
		discovery: discovery,
		schemaGen: schemaGen,
		routeGen:  routeGen,
	}
}

// DefaultEnhancedConfig returns a default configuration
func DefaultEnhancedConfig() EnhancedConfig {
	return EnhancedConfig{
		Title:               "PocketBase API",
		Version:             "1.0.0",
		Description:         "Auto-generated API documentation for PocketBase collections",
		ServerURL:           "http://localhost:8090",
		ExcludedCollections: []string{}, // Empty means exclude none
		IncludeSystem:       false,      // Don't include system collections by default
		CustomRoutes:        []CustomRoute{},
		EnableAuth:          true,
		IncludeExamples:     true,
	}
}

// GenerateSpec generates the complete OpenAPI specification
func (eg *EnhancedGenerator) GenerateSpec() (*EnhancedOpenAPISpec, error) {
	log.Printf("Starting enhanced OpenAPI specification generation...")

	// Step 1: Discover collections
	collections, err := eg.discovery.DiscoverCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to discover collections: %w", err)
	}
	log.Printf("Discovered %d collections", len(collections))

	// Step 2: Generate schemas for all collections
	schemas, err := eg.generateAllSchemas(collections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate schemas: %w", err)
	}
	log.Printf("Generated %d schemas", len(schemas))

	// Step 3: Generate all routes
	routes, err := eg.routeGen.GetAllRoutes(collections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes: %w", err)
	}
	log.Printf("Generated %d routes", len(routes))

	// Step 4: Build the OpenAPI specification
	spec := &EnhancedOpenAPISpec{
		OpenAPI: "3.0.0",
		Info: EnhancedInfo{
			Title:       eg.config.Title,
			Version:     eg.config.Version,
			Description: eg.config.Description,
		},
		Servers: []EnhancedServer{
			{
				URL:         eg.config.ServerURL,
				Description: "PocketBase server",
			},
		},
		Paths:      eg.buildPaths(routes),
		Components: eg.buildComponents(schemas),
		Tags:       eg.buildTags(collections, routes),
	}

	log.Printf("Enhanced OpenAPI specification generated successfully")
	return spec, nil
}

// generateAllSchemas generates schemas for all collections
func (eg *EnhancedGenerator) generateAllSchemas(collections []EnhancedCollectionInfo) (map[string]interface{}, error) {
	allSchemas := make(map[string]interface{})

	for _, collection := range collections {
		// Generate main collection schema
		schema, err := eg.schemaGen.GenerateCollectionSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate schema for collection %s: %v", collection.Name, err)
			continue
		}
		allSchemas[eg.schemaGen.GetSchemaName(collection)] = schema

		// Generate create schema
		createSchema, err := eg.schemaGen.GenerateCreateSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate create schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[eg.schemaGen.GetCreateSchemaName(collection)] = createSchema
		}

		// Generate update schema
		updateSchema, err := eg.schemaGen.GenerateUpdateSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate update schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[eg.schemaGen.GetUpdateSchemaName(collection)] = updateSchema
		}

		// Generate list response schema
		listResponseSchema, err := eg.schemaGen.GenerateListResponseSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate list response schema for collection %s: %v", collection.Name, err)
		} else {
			allSchemas[eg.schemaGen.GetListResponseSchemaName(collection)] = listResponseSchema
		}
	}

	return allSchemas, nil
}

// buildPaths builds the paths section of the OpenAPI spec
func (eg *EnhancedGenerator) buildPaths(routes []GeneratedRoute) map[string]map[string]interface{} {
	paths := make(map[string]map[string]interface{})

	for _, route := range routes {
		if paths[route.Path] == nil {
			paths[route.Path] = make(map[string]interface{})
		}

		operation := map[string]interface{}{
			"summary":     route.Summary,
			"description": route.Description,
			"tags":        route.Tags,
			"responses":   route.Responses,
		}

		if route.OperationID != "" {
			operation["operationId"] = route.OperationID
		}

		if len(route.Parameters) > 0 {
			operation["parameters"] = route.Parameters
		}

		if route.RequestBody != nil {
			operation["requestBody"] = route.RequestBody
		}

		if len(route.Security) > 0 {
			operation["security"] = route.Security
		}

		paths[route.Path][strings.ToLower(route.Method)] = operation
	}

	return paths
}

// buildComponents builds the components section of the OpenAPI spec
func (eg *EnhancedGenerator) buildComponents(schemas map[string]interface{}) *EnhancedComponents {
	components := &EnhancedComponents{
		Schemas: schemas,
	}

	// Add security schemes if auth is enabled
	if eg.config.EnableAuth {
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

// buildTags builds the tags section of the OpenAPI spec
func (eg *EnhancedGenerator) buildTags(collections []EnhancedCollectionInfo, routes []GeneratedRoute) []EnhancedTag {
	tagMap := make(map[string]string)

	// Add collection-based tags
	for _, collection := range collections {
		tagName := strings.Title(collection.Name)
		tagMap[tagName] = fmt.Sprintf("Operations for %s collection", collection.Name)
	}

	// Add standard tags
	// Removed hardcoded "Collections" tag as it was confusing
	if eg.config.EnableAuth {
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

	// Convert to slice
	var tags []EnhancedTag
	for name, description := range tagMap {
		tags = append(tags, EnhancedTag{
			Name:        name,
			Description: description,
		})
	}

	return tags
}

// RefreshCollections refreshes the collection cache
func (eg *EnhancedGenerator) RefreshCollections() error {
	if eg.discovery != nil {
		eg.discovery.RefreshCollectionCache()
	}
	return nil
}

// AddCustomRoute adds a custom route to the generator
func (eg *EnhancedGenerator) AddCustomRoute(route CustomRoute) {
	eg.config.CustomRoutes = append(eg.config.CustomRoutes, route)
	if eg.routeGen != nil {
		eg.routeGen.RegisterCustomRoute(route)
	}
}

// GetCollectionStats returns statistics about discovered collections
func (eg *EnhancedGenerator) GetCollectionStats() (map[string]int, error) {
	if eg.discovery == nil {
		return nil, fmt.Errorf("discovery service not initialized")
	}
	return eg.discovery.GetCollectionStats()
}

// ValidateConfiguration validates the generator configuration
func (eg *EnhancedGenerator) ValidateConfiguration() error {
	if eg.config.Title == "" {
		return fmt.Errorf("title is required")
	}
	if eg.config.Version == "" {
		return fmt.Errorf("version is required")
	}
	if eg.config.ServerURL == "" {
		return fmt.Errorf("server URL is required")
	}
	return nil
}

// GetConfiguration returns the current configuration
func (eg *EnhancedGenerator) GetConfiguration() EnhancedConfig {
	return eg.config
}

// UpdateConfiguration updates the generator configuration
func (eg *EnhancedGenerator) UpdateConfiguration(config EnhancedConfig) error {
	if err := eg.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	eg.config = config

	// Reinitialize components with new config
	eg.discovery = NewCollectionDiscoveryWithConfig(
		eg.app,
		[]string{}, // Empty allowed collections - we'll use excluded instead
		config.ExcludedCollections,
		config.IncludeSystem,
	)

	eg.schemaGen = NewSchemaGeneratorWithConfig(
		config.IncludeExamples,
		config.IncludeSystem,
	)

	eg.routeGen = NewRouteGenerator(eg.schemaGen)

	// Re-register custom routes
	for _, customRoute := range config.CustomRoutes {
		eg.routeGen.RegisterCustomRoute(customRoute)
	}

	return nil
}

// validateConfig validates a configuration
func (eg *EnhancedGenerator) validateConfig(config EnhancedConfig) error {
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
func (eg *EnhancedGenerator) GetHealthStatus() map[string]interface{} {
	status := map[string]interface{}{
		"status": "healthy",
		"components": map[string]interface{}{
			"discovery": eg.discovery != nil,
			"schemaGen": eg.schemaGen != nil,
			"routeGen":  eg.routeGen != nil,
			"app":       eg.app != nil,
		},
	}

	// Test collection access
	if eg.discovery != nil {
		if err := eg.discovery.ValidateCollectionAccess(); err != nil {
			status["status"] = "unhealthy"
			status["error"] = err.Error()
		}
	}

	return status
}
