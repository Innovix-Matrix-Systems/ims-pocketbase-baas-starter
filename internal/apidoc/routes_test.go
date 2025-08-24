package apidoc

import (
	"testing"

	"github.com/pocketbase/pocketbase"
)

// Mock schema generator for testing
type mockSchemaGen struct{}

func (m *mockSchemaGen) GenerateCollectionSchema(collection CollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"id":   {Type: "string", Required: true},
			"name": {Type: "string", Required: true},
		},
		Required: []string{"id", "name"},
	}, nil
}

func (m *mockSchemaGen) GenerateCollectionSchemas(collections []CollectionInfo) (map[string]*CollectionSchema, error) {
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

func (m *mockSchemaGen) GenerateCreateSchema(collection CollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"name": {Type: "string", Required: true},
		},
		Required: []string{"name"},
	}, nil
}

func (m *mockSchemaGen) GenerateUpdateSchema(collection CollectionInfo) (*CollectionSchema, error) {
	return &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"name": {Type: "string", Required: false},
		},
		Required: []string{},
	}, nil
}

func (m *mockSchemaGen) GenerateListResponseSchema(collection CollectionInfo) (map[string]any, error) {
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

func (m *mockSchemaGen) GetSchemaName(collection CollectionInfo) string {
	return collection.Name
}

func (m *mockSchemaGen) GetCreateSchemaName(collection CollectionInfo) string {
	return collection.Name + "Create"
}

func (m *mockSchemaGen) GetUpdateSchemaName(collection CollectionInfo) string {
	return collection.Name + "Update"
}

func (m *mockSchemaGen) GetListResponseSchemaName(collection CollectionInfo) string {
	return collection.Name + "ListResponse"
}

func TestNewRouteGenerator(t *testing.T) {
	app := pocketbase.New()
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGeneratorWithFullConfig(app, schemaGen, true, false) // Disable superuser route exclusion for tests

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.schemaGen == nil {
		t.Error("Expected schemaGen to be set")
	}

	if generator.customRoutes == nil {
		t.Error("Expected customRoutes to be initialized")
	}
}

func TestGenerateCollectionRoutes(t *testing.T) {
	app := pocketbase.New()
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGeneratorWithFullConfig(app, schemaGen, true, false) // Disable superuser route exclusion for tests

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "name", Type: "text", Required: true},
		},
	}

	routes, err := generator.GenerateCollectionRoutes(collection)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(routes) == 0 {
		t.Error("Expected routes to be generated")
	}

	// Basic check that we have CRUD operations
	hasGet := false
	hasPost := false
	hasPatch := false
	hasDelete := false

	for _, route := range routes {
		switch route.Method {
		case "GET":
			hasGet = true
		case "POST":
			hasPost = true
		case "PATCH":
			hasPatch = true
		case "DELETE":
			hasDelete = true
		}
	}

	if !hasGet || !hasPost || !hasPatch || !hasDelete {
		t.Error("Expected all CRUD operations to be present")
	}
}

func TestGenerateAuthRoutes(t *testing.T) {
	app := pocketbase.New()
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGeneratorWithFullConfig(app, schemaGen, true, false) // Disable superuser route exclusion for tests

	// Test with auth collection
	authCollection := CollectionInfo{
		Name: "users",
		Type: "auth",
	}

	routes, err := generator.GenerateAuthRoutes(authCollection)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(routes) == 0 {
		t.Error("Expected auth routes to be generated")
	}

	// Test with non-auth collection
	baseCollection := CollectionInfo{
		Name: "posts",
		Type: "base",
	}

	routes, err = generator.GenerateAuthRoutes(baseCollection)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(routes) != 0 {
		t.Error("Expected no auth routes for base collection")
	}
}

func TestRegisterCustomRoute(t *testing.T) {
	app := pocketbase.New()
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGeneratorWithFullConfig(app, schemaGen, true, false) // Disable superuser route exclusion for tests

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
}

func TestGetAllRoutes(t *testing.T) {
	app := pocketbase.New()
	schemaGen := &mockSchemaGen{}
	generator := NewRouteGeneratorWithFullConfig(app, schemaGen, true, false) // Disable superuser route exclusion for tests

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

	collections := []CollectionInfo{
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
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(routes) == 0 {
		t.Error("Expected routes to be generated")
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
