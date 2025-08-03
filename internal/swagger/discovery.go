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
	allowedCollections  []string
	excludedCollections []string
	includeSystem       bool
}

// EnhancedCollectionInfo holds complete collection metadata for OpenAPI generation
type EnhancedCollectionInfo struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // "base", "auth", "view"
	System     bool                   `json:"system"`
	Fields     []FieldInfo            `json:"fields"`
	ListRule   *string                `json:"listRule"`
	ViewRule   *string                `json:"viewRule"`
	CreateRule *string                `json:"createRule"`
	UpdateRule *string                `json:"updateRule"`
	DeleteRule *string                `json:"deleteRule"`
	Options    map[string]interface{} `json:"options"`
}

// FieldInfo holds field metadata for schema generation
type FieldInfo struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Required bool                   `json:"required"`
	System   bool                   `json:"system"`
	Options  map[string]interface{} `json:"options"`
}

// Discovery interface for collection discovery
type Discovery interface {
	DiscoverCollections() ([]EnhancedCollectionInfo, error)
	GetCollection(name string) (*EnhancedCollectionInfo, error)
	ShouldIncludeCollection(name string, collectionType string, system bool) bool
	GetCollectionStats() (map[string]int, error)
	ValidateCollectionAccess() error
	RefreshCollectionCache()
}

// NewCollectionDiscovery creates a new collection discovery service
func NewCollectionDiscovery(app *pocketbase.PocketBase, allowedCollections []string, includeSystem bool) *CollectionDiscovery {
	return &CollectionDiscovery{
		app:                 app,
		allowedCollections:  allowedCollections,
		excludedCollections: []string{},
		includeSystem:       includeSystem,
	}
}

// NewCollectionDiscoveryWithConfig creates a new collection discovery service with full configuration
func NewCollectionDiscoveryWithConfig(app *pocketbase.PocketBase, allowedCollections []string, excludedCollections []string, includeSystem bool) *CollectionDiscovery {
	return &CollectionDiscovery{
		app:                 app,
		allowedCollections:  allowedCollections,
		excludedCollections: excludedCollections,
		includeSystem:       includeSystem,
	}
}

// DiscoverCollections discovers all collections from PocketBase and returns their metadata
func (cd *CollectionDiscovery) DiscoverCollections() ([]EnhancedCollectionInfo, error) {
	if cd.app == nil {
		return nil, fmt.Errorf("PocketBase app is nil")
	}

	// Use PocketBase's built-in collection finder
	collections, err := cd.app.FindAllCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to find collections: %w", err)
	}

	var collectionInfos []EnhancedCollectionInfo
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
func (cd *CollectionDiscovery) GetCollection(name string) (*EnhancedCollectionInfo, error) {
	if cd.app == nil {
		return nil, fmt.Errorf("PocketBase app is nil")
	}

	var dbCol dbCollection
	err := cd.app.DB().NewQuery("SELECT name, type, system, schema, listRule, viewRule, createRule, updateRule, deleteRule, options FROM _collections WHERE name = {:name}").
		Bind(map[string]interface{}{"name": name}).
		One(&dbCol)
	if err != nil {
		return nil, fmt.Errorf("failed to find collection %s: %w", name, err)
	}

	if !cd.ShouldIncludeCollection(dbCol.Name, dbCol.Type, dbCol.System) {
		return nil, fmt.Errorf("collection %s is excluded from documentation", name)
	}

	return cd.extractCollectionInfo(dbCol)
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

// dbCollection represents the structure of a collection record from the database
type dbCollection struct {
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
func (cd *CollectionDiscovery) extractCollectionInfo(dbCol dbCollection) (*EnhancedCollectionInfo, error) {

	// Extract basic information
	collectionInfo := &EnhancedCollectionInfo{
		Name:   dbCol.Name,
		Type:   dbCol.Type,
		System: dbCol.System,
		Fields: []FieldInfo{},
	}

	// Extract API rules
	collectionInfo.ListRule = dbCol.ListRule
	collectionInfo.ViewRule = dbCol.ViewRule
	collectionInfo.CreateRule = dbCol.CreateRule
	collectionInfo.UpdateRule = dbCol.UpdateRule
	collectionInfo.DeleteRule = dbCol.DeleteRule

	// Initialize options
	collectionInfo.Options = make(map[string]interface{})

	// Parse schema JSON to extract fields
	if dbCol.Schema != "" {
		fields, err := cd.parseSchemaFields(dbCol.Schema)
		if err != nil {
			log.Printf("Warning: Failed to parse schema for collection %s: %v", dbCol.Name, err)
		} else {
			collectionInfo.Fields = fields
		}
	}

	// Parse options JSON
	if dbCol.Options != "" {
		options, err := cd.parseOptionsJSON(dbCol.Options)
		if err != nil {
			log.Printf("Warning: Failed to parse options for collection %s: %v", dbCol.Name, err)
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
	var schemaData []map[string]interface{}
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
func (cd *CollectionDiscovery) parseFieldInfo(fieldData map[string]interface{}) (*FieldInfo, error) {
	field := &FieldInfo{
		Options: make(map[string]interface{}),
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
	if options, ok := fieldData["options"].(map[string]interface{}); ok {
		field.Options = options
	}

	return field, nil
}

// parseOptionsJSON parses the options JSON string
func (cd *CollectionDiscovery) parseOptionsJSON(optionsJSON string) (map[string]interface{}, error) {
	if optionsJSON == "" {
		return make(map[string]interface{}), nil
	}

	var options map[string]interface{}
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
func (cd *CollectionDiscovery) GetCollectionsByType(collectionType string) ([]EnhancedCollectionInfo, error) {
	collections, err := cd.DiscoverCollections()
	if err != nil {
		return nil, err
	}

	var filtered []EnhancedCollectionInfo
	for _, collection := range collections {
		if collection.Type == collectionType {
			filtered = append(filtered, collection)
		}
	}

	return filtered, nil
}

// GetAuthCollections returns only auth-type collections
func (cd *CollectionDiscovery) GetAuthCollections() ([]EnhancedCollectionInfo, error) {
	return cd.GetCollectionsByType("auth")
}

// GetBaseCollections returns only base-type collections
func (cd *CollectionDiscovery) GetBaseCollections() ([]EnhancedCollectionInfo, error) {
	return cd.GetCollectionsByType("base")
}

// GetViewCollections returns only view-type collections
func (cd *CollectionDiscovery) GetViewCollections() ([]EnhancedCollectionInfo, error) {
	return cd.GetCollectionsByType("view")
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
			Options: map[string]interface{}{
				"description": "Unique record identifier",
			},
		},
		{
			Name:     "created",
			Type:     "date",
			Required: true,
			System:   true,
			Options: map[string]interface{}{
				"description": "Record creation timestamp",
			},
		},
		{
			Name:     "updated",
			Type:     "date",
			Required: true,
			System:   true,
			Options: map[string]interface{}{
				"description": "Record last update timestamp",
			},
		},
	}
}

// extractCollectionInfoFromPB extracts metadata from a PocketBase collection
func (cd *CollectionDiscovery) extractCollectionInfoFromPB(collection *core.Collection) (*EnhancedCollectionInfo, error) {
	collectionInfo := &EnhancedCollectionInfo{
		Name:    collection.Name,
		Type:    collection.Type,
		System:  collection.System,
		Fields:  []FieldInfo{},
		Options: make(map[string]interface{}),
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
		var collectionData map[string]interface{}
		if err := json.Unmarshal(schemaBytes, &collectionData); err == nil {
			if schema, ok := collectionData["fields"].([]interface{}); ok {
				for _, fieldData := range schema {
					if fieldMap, ok := fieldData.(map[string]interface{}); ok {
						fieldInfo := FieldInfo{
							Options: make(map[string]interface{}),
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
						if options, ok := fieldMap["options"].(map[string]interface{}); ok {
							fieldInfo.Options = options
						}
						
						collectionInfo.Fields = append(collectionInfo.Fields, fieldInfo)
					}
				}
			}
		}
	}

	// Add collection options if available
	if collection.IsAuth() {
		collectionInfo.Options["auth"] = true
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
		Bind(map[string]interface{}{"name": name}).
		One(&count)

	return err == nil && count > 0
}
