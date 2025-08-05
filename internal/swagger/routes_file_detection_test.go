package swagger

import (
	"reflect"
	"strings"
	"testing"
)

func TestHasFileFields(t *testing.T) {
	rg := NewRouteGenerator(nil)

	tests := []struct {
		name       string
		collection EnhancedCollectionInfo
		expected   bool
	}{
		{
			name: "collection with file field",
			collection: EnhancedCollectionInfo{
				Name: "users",
				Fields: []FieldInfo{
					{Name: "avatar", Type: "file", Required: false},
					{Name: "name", Type: "text", Required: true},
				},
			},
			expected: true,
		},
		{
			name: "collection without file field",
			collection: EnhancedCollectionInfo{
				Name: "posts",
				Fields: []FieldInfo{
					{Name: "title", Type: "text", Required: true},
					{Name: "content", Type: "editor", Required: true},
				},
			},
			expected: false,
		},
		{
			name: "collection with multiple file fields",
			collection: EnhancedCollectionInfo{
				Name: "documents",
				Fields: []FieldInfo{
					{Name: "avatar", Type: "file", Required: false},
					{Name: "attachments", Type: "file", Required: false},
					{Name: "title", Type: "text", Required: true},
				},
			},
			expected: true,
		},
		{
			name: "empty collection",
			collection: EnhancedCollectionInfo{
				Name:   "empty",
				Fields: []FieldInfo{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rg.hasFileFields(tt.collection)
			if result != tt.expected {
				t.Errorf("hasFileFields() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetFileFields(t *testing.T) {
	rg := NewRouteGenerator(nil)

	collection := EnhancedCollectionInfo{
		Name: "documents",
		Fields: []FieldInfo{
			{
				Name:     "avatar",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSize":   1048576, // 1MB
					"mimeTypes": []any{"image/jpeg", "image/png"},
				},
			},
			{
				Name:     "attachments",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSelect": 5,
					"maxSize":   5242880, // 5MB
				},
			},
			{
				Name:     "title",
				Type:     "text",
				Required: true,
			},
		},
	}

	fileFields := rg.getFileFields(collection)

	if len(fileFields) != 2 {
		t.Errorf("Expected 2 file fields, got %d", len(fileFields))
	}

	// Test first file field (avatar)
	avatar := fileFields[0]
	if avatar.Name != "avatar" {
		t.Errorf("Expected first file field name to be 'avatar', got '%s'", avatar.Name)
	}
	if avatar.IsMultiple {
		t.Error("Expected avatar to not be multiple")
	}
	if avatar.MaxSize != 1048576 {
		t.Errorf("Expected avatar max size to be 1048576, got %d", avatar.MaxSize)
	}
	if len(avatar.AllowedTypes) != 2 {
		t.Errorf("Expected 2 allowed types for avatar, got %d", len(avatar.AllowedTypes))
	}

	// Test second file field (attachments)
	attachments := fileFields[1]
	if attachments.Name != "attachments" {
		t.Errorf("Expected second file field name to be 'attachments', got '%s'", attachments.Name)
	}
	if !attachments.IsMultiple {
		t.Error("Expected attachments to be multiple")
	}
	if attachments.MaxSize != 5242880 {
		t.Errorf("Expected attachments max size to be 5242880, got %d", attachments.MaxSize)
	}
}

func TestGetFileFieldsNoFileFields(t *testing.T) {
	rg := NewRouteGenerator(nil)

	collection := EnhancedCollectionInfo{
		Name: "posts",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "content", Type: "editor", Required: true},
		},
	}

	fileFields := rg.getFileFields(collection)

	if len(fileFields) != 0 {
		t.Errorf("Expected 0 file fields, got %d", len(fileFields))
	}
}

func TestRouteGeneratorParseIntOption(t *testing.T) {
	rg := NewRouteGenerator(nil)

	tests := []struct {
		name     string
		value    any
		expected int
		hasError bool
	}{
		{
			name:     "int value",
			value:    42,
			expected: 42,
			hasError: false,
		},
		{
			name:     "float64 value",
			value:    42.0,
			expected: 42,
			hasError: false,
		},
		{
			name:     "string value",
			value:    "42",
			expected: 42,
			hasError: false,
		},
		{
			name:     "invalid string",
			value:    "not-a-number",
			expected: 0,
			hasError: true,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rg.parseIntOption(tt.value)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}
func TestGenerateRequestContentOperationSupport(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Collection with file field
	collection := EnhancedCollectionInfo{
		Name: "users",
		Fields: []FieldInfo{
			{Name: "avatar", Type: "file", Required: false},
			{Name: "name", Type: "text", Required: true},
		},
	}

	// Mock schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "User name",
			},
		},
		"required": []string{"name"},
	}

	tests := []struct {
		name           string
		operation      string
		expectFormData bool
		expectJSONOnly bool
	}{
		{
			name:           "create operation should support form-data",
			operation:      "create",
			expectFormData: true,
		},
		{
			name:           "update operation should support form-data",
			operation:      "update",
			expectFormData: true,
		},
		{
			name:           "list operation should be JSON-only",
			operation:      "list",
			expectJSONOnly: true,
		},
		{
			name:           "view operation should be JSON-only",
			operation:      "view",
			expectJSONOnly: true,
		},
		{
			name:           "delete operation should be JSON-only",
			operation:      "delete",
			expectJSONOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := rg.generateRequestContent(schema, collection, tt.operation)

			// Should always have JSON content type
			if _, hasJSON := content["application/json"]; !hasJSON {
				t.Error("Expected application/json content type")
			}

			// Check form-data expectations
			_, hasFormData := content["multipart/form-data"]

			if tt.expectFormData && !hasFormData {
				t.Errorf("Expected multipart/form-data content type for %s operation", tt.operation)
			}

			if tt.expectJSONOnly && hasFormData {
				t.Errorf("Expected JSON-only for %s operation, but got form-data", tt.operation)
			}
		})
	}
}

func TestGenerateRequestContentNoFileFields(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Collection without file fields
	collection := EnhancedCollectionInfo{
		Name: "posts",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "content", Type: "editor", Required: true},
		},
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type": "string",
			},
			"content": map[string]any{
				"type": "string",
			},
		},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should only have JSON content type
	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	if _, hasFormData := content["multipart/form-data"]; hasFormData {
		t.Error("Expected no multipart/form-data for collection without file fields")
	}

	if len(content) != 1 {
		t.Errorf("Expected only 1 content type, got %d", len(content))
	}
}

func TestGenerateRequestContentFormDataSchema(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Collection with mixed field types
	collection := EnhancedCollectionInfo{
		Name: "documents",
		Fields: []FieldInfo{
			{
				Name:     "avatar",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSize": 1048576,
				},
			},
			{
				Name:     "attachments",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSelect": 3,
				},
			},
			{Name: "title", Type: "text", Required: true},
		},
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"attachments": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":        "string",
					"format":      "binary",
					"description": "File upload",
				},
			},
			"title": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"title"},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should have both content types
	if len(content) != 2 {
		t.Errorf("Expected 2 content types, got %d", len(content))
	}

	// Check form-data schema
	formData, hasFormData := content["multipart/form-data"]
	if !hasFormData {
		t.Fatal("Expected multipart/form-data content type")
	}

	formSchema, ok := formData.Schema.(map[string]any)
	if !ok {
		t.Fatal("Expected form-data schema to be a map")
	}

	props, ok := formSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected form-data schema to have properties")
	}

	// Check that file fields are properly formatted as binary
	avatarProp, hasAvatar := props["avatar"]
	if !hasAvatar {
		t.Error("Expected avatar property in form-data schema")
	} else {
		avatarMap, ok := avatarProp.(map[string]any)
		if !ok {
			t.Error("Expected avatar property to be a map")
		} else {
			if avatarMap["format"] != "binary" {
				t.Error("Expected avatar to have binary format")
			}
		}
	}

	// Check that non-file fields retain their original schema
	titleProp, hasTitle := props["title"]
	if !hasTitle {
		t.Error("Expected title property in form-data schema")
	} else {
		titleMap, ok := titleProp.(map[string]any)
		if !ok {
			t.Error("Expected title property to be a map")
		} else {
			if titleMap["type"] != "string" {
				t.Error("Expected title to retain string type")
			}
		}
	}
}
func TestImprovedFormDataSchemaGeneration(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Collection with various field types
	collection := EnhancedCollectionInfo{
		Name: "mixed_fields",
		Fields: []FieldInfo{
			{
				Name:     "avatar",
				Type:     "file",
				Required: true,
				Options: map[string]any{
					"maxSize":   2097152, // 2MB
					"mimeTypes": []any{"image/jpeg", "image/png"},
				},
			},
			{
				Name:     "documents",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSelect": 3,
					"maxSize":   5242880, // 5MB
				},
			},
			{Name: "title", Type: "text", Required: true},
			{Name: "active", Type: "bool", Required: false},
			{Name: "metadata", Type: "json", Required: false},
		},
	}

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "Avatar file",
			},
			"documents": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":        "string",
					"format":      "binary",
					"description": "Document file",
				},
			},
			"title": map[string]any{
				"type":        "string",
				"description": "Title field",
			},
			"active": map[string]any{
				"type":        "boolean",
				"description": "Active status",
			},
			"metadata": map[string]any{
				"type":        "object",
				"description": "Metadata object",
			},
		},
		"required": []string{"title", "avatar"},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should have both content types
	if len(content) != 2 {
		t.Errorf("Expected 2 content types, got %d", len(content))
	}

	// Check form-data schema
	formData, hasFormData := content["multipart/form-data"]
	if !hasFormData {
		t.Fatal("Expected multipart/form-data content type")
	}

	formSchema, ok := formData.Schema.(map[string]any)
	if !ok {
		t.Fatal("Expected form-data schema to be a map")
	}

	props, ok := formSchema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Expected form-data schema to have properties")
	}

	// Test single file field with constraints
	avatarProp, hasAvatar := props["avatar"]
	if !hasAvatar {
		t.Error("Expected avatar property in form-data schema")
	} else {
		avatarMap, ok := avatarProp.(map[string]any)
		if !ok {
			t.Error("Expected avatar property to be a map")
		} else {
			if avatarMap["format"] != "binary" {
				t.Error("Expected avatar to have binary format")
			}
			if desc, ok := avatarMap["description"].(string); ok {
				if !strings.Contains(desc, "2097152 bytes") {
					t.Error("Expected avatar description to include max size")
				}
				if !strings.Contains(desc, "image/jpeg, image/png") {
					t.Error("Expected avatar description to include allowed types")
				}
			}
			if example, ok := avatarMap["example"].(string); ok {
				if !strings.Contains(example, "@avatar") {
					t.Error("Expected avatar to have file example")
				}
			}
		}
	}

	// Test multiple file field
	docsProp, hasDocs := props["documents"]
	if !hasDocs {
		t.Error("Expected documents property in form-data schema")
	} else {
		docsMap, ok := docsProp.(map[string]any)
		if !ok {
			t.Error("Expected documents property to be a map")
		} else {
			if docsMap["type"] != "array" {
				t.Error("Expected documents to be array type")
			}
			if maxItems, ok := docsMap["maxItems"].(int); ok {
				if maxItems != 3 {
					t.Errorf("Expected documents maxItems to be 3, got %d", maxItems)
				}
			}
		}
	}

	// Test boolean field conversion
	activeProp, hasActive := props["active"]
	if !hasActive {
		t.Error("Expected active property in form-data schema")
	} else {
		activeMap, ok := activeProp.(map[string]any)
		if !ok {
			t.Error("Expected active property to be a map")
		} else {
			if activeMap["type"] != "string" {
				t.Error("Expected active to be converted to string type for form-data")
			}
			if enum, ok := activeMap["enum"].([]any); ok {
				if len(enum) != 2 || enum[0] != "true" || enum[1] != "false" {
					t.Error("Expected active to have true/false enum values")
				}
			}
		}
	}

	// Test object field conversion
	metadataProp, hasMetadata := props["metadata"]
	if !hasMetadata {
		t.Error("Expected metadata property in form-data schema")
	} else {
		metadataMap, ok := metadataProp.(map[string]any)
		if !ok {
			t.Error("Expected metadata property to be a map")
		} else {
			if metadataMap["type"] != "string" {
				t.Error("Expected metadata to be converted to string type for form-data")
			}
			if desc, ok := metadataMap["description"].(string); ok {
				if !strings.Contains(desc, "JSON string") {
					t.Error("Expected metadata description to mention JSON string")
				}
			}
		}
	}

	// Test that required fields are preserved
	if required, ok := formSchema["required"].([]string); ok {
		expectedRequired := []string{"title", "avatar"}
		if len(required) != len(expectedRequired) {
			t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
		}
	}

	// Test that example is generated
	if _, hasExample := formSchema["example"]; !hasExample {
		t.Error("Expected form-data schema to have example")
	}
}
func TestConfigurationDisablesDynamicContentTypes(t *testing.T) {
	// Create route generator with dynamic content types disabled
	rg := NewRouteGeneratorWithConfig(nil, false)

	// Collection with file field
	collection := EnhancedCollectionInfo{
		Name: "users",
		Fields: []FieldInfo{
			{Name: "avatar", Type: "file", Required: false},
			{Name: "name", Type: "text", Required: true},
		},
	}

	// Mock schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "User name",
			},
		},
		"required": []string{"name"},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should only have JSON content type when disabled
	if len(content) != 1 {
		t.Errorf("Expected only 1 content type when disabled, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	if _, hasFormData := content["multipart/form-data"]; hasFormData {
		t.Error("Expected no multipart/form-data when dynamic content types are disabled")
	}
}

func TestConfigurationEnablesDynamicContentTypes(t *testing.T) {
	// Create route generator with dynamic content types enabled
	rg := NewRouteGeneratorWithConfig(nil, true)

	// Collection with file field
	collection := EnhancedCollectionInfo{
		Name: "users",
		Fields: []FieldInfo{
			{Name: "avatar", Type: "file", Required: false},
			{Name: "name", Type: "text", Required: true},
		},
	}

	// Mock schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "User name",
			},
		},
		"required": []string{"name"},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should have both content types when enabled
	if len(content) != 2 {
		t.Errorf("Expected 2 content types when enabled, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	if _, hasFormData := content["multipart/form-data"]; !hasFormData {
		t.Error("Expected multipart/form-data when dynamic content types are enabled")
	}
}

func TestDefaultRouteGeneratorEnablesDynamicContentTypes(t *testing.T) {
	// Default constructor should enable dynamic content types for backward compatibility
	rg := NewRouteGenerator(nil)

	// Collection with file field
	collection := EnhancedCollectionInfo{
		Name: "users",
		Fields: []FieldInfo{
			{Name: "avatar", Type: "file", Required: false},
			{Name: "name", Type: "text", Required: true},
		},
	}

	// Mock schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"avatar": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "User name",
			},
		},
		"required": []string{"name"},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should have both content types by default
	if len(content) != 2 {
		t.Errorf("Expected 2 content types by default, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	if _, hasFormData := content["multipart/form-data"]; !hasFormData {
		t.Error("Expected multipart/form-data by default")
	}
}
func TestErrorHandlingAndLogging(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Test with malformed field options
	collection := EnhancedCollectionInfo{
		Name: "test_errors",
		Fields: []FieldInfo{
			{
				Name:     "avatar",
				Type:     "file",
				Required: false,
				Options: map[string]any{
					"maxSize":   "invalid_size", // Invalid size
					"maxSelect": "not_a_number", // Invalid maxSelect
					"mimeTypes": "not_an_array", // Invalid mimeTypes format
				},
			},
			{Name: "title", Type: "text", Required: true},
		},
	}

	// This should not panic and should handle errors gracefully
	fileFields := rg.getFileFields(collection)

	// Should still find the file field despite errors
	if len(fileFields) != 1 {
		t.Errorf("Expected 1 file field despite errors, got %d", len(fileFields))
	}

	if fileFields[0].Name != "avatar" {
		t.Errorf("Expected file field name to be 'avatar', got '%s'", fileFields[0].Name)
	}

	// Should have default values due to parsing errors
	if fileFields[0].MaxSize != 0 {
		t.Errorf("Expected MaxSize to be 0 due to parsing error, got %d", fileFields[0].MaxSize)
	}

	if fileFields[0].IsMultiple {
		t.Error("Expected IsMultiple to be false due to parsing error")
	}

	if len(fileFields[0].AllowedTypes) != 0 {
		t.Errorf("Expected no allowed types due to parsing error, got %d", len(fileFields[0].AllowedTypes))
	}
}

func TestErrorHandlingEmptyCollection(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Test with empty collection
	collection := EnhancedCollectionInfo{
		Name:   "empty_collection",
		Fields: []FieldInfo{},
	}

	// Should handle empty collection gracefully
	fileFields := rg.getFileFields(collection)

	if len(fileFields) != 0 {
		t.Errorf("Expected 0 file fields for empty collection, got %d", len(fileFields))
	}

	// Should return JSON-only content
	schema := map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	if len(content) != 1 {
		t.Errorf("Expected 1 content type for empty collection, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	if _, hasFormData := content["multipart/form-data"]; hasFormData {
		t.Error("Expected no multipart/form-data for empty collection")
	}
}

func TestErrorHandlingInvalidSchema(t *testing.T) {
	rg := NewRouteGenerator(nil)

	collection := EnhancedCollectionInfo{
		Name: "users",
		Fields: []FieldInfo{
			{Name: "avatar", Type: "file", Required: false},
		},
	}

	// Test with invalid schema (not a map)
	invalidSchema := "not_a_map"
	content := rg.generateRequestContent(invalidSchema, collection, "create")

	// Should fallback to JSON-only
	if len(content) != 1 {
		t.Errorf("Expected 1 content type for invalid schema, got %d", len(content))
	}

	if _, hasJSON := content["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}

	// Test with schema missing properties
	schemaWithoutProps := map[string]any{
		"type": "object",
		// Missing "properties" field
	}

	content2 := rg.generateRequestContent(schemaWithoutProps, collection, "create")

	// Should fallback to JSON-only
	if len(content2) != 1 {
		t.Errorf("Expected 1 content type for schema without properties, got %d", len(content2))
	}

	if _, hasJSON := content2["application/json"]; !hasJSON {
		t.Error("Expected application/json content type")
	}
}
func TestRelationFieldExamplesInFormData(t *testing.T) {
	rg := NewRouteGenerator(nil)

	// Collection with relation fields and file fields
	collection := EnhancedCollectionInfo{
		Name: "posts",
		Fields: []FieldInfo{
			{Name: "title", Type: "text", Required: true},
			{Name: "author", Type: "relation", Required: true, Options: map[string]any{
				"maxSelect": 1, "collectionId": "users",
			}},
			{Name: "categories", Type: "relation", Required: false, Options: map[string]any{
				"maxSelect": 5, "collectionId": "categories",
			}},
			{Name: "featured_image", Type: "file", Required: false, Options: map[string]any{
				"maxSelect": 1,
			}},
			{Name: "is_published", Type: "bool", Required: false},
		},
	}

	// Mock schema with relation examples
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "Post title",
				"example":     "example_title",
			},
			"author": map[string]any{
				"type":        "string",
				"description": "Related record ID (references collection: users)",
				"example":     "RELATION_RECORD_ID",
			},
			"categories": map[string]any{
				"type":        "array",
				"description": "Relation field (references collection: categories)",
				"items": map[string]any{
					"type":        "string",
					"description": "Related record ID",
				},
				"example": []any{"RELATION_RECORD_ID"},
			},
			"featured_image": map[string]any{
				"type":        "string",
				"format":      "binary",
				"description": "File upload",
			},
			"is_published": map[string]any{
				"type":        "boolean",
				"description": "Boolean field",
				"example":     true,
			},
		},
		"required": []string{"title", "author"},
		"example": map[string]any{
			"title":        "example_title",
			"author":       "RELATION_RECORD_ID",
			"categories":   []any{"RELATION_RECORD_ID"},
			"is_published": true,
		},
	}

	content := rg.generateRequestContent(schema, collection, "create")

	// Should have both JSON and form-data content types
	if len(content) != 2 {
		t.Errorf("Expected 2 content types, got %d", len(content))
	}

	// Test JSON content
	jsonContent, hasJSON := content["application/json"]
	if !hasJSON {
		t.Fatal("Expected JSON content type")
	}

	jsonSchema := jsonContent.Schema.(map[string]any)
	jsonProps := jsonSchema["properties"].(map[string]any)

	// Verify author field in JSON
	authorProp := jsonProps["author"].(map[string]any)
	if authorProp["example"] != "RELATION_RECORD_ID" {
		t.Errorf("Expected author example 'RELATION_RECORD_ID' in JSON, got %v", authorProp["example"])
	}

	// Verify categories field in JSON
	categoriesProp := jsonProps["categories"].(map[string]any)
	expectedCategoriesExample := []any{"RELATION_RECORD_ID"}
	if !reflect.DeepEqual(categoriesProp["example"], expectedCategoriesExample) {
		t.Errorf("Expected categories example %v in JSON, got %v", expectedCategoriesExample, categoriesProp["example"])
	}

	// Test form-data content
	formContent, hasForm := content["multipart/form-data"]
	if !hasForm {
		t.Fatal("Expected multipart/form-data content type")
	}

	formSchema := formContent.Schema.(map[string]any)
	formProps := formSchema["properties"].(map[string]any)

	// Verify author field in form-data
	formAuthorProp := formProps["author"].(map[string]any)
	if formAuthorProp["example"] != "RELATION_RECORD_ID" {
		t.Errorf("Expected author example 'RELATION_RECORD_ID' in form-data, got %v", formAuthorProp["example"])
	}

	// Verify categories field in form-data
	formCategoriesProp := formProps["categories"].(map[string]any)
	if !reflect.DeepEqual(formCategoriesProp["example"], expectedCategoriesExample) {
		t.Errorf("Expected categories example %v in form-data, got %v", expectedCategoriesExample, formCategoriesProp["example"])
	}

	// Verify form-data example includes relation fields
	if formExample, hasExample := formSchema["example"]; hasExample {
		exampleMap := formExample.(map[string]any)

		if exampleMap["author"] != "RELATION_RECORD_ID" {
			t.Errorf("Expected author 'RELATION_RECORD_ID' in form-data example, got %v", exampleMap["author"])
		}

		if !reflect.DeepEqual(exampleMap["categories"], expectedCategoriesExample) {
			t.Errorf("Expected categories %v in form-data example, got %v", expectedCategoriesExample, exampleMap["categories"])
		}
	} else {
		t.Error("Expected form-data schema to have example")
	}
}
