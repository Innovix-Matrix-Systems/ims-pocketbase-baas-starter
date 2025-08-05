package swagger

import (
	"testing"
)

func TestGenerateHybridUpdateContent(t *testing.T) {
	// Create a route generator with dynamic content types enabled
	mockSchemaGen := &mockSchemaGen{}
	routeGen := NewRouteGeneratorWithConfig(mockSchemaGen, true)

	// Test collection with file fields
	collection := EnhancedCollectionInfo{
		Name: "posts",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "content", Type: "editor", Required: false},
			{Name: "avatar", Type: "file", Required: false, Options: map[string]any{"maxSelect": 1}},
			{Name: "images", Type: "file", Required: false, Options: map[string]any{"maxSelect": 5}},
			{Name: "status", Type: "select", Required: false},
		},
	}

	// Create a mock schema
	schema := &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"title": {
				Type:        "string",
				Description: "Post title",
				Required:    true,
				Example:     "My Post Title",
			},
			"content": {
				Type:        "string",
				Description: "Post content",
				Required:    false,
				Example:     "Post content here",
			},
			"avatar": {
				Type:        "string",
				Format:      "binary",
				Description: "Avatar image",
				Required:    false,
			},
			"images": {
				Type:        "array",
				Description: "Multiple images",
				Required:    false,
				Items: &FieldSchema{
					Type:   "string",
					Format: "binary",
				},
			},
			"status": {
				Type:        "string",
				Description: "Post status",
				Required:    false,
				Enum:        []any{"draft", "published"},
				Example:     "draft",
			},
		},
		Required: []string{"title"},
	}

	// Generate hybrid content
	content := routeGen.generateHybridUpdateContent(schema, collection)

	// Verify we have both content types
	if len(content) != 2 {
		t.Errorf("Expected 2 content types, got %d", len(content))
	}

	// Check JSON content
	jsonContent, hasJSON := content["application/json"]
	if !hasJSON {
		t.Fatal("Expected application/json content type")
	}

	jsonSchema, ok := jsonContent.Schema.(map[string]any)
	if !ok {
		t.Fatal("Expected JSON schema to be a map")
	}

	jsonProps, ok := jsonSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected JSON schema to have properties")
	}

	// JSON should contain non-file fields only
	expectedJSONFields := []string{"title", "content", "status"}
	if len(jsonProps) != len(expectedJSONFields) {
		t.Errorf("Expected %d JSON properties, got %d", len(expectedJSONFields), len(jsonProps))
	}

	for _, field := range expectedJSONFields {
		if _, exists := jsonProps[field]; !exists {
			t.Errorf("Expected JSON properties to contain %s", field)
		}
	}

	// JSON should NOT contain file fields
	fileFields := []string{"avatar", "images"}
	for _, field := range fileFields {
		if _, exists := jsonProps[field]; exists {
			t.Errorf("JSON properties should not contain file field %s", field)
		}
	}

	// Check multipart content
	multipartContent, hasMultipart := content["multipart/form-data"]
	if !hasMultipart {
		t.Fatal("Expected multipart/form-data content type")
	}

	multipartSchema, ok := multipartContent.Schema.(map[string]any)
	if !ok {
		t.Fatal("Expected multipart schema to be a map")
	}

	multipartProps, ok := multipartSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected multipart schema to have properties")
	}

	// Multipart should contain file fields only
	if len(multipartProps) != len(fileFields) {
		t.Errorf("Expected %d multipart properties, got %d", len(fileFields), len(multipartProps))
	}

	for _, field := range fileFields {
		if _, exists := multipartProps[field]; !exists {
			t.Errorf("Expected multipart properties to contain %s", field)
		}
	}

	// Multipart should NOT contain non-file fields
	for _, field := range expectedJSONFields {
		if _, exists := multipartProps[field]; exists {
			t.Errorf("Multipart properties should not contain non-file field %s", field)
		}
	}
}

func TestGenerateHybridUpdateContentNoFileFields(t *testing.T) {
	// Create a route generator with dynamic content types enabled
	mockSchemaGen := &mockSchemaGen{}
	routeGen := NewRouteGeneratorWithConfig(mockSchemaGen, true)

	// Test collection without file fields
	collection := EnhancedCollectionInfo{
		Name: "posts",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "content", Type: "editor", Required: false},
			{Name: "status", Type: "select", Required: false},
		},
	}

	schema := &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"title":   {Type: "string", Required: true},
			"content": {Type: "string", Required: false},
			"status":  {Type: "string", Required: false},
		},
		Required: []string{"title"},
	}

	// Generate hybrid content
	content := routeGen.generateHybridUpdateContent(schema, collection)

	// Should fallback to standard behavior (JSON + multipart with all fields)
	// Since there are no file fields, it should return the same as generateRequestContent
	standardContent := routeGen.generateRequestContent(schema, collection, "update")

	if len(content) != len(standardContent) {
		t.Errorf("Expected hybrid content to match standard content for collections without file fields")
	}

	// Should have JSON content
	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}
}

func TestGenerateHybridUpdateContentDisabledFeature(t *testing.T) {
	// Create a route generator with dynamic content types DISABLED
	mockSchemaGen := &mockSchemaGen{}
	routeGen := NewRouteGeneratorWithConfig(mockSchemaGen, false)

	// Test collection with file fields
	collection := EnhancedCollectionInfo{
		Name: "posts",
		Type: "base",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "avatar", Type: "file", Required: false, Options: map[string]any{"maxSelect": 1}},
		},
	}

	schema := &CollectionSchema{
		Type: "object",
		Properties: map[string]*FieldSchema{
			"title":  {Type: "string", Required: true},
			"avatar": {Type: "string", Format: "binary", Required: false},
		},
		Required: []string{"title"},
	}

	// Generate hybrid content
	content := routeGen.generateHybridUpdateContent(schema, collection)

	// Should fallback to standard behavior when feature is disabled
	standardContent := routeGen.generateRequestContent(schema, collection, "update")

	if len(content) != len(standardContent) {
		t.Errorf("Expected hybrid content to match standard content when feature is disabled")
	}

	// Should only have JSON content (since dynamic content types are disabled)
	if len(content) != 1 {
		t.Errorf("Expected 1 content type when feature is disabled, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}
}
