package middlewares

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"

	"ims-pocketbase-baas-starter/pkg/metrics"
)

// RegisterMiddlewares registers all application middlewares with the PocketBase router
// This function follows the same pattern as RegisterCrons and RegisterCustom routes
func RegisterMiddlewares(e *core.ServeEvent) {
	// Register metrics middleware first to capture all requests
	metricsProvider := metrics.GetInstance()
	metricsMiddleware := NewMetricsMiddleware(metricsProvider)

	e.Router.Bind(&hook.Handler[*core.RequestEvent]{
		Id:   "metricsCollection",
		Func: metricsMiddleware.RequireMetricsFunc(),
	})

	// Register the auth middleware
	authMiddleware := NewAuthMiddleware().WithApp(e.App)

	// Apply auth with exclusions to all PocketBase API endpoints
	e.Router.Bind(&hook.Handler[*core.RequestEvent]{
		Id:   "jwtAuth",
		Func: authMiddleware.RequireAuthWithExclusionsFunc,
	})
}
