package route

import (
	"ims-pocketbase-baas-starter/internal/swagger"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// HandleCacheStatus returns the current status of the global cache store
func HandleCacheStatus(e *core.RequestEvent) error {
	// Get the global swagger generator instance
	generator := swagger.GetGlobalGenerator()
	if generator == nil {
		return e.JSON(500, map[string]string{
			"error": "Swagger generator not initialized",
		})
	}

	// Create cached generator with 5-minute TTL (same as endpoints.go)
	cachedGenerator := swagger.NewCachedGenerator(generator, 5*time.Minute)

	// Get cache status
	status := cachedGenerator.GetCacheStatus()

	return e.JSON(200, status)
}
