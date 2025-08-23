package apidoc

import (
	"strings"
	"testing"
)

func TestNewSchemaGenerator(t *testing.T) {
	generator := NewSchemaGenerator()

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if !generator.includeExamples {
		t.Error("Expected includeExamples to be true by default")
	}

	if !generator.includeSystem {
		t.Error("Expected includeSystem to be true by default")
	}

	if generator.fieldMapper == nil {
		t.Error("Expected fieldMapper to be initialized")
	}
}

func TestNewSchemaGeneratorWithConfig(t *testing.T) {
	generator := NewSchemaGeneratorWithConfig(false, false)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.includeExamples {
		t.Error("Expected includeExamples to be false")
	}

	if generator.includeSystem {
		t.Error("Expected includeSystem to be false")
	}
}

func TestGenerateCollectionSchema(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{
				Name:     "name",
				Type:     "text",
				Required: true,
				Options:  map[string]any{"max": 100},
			},
			{
				Name:     "email",
				Type:     "email",
				Required: true,
			},
			{
				Name:     "age",
				Type:     "number",
				Required: false,
				Options:  map[string]any{"min": 0, "max": 120},
			},
		},
	}

	schema, err := generator.GenerateCollectionSchema(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	// Check system fields are included
	systemFields := []string{"id", "created", "updated"}
	for _, fieldName := range systemFields {
		if _, exists := schema.Properties[fieldName]; !exists {
			t.Errorf("Expected system field %s to be included", fieldName)
		}
	}

	// Check collection fields are included
	collectionFields := []string{"name", "email", "age"}
	for _, fieldName := range collectionFields {
		if _, exists := schema.Properties[fieldName]; !exists {
			t.Errorf("Expected collection field %s to be included", fieldName)
		}
	}

	// Check required fields
	expectedRequired := []string{"id", "created", "updated", "name", "email"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}

	// Check example is generated
	if schema.Example == nil {
		t.Error("Expected example to be generated")
	}

	// Validate field types
	if schema.Properties["name"].Type != "string" {
		t.Error("Expected name field to be string type")
	}

	if schema.Properties["email"].Format != "email" {
		t.Error("Expected email field to have email format")
	}

	if schema.Properties["age"].Type != "number" {
		t.Error("Expected age field to be number type")
	}
}

func TestGenerateCollectionSchemaWithoutSystemFields(t *testing.T) {
	generator := NewSchemaGeneratorWithConfig(true, false)

	collection := CollectionInfo{
		Name: "posts",
		Type: "base",
		Fields: []FieldInfo{
			{
				Name:     "title",
				Type:     "text",
				Required: true,
			},
		},
	}

	schema, err := generator.GenerateCollectionSchema(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check system fields are not included
	systemFields := []string{"id", "created", "updated"}
	for _, fieldName := range systemFields {
		if _, exists := schema.Properties[fieldName]; exists {
			t.Errorf("Expected system field %s to not be included", fieldName)
		}
	}

	// Check only collection fields are included
	if len(schema.Properties) != 1 {
		t.Errorf("Expected 1 property, got %d", len(schema.Properties))
	}

	if _, exists := schema.Properties["title"]; !exists {
		t.Error("Expected title field to be included")
	}
}

func TestGenerateCollectionSchemaEmptyName(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "",
		Type: "base",
	}

	schema, err := generator.GenerateCollectionSchema(collection)
	if err == nil {
		t.Error("Expected error for empty collection name, got nil")
	}

	if schema != nil {
		t.Error("Expected nil schema for empty collection name")
	}
}

func TestGenerateCollectionSchemas(t *testing.T) {
	generator := NewSchemaGenerator()

	collections := []CollectionInfo{
		{
			Name: "users",
			Type: "base",
			Fields: []FieldInfo{
				{Name: "name", Type: "text", Required: true},
			},
		},
		{
			Name: "posts",
			Type: "base",
			Fields: []FieldInfo{
				{Name: "title", Type: "text", Required: true},
			},
		},
	}

	schemas, err := generator.GenerateCollectionSchemas(collections)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("Expected 2 schemas, got %d", len(schemas))
	}

	if _, exists := schemas["users"]; !exists {
		t.Error("Expected users schema to be generated")
	}

	if _, exists := schemas["posts"]; !exists {
		t.Error("Expected posts schema to be generated")
	}
}

func TestGenerateCreateSchema(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{
				Name:     "name",
				Type:     "text",
				Required: true,
				System:   false,
			},
			{
				Name:     "email",
				Type:     "email",
				Required: true,
				System:   false,
			},
			{
				Name:     "internal_id",
				Type:     "text",
				Required: false,
				System:   true, // This should be excluded
			},
		},
	}

	schema, err := generator.GenerateCreateSchema(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	// Check system fields are not included
	if _, exists := schema.Properties["id"]; exists {
		t.Error("Expected system field 'id' to not be included in create schema")
	}

	if _, exists := schema.Properties["internal_id"]; exists {
		t.Error("Expected system field 'internal_id' to not be included in create schema")
	}

	// Check only non-system collection fields are included
	expectedFields := []string{"name", "email"}
	if len(schema.Properties) != len(expectedFields) {
		t.Errorf("Expected %d properties, got %d", len(expectedFields), len(schema.Properties))
	}

	for _, fieldName := range expectedFields {
		if _, exists := schema.Properties[fieldName]; !exists {
			t.Errorf("Expected field %s to be included in create schema", fieldName)
		}
	}

	// Check required fields
	expectedRequired := []string{"name", "email"}
	if len(schema.Required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(schema.Required))
	}
}

func TestGenerateUpdateSchema(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{
				Name:     "name",
				Type:     "text",
				Required: true,
				System:   false,
			},
			{
				Name:     "email",
				Type:     "email",
				Required: true,
				System:   false,
			},
		},
	}

	schema, err := generator.GenerateUpdateSchema(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	// Check no fields are required in update schema
	if len(schema.Required) != 0 {
		t.Errorf("Expected 0 required fields in update schema, got %d", len(schema.Required))
	}

	// Check all fields are optional
	for fieldName, fieldSchema := range schema.Properties {
		if fieldSchema.Required {
			t.Errorf("Expected field %s to be optional in update schema", fieldName)
		}
	}

	// Check collection fields are included
	expectedFields := []string{"name", "email"}
	for _, fieldName := range expectedFields {
		if _, exists := schema.Properties[fieldName]; !exists {
			t.Errorf("Expected field %s to be included in update schema", fieldName)
		}
	}
}

func TestGenerateListResponseSchema(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "name", Type: "text", Required: true},
		},
	}

	schema, err := generator.GenerateListResponseSchema(collection)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	// Check list response structure
	if schema["type"] != "object" {
		t.Error("Expected list response type to be 'object'")
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	// Check pagination fields
	paginationFields := []string{"page", "perPage", "totalItems", "totalPages", "items"}
	for _, fieldName := range paginationFields {
		if _, exists := properties[fieldName]; !exists {
			t.Errorf("Expected pagination field %s to be included", fieldName)
		}
	}

	// Check items field is array type
	items, ok := properties["items"].(map[string]any)
	if !ok {
		t.Fatal("Expected items to be a map")
	}

	if items["type"] != "array" {
		t.Error("Expected items type to be 'array'")
	}

	// Check example is generated
	if example, exists := schema["example"]; exists {
		exampleMap, ok := example.(map[string]any)
		if !ok {
			t.Error("Expected example to be a map")
		} else {
			if _, exists := exampleMap["items"]; !exists {
				t.Error("Expected example to include items")
			}
		}
	}
}

func TestGenerateBasicExample(t *testing.T) {
	generator := NewSchemaGenerator()

	tests := []struct {
		name        string
		fieldSchema *FieldSchema
		fieldName   string
		expected    any
	}{
		{
			name: "string field",
			fieldSchema: &FieldSchema{
				Type: "string",
			},
			fieldName: "title",
			expected:  "example_title",
		},
		{
			name: "email field",
			fieldSchema: &FieldSchema{
				Type:   "string",
				Format: "email",
			},
			fieldName: "email",
			expected:  "user@example.com",
		},
		{
			name: "number field",
			fieldSchema: &FieldSchema{
				Type: "number",
			},
			fieldName: "count",
			expected:  42,
		},
		{
			name: "boolean field",
			fieldSchema: &FieldSchema{
				Type: "boolean",
			},
			fieldName: "active",
			expected:  true,
		},
		{
			name: "array field",
			fieldSchema: &FieldSchema{
				Type: "array",
			},
			fieldName: "tags",
			expected:  "array", // We'll check the type instead of comparing slices
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.generateBasicExample(tt.fieldSchema, tt.fieldName)

			// Handle different types of comparisons
			if tt.fieldSchema.Type == "string" && tt.fieldSchema.Format == "" {
				if !strings.Contains(result.(string), tt.fieldName) {
					t.Errorf("Expected result to contain field name %s, got %v", tt.fieldName, result)
				}
			} else if tt.fieldSchema.Type == "array" {
				// For arrays, just check that we got a slice
				if _, ok := result.([]any); !ok {
					t.Errorf("Expected array result, got %T", result)
				}
			} else {
				if result != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
				}
			}
		})
	}
}

func TestGetSchemaNames(t *testing.T) {
	generator := NewSchemaGenerator()

	collection := CollectionInfo{
		Name: "users",
		Type: "base",
	}

	tests := []struct {
		name     string
		method   func(CollectionInfo) string
		expected string
	}{
		{"GetSchemaName", generator.GetSchemaName, "users"},
		{"GetCreateSchemaName", generator.GetCreateSchemaName, "usersCreate"},
		{"GetUpdateSchemaName", generator.GetUpdateSchemaName, "usersUpdate"},
		{"GetListResponseSchemaName", generator.GetListResponseSchemaName, "usersListResponse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method(collection)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateSchema(t *testing.T) {
	generator := NewSchemaGenerator()

	tests := []struct {
		name        string
		schema      *CollectionSchema
		expectError bool
		errorMsg    string
	}{
		{
			name:        "nil schema",
			schema:      nil,
			expectError: true,
			errorMsg:    "schema is nil",
		},
		{
			name: "invalid type",
			schema: &CollectionSchema{
				Type:       "array",
				Properties: map[string]*FieldSchema{},
			},
			expectError: true,
			errorMsg:    "schema type must be 'object'",
		},
		{
			name: "nil properties",
			schema: &CollectionSchema{
				Type:       "object",
				Properties: nil,
			},
			expectError: true,
			errorMsg:    "schema properties cannot be nil",
		},
		{
			name: "empty properties",
			schema: &CollectionSchema{
				Type:       "object",
				Properties: map[string]*FieldSchema{},
			},
			expectError: true,
			errorMsg:    "schema must have at least one property",
		},
		{
			name: "missing required field",
			schema: &CollectionSchema{
				Type: "object",
				Properties: map[string]*FieldSchema{
					"name": {Type: "string"},
				},
				Required: []string{"name", "email"}, // email is missing
			},
			expectError: true,
			errorMsg:    "required field 'email' not found in properties",
		},
		{
			name: "valid schema",
			schema: &CollectionSchema{
				Type: "object",
				Properties: map[string]*FieldSchema{
					"name":  {Type: "string"},
					"email": {Type: "string"},
				},
				Required: []string{"name"},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := generator.ValidateSchema(tt.schema)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestShouldIncludeInCreateExample(t *testing.T) {
	generator := NewSchemaGenerator()

	tests := []struct {
		name        string
		fieldName   string
		fieldSchema *FieldSchema
		expected    bool
	}{
		{
			name:        "required field",
			fieldName:   "any_field",
			fieldSchema: &FieldSchema{Required: true},
			expected:    true,
		},
		{
			name:        "name field",
			fieldName:   "user_name",
			fieldSchema: &FieldSchema{Required: false},
			expected:    true,
		},
		{
			name:        "title field",
			fieldName:   "post_title",
			fieldSchema: &FieldSchema{Required: false},
			expected:    true,
		},
		{
			name:        "random field",
			fieldName:   "random_field",
			fieldSchema: &FieldSchema{Required: false},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.shouldIncludeInCreateExample(tt.fieldName, tt.fieldSchema)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldIncludeInUpdateExample(t *testing.T) {
	generator := NewSchemaGenerator()

	tests := []struct {
		name        string
		fieldName   string
		fieldSchema *FieldSchema
		expected    bool
	}{
		{
			name:        "status field",
			fieldName:   "status",
			fieldSchema: &FieldSchema{Type: "string"},
			expected:    true,
		},
		{
			name:        "boolean field",
			fieldName:   "some_flag",
			fieldSchema: &FieldSchema{Type: "boolean"},
			expected:    true,
		},
		{
			name:        "string field",
			fieldName:   "description",
			fieldSchema: &FieldSchema{Type: "string"},
			expected:    true,
		},
		{
			name:        "number field",
			fieldName:   "count",
			fieldSchema: &FieldSchema{Type: "number"},
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.shouldIncludeInUpdateExample(tt.fieldName, tt.fieldSchema)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
