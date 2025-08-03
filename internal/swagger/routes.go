package swagger

import (
	"fmt"
	"strings"
)

// RouteGenerator handles automatic CRUD route generation for collections
type RouteGenerator struct {
	schemaGen    SchemaGen
	customRoutes []CustomRoute
}

// GeneratedRoute represents a complete OpenAPI route definition
type GeneratedRoute struct {
	Method      string                `json:"method"`
	Path        string                `json:"path"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Tags        []string              `json:"tags"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []SecurityRequirement `json:"security,omitempty"`
	OperationID string                `json:"operationId,omitempty"`
}

// Parameter represents an OpenAPI parameter
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"` // "path", "query", "header"
	Required    bool        `json:"required"`
	Schema      interface{} `json:"schema"`
	Description string      `json:"description,omitempty"`
	Example     interface{} `json:"example,omitempty"`
}

// RequestBody represents an OpenAPI request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// SecurityRequirement represents an OpenAPI security requirement
type SecurityRequirement map[string][]string

// CustomRoute represents a manually defined route
type CustomRoute struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Protected   bool     `json:"protected"`
}

// RouteGen interface for route generation
type RouteGen interface {
	GenerateCollectionRoutes(collection EnhancedCollectionInfo) ([]GeneratedRoute, error)
	GenerateAuthRoutes(collection EnhancedCollectionInfo) ([]GeneratedRoute, error)
	RegisterCustomRoute(route CustomRoute)
	GetAllRoutes(collections []EnhancedCollectionInfo) ([]GeneratedRoute, error)
}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator(schemaGen SchemaGen) *RouteGenerator {
	return &RouteGenerator{
		schemaGen:    schemaGen,
		customRoutes: []CustomRoute{},
	}
}

// GenerateCollectionRoutes generates standard CRUD routes for a collection
func (rg *RouteGenerator) GenerateCollectionRoutes(collection EnhancedCollectionInfo) ([]GeneratedRoute, error) {
	if collection.Name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	var routes []GeneratedRoute

	// Generate list route
	listRoute, err := rg.generateListRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate list route: %w", err)
	}
	routes = append(routes, *listRoute)

	// Generate create route
	createRoute, err := rg.generateCreateRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate create route: %w", err)
	}
	routes = append(routes, *createRoute)

	// Generate view route
	viewRoute, err := rg.generateViewRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate view route: %w", err)
	}
	routes = append(routes, *viewRoute)

	// Generate update route
	updateRoute, err := rg.generateUpdateRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate update route: %w", err)
	}
	routes = append(routes, *updateRoute)

	// Generate delete route
	deleteRoute, err := rg.generateDeleteRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate delete route: %w", err)
	}
	routes = append(routes, *deleteRoute)

	return routes, nil
}

// generateListRoute generates a list/search route for a collection
func (rg *RouteGenerator) generateListRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate list response schema
	listResponseSchema, err := rg.schemaGen.GenerateListResponseSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate list response schema: %w", err)
	}

	route := &GeneratedRoute{
		Method:      "GET",
		Path:        fmt.Sprintf("/api/collections/%s/records", collection.Name),
		Summary:     fmt.Sprintf("List %s records", collection.Name),
		Description: fmt.Sprintf("Fetch a paginated list of %s records with optional filtering and sorting", collection.Name),
		Tags:        []string{strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("list%s", strings.Title(collection.Name)),
		Parameters:  rg.generateListParameters(),
		Responses: map[string]Response{
			"200": {
				Description: "Successful response",
				Content: map[string]MediaType{
					"application/json": {
						Schema: listResponseSchema,
					},
				},
			},
			"400": {
				Description: "Bad request",
			},
			"403": {
				Description: "Forbidden",
			},
		},
	}

	// Add security if collection has list rule
	if collection.ListRule != nil && *collection.ListRule != "" {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateCreateRoute generates a create route for a collection
func (rg *RouteGenerator) generateCreateRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate create schema
	createSchema, err := rg.schemaGen.GenerateCreateSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate create schema: %w", err)
	}

	// Generate response schema (full collection schema)
	responseSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response schema: %w", err)
	}

	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/records", collection.Name),
		Summary:     fmt.Sprintf("Create %s record", collection.Name),
		Description: fmt.Sprintf("Create a new %s record", collection.Name),
		Tags:        []string{strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("create%s", strings.Title(collection.Name)),
		RequestBody: &RequestBody{
			Description: fmt.Sprintf("The %s record to create", collection.Name),
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: createSchema,
				},
			},
		},
		Responses: map[string]Response{
			"201": {
				Description: "Record created successfully",
				Content: map[string]MediaType{
					"application/json": {
						Schema: responseSchema,
					},
				},
			},
			"400": {
				Description: "Bad request",
			},
			"403": {
				Description: "Forbidden",
			},
		},
	}

	// Add security if collection has create rule
	if collection.CreateRule != nil && *collection.CreateRule != "" {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateViewRoute generates a view route for a single record
func (rg *RouteGenerator) generateViewRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema
	responseSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response schema: %w", err)
	}

	route := &GeneratedRoute{
		Method:      "GET",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("View %s record", collection.Name),
		Description: fmt.Sprintf("Fetch a single %s record by ID", collection.Name),
		Tags:        []string{strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("view%s", strings.Title(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]interface{}{"type": "string"},
				Description: "The record ID",
				Example:     "abc123def456",
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Successful response",
				Content: map[string]MediaType{
					"application/json": {
						Schema: responseSchema,
					},
				},
			},
			"404": {
				Description: "Record not found",
			},
			"403": {
				Description: "Forbidden",
			},
		},
	}

	// Add security if collection has view rule
	if collection.ViewRule != nil && *collection.ViewRule != "" {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateUpdateRoute generates an update route for a record
func (rg *RouteGenerator) generateUpdateRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate update schema
	updateSchema, err := rg.schemaGen.GenerateUpdateSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate update schema: %w", err)
	}

	// Generate response schema
	responseSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response schema: %w", err)
	}

	route := &GeneratedRoute{
		Method:      "PATCH",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("Update %s record", collection.Name),
		Description: fmt.Sprintf("Update an existing %s record", collection.Name),
		Tags:        []string{strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("update%s", strings.Title(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]interface{}{"type": "string"},
				Description: "The record ID",
				Example:     "abc123def456",
			},
		},
		RequestBody: &RequestBody{
			Description: fmt.Sprintf("The %s record fields to update", collection.Name),
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: updateSchema,
				},
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Record updated successfully",
				Content: map[string]MediaType{
					"application/json": {
						Schema: responseSchema,
					},
				},
			},
			"400": {
				Description: "Bad request",
			},
			"404": {
				Description: "Record not found",
			},
			"403": {
				Description: "Forbidden",
			},
		},
	}

	// Add security if collection has update rule
	if collection.UpdateRule != nil && *collection.UpdateRule != "" {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateDeleteRoute generates a delete route for a record
func (rg *RouteGenerator) generateDeleteRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	route := &GeneratedRoute{
		Method:      "DELETE",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("Delete %s record", collection.Name),
		Description: fmt.Sprintf("Delete an existing %s record", collection.Name),
		Tags:        []string{strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("delete%s", strings.Title(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]interface{}{"type": "string"},
				Description: "The record ID",
				Example:     "abc123def456",
			},
		},
		Responses: map[string]Response{
			"204": {
				Description: "Record deleted successfully",
			},
			"404": {
				Description: "Record not found",
			},
			"403": {
				Description: "Forbidden",
			},
		},
	}

	// Add security if collection has delete rule
	if collection.DeleteRule != nil && *collection.DeleteRule != "" {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateListParameters generates common query parameters for list operations
func (rg *RouteGenerator) generateListParameters() []Parameter {
	return []Parameter{
		{
			Name:        "page",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "integer", "minimum": 1, "default": 1},
			Description: "Page number for pagination",
			Example:     1,
		},
		{
			Name:        "perPage",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "integer", "minimum": 1, "maximum": 500, "default": 30},
			Description: "Number of records per page",
			Example:     30,
		},
		{
			Name:        "sort",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "string"},
			Description: "Sort records by field(s). Use '-' prefix for descending order",
			Example:     "-created",
		},
		{
			Name:        "filter",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "string"},
			Description: "Filter records using PocketBase filter syntax",
			Example:     "",
		},
		{
			Name:        "expand",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "string"},
			Description: "Expand relation fields",
			Example:     "",
		},
		{
			Name:        "fields",
			In:          "query",
			Required:    false,
			Schema:      map[string]interface{}{"type": "string"},
			Description: "Comma-separated list of fields to return",
			Example:     "",
		},
	}
}

// GenerateAuthRoutes generates authentication routes for auth collections
func (rg *RouteGenerator) GenerateAuthRoutes(collection EnhancedCollectionInfo) ([]GeneratedRoute, error) {
	if collection.Type != "auth" {
		return []GeneratedRoute{}, nil // No auth routes for non-auth collections
	}

	var routes []GeneratedRoute

	// Generate auth-with-password route
	authRoute, err := rg.generateAuthWithPasswordRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth-with-password route: %w", err)
	}
	routes = append(routes, *authRoute)

	// Generate auth-refresh route
	refreshRoute, err := rg.generateAuthRefreshRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth-refresh route: %w", err)
	}
	routes = append(routes, *refreshRoute)

	// Generate request-password-reset route
	resetRoute, err := rg.generateRequestPasswordResetRoute(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request-password-reset route: %w", err)
	}
	routes = append(routes, *resetRoute)

	return routes, nil
}

// generateAuthWithPasswordRoute generates the auth-with-password route
func (rg *RouteGenerator) generateAuthWithPasswordRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema (includes token and record)
	recordSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate record schema: %w", err)
	}

	authResponseSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"token": map[string]interface{}{
				"type":        "string",
				"description": "JWT authentication token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"record": recordSchema,
		},
		"required": []string{"token", "record"},
	}

	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/auth-with-password", collection.Name),
		Summary:     fmt.Sprintf("Authenticate %s with password", collection.Name),
		Description: fmt.Sprintf("Authenticate a %s using email/username and password", collection.Name),
		Tags:        []string{"Authentication", strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("auth%sWithPassword", strings.Title(collection.Name)),
		RequestBody: &RequestBody{
			Description: "Authentication credentials",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"identity": map[string]interface{}{
								"type":        "string",
								"description": "Email or username",
								"example":     "user@example.com",
							},
							"password": map[string]interface{}{
								"type":        "string",
								"description": "User password",
								"example":     "password123",
							},
						},
						"required": []string{"identity", "password"},
					},
				},
			},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Authentication successful",
				Content: map[string]MediaType{
					"application/json": {
						Schema: authResponseSchema,
					},
				},
			},
			"400": {
				Description: "Bad request",
			},
			"401": {
				Description: "Authentication failed",
			},
		},
	}

	return route, nil
}

// generateAuthRefreshRoute generates the auth-refresh route
func (rg *RouteGenerator) generateAuthRefreshRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema
	recordSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate record schema: %w", err)
	}

	authResponseSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"token": map[string]interface{}{
				"type":        "string",
				"description": "New JWT authentication token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"record": recordSchema,
		},
		"required": []string{"token", "record"},
	}

	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/auth-refresh", collection.Name),
		Summary:     fmt.Sprintf("Refresh %s authentication", collection.Name),
		Description: fmt.Sprintf("Refresh the authentication token for %s", collection.Name),
		Tags:        []string{"Authentication", strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("refresh%sAuth", strings.Title(collection.Name)),
		Security: []SecurityRequirement{
			{"BearerAuth": []string{}},
		},
		Responses: map[string]Response{
			"200": {
				Description: "Token refreshed successfully",
				Content: map[string]MediaType{
					"application/json": {
						Schema: authResponseSchema,
					},
				},
			},
			"401": {
				Description: "Invalid or expired token",
			},
		},
	}

	return route, nil
}

// generateRequestPasswordResetRoute generates the request-password-reset route
func (rg *RouteGenerator) generateRequestPasswordResetRoute(collection EnhancedCollectionInfo) (*GeneratedRoute, error) {
	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/request-password-reset", collection.Name),
		Summary:     fmt.Sprintf("Request password reset for %s", collection.Name),
		Description: fmt.Sprintf("Send a password reset email to %s", collection.Name),
		Tags:        []string{"Authentication", strings.Title(collection.Name)},
		OperationID: fmt.Sprintf("request%sPasswordReset", strings.Title(collection.Name)),
		RequestBody: &RequestBody{
			Description: "Email for password reset",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"email": map[string]interface{}{
								"type":        "string",
								"format":      "email",
								"description": "Email address for password reset",
								"example":     "user@example.com",
							},
						},
						"required": []string{"email"},
					},
				},
			},
		},
		Responses: map[string]Response{
			"204": {
				Description: "Password reset email sent (if email exists)",
			},
			"400": {
				Description: "Bad request",
			},
		},
	}

	return route, nil
}

// RegisterCustomRoute registers a custom route
func (rg *RouteGenerator) RegisterCustomRoute(route CustomRoute) {
	rg.customRoutes = append(rg.customRoutes, route)
}

// GetAllRoutes generates all routes for the given collections plus custom routes
func (rg *RouteGenerator) GetAllRoutes(collections []EnhancedCollectionInfo) ([]GeneratedRoute, error) {
	var allRoutes []GeneratedRoute

	// Generate collection routes
	for _, collection := range collections {
		// Generate CRUD routes
		crudRoutes, err := rg.GenerateCollectionRoutes(collection)
		if err != nil {
			return nil, fmt.Errorf("failed to generate CRUD routes for collection %s: %w", collection.Name, err)
		}
		allRoutes = append(allRoutes, crudRoutes...)

		// Generate auth routes if it's an auth collection
		if collection.Type == "auth" {
			authRoutes, err := rg.GenerateAuthRoutes(collection)
			if err != nil {
				return nil, fmt.Errorf("failed to generate auth routes for collection %s: %w", collection.Name, err)
			}
			allRoutes = append(allRoutes, authRoutes...)
		}
	}

	// Add custom routes
	for _, customRoute := range rg.customRoutes {
		generatedRoute := rg.convertCustomRoute(customRoute)
		allRoutes = append(allRoutes, generatedRoute)
	}

	return allRoutes, nil
}

// convertCustomRoute converts a CustomRoute to a GeneratedRoute
func (rg *RouteGenerator) convertCustomRoute(custom CustomRoute) GeneratedRoute {
	route := GeneratedRoute{
		Method:      custom.Method,
		Path:        custom.Path,
		Summary:     custom.Summary,
		Description: custom.Description,
		Tags:        custom.Tags,
		OperationID: rg.generateOperationID(custom.Method, custom.Path),
		Responses: map[string]Response{
			"200": {
				Description: "Successful response",
			},
		},
	}

	// Add security if protected
	if custom.Protected {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route
}

// generateOperationID generates an operation ID from method and path
func (rg *RouteGenerator) generateOperationID(method, path string) string {
	// Convert path to camelCase operation ID
	// e.g., "GET /api/v1/hello" -> "getApiV1Hello"
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var operationParts []string

	operationParts = append(operationParts, strings.ToLower(method))

	for _, part := range parts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		operationParts = append(operationParts, strings.Title(part))
	}

	return strings.Join(operationParts, "")
}
