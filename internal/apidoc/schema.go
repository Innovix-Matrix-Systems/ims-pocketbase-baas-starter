package apidoc

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// FieldSchema represents an OpenAPI field schema definition
type FieldSchema struct {
	Type        string         `json:"type"`
	Format      string         `json:"format,omitempty"`
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required"`
	Enum        []any          `json:"enum,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	Items       *FieldSchema   `json:"items,omitempty"`
	Minimum     *float64       `json:"minimum,omitempty"`
	Maximum     *float64       `json:"maximum,omitempty"`
	MinLength   *int           `json:"minLength,omitempty"`
	MaxLength   *int           `json:"maxLength,omitempty"`
	Pattern     string         `json:"pattern,omitempty"`
	Example     any            `json:"example,omitempty"`
}

// FieldSchemaMapper handles mapping PocketBase field types to OpenAPI schemas
type FieldSchemaMapper struct {
	// Configuration for schema generation
	includeExamples  bool
	strictValidation bool
}

// SchemaMapper interface for field schema mapping
type SchemaMapper interface {
	MapFieldToSchema(field FieldInfo) (*FieldSchema, error)
	GetSystemFieldSchemas() map[string]*FieldSchema
	GetFallbackSchema(fieldType string) *FieldSchema
}

// NewFieldSchemaMapper creates a new field schema mapper
func NewFieldSchemaMapper() *FieldSchemaMapper {
	return &FieldSchemaMapper{
		includeExamples:  true,
		strictValidation: true,
	}
}

const (
	DefaultEditorMaxLength   = 10000
	DefaultPasswordMinLength = 8
)

// NewFieldSchemaMapperWithConfig creates a new field schema mapper with configuration
func NewFieldSchemaMapperWithConfig(includeExamples, strictValidation bool) *FieldSchemaMapper {
	return &FieldSchemaMapper{
		includeExamples:  includeExamples,
		strictValidation: strictValidation,
	}
}

// MapFieldToSchema converts a PocketBase field to an OpenAPI schema with optimized memory allocation
func (fsm *FieldSchemaMapper) MapFieldToSchema(field FieldInfo) (*FieldSchema, error) {
	if field.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}

	schema := &FieldSchema{
		Required:    field.Required,
		Description: fsm.generateFieldDescription(field),
	}

	// Map field type to OpenAPI schema
	switch strings.ToLower(field.Type) {
	case "text":
		fsm.mapTextField(field, schema)
	case "number":
		fsm.mapNumberField(field, schema)
	case "bool", "boolean":
		fsm.mapBooleanField(field, schema)
	case "email":
		fsm.mapEmailField(field, schema)
	case "url":
		fsm.mapUrlField(field, schema)
	case "date":
		fsm.mapDateField(field, schema)
	case "select":
		fsm.mapSelectField(field, schema)
	case "relation":
		fsm.mapRelationField(field, schema)
	case "file":
		fsm.mapFileField(field, schema)
	case "json":
		fsm.mapJsonField(field, schema)
	case "editor":
		fsm.mapEditorField(field, schema)
	case "autodate":
		fsm.mapAutodateField(field, schema)
	case "password":
		fsm.mapPasswordField(field, schema)
	default:
		// If we can't access the app instance, we'll just comment out the log for now
		// log.Printf("Warning: Unknown field type '%s' for field '%s', using fallback", field.Type, field.Name)
		fsm.mapUnknownField(field, schema)
	}

	// Apply validation constraints from field options
	fsm.applyValidationConstraints(field, schema)

	// Add example if enabled
	if fsm.includeExamples {
		fsm.addFieldExample(field, schema)
	}

	return schema, nil
}

// mapTextField maps text field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapTextField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"

	// Check for specific text field options
	if options := field.Options; options != nil {
		// Handle max length
		if maxVal, ok := options["max"]; ok {
			if max, err := fsm.parseIntOption(maxVal); err == nil && max > 0 {
				schema.MaxLength = &max
			}
		}

		// Handle min length
		if minVal, ok := options["min"]; ok {
			if min, err := fsm.parseIntOption(minVal); err == nil && min > 0 {
				schema.MinLength = &min
			}
		}

		// Handle pattern
		if patternVal, ok := options["pattern"]; ok {
			if pattern, ok := patternVal.(string); ok && pattern != "" {
				schema.Pattern = pattern
			}
		}
	}
}

// mapNumberField maps number field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapNumberField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "number"

	if options := field.Options; options != nil {
		// Handle maximum value
		if maxVal, ok := options["max"]; ok {
			if max, err := fsm.parseFloatOption(maxVal); err == nil {
				schema.Maximum = &max
			}
		}

		// Handle minimum value
		if minVal, ok := options["min"]; ok {
			if min, err := fsm.parseFloatOption(minVal); err == nil {
				schema.Minimum = &min
			}
		}

		// Check if it should be integer
		if onlyInt, ok := options["onlyInt"]; ok {
			if isInt, ok := onlyInt.(bool); ok && isInt {
				schema.Type = "integer"
			}
		}
	}
}

// mapBooleanField maps boolean field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapBooleanField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "boolean"
}

// mapEmailField maps email field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapEmailField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Format = "email"
}

// mapUrlField maps URL field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapUrlField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Format = "uri"
}

// mapDateField maps date field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapDateField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Format = "date-time"
}

// mapSelectField maps select field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapSelectField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"

	if options := field.Options; options != nil {
		// Handle enum values
		if valuesVal, ok := options["values"]; ok {
			if values, ok := valuesVal.([]any); ok && len(values) > 0 {
				schema.Enum = values
			} else if valuesStr, ok := valuesVal.(string); ok && valuesStr != "" {
				// Handle comma-separated values
				valuesList := strings.Split(valuesStr, ",")
				schema.Enum = make([]any, len(valuesList))
				for i, v := range valuesList {
					schema.Enum[i] = strings.TrimSpace(v)
				}
			}
		}

		// Handle max select (for multi-select)
		if maxSelectVal, ok := options["maxSelect"]; ok {
			if maxSelect, err := fsm.parseIntOption(maxSelectVal); err == nil && maxSelect > 1 {
				// Multi-select field
				schema.Type = "array"
				schema.Items = &FieldSchema{
					Type: "string",
					Enum: schema.Enum,
				}
				schema.Enum = nil // Move enum to items
			}
		}
	}
}

// mapRelationField maps relation field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapRelationField(field FieldInfo, schema *FieldSchema) {
	if options := field.Options; options != nil {
		// Handle max select (for multi-relation)
		if maxSelectVal, ok := options["maxSelect"]; ok {
			if maxSelect, err := fsm.parseIntOption(maxSelectVal); err == nil && maxSelect > 1 {
				// Multi-relation field
				schema.Type = "array"
				schema.Items = &FieldSchema{
					Type:        "string",
					Description: "Related record ID",
				}
			} else {
				// Single relation field
				schema.Type = "string"
				schema.Description = "Related record ID"
			}
		} else {
			// Default to single relation
			schema.Type = "string"
			schema.Description = "Related record ID"
		}

		// Add collection reference if available
		if collectionIdVal, ok := options["collectionId"]; ok {
			if collectionId, ok := collectionIdVal.(string); ok && collectionId != "" {
				schema.Description += fmt.Sprintf(" (references collection: %s)", collectionId)
			}
		}
	} else {
		schema.Type = "string"
		schema.Description = "Related record ID"
	}
}

// mapFileField maps file field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapFileField(field FieldInfo, schema *FieldSchema) {
	if options := field.Options; options != nil {
		// Handle max select (for multiple files)
		if maxSelectVal, ok := options["maxSelect"]; ok {
			if maxSelect, err := fsm.parseIntOption(maxSelectVal); err == nil && maxSelect > 1 {
				// Multiple files
				schema.Type = "array"
				schema.Items = &FieldSchema{
					Type:   "string",
					Format: "binary",
				}
			} else {
				// Single file
				schema.Type = "string"
				schema.Format = "binary"
			}
		} else {
			// Default to single file
			schema.Type = "string"
			schema.Format = "binary"
		}

		// Add file size constraints
		if maxSizeVal, ok := options["maxSize"]; ok {
			if maxSize, err := fsm.parseIntOption(maxSizeVal); err == nil && maxSize > 0 {
				schema.Description += fmt.Sprintf(" (max size: %d bytes)", maxSize)
			}
		}

		// Add mime type constraints
		if mimeTypesVal, ok := options["mimeTypes"]; ok {
			if mimeTypes, ok := mimeTypesVal.([]any); ok && len(mimeTypes) > 0 {
				var types []string
				for _, mt := range mimeTypes {
					if mimeType, ok := mt.(string); ok {
						types = append(types, mimeType)
					}
				}
				if len(types) > 0 {
					schema.Description += fmt.Sprintf(" (allowed types: %s)", strings.Join(types, ", "))
				}
			}
		}
	} else {
		schema.Type = "string"
		schema.Format = "binary"
	}
}

// mapJsonField maps JSON field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapJsonField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "object"
	schema.Description = "JSON object"
}

// mapEditorField maps editor field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapEditorField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Description = "Rich text content"

	// Editor fields are typically longer
	if schema.MaxLength == nil {
		maxLen := DefaultEditorMaxLength // Default max length for editor fields
		schema.MaxLength = &maxLen
	}
}

// mapAutodateField maps autodate field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapAutodateField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Format = "date-time"
	schema.Description = "Auto-generated timestamp"

	// Check if it's a create or update timestamp
	if strings.ToLower(field.Name) == "created" {
		schema.Description = "Record creation timestamp (auto-generated)"
	} else if strings.ToLower(field.Name) == "updated" {
		schema.Description = "Record last update timestamp (auto-generated)"
	}
}

// mapPasswordField maps password field types to OpenAPI schema
func (fsm *FieldSchemaMapper) mapPasswordField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Format = "password"
	schema.Description = "Password field (write-only)"

	// Password fields typically have minimum length requirements
	if schema.MinLength == nil {
		minLen := DefaultPasswordMinLength // Default minimum length for passwords
		schema.MinLength = &minLen
	}
}

// mapUnknownField maps unknown field types to a fallback schema
func (fsm *FieldSchemaMapper) mapUnknownField(field FieldInfo, schema *FieldSchema) {
	schema.Type = "string"
	schema.Description = fmt.Sprintf("Unknown field type: %s", field.Type)
}

// applyValidationConstraints applies validation constraints from field options
func (fsm *FieldSchemaMapper) applyValidationConstraints(field FieldInfo, schema *FieldSchema) {
	if !fsm.strictValidation || field.Options == nil {
		return
	}

	// Apply additional validation constraints based on field options
	if requiredVal, ok := field.Options["required"]; ok {
		if required, ok := requiredVal.(bool); ok {
			schema.Required = required
		}
	}
}

// addFieldExample adds example values to the schema
func (fsm *FieldSchemaMapper) addFieldExample(field FieldInfo, schema *FieldSchema) {
	// Check if this is a relation field and generate relation-specific examples
	if field.Type == "relation" {
		fsm.addRelationFieldExample(field, schema)
		return
	}

	// Generate appropriate examples based on field type for non-relation fields
	fsm.addGenericFieldExample(field, schema)
}

// addRelationFieldExample generates examples specifically for relation fields
func (fsm *FieldSchemaMapper) addRelationFieldExample(field FieldInfo, schema *FieldSchema) {
	switch schema.Type {
	case "string":
		// Single relation field
		schema.Example = "RELATION_RECORD_ID"
	case "array":
		// Multi-relation field
		schema.Example = []any{"RELATION_RECORD_ID"}
	default:
		// If we can't access the app instance, we'll just comment out the log for now
		// log.Printf("Warning: Unexpected schema type %s for relation field %s, using fallback", schema.Type, field.Name)
		fsm.addGenericFieldExample(field, schema)
	}
}

// addGenericFieldExample generates examples for non-relation fields
func (fsm *FieldSchemaMapper) addGenericFieldExample(field FieldInfo, schema *FieldSchema) {
	switch schema.Type {
	case "string":
		if schema.Format == "email" {
			schema.Example = "user@example.com"
		} else if schema.Format == "uri" {
			schema.Example = "https://example.com"
		} else if schema.Format == "date-time" {
			schema.Example = "2024-01-01T12:00:00Z"
		} else if schema.Format == "password" {
			schema.Example = "********" // Don't show actual password examples
		} else if len(schema.Enum) > 0 {
			schema.Example = schema.Enum[0]
		} else {
			schema.Example = fmt.Sprintf("example_%s", field.Name)
		}
	case "number", "integer":
		if schema.Minimum != nil {
			schema.Example = *schema.Minimum + 1
		} else if schema.Maximum != nil {
			schema.Example = *schema.Maximum - 1
		} else {
			schema.Example = 42
		}
	case "boolean":
		schema.Example = true
	case "array":
		if schema.Items != nil && len(schema.Items.Enum) > 0 {
			schema.Example = []any{schema.Items.Enum[0]}
		} else {
			schema.Example = []any{"example_item"}
		}
	case "object":
		schema.Example = map[string]any{
			"key": "value",
		}
	}
}

// generateFieldDescription generates a description for the field
func (fsm *FieldSchemaMapper) generateFieldDescription(field FieldInfo) string {
	description := ""

	// Use description from options if available
	if field.Options != nil {
		if descVal, ok := field.Options["description"]; ok {
			if desc, ok := descVal.(string); ok && desc != "" {
				description = desc
			}
		}
	}

	// Generate default description if none provided
	if description == "" {
		caser := cases.Title(language.English)
		description = fmt.Sprintf("%s field", caser.String(field.Type))
	}

	// Add system field indicator
	if field.System {
		description += " (system field)"
	}

	return description
}

// GetSystemFieldSchemas returns schemas for standard system fields
func (fsm *FieldSchemaMapper) GetSystemFieldSchemas() map[string]*FieldSchema {
	return map[string]*FieldSchema{
		"id": {
			Type:        "string",
			Description: "Unique record identifier",
			Required:    true,
			Example:     "abc123def456",
		},
		"created": {
			Type:        "string",
			Format:      "date-time",
			Description: "Record creation timestamp",
			Required:    true,
			Example:     "2024-01-01T12:00:00Z",
		},
		"updated": {
			Type:        "string",
			Format:      "date-time",
			Description: "Record last update timestamp",
			Required:    true,
			Example:     "2024-01-01T12:00:00Z",
		},
	}
}

// GetFallbackSchema returns a fallback schema for unknown field types
func (fsm *FieldSchemaMapper) GetFallbackSchema(fieldType string) *FieldSchema {
	return &FieldSchema{
		Type:        "string",
		Description: fmt.Sprintf("Unknown field type: %s", fieldType),
		Required:    false,
		Example:     "unknown_value",
	}
}

// Helper methods for parsing options

// parseIntOption safely parses an integer option value
func (fsm *FieldSchemaMapper) parseIntOption(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot parse %v as int", value)
	}
}

// parseFloatOption safely parses a float option value
func (fsm *FieldSchemaMapper) parseFloatOption(value any) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot parse %v as float", value)
	}
}
