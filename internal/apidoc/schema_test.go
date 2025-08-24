package apidoc

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewFieldSchemaMapper(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	if mapper == nil {
		t.Fatal("Expected mapper to be created, got nil")
	}

	if !mapper.includeExamples {
		t.Error("Expected includeExamples to be true by default")
	}

	if !mapper.strictValidation {
		t.Error("Expected strictValidation to be true by default")
	}
}

func TestNewFieldSchemaMapperWithConfig(t *testing.T) {
	mapper := NewFieldSchemaMapperWithConfig(false, false)

	if mapper == nil {
		t.Fatal("Expected mapper to be created, got nil")
	}

	if mapper.includeExamples {
		t.Error("Expected includeExamples to be false")
	}

	if mapper.strictValidation {
		t.Error("Expected strictValidation to be false")
	}
}

func TestMapFieldToSchemaTextField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name:     "title",
		Type:     "text",
		Required: true,
		Options: map[string]any{
			"max": 100,
			"min": 5,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema == nil {
		t.Fatal("Expected schema, got nil")
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if !schema.Required {
		t.Error("Expected schema to be required")
	}

	if schema.MaxLength == nil || *schema.MaxLength != 100 {
		t.Errorf("Expected MaxLength 100, got %v", schema.MaxLength)
	}

	if schema.MinLength == nil || *schema.MinLength != 5 {
		t.Errorf("Expected MinLength 5, got %v", schema.MinLength)
	}

	if schema.Example == nil {
		t.Error("Expected example to be set")
	}
}

func TestMapFieldToSchemaNumberField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name:     "count",
		Type:     "number",
		Required: false,
		Options: map[string]any{
			"max": 1000.5,
			"min": 0.1,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "number" {
		t.Errorf("Expected type 'number', got %s", schema.Type)
	}

	if schema.Required {
		t.Error("Expected schema to not be required")
	}

	if schema.Maximum == nil || *schema.Maximum != 1000.5 {
		t.Errorf("Expected Maximum 1000.5, got %v", schema.Maximum)
	}

	if schema.Minimum == nil || *schema.Minimum != 0.1 {
		t.Errorf("Expected Minimum 0.1, got %v", schema.Minimum)
	}
}

func TestMapFieldToSchemaIntegerField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name:     "age",
		Type:     "number",
		Required: true,
		Options: map[string]any{
			"onlyInt": true,
			"max":     120,
			"min":     0,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "integer" {
		t.Errorf("Expected type 'integer', got %s", schema.Type)
	}
}

func TestMapFieldToSchemaBooleanField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name:     "active",
		Type:     "bool",
		Required: false,
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "boolean" {
		t.Errorf("Expected type 'boolean', got %s", schema.Type)
	}

	if schema.Example != true {
		t.Error("Expected example to be true")
	}
}

func TestMapFieldToSchemaEmailField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name:     "email",
		Type:     "email",
		Required: true,
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Format != "email" {
		t.Errorf("Expected format 'email', got %s", schema.Format)
	}

	if schema.Example != "user@example.com" {
		t.Errorf("Expected email example, got %v", schema.Example)
	}
}

func TestMapFieldToSchemaUrlField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "website",
		Type: "url",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Format != "uri" {
		t.Errorf("Expected format 'uri', got %s", schema.Format)
	}
}

func TestMapFieldToSchemaDateField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "birthday",
		Type: "date",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Format != "date-time" {
		t.Errorf("Expected format 'date-time', got %s", schema.Format)
	}
}

func TestMapFieldToSchemaSelectField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "status",
		Type: "select",
		Options: map[string]any{
			"values": []any{"active", "inactive", "pending"},
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if len(schema.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(schema.Enum))
	}

	if schema.Enum[0] != "active" {
		t.Errorf("Expected first enum value 'active', got %v", schema.Enum[0])
	}
}

func TestMapFieldToSchemaMultiSelectField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "tags",
		Type: "select",
		Options: map[string]any{
			"values":    []any{"tag1", "tag2", "tag3"},
			"maxSelect": 3,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "array" {
		t.Errorf("Expected type 'array', got %s", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected items to be set for array type")
	}

	if schema.Items.Type != "string" {
		t.Errorf("Expected items type 'string', got %s", schema.Items.Type)
	}

	if len(schema.Items.Enum) != 3 {
		t.Errorf("Expected 3 enum values in items, got %d", len(schema.Items.Enum))
	}
}

func TestMapFieldToSchemaRelationField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "user_id",
		Type: "relation",
		Options: map[string]any{
			"collectionId": "users",
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if !strings.Contains(schema.Description, "Related record ID") {
		t.Error("Expected description to mention related record ID")
	}

	if !strings.Contains(schema.Description, "users") {
		t.Error("Expected description to mention collection reference")
	}
}

func TestMapFieldToSchemaMultiRelationField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "categories",
		Type: "relation",
		Options: map[string]any{
			"collectionId": "categories",
			"maxSelect":    5,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "array" {
		t.Errorf("Expected type 'array', got %s", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected items to be set for array type")
	}

	if schema.Items.Type != "string" {
		t.Errorf("Expected items type 'string', got %s", schema.Items.Type)
	}
}

func TestMapFieldToSchemaFileField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "avatar",
		Type: "file",
		Options: map[string]any{
			"maxSize":   1048576, // 1MB
			"mimeTypes": []any{"image/jpeg", "image/png"},
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Format != "binary" {
		t.Errorf("Expected format 'binary', got %s", schema.Format)
	}

	if !strings.Contains(schema.Description, "max size") {
		t.Error("Expected description to mention max size")
	}

	if !strings.Contains(schema.Description, "allowed types") {
		t.Error("Expected description to mention allowed types")
	}
}

func TestMapFieldToSchemaMultiFileField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "attachments",
		Type: "file",
		Options: map[string]any{
			"maxSelect": 5,
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "array" {
		t.Errorf("Expected type 'array', got %s", schema.Type)
	}

	if schema.Items == nil {
		t.Fatal("Expected items to be set for array type")
	}

	if schema.Items.Format != "binary" {
		t.Errorf("Expected items format 'binary', got %s", schema.Items.Format)
	}
}

func TestMapFieldToSchemaJsonField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "metadata",
		Type: "json",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	if !strings.Contains(schema.Description, "JSON object") {
		t.Error("Expected description to mention JSON object")
	}
}

func TestMapFieldToSchemaEditorField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "content",
		Type: "editor",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.MaxLength == nil || *schema.MaxLength != 10000 {
		t.Errorf("Expected MaxLength 10000, got %v", schema.MaxLength)
	}

	if !strings.Contains(schema.Description, "Rich text") {
		t.Error("Expected description to mention rich text")
	}
}

func TestMapFieldToSchemaUnknownField(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "unknown_field",
		Type: "unknown_type",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if !strings.Contains(schema.Description, "Unknown field type") {
		t.Error("Expected description to mention unknown field type")
	}
}

func TestMapFieldToSchemaEmptyFieldName(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "",
		Type: "text",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err == nil {
		t.Error("Expected error for empty field name, got nil")
	}

	if schema != nil {
		t.Error("Expected nil schema for empty field name")
	}
}

func TestGetSystemFieldSchemas(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	schemas := mapper.GetSystemFieldSchemas()

	if len(schemas) != 3 {
		t.Errorf("Expected 3 system field schemas, got %d", len(schemas))
	}

	expectedFields := []string{"id", "created", "updated"}
	for _, fieldName := range expectedFields {
		schema, exists := schemas[fieldName]
		if !exists {
			t.Errorf("Expected system field schema for %s", fieldName)
			continue
		}

		if !schema.Required {
			t.Errorf("Expected system field %s to be required", fieldName)
		}

		if schema.Example == nil {
			t.Errorf("Expected system field %s to have example", fieldName)
		}
	}

	// Check specific field types
	if schemas["id"].Type != "string" {
		t.Error("Expected id field to be string type")
	}

	if schemas["created"].Format != "date-time" {
		t.Error("Expected created field to have date-time format")
	}

	if schemas["updated"].Format != "date-time" {
		t.Error("Expected updated field to have date-time format")
	}
}

func TestGetFallbackSchema(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	schema := mapper.GetFallbackSchema("unknown_type")

	if schema.Type != "string" {
		t.Errorf("Expected fallback type 'string', got %s", schema.Type)
	}

	if !strings.Contains(schema.Description, "Unknown field type") {
		t.Error("Expected fallback description to mention unknown field type")
	}

	if schema.Required {
		t.Error("Expected fallback schema to not be required")
	}

	if schema.Example != "unknown_value" {
		t.Errorf("Expected fallback example 'unknown_value', got %v", schema.Example)
	}
}

func TestMapFieldWithoutExamples(t *testing.T) {
	mapper := NewFieldSchemaMapperWithConfig(false, true)

	field := FieldInfo{
		Name: "title",
		Type: "text",
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Example != nil {
		t.Error("Expected no example when includeExamples is false")
	}
}

func TestParseIntOption(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	tests := []struct {
		name     string
		value    any
		expected int
		hasError bool
	}{
		{"int value", 42, 42, false},
		{"float64 value", 42.0, 42, false},
		{"string value", "42", 42, false},
		{"invalid string", "not_a_number", 0, true},
		{"nil value", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper.parseIntOption(tt.value)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, result)
				}
			}
		})
	}
}

func TestParseFloatOption(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	tests := []struct {
		name     string
		value    any
		expected float64
		hasError bool
	}{
		{"float64 value", 42.5, 42.5, false},
		{"int value", 42, 42.0, false},
		{"string value", "42.5", 42.5, false},
		{"invalid string", "not_a_number", 0, true},
		{"nil value", nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper.parseFloatOption(tt.value)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %f, got %f", tt.expected, result)
				}
			}
		})
	}
}
func TestRelationFieldExampleSingleRelation(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "user",
		Type: "relation",
		Options: map[string]any{
			"maxSelect":    1,
			"collectionId": "users",
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Example != "RELATION_RECORD_ID" {
		t.Errorf("Expected 'RELATION_RECORD_ID', got %v", schema.Example)
	}
}

func TestRelationFieldExampleMultiRelation(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "categories",
		Type: "relation",
		Options: map[string]any{
			"maxSelect":    5,
			"collectionId": "categories",
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "array" {
		t.Errorf("Expected type 'array', got %s", schema.Type)
	}

	expected := []any{"RELATION_RECORD_ID"}
	if !reflect.DeepEqual(schema.Example, expected) {
		t.Errorf("Expected %v, got %v", expected, schema.Example)
	}
}

func TestRelationFieldExampleNoMaxSelect(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	field := FieldInfo{
		Name: "owner",
		Type: "relation",
		Options: map[string]any{
			"collectionId": "users",
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Type != "string" {
		t.Errorf("Expected type 'string', got %s", schema.Type)
	}

	if schema.Example != "RELATION_RECORD_ID" {
		t.Errorf("Expected 'RELATION_RECORD_ID' for default single relation, got %v", schema.Example)
	}
}

func TestRelationFieldExampleEdgeCases(t *testing.T) {
	testCases := []struct {
		name            string
		field           FieldInfo
		expectedType    string
		expectedExample any
	}{
		{
			name: "zero maxSelect",
			field: FieldInfo{
				Name: "optional_relation",
				Type: "relation",
				Options: map[string]any{
					"maxSelect":    0,
					"collectionId": "users",
				},
			},
			expectedType:    "string",
			expectedExample: "RELATION_RECORD_ID",
		},
		{
			name: "negative maxSelect",
			field: FieldInfo{
				Name: "invalid_relation",
				Type: "relation",
				Options: map[string]any{
					"maxSelect":    -1,
					"collectionId": "users",
				},
			},
			expectedType:    "string",
			expectedExample: "RELATION_RECORD_ID",
		},
		{
			name: "string maxSelect",
			field: FieldInfo{
				Name: "string_max_select",
				Type: "relation",
				Options: map[string]any{
					"maxSelect":    "invalid",
					"collectionId": "users",
				},
			},
			expectedType:    "string",
			expectedExample: "RELATION_RECORD_ID",
		},
		{
			name: "nil options",
			field: FieldInfo{
				Name:    "no_options",
				Type:    "relation",
				Options: nil,
			},
			expectedType:    "string",
			expectedExample: "RELATION_RECORD_ID",
		},
		{
			name: "maxSelect exactly 1",
			field: FieldInfo{
				Name: "exactly_one",
				Type: "relation",
				Options: map[string]any{
					"maxSelect":    1,
					"collectionId": "users",
				},
			},
			expectedType:    "string",
			expectedExample: "RELATION_RECORD_ID",
		},
		{
			name: "maxSelect exactly 2",
			field: FieldInfo{
				Name: "exactly_two",
				Type: "relation",
				Options: map[string]any{
					"maxSelect":    2,
					"collectionId": "tags",
				},
			},
			expectedType:    "array",
			expectedExample: []any{"RELATION_RECORD_ID"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mapper := NewFieldSchemaMapper()
			schema, err := mapper.MapFieldToSchema(tc.field)

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if schema.Type != tc.expectedType {
				t.Errorf("Expected type '%s', got '%s'", tc.expectedType, schema.Type)
			}

			if !reflect.DeepEqual(schema.Example, tc.expectedExample) {
				t.Errorf("Expected example %v, got %v", tc.expectedExample, schema.Example)
			}
		})
	}
}

func TestRelationFieldExampleWithoutExamplesEnabled(t *testing.T) {
	mapper := NewFieldSchemaMapperWithConfig(false, true)

	field := FieldInfo{
		Name: "user",
		Type: "relation",
		Options: map[string]any{
			"maxSelect":    1,
			"collectionId": "users",
		},
	}

	schema, err := mapper.MapFieldToSchema(field)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Example != nil {
		t.Error("Expected no example when includeExamples is false")
	}
}

func TestNonRelationFieldExamplesUnchanged(t *testing.T) {
	mapper := NewFieldSchemaMapper()

	// Test that non-relation fields still get their original examples
	textField := FieldInfo{
		Name: "title",
		Type: "text",
	}

	schema, err := mapper.MapFieldToSchema(textField)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if schema.Example != "example_title" {
		t.Errorf("Expected 'example_title', got %v", schema.Example)
	}

	// Test email field
	emailField := FieldInfo{
		Name: "email",
		Type: "email",
	}

	emailSchema, err := mapper.MapFieldToSchema(emailField)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if emailSchema.Example != "user@example.com" {
		t.Errorf("Expected 'user@example.com', got %v", emailSchema.Example)
	}
}
