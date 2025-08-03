package swagger

import (
	"fmt"
	"log"
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
	Example    map[string]interface{}  `json:"example,omitempty"`
}

// SchemaGen interface for schema generation
type SchemaGen interface {
	GenerateCollectionSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error)
	GenerateCollectionSchemas(collections []EnhancedCollectionInfo) (map[string]*CollectionSchema, error)
	GenerateCreateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error)
	GenerateUpdateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error)
	GenerateListResponseSchema(collection EnhancedCollectionInfo) (map[string]interface{}, error)
	GetSchemaName(collection EnhancedCollectionInfo) string
	GetCreateSchemaName(collection EnhancedCollectionInfo) string
	GetUpdateSchemaName(collection EnhancedCollectionInfo) string
	GetListResponseSchemaName(collection EnhancedCollectionInfo) string
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
func (sg *SchemaGenerator) GenerateCollectionSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
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
			log.Printf("Warning: Failed to map field %s in collection %s: %v", field.Name, collection.Name, err)
			// Use fallback schema
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
func (sg *SchemaGenerator) GenerateCollectionSchemas(collections []EnhancedCollectionInfo) (map[string]*CollectionSchema, error) {
	schemas := make(map[string]*CollectionSchema)

	for _, collection := range collections {
		schema, err := sg.GenerateCollectionSchema(collection)
		if err != nil {
			log.Printf("Warning: Failed to generate schema for collection %s: %v", collection.Name, err)
			continue
		}
		schemas[collection.Name] = schema
	}

	return schemas, nil
}

// GenerateCreateSchema generates a schema for creating records (excludes system fields)
func (sg *SchemaGenerator) GenerateCreateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
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
		// Skip system fields in create schema
		if field.System {
			continue
		}

		fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
		if err != nil {
			log.Printf("Warning: Failed to map field %s in collection %s: %v", field.Name, collection.Name, err)
			fieldSchema = sg.fieldMapper.GetFallbackSchema(field.Type)
		}

		schema.Properties[field.Name] = fieldSchema
		if fieldSchema.Required {
			schema.Required = append(schema.Required, field.Name)
		}
	}

	// Generate example if enabled
	if sg.includeExamples {
		schema.Example = sg.generateCreateExample(collection, schema)
	}

	return schema, nil
}

// GenerateUpdateSchema generates a schema for updating records (all fields optional)
func (sg *SchemaGenerator) GenerateUpdateSchema(collection EnhancedCollectionInfo) (*CollectionSchema, error) {
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
		// Skip system fields in update schema
		if field.System {
			continue
		}

		fieldSchema, err := sg.fieldMapper.MapFieldToSchema(field)
		if err != nil {
			log.Printf("Warning: Failed to map field %s in collection %s: %v", field.Name, collection.Name, err)
			fieldSchema = sg.fieldMapper.GetFallbackSchema(field.Type)
		}

		// Make field optional for updates
		fieldSchema.Required = false
		schema.Properties[field.Name] = fieldSchema
	}

	// Generate example if enabled
	if sg.includeExamples {
		schema.Example = sg.generateUpdateExample(collection, schema)
	}

	return schema, nil
}

// GenerateListResponseSchema generates a schema for list responses with pagination
func (sg *SchemaGenerator) GenerateListResponseSchema(collection EnhancedCollectionInfo) (map[string]interface{}, error) {
	// Generate the item schema
	itemSchema, err := sg.GenerateCollectionSchema(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to generate item schema: %w", err)
	}

	// Create the list response schema
	listSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"page": map[string]interface{}{
				"type":        "integer",
				"description": "Current page number",
				"example":     1,
			},
			"perPage": map[string]interface{}{
				"type":        "integer",
				"description": "Number of items per page",
				"example":     30,
			},
			"totalItems": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of items",
				"example":     100,
			},
			"totalPages": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of pages",
				"example":     4,
			},
			"items": map[string]interface{}{
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
func (sg *SchemaGenerator) generateCollectionExample(collection EnhancedCollectionInfo, schema *CollectionSchema) map[string]interface{} {
	example := make(map[string]interface{})

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
func (sg *SchemaGenerator) generateCreateExample(collection EnhancedCollectionInfo, schema *CollectionSchema) map[string]interface{} {
	example := make(map[string]interface{})

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
func (sg *SchemaGenerator) generateUpdateExample(collection EnhancedCollectionInfo, schema *CollectionSchema) map[string]interface{} {
	example := make(map[string]interface{})

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
func (sg *SchemaGenerator) generateListResponseExample(collection EnhancedCollectionInfo, itemSchema *CollectionSchema) map[string]interface{} {
	// Generate a couple of item examples
	items := []interface{}{}
	if itemSchema.Example != nil {
		items = append(items, itemSchema.Example)

		// Create a second example with slight variations
		if secondExample := sg.createVariationExample(itemSchema.Example); secondExample != nil {
			items = append(items, secondExample)
		}
	}

	return map[string]interface{}{
		"page":       1,
		"perPage":    30,
		"totalItems": len(items),
		"totalPages": 1,
		"items":      items,
	}
}

// generateBasicExample generates a basic example value based on field schema
func (sg *SchemaGenerator) generateBasicExample(fieldSchema *FieldSchema, fieldName string) interface{} {
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
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
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
	commonFields := []string{"name", "title", "description", "email", "status"}
	for _, common := range commonFields {
		if strings.Contains(strings.ToLower(fieldName), common) {
			return true
		}
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
func (sg *SchemaGenerator) createVariationExample(original interface{}) interface{} {
	if originalMap, ok := original.(map[string]interface{}); ok {
		variation := make(map[string]interface{})
		for key, value := range originalMap {
			variation[key] = sg.createVariationValue(value, key)
		}
		return variation
	}
	return nil
}

// createVariationValue creates a variation of a single value
func (sg *SchemaGenerator) createVariationValue(value interface{}, key string) interface{} {
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
func (sg *SchemaGenerator) GetSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name
}

// GetCreateSchemaName returns the schema name for create operations
func (sg *SchemaGenerator) GetCreateSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name + "Create"
}

// GetUpdateSchemaName returns the schema name for update operations
func (sg *SchemaGenerator) GetUpdateSchemaName(collection EnhancedCollectionInfo) string {
	return collection.Name + "Update"
}

// GetListResponseSchemaName returns the schema name for list responses
func (sg *SchemaGenerator) GetListResponseSchemaName(collection EnhancedCollectionInfo) string {
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
