package swagger

import (
	"net/http"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// RegisterEndpoints registers Swagger documentation endpoints
func RegisterEndpoints(se *core.ServeEvent, app *pocketbase.PocketBase) {
	// OpenAPI JSON endpoint
	se.Router.GET("/api-docs/openapi.json", func(e *core.RequestEvent) error {
		spec, err := GenerateOpenAPI(app)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate OpenAPI specification",
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
}

// RegisterEnhancedEndpoints registers enhanced Swagger documentation endpoints
func RegisterEnhancedEndpoints(se *core.ServeEvent, generator *EnhancedGenerator) {
	// OpenAPI JSON endpoint
	se.Router.GET("/api-docs/openapi.json", func(e *core.RequestEvent) error {
		spec, err := generator.GenerateSpec()
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
}
