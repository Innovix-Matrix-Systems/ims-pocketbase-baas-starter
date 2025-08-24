package apidoc

import (
	"testing"
)

func TestNewCollectionDiscovery(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	if discovery == nil {
		t.Fatal("Expected discovery to be created, got nil")
	}

	if !discovery.includeSystem {
		t.Error("Expected includeSystem to be true")
	}
}

func TestNewCollectionDiscoveryWithConfig(t *testing.T) {
	excluded := []string{"internal", "temp"}
	discovery := NewCollectionDiscoveryWithConfig(nil, excluded, false)

	if discovery == nil {
		t.Fatal("Expected discovery to be created, got nil")
	}

	if len(discovery.excludedCollections) != 2 {
		t.Errorf("Expected 2 excluded collections, got %d", len(discovery.excludedCollections))
	}

	if discovery.includeSystem {
		t.Error("Expected includeSystem to be false")
	}
}

func TestShouldIncludeCollection(t *testing.T) {
	tests := []struct {
		name                string
		allowedCollections  []string
		excludedCollections []string
		includeSystem       bool
		collectionName      string
		collectionType      string
		system              bool
		expected            bool
	}{
		{
			name:                "Include regular collection with no filters",
			allowedCollections:  []string{},
			excludedCollections: []string{},
			includeSystem:       true,
			collectionName:      "users",
			collectionType:      "base",
			system:              false,
			expected:            true,
		},
		{
			name:                "Exclude system collection when includeSystem is false",
			allowedCollections:  []string{},
			excludedCollections: []string{},
			includeSystem:       false,
			collectionName:      "_admins",
			collectionType:      "auth",
			system:              true,
			expected:            false,
		},
		{
			name:                "Include system collection when includeSystem is true",
			allowedCollections:  []string{},
			excludedCollections: []string{},
			includeSystem:       true,
			collectionName:      "_admins",
			collectionType:      "auth",
			system:              true,
			expected:            true,
		},
		{
			name:                "Exclude collection in excluded list",
			allowedCollections:  []string{},
			excludedCollections: []string{"temp", "internal"},
			includeSystem:       true,
			collectionName:      "temp",
			collectionType:      "base",
			system:              false,
			expected:            false,
		},
		{
			name:                "Include regular collection when not excluded",
			allowedCollections:  []string{},
			excludedCollections: []string{},
			includeSystem:       true,
			collectionName:      "users",
			collectionType:      "base",
			system:              false,
			expected:            true,
		},
		{
			name:                "Include collection not in excluded list",
			allowedCollections:  []string{},
			excludedCollections: []string{"temp", "internal"},
			includeSystem:       true,
			collectionName:      "comments",
			collectionType:      "base",
			system:              false,
			expected:            true,
		},
		{
			name:                "Handle empty collection name",
			allowedCollections:  []string{},
			excludedCollections: []string{},
			includeSystem:       true,
			collectionName:      "",
			collectionType:      "base",
			system:              false,
			expected:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			discovery := NewCollectionDiscoveryWithConfig(nil, tt.excludedCollections, tt.includeSystem)
			result := discovery.ShouldIncludeCollection(tt.collectionName, tt.collectionType, tt.system)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for collection %s", tt.expected, result, tt.collectionName)
			}
		})
	}
}

func TestExtractCollectionInfo(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// Test with basic collection data
	listRule := "id != ''"
	createRule := "@request.auth.id != ''"

	dbCol := databaseCollection{
		Name:       "users",
		Type:       "base",
		System:     false,
		Schema:     "",
		ListRule:   &listRule,
		CreateRule: &createRule,
		Options:    "",
	}

	info, err := discovery.extractCollectionInfo(dbCol)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if info == nil {
		t.Fatal("Expected collection info, got nil")
	}

	if info.Name != "users" {
		t.Errorf("Expected name 'users', got %s", info.Name)
	}

	if info.Type != "base" {
		t.Errorf("Expected type 'base', got %s", info.Type)
	}

	if info.System {
		t.Error("Expected System to be false")
	}

	if info.ListRule == nil || *info.ListRule != listRule {
		t.Errorf("Expected ListRule %s, got %v", listRule, info.ListRule)
	}

	if info.CreateRule == nil || *info.CreateRule != createRule {
		t.Errorf("Expected CreateRule %s, got %v", createRule, info.CreateRule)
	}

	if info.Options == nil {
		t.Error("Expected Options to be initialized")
	}

	if info.Fields == nil {
		t.Error("Expected Fields to be initialized")
	}
}

func TestExtractCollectionInfoWithEmptyData(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// Test with empty collection data
	dbCol := databaseCollection{
		Name:    "",
		Type:    "",
		System:  false,
		Schema:  "",
		Options: "",
	}

	info, err := discovery.extractCollectionInfo(dbCol)
	if err != nil {
		t.Errorf("Expected no error for empty data, got %v", err)
	}

	if info == nil {
		t.Fatal("Expected collection info, got nil")
	}

	if info.Name != "" {
		t.Errorf("Expected empty name, got %s", info.Name)
	}
}

func TestDiscoverCollectionsWithNilApp(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	collections, err := discovery.DiscoverCollections()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collections != nil {
		t.Error("Expected nil collections for nil app")
	}
}

func TestGetCollectionWithNilApp(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	collection, err := discovery.GetCollection("users")
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collection != nil {
		t.Error("Expected nil collection for nil app")
	}
}

func TestValidateCollectionAccessWithNilApp(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	err := discovery.ValidateCollectionAccess()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}
}

func TestGetCollectionNames(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// This will fail with nil app, but we're testing the method exists
	names, err := discovery.GetCollectionNames()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if names != nil {
		t.Error("Expected nil names for nil app")
	}
}

func TestGetCollectionsByType(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// This will fail with nil app, but we're testing the method exists
	collections, err := discovery.GetCollectionsByType("auth")
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collections != nil {
		t.Error("Expected nil collections for nil app")
	}
}

func TestGetAuthCollections(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// This will fail with nil app, but we're testing the method exists
	collections, err := discovery.GetAuthCollections()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collections != nil {
		t.Error("Expected nil collections for nil app")
	}
}

func TestGetBaseCollections(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// This will fail with nil app, but we're testing the method exists
	collections, err := discovery.GetBaseCollections()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collections != nil {
		t.Error("Expected nil collections for nil app")
	}
}

func TestGetViewCollections(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// This will fail with nil app, but we're testing the method exists
	collections, err := discovery.GetViewCollections()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if collections != nil {
		t.Error("Expected nil collections for nil app")
	}
}

func TestParseSchemaFields(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	tests := []struct {
		name        string
		schemaJSON  string
		expectedLen int
		expectError bool
	}{
		{
			name:        "Empty schema",
			schemaJSON:  "",
			expectedLen: 0,
			expectError: false,
		},
		{
			name:        "Valid schema with one field",
			schemaJSON:  `[{"name":"title","type":"text","required":true,"system":false,"options":{"max":100}}]`,
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Valid schema with multiple fields",
			schemaJSON:  `[{"name":"title","type":"text","required":true,"system":false},{"name":"count","type":"number","required":false,"system":false}]`,
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			schemaJSON:  `invalid json`,
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields, err := discovery.parseSchemaFields(tt.schemaJSON)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(fields) != tt.expectedLen {
				t.Errorf("Expected %d fields, got %d", tt.expectedLen, len(fields))
			}

			// Test specific field properties for valid single field case
			if tt.name == "Valid schema with one field" && len(fields) > 0 {
				field := fields[0]
				if field.Name != "title" {
					t.Errorf("Expected field name 'title', got %s", field.Name)
				}
				if field.Type != "text" {
					t.Errorf("Expected field type 'text', got %s", field.Type)
				}
				if !field.Required {
					t.Error("Expected field to be required")
				}
				if field.System {
					t.Error("Expected field to not be system")
				}
				if field.Options == nil {
					t.Error("Expected field options to be initialized")
				}
			}
		})
	}
}

func TestParseFieldInfo(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	tests := []struct {
		name             string
		fieldData        map[string]any
		expectError      bool
		expectedName     string
		expectedType     string
		expectedRequired bool
	}{
		{
			name: "Valid field data",
			fieldData: map[string]any{
				"name":     "title",
				"type":     "text",
				"required": true,
				"system":   false,
				"options":  map[string]any{"max": 100},
			},
			expectError:      false,
			expectedName:     "title",
			expectedType:     "text",
			expectedRequired: true,
		},
		{
			name: "Missing field name",
			fieldData: map[string]any{
				"type":     "text",
				"required": true,
			},
			expectError: true,
		},
		{
			name: "Missing field type",
			fieldData: map[string]any{
				"name":     "title",
				"required": true,
			},
			expectError: true,
		},
		{
			name: "Field with minimal data",
			fieldData: map[string]any{
				"name": "simple",
				"type": "text",
			},
			expectError:      false,
			expectedName:     "simple",
			expectedType:     "text",
			expectedRequired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := discovery.parseFieldInfo(tt.fieldData)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if field == nil {
				t.Fatal("Expected field info, got nil")
			}

			if field.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, field.Name)
			}

			if field.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, field.Type)
			}

			if field.Required != tt.expectedRequired {
				t.Errorf("Expected required %v, got %v", tt.expectedRequired, field.Required)
			}

			if field.Options == nil {
				t.Error("Expected options to be initialized")
			}
		})
	}
}

func TestParseOptionsJSON(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	tests := []struct {
		name        string
		optionsJSON string
		expectError bool
		expectedLen int
	}{
		{
			name:        "Empty options",
			optionsJSON: "",
			expectError: false,
			expectedLen: 0,
		},
		{
			name:        "Valid options JSON",
			optionsJSON: `{"allowEmailAuth":true,"minPasswordLength":8}`,
			expectError: false,
			expectedLen: 2,
		},
		{
			name:        "Invalid JSON",
			optionsJSON: `invalid json`,
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, err := discovery.parseOptionsJSON(tt.optionsJSON)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if options == nil {
				t.Fatal("Expected options map, got nil")
			}

			if len(options) != tt.expectedLen {
				t.Errorf("Expected %d options, got %d", tt.expectedLen, len(options))
			}

			// Test specific options for valid case
			if tt.name == "Valid options JSON" {
				if allowEmailAuth, ok := options["allowEmailAuth"].(bool); !ok || !allowEmailAuth {
					t.Error("Expected allowEmailAuth to be true")
				}
				if minPasswordLength, ok := options["minPasswordLength"].(float64); !ok || minPasswordLength != 8 {
					t.Error("Expected minPasswordLength to be 8")
				}
			}
		})
	}
}

func TestExtractCollectionInfoWithSchema(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	// Test with collection that has schema
	schemaJSON := `[{"name":"title","type":"text","required":true,"system":false,"options":{"max":100}},{"name":"count","type":"number","required":false,"system":false}]`
	optionsJSON := `{"allowEmailAuth":true,"minPasswordLength":8}`

	dbCol := databaseCollection{
		Name:    "test_collection",
		Type:    "base",
		System:  false,
		Schema:  schemaJSON,
		Options: optionsJSON,
	}

	info, err := discovery.extractCollectionInfo(dbCol)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if info == nil {
		t.Fatal("Expected collection info, got nil")
	}

	if len(info.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(info.Fields))
	}

	// Check first field
	if len(info.Fields) > 0 {
		field := info.Fields[0]
		if field.Name != "title" {
			t.Errorf("Expected first field name 'title', got %s", field.Name)
		}
		if field.Type != "text" {
			t.Errorf("Expected first field type 'text', got %s", field.Type)
		}
		if !field.Required {
			t.Error("Expected first field to be required")
		}
	}

	// Check options
	if len(info.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(info.Options))
	}

	if allowEmailAuth, ok := info.Options["allowEmailAuth"].(bool); !ok || !allowEmailAuth {
		t.Error("Expected allowEmailAuth option to be true")
	}
}
func TestGetSystemFields(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	systemFields := discovery.GetSystemFields()

	if len(systemFields) != 3 {
		t.Errorf("Expected 3 system fields, got %d", len(systemFields))
	}

	expectedFields := []string{"id", "created", "updated"}
	for i, expectedName := range expectedFields {
		if i >= len(systemFields) {
			t.Errorf("Missing system field: %s", expectedName)
			continue
		}

		field := systemFields[i]
		if field.Name != expectedName {
			t.Errorf("Expected system field %s, got %s", expectedName, field.Name)
		}

		if !field.Required {
			t.Errorf("Expected system field %s to be required", expectedName)
		}

		if !field.System {
			t.Errorf("Expected system field %s to be marked as system", expectedName)
		}

		if field.Options == nil {
			t.Errorf("Expected system field %s to have options", expectedName)
		}
	}
}

func TestGetCollectionStatsWithNilApp(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	stats, err := discovery.GetCollectionStats()
	if err == nil {
		t.Error("Expected error for nil app, got nil")
	}

	if stats != nil {
		t.Error("Expected nil stats for nil app")
	}
}

func TestIsCollectionAccessibleWithNilApp(t *testing.T) {
	discovery := NewCollectionDiscovery(nil, true)

	accessible := discovery.IsCollectionAccessible("users")
	if accessible {
		t.Error("Expected collection to not be accessible with nil app")
	}
}
