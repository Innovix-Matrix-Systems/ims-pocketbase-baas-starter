package swagger

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// OpenAPISpec represents the OpenAPI 3.0 specification
type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components *Components         `json:"components,omitempty"`
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
	Summary     string              `json:"summary,omitempty"`
	Description string              `json:"description,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Tags        []string            `json:"tags,omitempty"`
}

// Response represents a response in OpenAPI
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents an OpenAPI media type
type MediaType struct {
	Schema interface{} `json:"schema"`
}

// Components represents OpenAPI components
type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

// SecurityScheme represents a security scheme in OpenAPI
type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description,omitempty"`
}

// Config holds the generator configuration
type Config struct {
	Title              string
	Version            string
	Description        string
	ServerURL          string
	EnableDiscovery    bool                 // Enable automatic route discovery
	FallbackRoutes     map[string]RouteInfo // Manual route definitions
	CustomTags         map[string][]string  // Custom tags for specific paths
	AllowedCollections []string             // Collections to include in documentation (empty = all)
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

// CollectionError represents an error during route collection
type CollectionError struct {
	Phase  string // "discovery", "parsing", "registration"
	Route  string
	Method string
	Err    error
}

func (e CollectionError) Error() string {
	return fmt.Sprintf("route collection error in %s phase for %s %s: %v", e.Phase, e.Method, e.Route, e.Err)
}

// Generator handles OpenAPI specification generation
type Generator struct {
	app       *pocketbase.PocketBase
	config    Config
	registry  *RouteRegistry
	collector *SimpleRouteCollector
}

// NewGenerator creates a new OpenAPI generator
func NewGenerator(app *pocketbase.PocketBase, config Config) *Generator {
	registry := NewRouteRegistry()
	collector := NewSimpleRouteCollectorWithConfig(app, registry, config.AllowedCollections)

	return &Generator{
		app:       app,
		config:    config,
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

// GenerateSpec generates the OpenAPI specification
func (g *Generator) GenerateSpec() (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       g.config.Title,
			Version:     g.config.Version,
			Description: g.config.Description,
		},
		Paths: make(map[string]PathItem),
		Components: &Components{
			SecuritySchemes: map[string]SecurityScheme{
				"BearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "JWT token authentication",
				},
			},
		},
	}

	// Add routes from registry if available, otherwise use fallback
	if g.registry.IsCollected() && g.registry.Count() > 0 {
		// Add collected routes to spec
		for path, methods := range g.registry.GetAllRoutes() {
			pathItem := make(PathItem)
			for method, routeInfo := range methods {
				pathItem[method] = Operation{
					Summary:     routeInfo.Summary,
					Description: routeInfo.Description,
					Responses:   g.generateResponses(routeInfo.Method),
					Tags:        routeInfo.Tags,
				}
			}
			spec.Paths[path] = pathItem
		}
	} else {
		// No routes collected - this is expected if no collections exist or route discovery fails
		fmt.Printf("No routes collected from registry\n")
	}

	// Custom routes are now discovered automatically through route collection

	return spec, nil
}

// addFallbackRoutes method removed - we now only use discovered routes

// generateResponses generates default responses for a route
func (g *Generator) generateResponses(method string) map[string]Response {
	responses := map[string]Response{
		"200": {Description: "Successful response"},
	}

	// Add specific responses based on method
	switch strings.ToLower(method) {
	case "post":
		responses["201"] = Response{Description: "Resource created"}
	case "delete":
		responses["204"] = Response{Description: "Resource deleted"}
	}

	return responses
}

// Global generator instance for sharing between OnServe hook and endpoints
var globalGenerator *Generator

// InitializeGenerator creates and stores a global generator instance
func InitializeGenerator(app *pocketbase.PocketBase) *Generator {
	config := Config{
		Title:              "IMS PocketBase BaaS API",
		Version:            "1.0.0",
		Description:        "Automatically generated API documentation",
		ServerURL:          "http://localhost:8090",
		EnableDiscovery:    true,
		FallbackRoutes:     make(map[string]RouteInfo),
		CustomTags:         make(map[string][]string),
		AllowedCollections: nil, // nil means include all collections found in database
		// Example: []string{"users", "posts", "categories"} to only include specific collections
	}

	globalGenerator = NewGenerator(app, config)
	return globalGenerator
}

// GetGlobalGenerator returns the global generator instance
func GetGlobalGenerator() *Generator {
	return globalGenerator
}

// GenerateOpenAPI generates OpenAPI specification from PocketBase app
func GenerateOpenAPI(app *pocketbase.PocketBase) (*OpenAPISpec, error) {
	// Use global generator if available, otherwise create a new one
	var generator *Generator
	if globalGenerator != nil {
		generator = globalGenerator
	} else {
		generator = InitializeGenerator(app)
	}

	return generator.GenerateSpec()
}
