package swagger

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CachedGenerator wraps the Generator with thread-safe caching
type CachedGenerator struct {
	*Generator
	cache           *CombinedOpenAPISpec
	cacheTime       time.Time
	cacheTTL        time.Duration
	collectionsHash string // Hash of collection metadata for change detection
	mutex           sync.RWMutex
	generating      sync.Mutex // Prevents multiple simultaneous generations
}

// NewCachedGenerator creates a new cached generator
func NewCachedGenerator(generator *Generator, ttl time.Duration) *CachedGenerator {
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default 5 minutes cache
	}

	return &CachedGenerator{
		Generator: generator,
		cacheTTL:  ttl,
	}
}

// GenerateSpec generates OpenAPI spec with caching and automatic invalidation
func (cg *CachedGenerator) GenerateSpec() (*CombinedOpenAPISpec, error) {
	// Check if collections have changed (this invalidates cache automatically)
	collectionsChanged, err := cg.hasCollectionsChanged()
	if err != nil {
		// If we can't check for changes, log warning but continue with cache logic
		// This ensures the system remains functional even if change detection fails
		fmt.Printf("Warning: Failed to check for collection changes: %v\n", err)
	}

	// Check if we have a valid cached version (considering both TTL and collection changes)
	cg.mutex.RLock()
	cacheValid := cg.cache != nil &&
		time.Since(cg.cacheTime) < cg.cacheTTL &&
		!collectionsChanged

	if cacheValid {
		cached := cg.cache
		cg.mutex.RUnlock()
		return cached, nil
	}
	cg.mutex.RUnlock()

	// Prevent multiple simultaneous generations
	cg.generating.Lock()
	defer cg.generating.Unlock()

	// Double-check after acquiring the generation lock
	// Re-check collections changed status in case another goroutine updated it
	collectionsChanged, err = cg.hasCollectionsChanged()
	if err != nil {
		fmt.Printf("Warning: Failed to re-check for collection changes: %v\n", err)
		collectionsChanged = true // Assume changed to be safe
	}

	cg.mutex.RLock()
	cacheValid = cg.cache != nil &&
		time.Since(cg.cacheTime) < cg.cacheTTL &&
		!collectionsChanged

	if cacheValid {
		cached := cg.cache
		cg.mutex.RUnlock()
		return cached, nil
	}
	cg.mutex.RUnlock()

	// Generate new spec
	spec, err := cg.Generator.GenerateSpec()
	if err != nil {
		return nil, err
	}

	// Update cache and collections hash
	cg.mutex.Lock()
	cg.cache = spec
	cg.cacheTime = time.Now()

	// Update the collections hash to reflect current state
	if hashErr := cg.updateCollectionsHash(); hashErr != nil {
		fmt.Printf("Warning: Failed to update collections hash: %v\n", hashErr)
	}
	cg.mutex.Unlock()

	return spec, nil
}

// InvalidateCache clears the cached spec and collections hash
func (cg *CachedGenerator) InvalidateCache() {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()
	cg.cache = nil
	cg.cacheTime = time.Time{}
	cg.collectionsHash = "" // Clear hash to force regeneration
}

// GetCacheStatus returns cache information including collection change detection
func (cg *CachedGenerator) GetCacheStatus() map[string]any {
	cg.mutex.RLock()
	collectionsHash := cg.collectionsHash
	cg.mutex.RUnlock()

	status := map[string]any{
		"cached":           cg.cache != nil,
		"cache_ttl":        cg.cacheTTL.String(),
		"collections_hash": collectionsHash,
	}

	cg.mutex.RLock()
	if cg.cache != nil {
		status["cache_age"] = time.Since(cg.cacheTime).String()
		status["expires_in"] = (cg.cacheTTL - time.Since(cg.cacheTime)).String()
	}
	cg.mutex.RUnlock()

	// Check if collections have changed (without holding the lock too long)
	collectionsChanged, err := cg.hasCollectionsChanged()
	if err != nil {
		status["collections_check_error"] = err.Error()
	} else {
		status["collections_changed"] = collectionsChanged
	}

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
func (cg *CachedGenerator) hasCollectionsChanged() (bool, error) {
	currentHash, err := cg.generateCollectionsHash()
	if err != nil {
		return false, err
	}

	// If we don't have a previous hash, consider it changed
	if cg.collectionsHash == "" {
		return true, nil
	}

	return currentHash != cg.collectionsHash, nil
}

// updateCollectionsHash updates the stored collections hash
func (cg *CachedGenerator) updateCollectionsHash() error {
	hash, err := cg.generateCollectionsHash()
	if err != nil {
		return err
	}
	cg.collectionsHash = hash
	return nil
}

// CheckAndInvalidateIfChanged checks for collection changes and invalidates cache if needed
func (cg *CachedGenerator) CheckAndInvalidateIfChanged() (bool, error) {
	changed, err := cg.hasCollectionsChanged()
	if err != nil {
		return false, fmt.Errorf("failed to check for collection changes: %w", err)
	}

	if changed {
		cg.InvalidateCache()
		return true, nil
	}

	return false, nil
}
