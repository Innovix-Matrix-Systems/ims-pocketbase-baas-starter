package route

import (
	"ims-pocketbase-baas-starter/pkg/cache"
	"ims-pocketbase-baas-starter/pkg/common"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// HandleCacheStatus returns the current status of the global cache store
func HandleCacheStatus(e *core.RequestEvent) error {
	// Get the cache service instance
	cacheService := cache.GetInstance()

	// Get cache statistics
	stats := cacheService.GetStats()

	// Return cache status using common response helper
	return common.Response.OK(e, "Cache status retrieved successfully", map[string]any{
		"status": "ok",
		"stats":  stats,
	})
}

// HandleCacheClear clears all cache entries in the system
func HandleCacheClear(e *core.RequestEvent) error {
	// Get the cache service instance
	cacheService := cache.GetInstance()

	// Clear all cache entries
	cacheService.Flush()

	// Return success response using common response helper
	return common.Response.OK(e, "Cache cleared successfully", map[string]any{
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
