package swagger

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// CollectionDiscovery handles PocketBase collection discovery and metadata extraction
type CollectionDiscovery struct {
	app                 *pocketbase.PocketBase
	excludedCollections []string
	includeSystem       bool
}

const (
	CollectionTypeBase = "base"
	CollectionTypeAuth = "auth"
	CollectionTypeView = "view"
)

// CollectionInfo holds complete collection metadata for OpenAPI generation
type CollectionInfo struct {
	Name       string         `json:"name"`
	Type       string         `json:"type"` // "base", "auth", "view"
	System     bool           `json:"system"`
	Fields     []FieldInfo    `json:"fields"`
	ListRule   *string        `json:"listRule"`
	ViewRule   *string        `json:"viewRule"`
	CreateRule *string        `json:"createRule"`
	UpdateRule *string        `json:"updateRule"`
	DeleteRule *string        `json:"deleteRule"`
	Options    map[string]any `json:"options"`
}

// FieldInfo holds field metadata for schema generation
type FieldInfo struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Required bool           `json:"required"`
	System   bool           `json:"system"`
	Options  map[string]any `json:"options"`
}

// Discovery interface for collection discovery
type Discovery interface {
	DiscoverCollections() ([]CollectionInfo, error)
	GetCollection(name string) (*CollectionInfo, error)
	ShouldIncludeCollection(name string, collectionType string, system bool) bool
	GetCollectionStats() (map[string]int, error)
	ValidateCollectionAccess() error
	RefreshCollectionCache()
}

// NewCollectionDiscovery creates a new collection discovery service
func NewCollectionDiscovery(app *pocketbase.PocketBase, includeSystem bool) *CollectionDiscovery {
	return &CollectionDiscovery{
		app:                 app,
		excludedCollections: []string{},
		includeSystem:       includeSystem,
	}
}

// NewCollectionDiscoveryWithConfig creates a new collection discovery service with full configuration
func NewCollectionDiscoveryWithConfig(app *pocketbase.PocketBase, excludedCollections []string, includeSystem bool) *CollectionDiscovery {
	return &CollectionDiscovery{
		app:                 app,
		excludedCollections: excludedCollections,
		includeSystem:       includeSystem,
	}
}

// DiscoverCollections discovers all collections from PocketBase and returns their metadata
func (cd *CollectionDiscovery) DiscoverCollections() ([]CollectionInfo, error) {
	if cd.app == nil {
		return nil, fmt.Errorf("PocketBase app is nil")
	}

	// Use PocketBase's built-in collection finder
	collections, err := cd.app.FindAllCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to find collections: %w", err)
	}

	var collectionInfos []CollectionInfo
	var skippedCollections []string

	for _, collection := range collections {
		// Check if collection should be included
		if !cd.ShouldIncludeCollection(collection.Name, collection.Type, collection.System) {
			skippedCollections = append(skippedCollections, collection.Name)
			continue
		}

		// Extract collection metadata from PocketBase collection
		collectionInfo, err := cd.extractCollectionInfoFromPB(collection)
		if err != nil {
			log.Printf("Warning: Failed to extract metadata for collection %s: %v", collection.Name, err)
			continue
		}

		collectionInfos = append(collectionInfos, *collectionInfo)
	}

	log.Printf("Discovered %d collections, skipped %d collections", len(collectionInfos), len(skippedCollections))
	if len(skippedCollections) > 0 {
		log.Printf("Skipped collections: %s", strings.Join(skippedCollections, ", "))
	}

	return collectionInfos, nil
}

// GetCollection retrieves a specific collection by name
func (cd *CollectionDiscovery) GetCollection(name string) (*CollectionInfo, error) {
	if cd.app == nil {
		return nil, fmt.Errorf("PocketBase app is nil")
	}

	var dbCollection databaseCollection
	err := cd.app.DB().NewQuery("SELECT name, type, system, schema, listRule, viewRule, createRule, updateRule, deleteRule, options FROM _collections WHERE name = {:name}").
		Bind(map[string]any{"name": name}).
		One(&dbCollection)
	if err != nil {
		return nil, fmt.Errorf("failed to find collection %s: %w", name, err)
	}

	if !cd.ShouldIncludeCollection(dbCollection.Name, dbCollection.Type, dbCollection.System) {
		return nil, fmt.Errorf("collection %s is excluded from documentation", name)
	}

	return cd.extractCollectionInfo(dbCollection)
}

// ShouldIncludeCollection determines if a collection should be included in documentation
func (cd *CollectionDiscovery) ShouldIncludeCollection(name string, collectionType string, system bool) bool {
	if name == "" {
		return false
	}

	// Check if it's a system collection and system collections are disabled
	if system && !cd.includeSystem {
		return false
	}

	// Check excluded collections list
	for _, excluded := range cd.excludedCollections {
		if name == excluded {
			return false
		}
	}

	// Include by default if not explicitly excluded
	return true
}

// databaseCollection represents the structure of a collection record from the database
type databaseCollection struct {
	Name       string  `db:"name"`
	Type       string  `db:"type"`
	System     bool    `db:"system"`
	Schema     string  `db:"schema"`
	ListRule   *string `db:"listRule"`
	ViewRule   *string `db:"viewRule"`
	CreateRule *string `db:"createRule"`
	UpdateRule *string `db:"updateRule"`
	DeleteRule *string `db:"deleteRule"`
	Options    string  `db:"options"`
}

// extractCollectionInfo extracts complete metadata from a database collection record
func (cd *CollectionDiscovery) extractCollectionInfo(dbCollection databaseCollection) (*CollectionInfo, error) {

	// Extract basic information
	collectionInfo := &CollectionInfo{
		Name:   dbCollection.Name,
		Type:   dbCollection.Type,
		System: dbCollection.System,
		Fields: []FieldInfo{},
	}

	// Extract API rules
	collectionInfo.ListRule = dbCollection.ListRule
	collectionInfo.ViewRule = dbCollection.ViewRule
	collectionInfo.CreateRule = dbCollection.CreateRule
	collectionInfo.UpdateRule = dbCollection.UpdateRule
	collectionInfo.DeleteRule = dbCollection.DeleteRule

	// Initialize options
	collectionInfo.Options = make(map[string]any)

	// Parse schema JSON to extract fields
	if dbCollection.Schema != "" {
		fields, err := cd.parseSchemaFields(dbCollection.Schema)
		if err != nil {
			log.Printf("Warning: Failed to parse schema for collection %s: %v", dbCollection.Name, err)
		} else {
			collectionInfo.Fields = fields
		}
	}

	// Parse options JSON
	if dbCollection.Options != "" {
		options, err := cd.parseOptionsJSON(dbCollection.Options)
		if err != nil {
			log.Printf("Warning: Failed to parse options for collection %s: %v", dbCollection.Name, err)
		} else {
			collectionInfo.Options = options
		}
	}

	return collectionInfo, nil
}

// parseSchemaFields parses the schema JSON string to extract field information
func (cd *CollectionDiscovery) parseSchemaFields(schemaJSON string) ([]FieldInfo, error) {
	if schemaJSON == "" {
		return []FieldInfo{}, nil
	}

	// Parse the schema JSON
	var schemaData []map[string]any
	if err := json.Unmarshal([]byte(schemaJSON), &schemaData); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	var fields []FieldInfo
	for _, fieldData := range schemaData {
		field, err := cd.parseFieldInfo(fieldData)
		if err != nil {
			log.Printf("Warning: Failed to parse field %v: %v", fieldData, err)
			continue
		}
		fields = append(fields, *field)
	}

	return fields, nil
}

// parseFieldInfo parses a single field from the schema data
func (cd *CollectionDiscovery) parseFieldInfo(fieldData map[string]any) (*FieldInfo, error) {
	field := &FieldInfo{
		Options: make(map[string]any),
	}

	// Extract field name
	if name, ok := fieldData["name"].(string); ok {
		field.Name = name
	} else {
		return nil, fmt.Errorf("field name is missing or not a string")
	}

	// Extract field type
	if fieldType, ok := fieldData["type"].(string); ok {
		field.Type = fieldType
	} else {
		return nil, fmt.Errorf("field type is missing or not a string")
	}

	// Extract required flag
	if required, ok := fieldData["required"].(bool); ok {
		field.Required = required
	}

	// Extract system flag
	if system, ok := fieldData["system"].(bool); ok {
		field.System = system
	}

	// Extract options
	if options, ok := fieldData["options"].(map[string]any); ok {
		field.Options = options
	}

	// For relation fields, copy field-level properties to options for consistency
	if field.Type == "relation" {
		// Copy relation-specific properties from field level to options
		relationProps := []string{"maxSelect", "minSelect", "collectionId", "cascadeDelete", "presentable", "hidden"}
		for _, prop := range relationProps {
			if value, exists := fieldData[prop]; exists {
				field.Options[prop] = value
			}
		}
	}

	return field, nil
}

// parseOptionsJSON parses the options JSON string
func (cd *CollectionDiscovery) parseOptionsJSON(optionsJSON string) (map[string]any, error) {
	if optionsJSON == "" {
		return make(map[string]any), nil
	}

	var options map[string]any
	if err := json.Unmarshal([]byte(optionsJSON), &options); err != nil {
		return nil, fmt.Errorf("failed to parse options JSON: %w", err)
	}

	return options, nil
}

// GetCollectionNames returns a list of all discoverable collection names
func (cd *CollectionDiscovery) GetCollectionNames() ([]string, error) {
	collections, err := cd.DiscoverCollections()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(collections))
	for i, collection := range collections {
		names[i] = collection.Name
	}

	return names, nil
}

// GetCollectionsByType returns collections filtered by type
func (cd *CollectionDiscovery) GetCollectionsByType(collectionType string) ([]CollectionInfo, error) {
	collections, err := cd.DiscoverCollections()
	if err != nil {
		return nil, err
	}

	var filtered []CollectionInfo
	for _, collection := range collections {
		if collection.Type == collectionType {
			filtered = append(filtered, collection)
		}
	}

	return filtered, nil
}

// GetAuthCollections returns only auth-type collections
func (cd *CollectionDiscovery) GetAuthCollections() ([]CollectionInfo, error) {
	return cd.GetCollectionsByType(CollectionTypeAuth)
}

// GetBaseCollections returns only base-type collections
func (cd *CollectionDiscovery) GetBaseCollections() ([]CollectionInfo, error) {
	return cd.GetCollectionsByType(CollectionTypeBase)
}

// GetViewCollections returns only view-type collections
func (cd *CollectionDiscovery) GetViewCollections() ([]CollectionInfo, error) {
	return cd.GetCollectionsByType(CollectionTypeView)
}

// ValidateCollectionAccess validates that the discovery service can access collections
func (cd *CollectionDiscovery) ValidateCollectionAccess() error {
	if cd.app == nil {
		return fmt.Errorf("PocketBase app is nil")
	}

	// Try to access collections to validate the connection
	var count int
	err := cd.app.DB().NewQuery("SELECT COUNT(*) FROM _collections").One(&count)
	if err != nil {
		return fmt.Errorf("failed to access collections table: %w", err)
	}

	log.Printf("Collection access validated: found %d total collections", count)
	return nil
}

// GetCollectionStats returns statistics about discovered collections
func (cd *CollectionDiscovery) GetCollectionStats() (map[string]int, error) {
	if cd.app == nil {
		return nil, fmt.Errorf("PocketBase app is nil")
	}

	// Query for collection statistics
	var stats struct {
		Total int `db:"total"`
		Base  int `db:"base"`
		Auth  int `db:"auth"`
		View  int `db:"view"`
	}

	err := cd.app.DB().NewQuery(`
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN type = 'base' THEN 1 END) as base,
			COUNT(CASE WHEN type = 'auth' THEN 1 END) as auth,
			COUNT(CASE WHEN type = 'view' THEN 1 END) as view
		FROM _collections
	`).One(&stats)

	if err != nil {
		return nil, fmt.Errorf("failed to get collection statistics: %w", err)
	}

	// Filter by included/excluded collections if configured
	collections, err := cd.DiscoverCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to discover collections: %w", err)
	}

	// Count filtered collections by type
	filteredStats := map[string]int{
		"total":    len(collections),
		"base":     0,
		"auth":     0,
		"view":     0,
		"system":   0,
		"included": len(collections),
		"excluded": stats.Total - len(collections),
	}

	for _, col := range collections {
		filteredStats[col.Type]++
		if col.System {
			filteredStats["system"]++
		}
	}

	return filteredStats, nil
}

// RefreshCollectionCache refreshes any cached collection data
func (cd *CollectionDiscovery) RefreshCollectionCache() {
	// Currently, we don't maintain a cache, but this method is here for future use
	// In the future, we could implement caching to improve performance
	log.Printf("Collection cache refresh requested (no-op - caching not implemented)")
}

// GetSystemFields returns the standard system fields that are present in all collections
func (cd *CollectionDiscovery) GetSystemFields() []FieldInfo {
	return []FieldInfo{
		{
			Name:     "id",
			Type:     "text",
			Required: true,
			System:   true,
			Options: map[string]any{
				"description": "Unique record identifier",
			},
		},
		{
			Name:     "created",
			Type:     "date",
			Required: true,
			System:   true,
			Options: map[string]any{
				"description": "Record creation timestamp",
			},
		},
		{
			Name:     "updated",
			Type:     "date",
			Required: true,
			System:   true,
			Options: map[string]any{
				"description": "Record last update timestamp",
			},
		},
	}
}

// extractCollectionInfoFromPB extracts metadata from a PocketBase collection
func (cd *CollectionDiscovery) extractCollectionInfoFromPB(collection *core.Collection) (*CollectionInfo, error) {
	collectionInfo := &CollectionInfo{
		Name:    collection.Name,
		Type:    collection.Type,
		System:  collection.System,
		Fields:  []FieldInfo{},
		Options: make(map[string]any),
	}

	// Extract API rules
	collectionInfo.ListRule = collection.ListRule
	collectionInfo.ViewRule = collection.ViewRule
	collectionInfo.CreateRule = collection.CreateRule
	collectionInfo.UpdateRule = collection.UpdateRule
	collectionInfo.DeleteRule = collection.DeleteRule

	// Extract fields from collection schema
	// For PocketBase v0.29, we'll marshal and unmarshal the schema to get field info
	schemaBytes, err := collection.MarshalJSON()
	if err == nil {
		var collectionData map[string]any
		if err := json.Unmarshal(schemaBytes, &collectionData); err == nil {
			if schema, ok := collectionData["fields"].([]any); ok {
				for _, fieldData := range schema {
					if fieldMap, ok := fieldData.(map[string]any); ok {

						fieldInfo := FieldInfo{
							Options: make(map[string]any),
						}

						if name, ok := fieldMap["name"].(string); ok {
							fieldInfo.Name = name
						}
						if fieldType, ok := fieldMap["type"].(string); ok {
							fieldInfo.Type = fieldType
						}
						if required, ok := fieldMap["required"].(bool); ok {
							fieldInfo.Required = required
						}
						if system, ok := fieldMap["system"].(bool); ok {
							fieldInfo.System = system
						}
						if options, ok := fieldMap["options"].(map[string]any); ok {
							fieldInfo.Options = options
						}

						// For relation fields, copy field-level properties to options for consistency
						if fieldInfo.Type == "relation" {
							// Copy relation-specific properties from field level to options
							relationProps := []string{"maxSelect", "minSelect", "collectionId", "cascadeDelete", "presentable", "hidden"}
							for _, prop := range relationProps {
								if value, exists := fieldMap[prop]; exists {
									fieldInfo.Options[prop] = value
								}
							}
						}

						// For file fields, copy field-level properties to options for consistency
						if fieldInfo.Type == "file" {
							// Copy file-specific properties from field level to options
							fileProps := []string{"maxSelect", "maxSize", "mimeTypes", "thumbs", "protected"}
							for _, prop := range fileProps {
								if value, exists := fieldMap[prop]; exists {
									fieldInfo.Options[prop] = value
								}
							}
						}

						collectionInfo.Fields = append(collectionInfo.Fields, fieldInfo)
					}
				}
			}
		}
	}

	// Add collection options if available
	if collection.IsAuth() {
		collectionInfo.Options[CollectionTypeAuth] = true
	}

	return collectionInfo, nil
}

// IsCollectionAccessible checks if a specific collection can be accessed
func (cd *CollectionDiscovery) IsCollectionAccessible(name string) bool {
	if cd.app == nil {
		return false
	}

	var count int
	err := cd.app.DB().NewQuery("SELECT COUNT(*) FROM _collections WHERE name = {:name}").
		Bind(map[string]any{"name": name}).
		One(&count)

	return err == nil && count > 0
}
