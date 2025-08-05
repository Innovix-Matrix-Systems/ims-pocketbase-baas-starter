package swagger

import (
	"strings"
	"testing"
)

// Mock schema generator for testing
type mockSchemaGen struct{}

func (m *mockSchemaGen) GenerateCollectionSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"id":   {Type: "string", Required: true},
			"name": {Type: "string", Required: true},
		},
		Required: []string{"id", "name"},
	}, nil
}

func (m *mockSchemaGen) GenerateCollectionSchemas(collections []EnhancedCollectionInfo) (map[string]*CollectionSchema, error) {
	schemas := make(map[string]*CollectionSchema)
	for _, collection := range collections {
		schema, err := m.GenerateCollectionSchema(collection)
		if err != nil {
			return nil, err
		}
		schemas[collection.Name] = schema
	}
	return schemas, nil
}

func (m *mockSchemaGen) GenerateCreateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"name": {Type: "string", Required: true},
		},
		Required: []string{"name"},
	}, nil
}

func (m *mockSchemaGen) GenerateUpdateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"name": {Type: "string", Required: false},
		},
		Required: []string{},
	}, nil
}

func (m *mockSchemaGen) GenerateListResponseSchema(collection EnhancedCollectionInfo) (map[string]any, error) {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"page":       map[string]any{"type": "integer"},
			"perPage":    map[string]any{"type": "integer"},
			"totalItems": map[string]any{"type": "integer"},
			"totalPages": map[string]any{"type": "integer"},
			"items": map[string]any{
				"type": "array",
				"items": &CollectionSchema{
					Type: "object",
					Properties: map[string]*FieldSchema{
						"id":   {Type: "string"},
						"name": {Type: "string"},
					},
				},
			},
		},
	}, nil
}

func (m *mockSchemaGen) GetSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name
}

func (m *mockSchemaGen) GetCreateSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name + "Create"
}

func (m *mockSchemaGen) GetUpdateSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name + "Update"
}

func (m *mockSchemaGen) GetListResponseSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name + "ListResponse"
}

func TestNewRouteGenerator(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.schemaGen == nil {
		t.Error("Expected schemaGen to be set")
	}

	if generator.customRoutes == nil {
		t.Error("Expected customRoutes to be initialized")
	}

	if len(generator.customRoutes) != 0 {
		t.Error("Expected customRoutes to be empty initially")
	}
}

func TestGenerateCollectionRoutes(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "name", Type: "text", Required: true},
		},
	}

	routes, err := generator.GenerateCollectionRoutes(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(routes) != 5 {
		t.Errorf("Expected 5 CRUD routes, got %d", len(routes))
	}

	// Check that we have all CRUD operations
	expectedMethods := []string{"GET", "POST", "GET", "PATCH", "DELETE"}
	expectedPaths := []string{
		"/api/collections/users/records",
		"/api/collections/users/records",
		"/api/collections/users/records/{id}",
		"/api/collections/users/records/{id}",
		"/api/collections/users/records/{id}",
	}

	for i, route := range routes {
		if route.Method != expectedMethods[i] {
			t.Errorf("Expected method %s at index %d, got %s", expectedMethods[i], i, route.Method)
		}

		if route.Path != expectedPaths[i] {
			t.Errorf("Expected path %s at index %d, got %s", expectedPaths[i], i, route.Path)
		}

		if len(route.Tags) == 0 {
			t.Errorf("Expected tags to be set for route %d", i)
		}

		if route.OperationID == "" {
			t.Errorf("Expected operationId to be set for route %d", i)
		}

		if len(route.Responses) == 0 {
			t.Errorf("Expected responses to be set for route %d", i)
		}
	}
}

func TestGenerateCollectionRoutesEmptyName(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name: "",
		Type: "base",
	}

	routes, err := generator.GenerateCollectionRoutes(collection)
	if err == nil {
		t.Error("Expected error for empty collection name, got nil")
	}

	if routes != nil {
		t.Error("Expected nil routes for empty collection name")
	}
}

func TestGenerateListRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name:     "posts",
		Type:     "base",
		ListRule: stringPtr("id != ''"),
	}

	route, err := generator.generateListRoute(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if route.Method != "GET" {
		t.Errorf("Expected method GET, got %s", route.Method)
	}

	if route.Path != "/api/collections/posts/records" {
		t.Errorf("Expected path '/api/collections/posts/records', got %s", route.Path)
	}

	if !strings.Contains(route.Summary, "posts") {
		t.Error("Expected summary to contain collection name")
	}

	if len(route.Parameters) == 0 {
		t.Error("Expected list parameters to be set")
	}

	// Check that security is added when list rule exists
	if len(route.Security) == 0 {
		t.Error("Expected security to be set when list rule exists")
	}

	// Check for pagination parameters
	paramNames := make(map[string]bool)
	for _, param := range route.Parameters {
		paramNames[param.Name] = true
	}

	expectedParams := []string{"page", "perPage", "sort", "filter", "expand", "fields"}
	for _, expectedParam := range expectedParams {
		if !paramNames[expectedParam] {
			t.Errorf("Expected parameter %s to be present", expectedParam)
		}
	}
}

func TestGenerateCreateRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name:       "posts",
		Type:       "base",
		CreateRule: stringPtr("@request.auth.id != ''"),
	}

	route, err := generator.generateCreateRoute(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if route.Method != "POST" {
		t.Errorf("Expected method POST, got %s", route.Method)
	}

	if route.RequestBody == nil {
		t.Fatal("Expected request body to be set")
	}

	if !route.RequestBody.Required {
		t.Error("Expected request body to be required")
	}

	if route.RequestBody.Content == nil {
		t.Error("Expected request body content to be set")
	}

	// Check that security is added when create rule exists
	if len(route.Security) == 0 {
		t.Error("Expected security to be set when create rule exists")
	}

	// Check response status codes
	if _, exists := route.Responses["201"]; !exists {
		t.Error("Expected 201 response for create operation")
	}
}

func TestGenerateViewRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name:     "posts",
		Type:     "base",
		ViewRule: stringPtr("id != ''"),
	}

	route, err := generator.generateViewRoute(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if route.Method != "GET" {
		t.Errorf("Expected method GET, got %s", route.Method)
	}

	if !strings.Contains(route.Path, "{id}") {
		t.Error("Expected path to contain {id} parameter")
	}

	if len(route.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(route.Parameters))
	}

	if route.Parameters[0].Name != "id" {
		t.Errorf("Expected parameter name 'id', got %s", route.Parameters[0].Name)
	}

	if route.Parameters[0].In != "path" {
		t.Errorf("Expected parameter in 'path', got %s", route.Parameters[0].In)
	}

	if !route.Parameters[0].Required {
		t.Error("Expected path parameter to be required")
	}

	// Check that security is added when view rule exists
	if len(route.Security) == 0 {
		t.Error("Expected security to be set when view rule exists")
	}
}

func TestGenerateUpdateRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name:       "posts",
		Type:       "base",
		UpdateRule: stringPtr("@request.auth.id != ''"),
	}

	route, err := generator.generateUpdateRoute(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if route.Method != "PATCH" {
		t.Errorf("Expected method PATCH, got %s", route.Method)
	}

	if route.RequestBody == nil {
		t.Fatal("Expected request body to be set")
	}

	if len(route.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(route.Parameters))
	}

	// Check that security is added when update rule exists
	if len(route.Security) == 0 {
		t.Error("Expected security to be set when update rule exists")
	}

	// Check response status codes
	if _, exists := route.Responses["200"]; !exists {
		t.Error("Expected 200 response for update operation")
	}
}

func TestGenerateDeleteRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	collection := EnhancedCollectionInfo{
		Name:       "posts",
		Type:       "base",
		DeleteRule: stringPtr("@request.auth.id != ''"),
	}

	route, err := generator.generateDeleteRoute(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if route.Method != "DELETE" {
		t.Errorf("Expected method DELETE, got %s", route.Method)
	}

	if route.RequestBody != nil {
		t.Error("Expected no request body for delete operation")
	}

	if len(route.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(route.Parameters))
	}

	// Check that security is added when delete rule exists
	if len(route.Security) == 0 {
		t.Error("Expected security to be set when delete rule exists")
	}

	// Check response status codes
	if _, exists := route.Responses["204"]; !exists {
		t.Error("Expected 204 response for delete operation")
	}
}

func TestGenerateAuthRoutes(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	// Test with auth collection
	authCollection := EnhancedCollectionInfo{
		Name: "users",
		Type: "auth",
	}

	routes, err := generator.GenerateAuthRoutes(authCollection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(routes) != 3 {
		t.Errorf("Expected 3 auth routes, got %d", len(routes))
	}

	// Check auth routes
	expectedPaths := []string{
		"/api/collections/users/auth-with-password",
		"/api/collections/users/auth-refresh",
		"/api/collections/users/request-password-reset",
	}

	for i, route := range routes {
		if route.Path != expectedPaths[i] {
			t.Errorf("Expected path %s at index %d, got %s", expectedPaths[i], i, route.Path)
		}

		if route.Method != "POST" {
			t.Errorf("Expected method POST for auth route %d, got %s", i, route.Method)
		}

		if len(route.Tags) == 0 {
			t.Errorf("Expected tags to be set for auth route %d", i)
		}
	}

	// Test with non-auth collection
	baseCollection := EnhancedCollectionInfo{
		Name: "posts",
		Type: "base",
	}

	routes, err = generator.GenerateAuthRoutes(baseCollection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(routes) != 0 {
		t.Errorf("Expected 0 auth routes for base collection, got %d", len(routes))
	}
}

func TestRegisterCustomRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	customRoute := CustomRoute{
		Method:      "GET",
		Path:        "/api/v1/hello",
		Summary:     "Hello World",
		Description: "Returns a greeting",
		Tags:        []string{"Custom"},
		Protected:   false,
	}

	generator.RegisterCustomRoute(customRoute)

	if len(generator.customRoutes) != 1 {
		t.Errorf("Expected 1 custom route, got %d", len(generator.customRoutes))
	}

	if generator.customRoutes[0].Path != "/api/v1/hello" {
		t.Errorf("Expected custom route path '/api/v1/hello', got %s", generator.customRoutes[0].Path)
	}
}

func TestGetAllRoutes(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	// Add a custom route
	customRoute := CustomRoute{
		Method:      "GET",
		Path:        "/api/v1/hello",
		Summary:     "Hello World",
		Description: "Returns a greeting",
		Tags:        []string{"Custom"},
		Protected:   false,
	}
	generator.RegisterCustomRoute(customRoute)

	collections := []EnhancedCollectionInfo{
		{
			Name: "posts",
			Type: "base",
		},
		{
			Name: "users",
			Type: "auth",
		},
	}

	routes, err := generator.GetAllRoutes(collections)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Expected: 5 CRUD routes for posts + 5 CRUD routes for users + 3 auth routes for users + 1 custom route = 14 routes
	expectedRouteCount := 5 + 5 + 3 + 1
	if len(routes) != expectedRouteCount {
		t.Errorf("Expected %d total routes, got %d", expectedRouteCount, len(routes))
	}

	// Check that custom route is included
	customRouteFound := false
	for _, route := range routes {
		if route.Path == "/api/v1/hello" {
			customRouteFound = true
			break
		}
	}

	if !customRouteFound {
		t.Error("Expected custom route to be included in all routes")
	}
}

func TestGenerateOperationID(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			method:   "GET",
			path:     "/api/v1/hello",
			expected: "getApiV1Hello",
		},
		{
			name:     "path with parameter",
			method:   "POST",
			path:     "/api/collections/{collection}/records",
			expected: "postApiCollectionsRecords",
		},
		{
			name:     "path with multiple parameters",
			method:   "DELETE",
			path:     "/api/collections/{collection}/records/{id}",
			expected: "deleteApiCollectionsRecords",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.generateOperationID(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestConvertCustomRoute(t *testing.T) {
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGenerator(schemaGen)

	customRoute := CustomRoute{
		Method:      "GET",
		Path:        "/api/v1/hello",
		Summary:     "Hello World",
		Description: "Returns a greeting",
		Tags:        []string{"Custom"},
		Protected:   true,
	}

	generatedRoute := generator.convertCustomRoute(customRoute)

	if generatedRoute.Method != customRoute.Method {
		t.Errorf("Expected method %s, got %s", customRoute.Method, generatedRoute.Method)
	}

	if generatedRoute.Path != customRoute.Path {
		t.Errorf("Expected path %s, got %s", customRoute.Path, generatedRoute.Path)
	}

	if generatedRoute.Summary != customRoute.Summary {
		t.Errorf("Expected summary %s, got %s", customRoute.Summary, generatedRoute.Summary)
	}

	if len(generatedRoute.Security) == 0 {
		t.Error("Expected security to be set for protected custom route")
	}

	if generatedRoute.OperationID == "" {
		t.Error("Expected operation ID to be generated")
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
