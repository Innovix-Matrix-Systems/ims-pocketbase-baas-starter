package apidoc

import (
	"fmt"
	"strings"
)

// SchemaGenerator handles OpenAPI schema generation from collection metadata
type SchemaGenerator struct {
	fieldMapper     SchemaMapper
	includeExamples bool
	includeSystem   bool
}

// CollectionSchema represents a complete OpenAPI schema for a collection
type CollectionSchema struct {
	Type       string                  `json:"type"`
	Properties map[string]*FieldSchema `json:"properties"`
	Required   []string                `json:"required"`
	Example    map[string]any          `json:"example,omitempty"`
}

// SchemaGen interface for schema generation
type SchemaGen interface {
	GenerateCollectionSchema(collection CollectionInfo) (*CollectionSchema, error)
	GenerateCollectionSchemas(collections []CollectionInfo) (map[string]*CollectionSchema, error)
	GenerateCreateSchema(collection CollectionInfo) (*CollectionSchema, error)
	GenerateUpdateSchema(collection CollectionInfo) (*CollectionSchema, error)
	GenerateListResponseSchema(collection CollectionInfo) (map[string]any, error)
	GetSchemaName(collection CollectionInfo) string
	GetCreateSchemaName(collection CollectionInfo) string
	GetUpdateSchemaName(collection CollectionInfo) string
	GetListResponseSchemaName(collection CollectionInfo) string
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator() *SchemaGenerator {
	return &SchemaGenerator{
		fieldMapper:     NewFieldSchemaMapper(),
		includeExamples: true,
		includeSystem:   true,
	}
}

// NewSchemaGeneratorWithConfig creates a new schema generator with configuration
func NewSchemaGeneratorWithConfig(includeExamples, includeSystem bool) *SchemaGenerator {
	return &SchemaGenerator{
		fieldMapper:     NewFieldSchemaMapperWithConfig(includeExamples, true),
		includeExamples: includeExamples,
		includeSystem:   includeSystem,
	}
}

// GenerateCollectionSchema generates a complete OpenAPI schema for a collection
func (sg *SchemaGenerator) GenerateCollectionSchema(collection CollectionInfo) (*CollectionSchema, error) {
	if collection.Name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	schema := &CollectionSchema{
		Type:       "object",
		Properties: make(map[string]*FieldSchema),
		Required:   []string{},
	}

	// Add system fields if enabled
	if sg.includeSystem {
		systemFields := sg.fieldMapper.GetSystemFieldSchemas()
		for fieldName, fieldSchema := range systemFields {
			schema.Properties[fieldName] = fieldSchema
			if fieldSchema.Required {
				schema.Required = append(schema.Required, fieldName)
			}
		}
	}

	// Add collection fields
	for _, field := range collection.Fields {
		fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
		if err != nil {
			// If we can't access the app instance, we'll just comment out the log for now
			fieldSchema = sg.fieldMapper.GetFallbackSchema(field.Type)
		}

		schema.Properties[field.Name] = fieldSchema
		if fieldSchema.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	// Generate example if enabled
	if sg.includeExamples {
		schema.Example = sg.generateCollectionExample(collection, schema)
	}

	return schema, nil
}

// GenerateCollectionSchemas generates schemas for multiple collections
func (sg *SchemaGenerator) GenerateCollectionSchemas(collections []CollectionInfo) (map[string]*CollectionSchema, error) {
	schemas := make(map[string]*CollectionSchema)

	for _, collection := range collections {
		schema, err := sg.GenerateCollectionSchema(collection)
		if err != nil {
			// If we can't access the app instance, we'll just comment out the log for now
			continue
		}
		schemas[collection.Name] = schema
	}

	return schemas, nil
}

// GenerateCreateSchema generates a schema for creating records (excludes system fields)
func (sg *SchemaGenerator) GenerateCreateSchema(collection CollectionInfo) (*CollectionSchema, error) {
	if collection.Name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	schema := &CollectionSchema{
		Type:       "object",
		Properties: make(map[string]*FieldSchema),
		Required:   []string{},
	}

	// Only add collection fields (no system fields for create operations)
	for _, field := range collection.Fields {

		// Skip system fields in create schema, except for password and email fields in auth collections
		if field.System && field.Type != "password" && !(collection.Type == "auth" && field.Type == "email") {
			continue
		}

		// Also skip common system fields by name (created, updated, id)
		if field.Name == "created" || field.Name == "updated" || field.Name == "id" {
			continue
		}

		fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
		if err != nil {
			// If we can't access the app instance, we'll just comment out the log for now
			// log.Printf("Warning: Failed to map field %s in collection %s: %v", field.Name, collection.Name, err)
			fieldSchema = sg.fieldMapper.GetFallbackSchema(field.Type)
		}

		schema.Properties[field.Name] = fieldSchema
		if fieldSchema.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	// For auth collections, add passwordConfirm field if password field exists
	if collection.Type == "auth" {
		if _, hasPassword := schema.Properties["password"]; hasPassword {
			passwordConfirmSchema := &FieldSchema{
				Type:        "string",
				Format:      "password",
				Description: "Password confirmation (must match password)",
				Required:    true,
			}
			schema.Properties["passwordConfirm"] = passwordConfirmSchema
			schema.Required = append(schema.Required, "passwordConfirm")
		}
	}

	// Special case: For auth collections, ensure email field is included if it exists
	if collection.Type == "auth" {
		for _, field := range collection.Fields {
			if field.Type == "email" && field.Name == "email" {
				fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
				if err != nil {
					// If we can't access the app instance, we'll just comment out the log for now
					fieldSchema = &FieldSchema{
						Type:        "string",
						Format:      "email",
						Description: "Email address",
						Required:    field.Required,
						Example:     "user@example.com",
					}
				}
				schema.Properties["email"] = fieldSchema
				if fieldSchema.Required {
					if !contains(schema.Required, "email") {
						schema.Required = append(schema.Required, "email")
					}
				}
				break
			}
		}
	}

	// Generate example if enabled
	if sg.includeExamples {
		schema.Example = sg.generateCreateExample(collection, schema)
	}

	return schema, nil
}

// GenerateUpdateSchema generates a schema for updating records (all fields optional)
func (sg *SchemaGenerator) GenerateUpdateSchema(collection CollectionInfo) (*CollectionSchema, error) {
	if collection.Name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	schema := &CollectionSchema{
		Type:       "object",
		Properties: make(map[string]*FieldSchema),
		Required:   []string{}, // No required fields for updates
	}

	// Add collection fields (no system fields for update operations)
	for _, field := range collection.Fields {
		// Skip system fields in update schema, except for password and email fields in auth collections
		if field.System && field.Type != "password" && !(collection.Type == "auth" && field.Type == "email") {
			continue
		}

		// Also skip common system fields by name (created, updated, id)
		if field.Name == "created" || field.Name == "updated" || field.Name == "id" {
			continue
		}

		fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
		if err != nil {
			// If we can't access the app instance, we'll just comment out the log for now
			fieldSchema = sg.fieldMapper.GetFallbackSchema(field.Type)
		}

		// Make field optional for updates
		fieldSchema.Required = false
		schema.Properties[field.Name] = fieldSchema
	}

	// Special case: For auth collections, ensure email field is included if it exists
	if collection.Type == "auth" {
		for _, field := range collection.Fields {
			if field.Type == "email" && field.Name == "email" {
				fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
				if err != nil {
					// If we can't access the app instance, we'll just comment out the log for now
					fieldSchema = &FieldSchema{
						Type:        "string",
						Format:      "email",
						Description: "Email address",
						Required:    false, // Always optional for updates
						Example:     "user@example.com",
					}
				}
				fieldSchema.Required = false // Ensure it's optional for updates
				schema.Properties["email"] = fieldSchema
				break
			}
		}
	}

	// Generate example if enabled
	if sg.includeExamples {
		schema.Example = sg.generateUpdateExample(collection, schema)
	}

	return schema, nil
}

// GenerateListResponseSchema generates a schema for list responses with pagination
func (sg *SchemaGenerator) GenerateListResponseSchema(collection CollectionInfo) (map[string]any, error) {
	// Generate the item schema
	itemSchema, err := sg.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate item schema: %w", err)
	}

	// Create the list response schema
	listSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"page": map[string]any{
				"type":        "integer",
				"description": "Current page number",
				"example":     1,
			},
			"perPage": map[string]any{
				"type":        "integer",
				"description": "Number of items per page",
				"example":     30,
			},
			"totalItems": map[string]any{
				"type":        "integer",
				"description": "Total number of items",
				"example":     100,
			},
			"totalPages": map[string]any{
				"type":        "integer",
				"description": "Total number of pages",
				"example":     4,
			},
			"items": map[string]any{
				"type":        "array",
				"description": fmt.Sprintf("Array of %s records", collection.Name),
				"items":       itemSchema,
			},
		},
		"required": []string{"page", "perPage", "totalItems", "totalPages", "items"},
	}

	// Add example if enabled
	if sg.includeExamples {
		listSchema["example"] = sg.generateListResponseExample(collection, itemSchema)
	}

	return listSchema, nil
}

// generateCollectionExample generates an example object for a collection schema
func (sg *SchemaGenerator) generateCollectionExample(collection CollectionInfo, schema *CollectionSchema) map[string]any {
	example := make(map[string]any)

	for fieldName, fieldSchema := range schema.Properties {
		if fieldSchema.Example != nil {
			example[fieldName] = fieldSchema.Example
		} else {
			// Generate a basic example based on type
			example[fieldName] = sg.generateBasicExample(fieldSchema, fieldName)
		}
	}

	return example
}

// generateCreateExample generates an example object for create operations
func (sg *SchemaGenerator) generateCreateExample(collection CollectionInfo, schema *CollectionSchema) map[string]any {
	example := make(map[string]any)

	for fieldName, fieldSchema := range schema.Properties {
		// Only include required fields and some optional fields in create examples
		if fieldSchema.Required || sg.shouldIncludeInCreateExample(fieldName, fieldSchema) {
			if fieldSchema.Example != nil {
				example[fieldName] = fieldSchema.Example
			} else {
				example[fieldName] = sg.generateBasicExample(fieldSchema, fieldName)
			}
		}
	}

	return example
}

// generateUpdateExample generates an example object for update operations
func (sg *SchemaGenerator) generateUpdateExample(collection CollectionInfo, schema *CollectionSchema) map[string]any {
	example := make(map[string]any)

	// Include a subset of fields for update examples
	count := 0
	maxFields := 3 // Limit update examples to a few fields

	for fieldName, fieldSchema := range schema.Properties {
		if count >= maxFields {
			break
		}

		if sg.shouldIncludeInUpdateExample(fieldName, fieldSchema) {
			if fieldSchema.Example != nil {
				example[fieldName] = fieldSchema.Example
			} else {
				example[fieldName] = sg.generateBasicExample(fieldSchema, fieldName)
			}
			count++
		}
	}

	return example
}

// generateListResponseExample generates an example for list responses
func (sg *SchemaGenerator) generateListResponseExample(collection CollectionInfo, itemSchema *CollectionSchema) map[string]any {
	// Generate a couple of item examples
	items := []any{}
	if itemSchema.Example != nil {
		items = append(items, itemSchema.Example)

		// Create a second example with slight variations
		if secondExample := sg.createVariationExample(itemSchema.Example); secondExample != nil {
			items = append(items, secondExample)
		}
	}

	return map[string]any{
		"page":       1,
		"perPage":    30,
		"totalItems": len(items),
		"totalPages": 1,
		"items":      items,
	}
}

// generateBasicExample generates a basic example value based on field schema
func (sg *SchemaGenerator) generateBasicExample(fieldSchema *FieldSchema, fieldName string) any {
	switch fieldSchema.Type {
	case "string":
		if fieldSchema.Format == "email" {
			return "user@example.com"
		} else if fieldSchema.Format == "uri" {
			return "https://example.com"
		} else if fieldSchema.Format == "date-time" {
			return "2024-01-01T12:00:00Z"
		} else if len(fieldSchema.Enum) > 0 {
			return fieldSchema.Enum[0]
		}
		return fmt.Sprintf("example_%s", fieldName)
	case "number", "integer":
		if fieldSchema.Minimum != nil {
			return *fieldSchema.Minimum + 1
		} else if fieldSchema.Maximum != nil {
			return *fieldSchema.Maximum - 1
		}
		return 42
	case "boolean":
		return true
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	default:
		return fmt.Sprintf("example_%s", fieldName)
	}
}

// shouldIncludeInCreateExample determines if a field should be included in create examples
func (sg *SchemaGenerator) shouldIncludeInCreateExample(fieldName string, fieldSchema *FieldSchema) bool {
	// Always include required fields
	if fieldSchema.Required {
		return true
	}

	// Include some common optional fields
	commonFields := []string{"name", "title", "description", "email", "status", "active", "enabled", "password"}
	for _, common := range commonFields {
		if strings.Contains(strings.ToLower(fieldName), common) {
			return true
		}
	}

	// Include password fields (they're important for auth collections)
	if fieldSchema.Format == "password" {
		return true
	}

	// Include relation fields (they're important for understanding the API)
	if isRelationFieldSchema(fieldSchema) {
		return true
	}

	// Include boolean fields (they're usually simple and helpful)
	if fieldSchema.Type == "boolean" {
		return true
	}

	return false
}

// isRelationFieldSchema checks if a field schema represents a relation field
func isRelationFieldSchema(fieldSchema *FieldSchema) bool {
	if fieldSchema.Description != "" {
		return strings.Contains(fieldSchema.Description, "Related record ID") ||
			strings.Contains(fieldSchema.Description, "Relation field")
	}
	return false
}

// shouldIncludeInUpdateExample determines if a field should be included in update examples
func (sg *SchemaGenerator) shouldIncludeInUpdateExample(fieldName string, fieldSchema *FieldSchema) bool {
	// Include commonly updated fields
	commonFields := []string{"name", "title", "description", "status", "active", "enabled"}
	for _, common := range commonFields {
		if strings.Contains(strings.ToLower(fieldName), common) {
			return true
		}
	}

	// Include text and boolean fields as they're commonly updated
	if fieldSchema.Type == "string" || fieldSchema.Type == "boolean" {
		return true
	}

	return false
}

// createVariationExample creates a variation of an example for list responses
func (sg *SchemaGenerator) createVariationExample(original any) any {
	if originalMap, ok := original.(map[string]any); ok {
		variation := make(map[string]any)
		for key, value := range originalMap {
			variation[key] = sg.createVariationValue(value, key)
		}
		return variation
	}
	return nil
}

// createVariationValue creates a variation of a single value
func (sg *SchemaGenerator) createVariationValue(value any, key string) any {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "example_") {
			return strings.Replace(v, "example_", "sample_", 1)
		} else if v == "user@example.com" {
			return "admin@example.com"
		} else if v == "https://example.com" {
			return "https://sample.com"
		}
		return v + "_2"
	case int:
		return v + 1
	case float64:
		return v + 1.0
	case bool:
		return !v
	default:
		return v
	}
}

// GetSchemaName returns the schema name for a collection
func (sg *SchemaGenerator) GetSchemaName(collection CollectionInfo) string {
	return collection.Name
}

// GetCreateSchemaName returns the schema name for create operations
func (sg *SchemaGenerator) GetCreateSchemaName(collection CollectionInfo) string {
	return collection.Name + "Create"
}

// GetUpdateSchemaName returns the schema name for update operations
func (sg *SchemaGenerator) GetUpdateSchemaName(collection CollectionInfo) string {
	return collection.Name + "Update"
}

// GetListResponseSchemaName returns the schema name for list responses
func (sg *SchemaGenerator) GetListResponseSchemaName(collection CollectionInfo) string {
	return collection.Name + "ListResponse"
}

// ValidateSchema performs basic validation on a generated schema
func (sg *SchemaGenerator) ValidateSchema(schema *CollectionSchema) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if schema.Type != "object" {
		return fmt.Errorf("schema type must be 'object', got '%s'", schema.Type)
	}

	if schema.Properties == nil {
		return fmt.Errorf("schema properties cannot be nil")
	}

	if len(schema.Properties) == 0 {
		return fmt.Errorf("schema must have at least one property")
	}

	// Validate that all required fields exist in properties
	for _, requiredField := range schema.Required {
		if _, exists := schema.Properties[requiredField]; !exists {
			return fmt.Errorf("required field '%s' not found in properties", requiredField)
		}
	}

	return nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
