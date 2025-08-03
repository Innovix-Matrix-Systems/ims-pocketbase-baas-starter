package swagger

import (
	"fmt"
	"strings"
	"testing"
)

// Mock implementations for testing
type mockDiscovery struct {
	collections []EnhancedCollectionInfo
	stats       map[string]int
	shouldError bool
}

func (m *mockDiscovery) DiscoverCollections() ([]EnhancedCollectionInfo, error) {
	if m.shouldError {
		return nil, fmt.Errorf("mock discovery error")
	}
	return m.collections, nil
}

func (m *mockDiscovery) GetCollection(name string) (*EnhancedCollectionInfo, error) {
	for _, collection := range m.collections {
		if collection.Name == name {
			return &collection, nil
		}
	}
	return nil, fmt.Errorf("collection not found")
}

func (m *mockDiscovery) ShouldIncludeCollection(name string, collectionType string, system bool) bool {
	return true
}

func (m *mockDiscovery) GetCollectionStats() (map[string]int, error) {
	if m.stats != nil {
		return m.stats, nil
	}
	return map[string]int{
		"total": len(m.collections),
		"base":  1,
		"auth":  1,
	}, nil
}

func (m *mockDiscovery) ValidateCollectionAccess() error {
	if m.shouldError {
		return fmt.Errorf("mock validation error")
	}
	return nil
}

func (m *mockDiscovery) RefreshCollectionCache() {
	// Mock implementation
}

func TestNewEnhancedGenerator(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	if generator == nil {
		t.Fatal("Expected generator to be created, got nil")
	}

	if generator.config.Title != config.Title {
		t.Errorf("Expected title %s, got %s", config.Title, generator.config.Title)
	}

	if generator.discovery == nil {
		t.Error("Expected discovery to be initialized")
	}

	if generator.schemaGen == nil {
		t.Error("Expected schemaGen to be initialized")
	}

	if generator.routeGen == nil {
		t.Error("Expected routeGen to be initialized")
	}
}

func TestDefaultEnhancedConfig(t *testing.T) {
	config := DefaultEnhancedConfig()

	if config.Title == "" {
		t.Error("Expected default title to be set")
	}

	if config.Version == "" {
		t.Error("Expected default version to be set")
	}

	if config.Description == "" {
		t.Error("Expected default description to be set")
	}

	if config.ServerURL == "" {
		t.Error("Expected default server URL to be set")
	}

	if config.IncludeSystem {
		t.Error("Expected IncludeSystem to be false by default")
	}

	if !config.EnableAuth {
		t.Error("Expected EnableAuth to be true by default")
	}

	if !config.IncludeExamples {
		t.Error("Expected IncludeExamples to be true by default")
	}
}

func TestGenerateSpec(t *testing.T) {
	// Create mock collections
	collections := []EnhancedCollectionInfo{
		{
			Name: "users",
			Type: "auth",
			Fields: []FieldInfo{
				{Name: "email", Type: "email", Required: true},
				{Name: "name", Type: "text", Required: true},
			},
		},
		{
			Name: "posts",
			Type: "base",
			Fields: []FieldInfo{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "text", Required: false},
			},
		},
	}

	// Create enhanced generator with mock discovery
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Replace discovery with mock
	mockDisc := &mockDiscovery{collections: collections}
	generator.discovery = mockDisc

	spec, err := generator.GenerateSpec()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if spec == nil {
		t.Fatal("Expected spec to be generated, got nil")
	}

	// Check basic spec structure
	if spec.OpenAPI != "3.0.0" {
		t.Errorf("Expected OpenAPI version 3.0.0, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != config.Title {
		t.Errorf("Expected title %s, got %s", config.Title, spec.Info.Title)
	}

	if len(spec.Servers) == 0 {
		t.Error("Expected at least one server to be defined")
	}

	if spec.Paths == nil {
		t.Error("Expected paths to be defined")
	}

	if spec.Components == nil {
		t.Error("Expected components to be defined")
	}

	if spec.Components.Schemas == nil {
		t.Error("Expected schemas to be defined")
	}

	if spec.Components.SecuritySchemes == nil {
		t.Error("Expected security schemes to be defined when auth is enabled")
	}

	// Check that we have paths for both collections
	foundUsersPaths := false
	foundPostsPaths := false

	for path := range spec.Paths {
		if strings.Contains(path, "users") {
			foundUsersPaths = true
		}
		if strings.Contains(path, "posts") {
			foundPostsPaths = true
		}
	}

	if !foundUsersPaths {
		t.Error("Expected to find users collection paths")
	}

	if !foundPostsPaths {
		t.Error("Expected to find posts collection paths")
	}

	// Check that we have tags
	if len(spec.Tags) == 0 {
		t.Error("Expected tags to be generated")
	}
}

func TestGenerateSpecWithError(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Replace discovery with error-prone mock
	mockDisc := &mockDiscovery{shouldError: true}
	generator.discovery = mockDisc

	spec, err := generator.GenerateSpec()
	if err == nil {
		t.Error("Expected error when discovery fails, got nil")
	}

	if spec != nil {
		t.Error("Expected nil spec when discovery fails")
	}
}

func TestAddCustomRoute(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	customRoute := CustomRoute{
		Method:      "GET",
		Path:        "/api/v1/hello",
		Summary:     "Hello World",
		Description: "Returns a greeting",
		Tags:        []string{"Custom"},
		Protected:   false,
	}

	initialCount := len(generator.config.CustomRoutes)
	generator.AddCustomRoute(customRoute)

	if len(generator.config.CustomRoutes) != initialCount+1 {
		t.Errorf("Expected %d custom routes, got %d", initialCount+1, len(generator.config.CustomRoutes))
	}

	addedRoute := generator.config.CustomRoutes[len(generator.config.CustomRoutes)-1]
	if addedRoute.Path != customRoute.Path {
		t.Errorf("Expected custom route path %s, got %s", customRoute.Path, addedRoute.Path)
	}
}

func TestGetCollectionStats(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Replace discovery with mock
	expectedStats := map[string]int{
		"total": 5,
		"base":  3,
		"auth":  2,
	}
	mockDisc := &mockDiscovery{stats: expectedStats}
	generator.discovery = mockDisc

	stats, err := generator.GetCollectionStats()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stats["total"] != expectedStats["total"] {
		t.Errorf("Expected total %d, got %d", expectedStats["total"], stats["total"])
	}

	if stats["base"] != expectedStats["base"] {
		t.Errorf("Expected base %d, got %d", expectedStats["base"], stats["base"])
	}

	if stats["auth"] != expectedStats["auth"] {
		t.Errorf("Expected auth %d, got %d", expectedStats["auth"], stats["auth"])
	}
}

func TestValidateConfiguration(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Valid configuration should pass
	err := generator.ValidateConfiguration()
	if err != nil {
		t.Errorf("Expected no error for valid config, got %v", err)
	}

	// Test invalid configurations
	invalidConfigs := []struct {
		name   string
		modify func(*EnhancedConfig)
	}{
		{
			name: "empty title",
			modify: func(c *EnhancedConfig) {
				c.Title = ""
			},
		},
		{
			name: "empty version",
			modify: func(c *EnhancedConfig) {
				c.Version = ""
			},
		},
		{
			name: "empty server URL",
			modify: func(c *EnhancedConfig) {
				c.ServerURL = ""
			},
		},
	}

	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			invalidConfig := config
			tc.modify(&invalidConfig)

			generator.config = invalidConfig
			err := generator.ValidateConfiguration()
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestUpdateConfiguration(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Update with valid configuration
	newConfig := config
	newConfig.Title = "Updated Title"
	newConfig.Version = "2.0.0"

	err := generator.UpdateConfiguration(newConfig)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if generator.config.Title != "Updated Title" {
		t.Errorf("Expected title to be updated to 'Updated Title', got %s", generator.config.Title)
	}

	if generator.config.Version != "2.0.0" {
		t.Errorf("Expected version to be updated to '2.0.0', got %s", generator.config.Version)
	}

	// Update with invalid configuration
	invalidConfig := newConfig
	invalidConfig.Title = ""

	err = generator.UpdateConfiguration(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid config, got nil")
	}

	// Configuration should not be updated when invalid
	if generator.config.Title != "Updated Title" {
		t.Error("Configuration should not be updated when validation fails")
	}
}

func TestGetConfiguration(t *testing.T) {
	config := DefaultEnhancedConfig()
	config.Title = "Test Title"
	generator := NewEnhancedGenerator(nil, config)

	retrievedConfig := generator.GetConfiguration()
	if retrievedConfig.Title != config.Title {
		t.Errorf("Expected title %s, got %s", config.Title, retrievedConfig.Title)
	}
}

func TestRefreshCollections(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Should not error even with nil discovery
	err := generator.RefreshCollections()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// With mock discovery
	mockDisc := &mockDiscovery{}
	generator.discovery = mockDisc

	err = generator.RefreshCollections()
	if err != nil {
		t.Errorf("Expected no error with mock discovery, got %v", err)
	}
}

func TestGetHealthStatus(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	status := generator.GetHealthStatus()

	if status["status"] == nil {
		t.Error("Expected status to be set")
	}

	components, ok := status["components"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected components to be a map")
	}

	expectedComponents := []string{"discovery", "schemaGen", "routeGen", "app"}
	for _, component := range expectedComponents {
		if _, exists := components[component]; !exists {
			t.Errorf("Expected component %s to be present", component)
		}
	}
}

func TestGetHealthStatusWithError(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	// Replace discovery with error-prone mock
	mockDisc := &mockDiscovery{shouldError: true}
	generator.discovery = mockDisc

	status := generator.GetHealthStatus()

	if status["status"] != "unhealthy" {
		t.Error("Expected status to be unhealthy when validation fails")
	}

	if status["error"] == nil {
		t.Error("Expected error to be set when status is unhealthy")
	}
}

func TestBuildPaths(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	routes := []GeneratedRoute{
		{
			Method:      "GET",
			Path:        "/api/collections/users/records",
			Summary:     "List users",
			Description: "Get all users",
			Tags:        []string{"Users"},
			OperationID: "listUsers",
			Responses: map[string]Response{
				"200": {Description: "Success"},
			},
		},
		{
			Method:      "POST",
			Path:        "/api/collections/users/records",
			Summary:     "Create user",
			Description: "Create a new user",
			Tags:        []string{"Users"},
			OperationID: "createUser",
			Responses: map[string]Response{
				"201": {Description: "Created"},
			},
		},
	}

	paths := generator.buildPaths(routes)

	if len(paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(paths))
	}

	userPath, exists := paths["/api/collections/users/records"]
	if !exists {
		t.Fatal("Expected users path to exist")
	}

	if len(userPath) != 2 {
		t.Errorf("Expected 2 operations for users path, got %d", len(userPath))
	}

	if _, exists := userPath["get"]; !exists {
		t.Error("Expected GET operation to exist")
	}

	if _, exists := userPath["post"]; !exists {
		t.Error("Expected POST operation to exist")
	}
}

func TestBuildComponents(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	schemas := map[string]interface{}{
		"User": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{"type": "string"},
			},
		},
	}

	components := generator.buildComponents(schemas)

	if components == nil {
		t.Fatal("Expected components to be created")
	}

	if components.Schemas == nil {
		t.Error("Expected schemas to be set")
	}

	if len(components.Schemas) != 1 {
		t.Errorf("Expected 1 schema, got %d", len(components.Schemas))
	}

	if components.SecuritySchemes == nil {
		t.Error("Expected security schemes to be set when auth is enabled")
	}

	if _, exists := components.SecuritySchemes["BearerAuth"]; !exists {
		t.Error("Expected BearerAuth security scheme to exist")
	}
}

func TestBuildTags(t *testing.T) {
	config := DefaultEnhancedConfig()
	generator := NewEnhancedGenerator(nil, config)

	collections := []EnhancedCollectionInfo{
		{Name: "users", Type: "auth"},
		{Name: "posts", Type: "base"},
	}

	routes := []GeneratedRoute{
		{Tags: []string{"Custom", "API"}},
		{Tags: []string{"Users"}},
	}

	tags := generator.buildTags(collections, routes)

	if len(tags) == 0 {
		t.Error("Expected tags to be generated")
	}

	// Check for expected tags
	tagNames := make(map[string]bool)
	for _, tag := range tags {
		tagNames[tag.Name] = true
	}

	expectedTags := []string{"Users", "Posts", "Collections", "Authentication", "Custom", "API"}
	for _, expectedTag := range expectedTags {
		if !tagNames[expectedTag] {
			t.Errorf("Expected tag %s to be present", expectedTag)
		}
	}
}
