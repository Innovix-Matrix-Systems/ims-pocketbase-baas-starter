package swagger

import (
	"net/http"

	"ims-pocketbase-baas-starter/pkg/common"

	"github.com/pocketbase/pocketbase/core"
)

// RegisterEndpoints registers Swagger documentation endpoints
func RegisterEndpoints(se *core.ServeEvent, generator *Generator) {
	// OpenAPI JSON endpoint
	se.Router.GET("/api-docs/openapi.json", func(e *core.RequestEvent) error {
		spec, err := GenerateSpecWithCache(generator)
		if err != nil {
			return common.Response.InternalServerError(e, "Failed to generate OpenAPI specification", map[string]any{
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

	se.Router.GET("/api-docs/scalar", func(e *core.RequestEvent) error {
		html := GetScalarHTML()
		return e.HTML(http.StatusOK, html)
	})

	// Collection stats endpoint
	se.Router.GET("/api-docs/stats", func(e *core.RequestEvent) error {
		stats, err := generator.GetCollectionStats()
		if err != nil {
			return common.Response.InternalServerError(e, "Failed to get collection statistics", map[string]any{
				"details": err.Error(),
			})
		}
		return e.JSON(http.StatusOK, stats)
	})

	// Cache invalidation endpoint (for development)
	se.Router.POST("/api-docs/invalidate-cache", func(e *core.RequestEvent) error {
		InvalidateCache()
		return common.Response.OK(e, "Cache invalidated successfully", nil)
	})

	// Check for collection changes endpoint
	se.Router.POST("/api-docs/check-collections", func(e *core.RequestEvent) error {
		invalidated := CheckAndInvalidateIfChanged(generator)

		response := map[string]any{
			"collections_changed": invalidated,
			"cache_invalidated":   invalidated,
		}

		if invalidated {
			return common.Response.OK(e, "Collections changed, cache invalidated", response)
		} else {
			return common.Response.OK(e, "No collection changes detected", response)
		}
	})
}
