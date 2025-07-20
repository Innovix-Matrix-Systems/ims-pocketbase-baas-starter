package routes

import (
	"ims-pocketbase-baas-starter/internal/middlewares"

	"github.com/pocketbase/pocketbase/core"
)

func RegisterCustom(e *core.ServeEvent) {
	middleware := middlewares.NewAuthMiddleware()

	g := e.Router.Group("/api/v1")

	//public route
	g.GET("/hello", func(e *core.RequestEvent) error {
		return e.JSON(200, map[string]string{"msg": "Hello from custom route"})
	})

	//auth protected route
	g.GET("/protected", func(e *core.RequestEvent) error {
		// Apply authentication middleware
		authFunc := middleware.RequireAuthFunc()
		if err := authFunc(e); err != nil {
			return err
		}

		// Your protected handler logic
		return e.JSON(200, map[string]string{"msg": "You are authenticated!"})
	})
}
