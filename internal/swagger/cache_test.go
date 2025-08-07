package swagger

import (
	"ims-pocketbase-baas-starter/pkg/cache"
	"testing"
	"time"
)

func TestCachedGeneratorBasicFunctionality(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test cache invalidation
	cachedGen.InvalidateCache()

	// Check that cache keys are cleared
	specKey := cachedGen.cacheKey.SwaggerSpec()
	hashKey := cachedGen.cacheKey.SwaggerCollectionsHash()

	if _, found := cachedGen.cache.Get(specKey); found {
		t.Error("Expected swagger spec cache to be cleared after invalidation")
	}

	if _, found := cachedGen.cache.Get(hashKey); found {
		t.Error("Expected collections hash cache to be cleared after invalidation")
	}
}

func TestCachedGeneratorCacheStatus(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 5*time.Minute)

	// Test initial status
	status := cachedGen.GetCacheStatus()

	if status["cached"].(bool) {
		t.Error("Expected cached to be false initially")
	}

	if status["cache_ttl"].(string) != "5m0s" {
		t.Errorf("Expected cache_ttl to be '5m0s', got %s", status["cache_ttl"])
	}

	// Check that cache_stats is included
	if _, exists := status["cache_stats"]; !exists {
		t.Error("Expected cache_stats to be included in status")
	}
}

func TestCachedGeneratorTTL(t *testing.T) {
	// Create a basic generator with very short TTL
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 1*time.Millisecond)

	// Set some fake cache data
	specKey := cachedGen.cacheKey.SwaggerSpec()
	cachedGen.cache.SetWithExpiration(specKey, &CombinedOpenAPISpec{}, 1*time.Millisecond)

	// Wait for TTL to expire
	time.Sleep(2 * time.Millisecond)

	// Check that cache is expired
	if _, found := cachedGen.cache.Get(specKey); found {
		t.Error("Expected cache to be expired after TTL")
	}
}

func TestCachedGeneratorCollectionChanges(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test collection change detection
	changed := cachedGen.hasCollectionsChanged()

	// Should be true initially (no previous hash)
	if !changed {
		t.Error("Expected collections to be considered changed initially")
	}

	// Test CheckAndInvalidateIfChanged
	invalidated := cachedGen.CheckAndInvalidateIfChanged()

	// Should be true since collections changed
	if !invalidated {
		t.Error("Expected cache to be invalidated due to collection changes")
	}
}

func TestCachedGeneratorCacheService(t *testing.T) {
	// Create a basic generator
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 5*time.Minute)

	// Test that it uses the centralized cache service
	if cachedGen.cache != cache.GetInstance() {
		t.Error("Expected cached generator to use the centralized cache service")
	}

	// Test cache key generation
	specKey := cachedGen.cacheKey.SwaggerSpec()
	expectedSpecKey := "swagger_spec"
	if specKey != expectedSpecKey {
		t.Errorf("Expected spec key to be '%s', got '%s'", expectedSpecKey, specKey)
	}

	hashKey := cachedGen.cacheKey.SwaggerCollectionsHash()
	expectedHashKey := "swagger_collections_hash"
	if hashKey != expectedHashKey {
		t.Errorf("Expected hash key to be '%s', got '%s'", expectedHashKey, hashKey)
	}
}
