package apidoc

import (
	"ims-pocketbase-baas-starter/pkg/cache"
	"testing"
	"time"
)

func TestCacheBasicFunctionality(t *testing.T) {
	// Test cache invalidation
	InvalidateCache()

	// Check that cache keys are cleared
	if _, found := cache.GetInstance().Get(SwaggerSpecKey); found {
		t.Error("Expected swagger spec cache to be cleared after invalidation")
	}

	if _, found := cache.GetInstance().Get(SwaggerCollectionsHash); found {
		t.Error("Expected collections hash cache to be cleared after invalidation")
	}
}

func TestCacheStatus(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())

	// Test initial status
	status := GetCacheStatus(generator)

	if status["cached"].(bool) {
		t.Error("Expected cached to be false initially")
	}

	// Check that cache_stats is included
	if _, exists := status["cache_stats"]; !exists {
		t.Error("Expected cache_stats to be included in status")
	}
}

func TestCacheTTL(t *testing.T) {
	// Save original TTL
	originalTTL := cacheTTL
	defer func() { cacheTTL = originalTTL }()

	// Set a very short TTL for testing
	cacheTTL = 1 * time.Millisecond

	// Set some fake cache data
	cache.GetInstance().SetWithExpiration(SwaggerSpecKey, &CombinedOpenAPISpec{}, 1*time.Millisecond)

	// Wait for TTL to expire
	time.Sleep(2 * time.Millisecond)

	// Check that cache is expired
	if _, found := cache.GetInstance().Get(SwaggerSpecKey); found {
		t.Error("Expected cache to be expired after TTL")
	}
}

func TestCollectionChanges(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())

	// Test collection change detection
	changed := hasCollectionsChanged(generator)

	// Should be true initially (no previous hash)
	if !changed {
		t.Error("Expected collections to be considered changed initially")
	}

	// Test CheckAndInvalidateIfChanged
	invalidated := CheckAndInvalidateIfChanged(generator)

	// Should be true since collections changed
	if !invalidated {
		t.Error("Expected cache to be invalidated due to collection changes")
	}
}

func TestCacheService(t *testing.T) {
	// Test that it uses the centralized cache service
	cacheService := cache.GetInstance()
	if cacheService == nil {
		t.Error("Expected to get the centralized cache service")
	}

	// Test cache key constants
	expectedSpecKey := "swagger_spec"
	if SwaggerSpecKey != expectedSpecKey {
		t.Errorf("Expected spec key to be '%s', got '%s'", expectedSpecKey, SwaggerSpecKey)
	}

	expectedHashKey := "swagger_collections_hash"
	if SwaggerCollectionsHash != expectedHashKey {
		t.Errorf("Expected hash key to be '%s', got '%s'", expectedHashKey, SwaggerCollectionsHash)
	}
}

func TestGetCachedSpec(t *testing.T) {
	// Clear any existing cache
	InvalidateCache()

	// Should return nil when no cache exists
	spec := getCachedSpec()
	if spec != nil {
		t.Error("Expected nil when no cache exists")
	}

	// Set a cached spec
	testSpec := &CombinedOpenAPISpec{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
	}
	cache.GetInstance().SetWithExpiration(SwaggerSpecKey, testSpec, 5*time.Minute)

	// Should return the cached spec
	cachedSpec := getCachedSpec()
	if cachedSpec == nil {
		t.Error("Expected cached spec to be returned")
		return // Add early return to prevent nil pointer dereference
	}

	if cachedSpec.OpenAPI != "3.0.0" {
		t.Error("Expected cached spec to have correct OpenAPI version")
	}
}
