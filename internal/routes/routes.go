package routes

import (
	"ims-pocketbase-baas-starter/internal/handlers/route"
	"ims-pocketbase-baas-starter/internal/middlewares"

	"github.com/pocketbase/pocketbase/core"
)

func RegisterCustom(e *core.ServeEvent) {
	authMiddleware := middlewares.NewAuthMiddleware()
	permissionMiddleware := middlewares.NewPermissionMiddleware()

	g := e.Router.Group("/api/v1")

	//public route
	g.GET("/hello", func(e *core.RequestEvent) error {
		return e.JSON(200, map[string]string{"msg": "Hello from custom route"})
	})

	//auth protected route
	g.GET("/protected", func(e *core.RequestEvent) error {
		// Apply authentication middleware
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(e); err != nil {
			return err
		}
		// Your protected handler logic
		return e.JSON(200, map[string]string{"msg": "You are authenticated!"})
	})

	//Permission protected route
	g.GET("/permission-test", func(e *core.RequestEvent) error {
		//apply auth middleware
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(e); err != nil {
			return err
		}
		// Apply permission middleware
		permissionFunc := permissionMiddleware.RequirePermission("user.create")
		if err := permissionFunc(e); err != nil {
			return err
		}
		// Your protected handler logic
		return e.JSON(200, map[string]string{"msg": "You have the User create permission!"})
	})

	//user export
	g.POST("/users/export", func(e *core.RequestEvent) error {
		//apply auth middleware
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(e); err != nil {
			return err
		}
		// Apply permission middleware
		permissionFunc := permissionMiddleware.RequirePermission("user.export")
		if err := permissionFunc(e); err != nil {
			return err
		}

		return route.HandleUserExport(e)
	})
}
