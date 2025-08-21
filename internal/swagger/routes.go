package swagger

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RouteGenerator handles automatic CRUD route generation for collections
type RouteGenerator struct {
	schemaGen                 SchemaGen
	customRoutes              []CustomRoute
	enableDynamicContentTypes bool
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
	Name        string `json:"name"`
	In          string `json:"in"` // "path", "query", "header"
	Required    bool   `json:"required"`
	Schema      any    `json:"schema"`
	Description string `json:"description,omitempty"`
	Example     any    `json:"example,omitempty"`
}

// RequestBody represents an OpenAPI request body
type RequestBody struct {
	Description string               `json:"description,omitempty"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// Response represents a response in OpenAPI
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// MediaType represents an OpenAPI media type
type MediaType struct {
	Schema any `json:"schema"`
}

// SecurityRequirement represents an OpenAPI security requirement
type SecurityRequirement map[string][]string

// CustomRoute represents a manually defined route
type CustomRoute struct {
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags"`
	Protected   bool        `json:"protected"`
	Parameters  []Parameter `json:"parameters,omitempty"`
}

// RouteGen interface for route generation
type RouteGen interface {
	GenerateCollectionRoutes(collection CollectionInfo) ([]GeneratedRoute, error)
	GenerateAuthRoutes(collection CollectionInfo) ([]GeneratedRoute, error)
	RegisterCustomRoute(route CustomRoute)
	GetAllRoutes(collections []CollectionInfo) ([]GeneratedRoute, error)
}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator(schemaGen SchemaGen) *RouteGenerator {

	return &RouteGenerator{
		schemaGen:                 schemaGen,
		customRoutes:              []CustomRoute{},
		enableDynamicContentTypes: true, // Default to enabled for backward compatibility
	}
}

// NewRouteGeneratorWithConfig creates a new route generator with configuration
func NewRouteGeneratorWithConfig(schemaGen SchemaGen, enableDynamicContentTypes bool) *RouteGenerator {

	return &RouteGenerator{
		schemaGen:                 schemaGen,
		customRoutes:              []CustomRoute{},
		enableDynamicContentTypes: enableDynamicContentTypes,
	}
}

// GenerateCollectionRoutes generates standard CRUD routes for a collection
func (rg *RouteGenerator) GenerateCollectionRoutes(collection CollectionInfo) ([]GeneratedRoute, error) {
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

// requiresAuthentication checks if a rule requires authentication
// Returns true if the rule contains @request.auth.id != ” or is not public
func (rg *RouteGenerator) requiresAuthentication(rule *string) bool {
	// If rule is nil, it's locked and requires auth
	if rule == nil {
		return true
	}

	// If rule is empty string, it's public and doesn't require auth
	if *rule == "" {
		return false
	}

	// Check if the rule contains the auth pattern with single or double quotes
	return strings.Contains(*rule, "@request.auth.id != ''") ||
		strings.Contains(*rule, "@request.auth.id != \"\"")
}

// generateListRoute generates a list/search route for a collection
func (rg *RouteGenerator) generateListRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	// Generate list response schema
	listResponseSchema, err := rg.schemaGen.GenerateListResponseSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate list response schema: %w", err)
	}
	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "GET",
		Path:        fmt.Sprintf("/api/collections/%s/records", collection.Name),
		Summary:     fmt.Sprintf("List %s records", collection.Name),
		Description: fmt.Sprintf("Fetch a paginated list of %s records with optional filtering and sorting", collection.Name),
		Tags:        []string{caser.String(collection.Name)},
		OperationID: fmt.Sprintf("list%s", caser.String(collection.Name)),
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

	// Add security if collection requires authentication for list operations
	if rg.requiresAuthentication(collection.ListRule) {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateCreateRoute generates a create route for a collection
func (rg *RouteGenerator) generateCreateRoute(collection CollectionInfo) (*GeneratedRoute, error) {
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

	caser := cases.Title(language.English)
	requestContent := rg.generateHybridCreateContent(createSchema, collection)

	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/records", collection.Name),
		Summary:     fmt.Sprintf("Create %s record", collection.Name),
		Description: fmt.Sprintf("Create a new %s record", collection.Name),
		Tags:        []string{caser.String(collection.Name)},
		OperationID: fmt.Sprintf("create%s", caser.String(collection.Name)),
		RequestBody: &RequestBody{
			Description: fmt.Sprintf("The %s record to create", collection.Name),
			Required:    true,
			Content:     requestContent,
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

	// Add security if collection requires authentication for create operations
	if rg.requiresAuthentication(collection.CreateRule) {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateViewRoute generates a view route for a single record
func (rg *RouteGenerator) generateViewRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema
	responseSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response schema: %w", err)
	}

	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "GET",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("View %s record", collection.Name),
		Description: fmt.Sprintf("Fetch a single %s record by ID", collection.Name),
		Tags:        []string{caser.String(collection.Name)},
		OperationID: fmt.Sprintf("view%s", caser.String(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]any{"type": "string"},
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

	// Add security if collection requires authentication for view operations
	if rg.requiresAuthentication(collection.ViewRule) {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateUpdateRoute generates an update route for a record
func (rg *RouteGenerator) generateUpdateRoute(collection CollectionInfo) (*GeneratedRoute, error) {
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

	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "PATCH",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("Update %s record", collection.Name),
		Description: fmt.Sprintf("Update an existing %s record", collection.Name),
		Tags:        []string{caser.String(collection.Name)},
		OperationID: fmt.Sprintf("update%s", caser.String(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]any{"type": "string"},
				Description: "The record ID",
				Example:     "abc123def456",
			},
		},
		RequestBody: &RequestBody{
			Description: fmt.Sprintf("The %s record fields to update", collection.Name),
			Required:    true,
			Content:     rg.generateHybridUpdateContent(updateSchema, collection),
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

	// Add security if collection requires authentication for update operations
	if rg.requiresAuthentication(collection.UpdateRule) {
		route.Security = []SecurityRequirement{
			{"BearerAuth": []string{}},
		}
	}

	return route, nil
}

// generateDeleteRoute generates a delete route for a record
func (rg *RouteGenerator) generateDeleteRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "DELETE",
		Path:        fmt.Sprintf("/api/collections/%s/records/{id}", collection.Name),
		Summary:     fmt.Sprintf("Delete %s record", collection.Name),
		Description: fmt.Sprintf("Delete an existing %s record", collection.Name),
		Tags:        []string{caser.String(collection.Name)},
		OperationID: fmt.Sprintf("delete%s", caser.String(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "id",
				In:          "path",
				Required:    true,
				Schema:      map[string]any{"type": "string"},
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

	// Add security if collection requires authentication for delete operations
	if rg.requiresAuthentication(collection.DeleteRule) {
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
			Schema:      map[string]any{"type": "integer", "minimum": 1, "default": 1},
			Description: "Page number for pagination",
			Example:     1,
		},
		{
			Name:        "perPage",
			In:          "query",
			Required:    false,
			Schema:      map[string]any{"type": "integer", "minimum": 1, "maximum": 500, "default": 30},
			Description: "Number of records per page",
			Example:     30,
		},
		{
			Name:        "sort",
			In:          "query",
			Required:    false,
			Schema:      map[string]any{"type": "string"},
			Description: "Sort records by field(s). Use '-' prefix for descending order",
			Example:     "-created",
		},
		{
			Name:        "filter",
			In:          "query",
			Required:    false,
			Schema:      map[string]any{"type": "string"},
			Description: "Filter records using PocketBase filter syntax",
			Example:     "",
		},
		{
			Name:        "expand",
			In:          "query",
			Required:    false,
			Schema:      map[string]any{"type": "string"},
			Description: "Expand relation fields",
			Example:     "",
		},
		{
			Name:        "fields",
			In:          "query",
			Required:    false,
			Schema:      map[string]any{"type": "string"},
			Description: "Comma-separated list of fields to return",
			Example:     "",
		},
	}
}

// GenerateAuthRoutes generates authentication routes for auth collections
func (rg *RouteGenerator) GenerateAuthRoutes(collection CollectionInfo) ([]GeneratedRoute, error) {
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
func (rg *RouteGenerator) generateAuthWithPasswordRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema (includes token and record)
	recordSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate record schema: %w", err)
	}

	authResponseSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token": map[string]any{
				"type":        "string",
				"description": "JWT authentication token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"record": recordSchema,
		},
		"required": []string{"token", "record"},
	}

	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/auth-with-password", collection.Name),
		Summary:     fmt.Sprintf("Authenticate %s with password", collection.Name),
		Description: fmt.Sprintf("Authenticate a %s using email/username and password", collection.Name),
		Tags:        []string{"Authentication"},
		OperationID: fmt.Sprintf("auth%sWithPassword", caser.String(collection.Name)),
		Parameters: []Parameter{
			{
				Name:        "expand",
				In:          "query",
				Description: "Auto expand record relations. Ex: `roles.permissions,permissions`",
				Required:    false,
				Schema: map[string]any{
					"type":    "string",
					"example": "roles.permissions,permissions",
				},
			},
		},
		RequestBody: &RequestBody{
			Description: "Authentication credentials",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"identity": map[string]any{
								"type":        "string",
								"description": "Email or username",
								"example":     "superadminuser@example.com",
							},
							"password": map[string]any{
								"type":        "string",
								"description": "User password",
								"example":     "superadmin123",
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
func (rg *RouteGenerator) generateAuthRefreshRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	// Generate response schema
	recordSchema, err := rg.schemaGen.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate record schema: %w", err)
	}

	authResponseSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"token": map[string]any{
				"type":        "string",
				"description": "New JWT authentication token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"record": recordSchema,
		},
		"required": []string{"token", "record"},
	}

	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/auth-refresh", collection.Name),
		Summary:     fmt.Sprintf("Refresh %s authentication", collection.Name),
		Description: fmt.Sprintf("Refresh the authentication token for %s", collection.Name),
		Tags:        []string{"Authentication"},
		OperationID: fmt.Sprintf("refresh%sAuth", caser.String(collection.Name)),
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
func (rg *RouteGenerator) generateRequestPasswordResetRoute(collection CollectionInfo) (*GeneratedRoute, error) {
	caser := cases.Title(language.English)
	route := &GeneratedRoute{
		Method:      "POST",
		Path:        fmt.Sprintf("/api/collections/%s/request-password-reset", collection.Name),
		Summary:     fmt.Sprintf("Request password reset for %s", collection.Name),
		Description: fmt.Sprintf("Send a password reset email to %s", collection.Name),
		Tags:        []string{"Authentication"},
		OperationID: fmt.Sprintf("request%sPasswordReset", caser.String(collection.Name)),
		RequestBody: &RequestBody{
			Description: "Email for password reset",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"email": map[string]any{
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
func (rg *RouteGenerator) GetAllRoutes(collections []CollectionInfo) ([]GeneratedRoute, error) {
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
		Parameters:  custom.Parameters,
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

	caser := cases.Title(language.English)
	for _, part := range parts {
		// Skip path parameters
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			continue
		}
		operationParts = append(operationParts, caser.String(part))
	}

	return strings.Join(operationParts, "")
}

// FileFieldInfo holds information about a file field
type FileFieldInfo struct {
	Name         string
	IsMultiple   bool
	MaxSize      int64
	AllowedTypes []string
	Required     bool
}

// hasFileFields checks if a collection contains any file fields
func (rg *RouteGenerator) hasFileFields(collection CollectionInfo) bool {
	for _, field := range collection.Fields {
		if strings.ToLower(field.Type) == "file" {
			return true
		}
	}
	return false
}

// getFileFields returns a list of file fields with their options
func (rg *RouteGenerator) getFileFields(collection CollectionInfo) []FileFieldInfo {
	var fileFields []FileFieldInfo

	if len(collection.Fields) == 0 {
		return fileFields
	}

	for _, field := range collection.Fields {
		if strings.ToLower(field.Type) != "file" {
			continue
		}

		fileField := FileFieldInfo{
			Name:     field.Name,
			Required: field.Required,
		}

		// Extract file-specific options with error handling
		if field.Options != nil {
			// Check for multiple file uploads
			if maxSelect, ok := field.Options["maxSelect"]; ok {
				log.Printf("Debug: Field %s in collection %s has maxSelect: %v (type: %T)", field.Name, collection.Name, maxSelect, maxSelect)
				if ms, err := rg.parseIntOption(maxSelect); err == nil && ms > 1 {
					fileField.IsMultiple = true
					log.Printf("Debug: Field %s set as multiple (maxSelect: %d)", field.Name, ms)
				} else if err != nil {
					log.Printf("Warning: Failed to parse maxSelect for field %s in collection %s: %v", field.Name, collection.Name, err)
				} else {
					log.Printf("Debug: Field %s not multiple (maxSelect: %d <= 1)", field.Name, ms)
				}
			} else {
				log.Printf("Debug: Field %s in collection %s has no maxSelect option", field.Name, collection.Name)
			}

			// Extract max file size
			if maxSize, ok := field.Options["maxSize"]; ok {
				if ms, err := rg.parseIntOption(maxSize); err == nil {
					fileField.MaxSize = int64(ms)

				} else {
					log.Printf("Warning: Failed to parse maxSize for field %s in collection %s: %v", field.Name, collection.Name, err)
				}
			}

			// Extract allowed MIME types
			if mimeTypes, ok := field.Options["mimeTypes"]; ok {
				if types, ok := mimeTypes.([]any); ok {
					for _, t := range types {
						if typeStr, ok := t.(string); ok {
							fileField.AllowedTypes = append(fileField.AllowedTypes, typeStr)
						} else {
							log.Printf("Warning: Invalid MIME type in field %s of collection %s: %v", field.Name, collection.Name, t)
						}
					}

				} else {
					log.Printf("Warning: Invalid mimeTypes format for field %s in collection %s", field.Name, collection.Name)
				}
			}
		}

		fileFields = append(fileFields, fileField)
	}

	return fileFields
}

// getStringOrDefault safely gets a string value or returns a default
func getStringOrDefault(value any, defaultValue string) string {
	if str, ok := value.(string); ok && str != "" {
		return str
	}
	return defaultValue
}

// parseIntOption safely parses an integer option value
func (rg *RouteGenerator) parseIntOption(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		var result int
		if i, err := fmt.Sscanf(v, "%d", &result); err == nil && i == 1 {
			return result, nil
		}
		return 0, fmt.Errorf("cannot parse string %s as int", v)
	default:
		return 0, fmt.Errorf("cannot parse %v as int", value)
	}
}

// isRelationExample checks if an example value looks like a relation field example
func isRelationExample(example any) bool {
	// Check for single relation example
	if str, ok := example.(string); ok && str == "RELATION_RECORD_ID" {
		return true
	}

	// Check for multi-relation example
	if arr, ok := example.([]any); ok && len(arr) == 1 {
		if str, ok := arr[0].(string); ok && str == "RELATION_RECORD_ID" {
			return true
		}
	}

	return false
}

// isRelationField checks if a field schema represents a relation field
func isRelationField(fieldSchema map[string]any) bool {
	if desc, ok := fieldSchema["description"].(string); ok {
		return strings.Contains(desc, "Related record ID") || strings.Contains(desc, "Relation field")
	}
	return false
}

// generateRequestContent generates request body content with appropriate media types
// If collection has file fields, it adds multipart/form-data support for create and update operations
func (rg *RouteGenerator) generateRequestContent(schema any, collection CollectionInfo, operation string) map[string]MediaType {
	content := make(map[string]MediaType)

	// Always include application/json
	content["application/json"] = MediaType{
		Schema: schema,
	}

	// Check if dynamic content types are enabled
	if !rg.enableDynamicContentTypes {
		return content // Return JSON-only if feature is disabled
	}

	// Only support form-data for create and update operations
	supportedOperations := []string{"create", "update"}
	isSupported := false
	for _, op := range supportedOperations {
		if strings.ToLower(operation) == op {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return content // Return JSON-only for non-create/update operations
	}

	// Check if collection has file fields using our helper function
	if !rg.hasFileFields(collection) {
		return content // Return JSON-only if no file fields
	}

	// Get file fields with their options
	fileFields := rg.getFileFields(collection)
	if len(fileFields) == 0 {
		return content // Fallback to JSON-only if detection failed
	}

	// Generate form data schema based on the original schema
	formDataProps := make(map[string]any)

	// Handle the schema to extract properties - support both map and CollectionSchema types
	var props map[string]any
	var required []string

	switch s := schema.(type) {
	case map[string]any:
		// Handle raw map schema
		if p, ok := s["properties"].(map[string]any); ok {
			props = p
		} else {
			log.Printf("Warning: Cannot parse schema properties for collection %s (operation: %s), falling back to JSON-only", collection.Name, operation)
			return content
		}
		if r, ok := s["required"].([]string); ok {
			required = r
		}
	case *CollectionSchema:
		// Handle CollectionSchema struct
		props = make(map[string]any)
		for propName, fieldSchema := range s.Properties {
			// Convert FieldSchema to map[string]any
			propMap := map[string]any{
				"type": fieldSchema.Type,
			}
			if fieldSchema.Format != "" {
				propMap["format"] = fieldSchema.Format
			}
			if fieldSchema.Description != "" {
				propMap["description"] = fieldSchema.Description
			}
			if len(fieldSchema.Enum) > 0 {
				propMap["enum"] = fieldSchema.Enum
			}
			if fieldSchema.Minimum != nil {
				propMap["minimum"] = *fieldSchema.Minimum
			}
			if fieldSchema.Maximum != nil {
				propMap["maximum"] = *fieldSchema.Maximum
			}
			if fieldSchema.MinLength != nil {
				propMap["minLength"] = *fieldSchema.MinLength
			}
			if fieldSchema.MaxLength != nil {
				propMap["maxLength"] = *fieldSchema.MaxLength
			}
			if fieldSchema.Pattern != "" {
				propMap["pattern"] = fieldSchema.Pattern
			}
			if fieldSchema.Example != nil {
				propMap["example"] = fieldSchema.Example
			}
			if fieldSchema.Items != nil {
				// Handle array items
				itemMap := map[string]any{
					"type": fieldSchema.Items.Type,
				}
				if fieldSchema.Items.Format != "" {
					itemMap["format"] = fieldSchema.Items.Format
				}
				if fieldSchema.Items.Description != "" {
					itemMap["description"] = fieldSchema.Items.Description
				}
				propMap["items"] = itemMap
			}
			props[propName] = propMap
		}
		required = s.Required
	default:
		log.Printf("Warning: Cannot parse schema for collection %s (operation: %s), schema type: %T, falling back to JSON-only", collection.Name, operation, schema)
		return content
	}

	// Create a map of file field names for quick lookup
	fileFieldMap := make(map[string]FileFieldInfo)
	for _, fileField := range fileFields {
		fileFieldMap[fileField.Name] = fileField
	}

	// Process each property
	processedFields := 0
	for propName, propValue := range props {
		if fileField, isFileField := fileFieldMap[propName]; isFileField {
			// Handle file fields with proper binary format
			if fileField.IsMultiple {
				// Multiple files - array of binary
				itemSchema := map[string]any{
					"type":        "string",
					"format":      "binary",
					"description": "Individual file upload",
				}

				// Add file constraints to items if specified
				if fileField.MaxSize > 0 {
					itemSchema["description"] = fmt.Sprintf("Individual file upload (max size: %d bytes)", fileField.MaxSize)
				}

				if len(fileField.AllowedTypes) > 0 {
					itemSchema["description"] = fmt.Sprintf("%s (allowed types: %s)",
						itemSchema["description"], strings.Join(fileField.AllowedTypes, ", "))
				}

				multipleFileSchema := map[string]any{
					"type":        "array",
					"items":       itemSchema,
					"description": fmt.Sprintf("Multiple file uploads for %s", propName),
				}

				// Add array constraints if available
				if fileField.IsMultiple {
					// We can infer maxItems from the maxSelect option
					for _, field := range collection.Fields {
						if field.Name == propName && field.Options != nil {
							if maxSelect, ok := field.Options["maxSelect"]; ok {
								if ms, err := rg.parseIntOption(maxSelect); err == nil && ms > 1 {
									multipleFileSchema["maxItems"] = ms
									multipleFileSchema["description"] = fmt.Sprintf("%s (max %d files)",
										multipleFileSchema["description"], ms)
								}
							}
						}
					}
				}

				formDataProps[propName] = multipleFileSchema
			} else {
				// Single file - binary format
				fileSchema := map[string]any{
					"type":        "string",
					"format":      "binary",
					"description": fmt.Sprintf("File upload for %s", propName),
				}

				// Add file size constraint if specified
				if fileField.MaxSize > 0 {
					fileSchema["description"] = fmt.Sprintf("File upload for %s (max size: %d bytes)",
						propName, fileField.MaxSize)
				}

				// Add allowed types if specified
				if len(fileField.AllowedTypes) > 0 {
					fileSchema["description"] = fmt.Sprintf("%s (allowed types: %s)",
						fileSchema["description"], strings.Join(fileField.AllowedTypes, ", "))
				}

				// Add example for single file
				fileSchema["example"] = fmt.Sprintf("@%s.jpg", propName)

				formDataProps[propName] = fileSchema
			}
			processedFields++
		} else {
			// Non-file fields retain their original schema but may need adjustments for form-data
			if propMap, ok := propValue.(map[string]any); ok {
				// Create a copy to avoid modifying the original schema
				formFieldSchema := make(map[string]any)
				for k, v := range propMap {
					formFieldSchema[k] = v
				}

				// For form-data, complex types should be handled as strings
				if fieldType, ok := formFieldSchema["type"].(string); ok {
					switch fieldType {
					case "object", "array":
						// Check if this is a relation field array
						if fieldType == "array" && isRelationField(formFieldSchema) {
							// For relation arrays, keep as array but add note about form-data format
							formFieldSchema["description"] = fmt.Sprintf("%s (multiple values supported)",
								getStringOrDefault(formFieldSchema["description"], fmt.Sprintf("%s field", propName)))
							// Keep the array example as-is for relation fields
						} else {
							// Complex types in form-data are typically sent as JSON strings
							formFieldSchema["type"] = "string"
							formFieldSchema["description"] = fmt.Sprintf("%s (JSON string)",
								getStringOrDefault(formFieldSchema["description"], fmt.Sprintf("%s field", propName)))

							// Preserve relation field examples, otherwise use generic example
							if existingExample, hasExample := formFieldSchema["example"]; hasExample {
								// Check if this looks like a relation field example
								if isRelationExample(existingExample) {
									// Keep the relation example as-is for form-data
									formFieldSchema["example"] = existingExample
								} else {
									formFieldSchema["example"] = `{"key": "value"}`
								}
							} else {
								formFieldSchema["example"] = `{"key": "value"}`
							}
						}
					case "boolean":
						// Booleans in form-data are typically sent as strings
						formFieldSchema["type"] = "string"
						formFieldSchema["enum"] = []any{"true", "false"}
						formFieldSchema["description"] = fmt.Sprintf("%s (boolean as string)",
							getStringOrDefault(formFieldSchema["description"], fmt.Sprintf("%s field", propName)))
						formFieldSchema["example"] = "true"
					case "integer", "number":
						// Numbers in form-data are sent as strings but we can keep the type
						// and add a note in the description
						if desc, ok := formFieldSchema["description"].(string); ok {
							formFieldSchema["description"] = fmt.Sprintf("%s (sent as string in form-data)", desc)
						} else {
							formFieldSchema["description"] = fmt.Sprintf("%s field (sent as string in form-data)", propName)
						}
					}
				}

				formDataProps[propName] = formFieldSchema
			} else {
				// Fallback: use the original property as-is
				log.Printf("Warning: Could not parse property %s for form-data, using original schema", propName)
				formDataProps[propName] = propValue
			}
			processedFields++
		}
	}

	// Create the form data schema
	formDataSchema := map[string]any{
		"type":       "object",
		"properties": formDataProps,
	}

	// Copy required fields if present
	if len(required) > 0 {
		formDataSchema["required"] = required
	}

	// Add example for form-data
	formDataExample := make(map[string]any)
	for propName, propSchema := range formDataProps {
		if propMap, ok := propSchema.(map[string]any); ok {
			if example, hasExample := propMap["example"]; hasExample {
				formDataExample[propName] = example
			} else {
				// Generate appropriate example based on type
				if propType, ok := propMap["type"].(string); ok {
					switch propType {
					case "string":
						if format, ok := propMap["format"].(string); ok && format == "binary" {
							formDataExample[propName] = fmt.Sprintf("@%s_file", propName)
						} else {
							formDataExample[propName] = fmt.Sprintf("example_%s", propName)
						}
					case "array":
						if items, ok := propMap["items"].(map[string]any); ok {
							if itemFormat, ok := items["format"].(string); ok && itemFormat == "binary" {
								formDataExample[propName] = []any{
									fmt.Sprintf("@%s_file1", propName),
									fmt.Sprintf("@%s_file2", propName),
								}
							}
						}
					default:
						formDataExample[propName] = fmt.Sprintf("example_%s", propName)
					}
				}
			}
		}
	}

	if len(formDataExample) > 0 {
		formDataSchema["example"] = formDataExample
	}

	// Validate that we have properties in the form-data schema
	if len(formDataProps) == 0 {
		log.Printf("Warning: No properties generated for form-data schema in collection %s, falling back to JSON-only", collection.Name)
		return content
	}

	content["multipart/form-data"] = MediaType{
		Schema: formDataSchema,
	}

	return content
}

// generateHybridCreateContent generates hybrid content for POST operations:
// - application/json for non-file fields only
// - multipart/form-data for file fields only
func (rg *RouteGenerator) generateHybridCreateContent(schema any, collection CollectionInfo) map[string]MediaType {
	content := make(map[string]MediaType)

	// Check if dynamic content types are enabled and collection has file fields
	if !rg.enableDynamicContentTypes || !rg.hasFileFields(collection) {
		// Fallback to standard behavior if no file fields or feature disabled
		return rg.generateRequestContent(schema, collection, "create")
	}

	// Get file fields for filtering
	fileFields := rg.getFileFields(collection)
	fileFieldMap := make(map[string]bool)
	for _, fileField := range fileFields {
		fileFieldMap[fileField.Name] = true
	}

	// Extract properties from schema
	var props map[string]any
	var required []string

	switch s := schema.(type) {
	case map[string]any:
		if p, ok := s["properties"].(map[string]any); ok {
			props = p
		}
		if r, ok := s["required"].([]string); ok {
			required = r
		}
	case *CollectionSchema:
		props = make(map[string]any)
		for propName, fieldSchema := range s.Properties {
			propMap := map[string]any{
				"type": fieldSchema.Type,
			}
			if fieldSchema.Format != "" {
				propMap["format"] = fieldSchema.Format
			}
			if fieldSchema.Description != "" {
				propMap["description"] = fieldSchema.Description
			}
			if fieldSchema.Example != nil {
				propMap["example"] = fieldSchema.Example
			}
			props[propName] = propMap
		}
		required = s.Required
	default:
		// Fallback to standard behavior if schema type is unknown
		return rg.generateRequestContent(schema, collection, "create")
	}

	// 1. Generate JSON content with NON-file fields only
	jsonProps := make(map[string]any)
	jsonRequired := []string{}

	for propName, propValue := range props {
		if !fileFieldMap[propName] {
			jsonProps[propName] = propValue
		}
	}

	for _, reqField := range required {
		if !fileFieldMap[reqField] {
			jsonRequired = append(jsonRequired, reqField)
		}
	}

	jsonSchema := map[string]any{
		"type":       "object",
		"properties": jsonProps,
	}
	if len(jsonRequired) > 0 {
		jsonSchema["required"] = jsonRequired
	}

	content["application/json"] = MediaType{
		Schema: jsonSchema,
	}

	// 2. Generate multipart content with FILE fields only
	return rg.addMultipartContent(content, fileFields)
}

// addMultipartContent adds multipart/form-data content type with file fields to existing content
func (rg *RouteGenerator) addMultipartContent(content map[string]MediaType, fileFields []FileFieldInfo) map[string]MediaType {
	multipartProps := make(map[string]any)

	for _, fileField := range fileFields {
		if fileField.IsMultiple {
			// Multiple files - array of binary with proper UI support
			itemDescription := "File upload"
			if fileField.MaxSize > 0 {
				itemDescription = fmt.Sprintf("File upload (max size: %d bytes)", fileField.MaxSize)
			}
			if len(fileField.AllowedTypes) > 0 {
				itemDescription = fmt.Sprintf("%s (allowed types: %s)", itemDescription, strings.Join(fileField.AllowedTypes, ", "))
			}

			itemSchema := map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": itemDescription,
				"title":       "File",
			}

			multipleFileSchema := map[string]any{
				"type":        "array",
				"items":       itemSchema,
				"description": fmt.Sprintf("Multiple file uploads for %s (multiple values supported)", fileField.Name),
				"title":       fmt.Sprintf("%s files", fileField.Name),
				"minItems":    0,
				"default":     []any{},
			}

			multipartProps[fileField.Name] = multipleFileSchema
		} else {
			// Single file
			singleFileSchema := map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": fmt.Sprintf("File upload for %s", fileField.Name),
			}

			if fileField.MaxSize > 0 {
				singleFileSchema["description"] = fmt.Sprintf("%s (max size: %d bytes)",
					singleFileSchema["description"], fileField.MaxSize)
			}

			if len(fileField.AllowedTypes) > 0 {
				singleFileSchema["description"] = fmt.Sprintf("%s (allowed types: %s)",
					singleFileSchema["description"], strings.Join(fileField.AllowedTypes, ", "))
			}

			multipartProps[fileField.Name] = singleFileSchema
		}
	}

	multipartSchema := map[string]any{
		"type":        "object",
		"properties":  multipartProps,
		"description": "File fields only - use this content type when uploading files",
	}

	content["multipart/form-data"] = MediaType{
		Schema: multipartSchema,
	}

	return content
}

// generateHybridUpdateContent generates hybrid content for PATCH operations:
// - application/json for non-file fields only
// - multipart/form-data for file fields only
func (rg *RouteGenerator) generateHybridUpdateContent(schema any, collection CollectionInfo) map[string]MediaType {
	content := make(map[string]MediaType)

	// Check if dynamic content types are enabled and collection has file fields
	if !rg.enableDynamicContentTypes || !rg.hasFileFields(collection) {
		// Fallback to standard behavior if no file fields or feature disabled
		return rg.generateRequestContent(schema, collection, "update")
	}

	// Get file fields for filtering
	fileFields := rg.getFileFields(collection)
	fileFieldMap := make(map[string]bool)
	for _, fileField := range fileFields {
		fileFieldMap[fileField.Name] = true
	}

	// Extract properties from schema
	var props map[string]any
	var required []string

	switch s := schema.(type) {
	case map[string]any:
		if p, ok := s["properties"].(map[string]any); ok {
			props = p
		}
		if r, ok := s["required"].([]string); ok {
			required = r
		}
	case *CollectionSchema:
		props = make(map[string]any)
		for propName, fieldSchema := range s.Properties {
			propMap := map[string]any{
				"type": fieldSchema.Type,
			}
			if fieldSchema.Format != "" {
				propMap["format"] = fieldSchema.Format
			}
			if fieldSchema.Description != "" {
				propMap["description"] = fieldSchema.Description
			}
			if fieldSchema.Example != nil {
				propMap["example"] = fieldSchema.Example
			}
			props[propName] = propMap
		}
		required = s.Required
	default:
		// Fallback to standard behavior if schema type is unknown
		return rg.generateRequestContent(schema, collection, "update")
	}

	// 1. Generate JSON content with NON-file fields only
	jsonProps := make(map[string]any)
	jsonRequired := []string{}

	for propName, propValue := range props {
		if !fileFieldMap[propName] {
			jsonProps[propName] = propValue
		}
	}

	for _, reqField := range required {
		if !fileFieldMap[reqField] {
			jsonRequired = append(jsonRequired, reqField)
		}
	}

	jsonSchema := map[string]any{
		"type":       "object",
		"properties": jsonProps,
	}
	if len(jsonRequired) > 0 {
		jsonSchema["required"] = jsonRequired
	}

	content["application/json"] = MediaType{
		Schema: jsonSchema,
	}

	// 2. Generate multipart content with FILE fields only
	return rg.addMultipartContent(content, fileFields)
}
