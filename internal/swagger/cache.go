package swagger

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"ims-pocketbase-baas-starter/pkg/cache"
	"sync"
	"time"
)

// CachedGenerator wraps the Generator with centralized caching
type CachedGenerator struct {
	*Generator
	cache      *cache.CacheService
	cacheKey   cache.CacheKey
	cacheTTL   time.Duration
	generating sync.Mutex // Prevents multiple simultaneous generations
}

// NewCachedGenerator creates a new cached generator with centralized caching
func NewCachedGenerator(generator *Generator, ttl time.Duration) *CachedGenerator {
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default 5 minutes cache
	}

	return &CachedGenerator{
		Generator: generator,
		cache:     cache.GetInstance(),
		cacheKey:  cache.CacheKey{},
		cacheTTL:  ttl,
	}
}

// GenerateSpec generates OpenAPI spec with centralized caching and automatic invalidation
func (cg *CachedGenerator) GenerateSpec() (*CombinedOpenAPISpec, error) {
	// Check cache first
	specKey := cg.cacheKey.SwaggerSpec()
	if cachedSpec, found := cg.cache.Get(specKey); found {
		if spec, ok := cachedSpec.(*CombinedOpenAPISpec); ok {
			// Check if collections have changed
			if !cg.hasCollectionsChanged() {
				return spec, nil
			}
			// Collections changed, invalidate cache
			cg.InvalidateCache()
		}
	}

	// Prevent multiple simultaneous generations
	cg.generating.Lock()
	defer cg.generating.Unlock()

	// Double-check cache after acquiring lock
	if cachedSpec, found := cg.cache.Get(specKey); found {
		if spec, ok := cachedSpec.(*CombinedOpenAPISpec); ok {
			if !cg.hasCollectionsChanged() {
				return spec, nil
			}
		}
	}

	// Generate new spec
	spec, err := cg.Generator.GenerateSpec()
	if err != nil {
		return nil, err
	}

	// Cache the spec and update collections hash
	cg.cache.SetWithExpiration(specKey, spec, cg.cacheTTL)
	cg.updateCollectionsHash()

	return spec, nil
}

// InvalidateCache clears the cached spec and collections hash
func (cg *CachedGenerator) InvalidateCache() {
	cg.cache.Delete(cg.cacheKey.SwaggerSpec())
	cg.cache.Delete(cg.cacheKey.SwaggerCollectionsHash())
}

// GetCacheStatus returns cache information including collection change detection
func (cg *CachedGenerator) GetCacheStatus() map[string]any {
	specKey := cg.cacheKey.SwaggerSpec()
	hashKey := cg.cacheKey.SwaggerCollectionsHash()

	_, specCached := cg.cache.Get(specKey)
	collectionsHash, _ := cg.cache.GetString(hashKey)

	status := map[string]any{
		"cached":           specCached,
		"cache_ttl":        cg.cacheTTL.String(),
		"collections_hash": collectionsHash,
	}

	// Check if collections have changed
	collectionsChanged := cg.hasCollectionsChanged()
	status["collections_changed"] = collectionsChanged

	// Add general cache stats
	status["cache_stats"] = cg.cache.GetStats()

	return status
}

// generateCollectionsHash creates a hash of collection metadata for change detection
// Uses the full CollectionInfo struct - any change will invalidate the cache
func (cg *CachedGenerator) generateCollectionsHash() (string, error) {
	collections, err := cg.discovery.DiscoverCollections()
	if err != nil {
		return "", fmt.Errorf("failed to discover collections for hashing: %w", err)
	}

	// Use the full CollectionInfo structs for hashing
	// This ensures any change (fields, rules, options) invalidates the cache
	jsonData, err := json.Marshal(collections)
	if err != nil {
		return "", fmt.Errorf("failed to marshal collections for hashing: %w", err)
	}

	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash), nil
}

// hasCollectionsChanged checks if collections have changed since last cache
func (cg *CachedGenerator) hasCollectionsChanged() bool {
	currentHash, err := cg.generateCollectionsHash()
	if err != nil {
		fmt.Printf("Warning: Failed to generate collections hash: %v\n", err)
		return true // Assume changed to be safe
	}

	hashKey := cg.cacheKey.SwaggerCollectionsHash()
	cachedHash, found := cg.cache.GetString(hashKey)

	// If we don't have a previous hash, consider it changed
	if !found || cachedHash == "" {
		return true
	}

	return currentHash != cachedHash
}

// updateCollectionsHash updates the stored collections hash
func (cg *CachedGenerator) updateCollectionsHash() {
	hash, err := cg.generateCollectionsHash()
	if err != nil {
		fmt.Printf("Warning: Failed to generate collections hash: %v\n", err)
		return
	}

	hashKey := cg.cacheKey.SwaggerCollectionsHash()
	cg.cache.SetWithExpiration(hashKey, hash, cg.cacheTTL)
}

// CheckAndInvalidateIfChanged checks for collection changes and invalidates cache if needed
func (cg *CachedGenerator) CheckAndInvalidateIfChanged() bool {
	if cg.hasCollectionsChanged() {
		cg.InvalidateCache()
		return true
	}
	return false
}
