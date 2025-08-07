package swagger

import (
	"testing"
	"time"
)

func TestNewCachedGenerator(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	if cachedGen == nil {
		t.Fatal("Expected cached generator to be created, got nil")
	}

	if cachedGen.Generator != generator {
		t.Error("Expected wrapped generator to be set")
	}

	if cachedGen.cacheTTL != 10*time.Minute {
		t.Errorf("Expected cache TTL to be 10 minutes, got %v", cachedGen.cacheTTL)
	}
}

func TestNewCachedGeneratorDefaultTTL(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 0) // Should use default

	if cachedGen.cacheTTL != 5*time.Minute {
		t.Errorf("Expected default cache TTL to be 5 minutes, got %v", cachedGen.cacheTTL)
	}
}

func TestInvalidateCache(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Set some fake cache data
	cachedGen.cache = &CombinedOpenAPISpec{}
	cachedGen.cacheTime = time.Now()
	cachedGen.collectionsHash = "test-hash"

	// Invalidate cache
	cachedGen.InvalidateCache()

	if cachedGen.cache != nil {
		t.Error("Expected cache to be nil after invalidation")
	}

	if !cachedGen.cacheTime.IsZero() {
		t.Error("Expected cache time to be zero after invalidation")
	}

	if cachedGen.collectionsHash != "" {
		t.Error("Expected collections hash to be empty after invalidation")
	}
}

func TestGetCacheStatus(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test with no cache
	status := cachedGen.GetCacheStatus()
	if status["cached"] != false {
		t.Error("Expected cached status to be false when no cache exists")
	}

	if status["cache_ttl"] != "10m0s" {
		t.Errorf("Expected cache_ttl to be '10m0s', got %v", status["cache_ttl"])
	}

	if status["collections_hash"] != "" {
		t.Error("Expected collections_hash to be empty when no cache exists")
	}

	// Test with cache
	cachedGen.cache = &CombinedOpenAPISpec{}
	cachedGen.cacheTime = time.Now().Add(-2 * time.Minute) // 2 minutes ago
	cachedGen.collectionsHash = "test-hash-123"

	status = cachedGen.GetCacheStatus()
	if status["cached"] != true {
		t.Error("Expected cached status to be true when cache exists")
	}

	if status["cache_age"] == nil {
		t.Error("Expected cache_age to be present when cache exists")
	}

	if status["expires_in"] == nil {
		t.Error("Expected expires_in to be present when cache exists")
	}

	if status["collections_hash"] != "test-hash-123" {
		t.Errorf("Expected collections_hash to be 'test-hash-123', got %v", status["collections_hash"])
	}

	// collections_changed should be present (even if there's an error checking)
	if _, exists := status["collections_changed"]; !exists {
		if _, errorExists := status["collections_check_error"]; !errorExists {
			t.Error("Expected either collections_changed or collections_check_error to be present")
		}
	}
}

func TestCacheStatusThreadSafety(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Run multiple goroutines to test thread safety
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()

			// Simulate concurrent access
			for j := 0; j < 100; j++ {
				cachedGen.GetCacheStatus()
				if j%10 == 0 {
					cachedGen.InvalidateCache()
				}
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock, the test passes
}

func TestCollectionChangeDetection(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test with nil app (should handle gracefully)
	_, err := cachedGen.hasCollectionsChanged()
	if err == nil {
		t.Error("Expected error when checking collections with nil app")
	}

	// Test CheckAndInvalidateIfChanged with nil app
	invalidated, err := cachedGen.CheckAndInvalidateIfChanged()
	if err == nil {
		t.Error("Expected error when checking collections with nil app")
	}

	if invalidated {
		t.Error("Expected invalidated to be false when there's an error")
	}
}

func TestGenerateCollectionsHash(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test with nil app (should return error)
	hash, err := cachedGen.generateCollectionsHash()
	if err == nil {
		t.Error("Expected error when generating hash with nil app")
	}

	if hash != "" {
		t.Error("Expected empty hash when there's an error")
	}
}

func TestUpdateCollectionsHash(t *testing.T) {
	generator := NewGenerator(nil, DefaultConfig())
	cachedGen := NewCachedGenerator(generator, 10*time.Minute)

	// Test with nil app (should return error)
	err := cachedGen.updateCollectionsHash()
	if err == nil {
		t.Error("Expected error when updating hash with nil app")
	}

	// Hash should remain empty after failed update
	if cachedGen.collectionsHash != "" {
		t.Error("Expected collections hash to remain empty after failed update")
	}
}
