package swagger

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterEndpoints registers Swagger documentation endpoints
func RegisterEndpoints(se *core.ServeEvent, generator *Generator) {
	// Create cached generator with 5-minute TTL
	cachedGenerator := NewCachedGenerator(generator, 5*time.Minute)

	// OpenAPI JSON endpoint
	se.Router.GET("/api-docs/openapi.json", func(e *core.RequestEvent) error {
		spec, err := cachedGenerator.GenerateSpec()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Failed to generate OpenAPI specification",
				"details": err.Error(),
			})
		}
		return e.JSON(http.StatusOK, spec)
	})

	// Swagger UI endpoint
	se.Router.GET("/api-docs", func(e *core.RequestEvent) error {
		html := GetSwaggerUIHTML()
		return e.HTML(http.StatusOK, html)
	})

	// ReDoc endpoint
	se.Router.GET("/api-docs/redoc", func(e *core.RequestEvent) error {
		html := GetRedocHTML()
		return e.HTML(http.StatusOK, html)
	})

	// Collection stats endpoint
	se.Router.GET("/api-docs/stats", func(e *core.RequestEvent) error {
		stats, err := generator.GetCollectionStats()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Failed to get collection statistics",
				"details": err.Error(),
			})
		}
		return e.JSON(http.StatusOK, stats)
	})

	// Cache status endpoint
	se.Router.GET("/api-docs/cache-status", func(e *core.RequestEvent) error {
		status := cachedGenerator.GetCacheStatus()
		return e.JSON(http.StatusOK, status)
	})

	// Cache invalidation endpoint (for development)
	se.Router.POST("/api-docs/invalidate-cache", func(e *core.RequestEvent) error {
		cachedGenerator.InvalidateCache()
		return e.JSON(http.StatusOK, map[string]string{
			"message": "Cache invalidated successfully",
		})
	})

	// Check for collection changes endpoint
	se.Router.POST("/api-docs/check-collections", func(e *core.RequestEvent) error {
		invalidated, err := cachedGenerator.CheckAndInvalidateIfChanged()
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error":   "Failed to check for collection changes",
				"details": err.Error(),
			})
		}

		response := map[string]any{
			"collections_changed": invalidated,
			"cache_invalidated":   invalidated,
		}

		if invalidated {
			response["message"] = "Collections changed, cache invalidated"
		} else {
			response["message"] = "No collection changes detected"
		}

		return e.JSON(http.StatusOK, response)
	})
}
