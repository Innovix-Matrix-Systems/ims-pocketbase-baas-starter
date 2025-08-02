package routes

import (
	"ims-pocketbase-baas-starter/internal/handlers/route"
	"ims-pocketbase-baas-starter/internal/middlewares"
	"ims-pocketbase-baas-starter/pkg/permission"

	"github.com/pocketbase/pocketbase/core"
)

func applyAuthAndPermissionCheck(e *core.RequestEvent, authMiddleware *middlewares.AuthMiddleware, permissionMiddleware *middlewares.PermissionMiddleware, permission string) error {
	// Apply authentication middleware
	authFunc := authMiddleware.RequireAuthFunc()
	if err := authFunc(e); err != nil {
		return err
	}

	// Apply permission middleware if permission is provided
	if permission != "" {
		permissionFunc := permissionMiddleware.RequirePermission(permission)
		if err := permissionFunc(e); err != nil {
			return err
		}
	}

	return nil
}

// Register Custom made routes here
func RegisterCustom(e *core.ServeEvent) {
	authMiddleware := middlewares.NewAuthMiddleware()
	permissionMiddleware := middlewares.NewPermissionMiddleware()

	g := e.Router.Group("/api/v1")

	//public route
	g.GET("/hello", func(request *core.RequestEvent) error {
		return request.JSON(200, map[string]string{"msg": "Hello from custom route"})
	})

	//auth protected route
	g.GET("/protected", func(request *core.RequestEvent) error {
		// Apply authentication middleware
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(request); err != nil {
			return err
		}
		// Your protected handler logic
		return request.JSON(200, map[string]string{"msg": "You are authenticated!"})
	})

	//Permission protected route
	g.GET("/permission-test", func(request *core.RequestEvent) error {
		if err := applyAuthAndPermissionCheck(request, authMiddleware, permissionMiddleware, permission.UserCreate); err != nil {
			return err
		}
		// Your protected handler logic
		return request.JSON(200, map[string]string{"msg": "You have the User create permission!"})
	})

	//user export
	g.POST("/users/export", func(request *core.RequestEvent) error {
		if err := applyAuthAndPermissionCheck(request, authMiddleware, permissionMiddleware, permission.UserExport); err != nil {
			return err
		}

		return route.HandleUserExport(request)
	})

	g.GET("/jobs/{id}/status", func(request *core.RequestEvent) error {
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(request); err != nil {
			return err
		}

		return route.HandleGetJobStatus(request)
	})

	g.POST("/jobs/{id}/download", func(request *core.RequestEvent) error {
		authFunc := authMiddleware.RequireAuthFunc()
		if err := authFunc(request); err != nil {
			return err
		}

		return route.HandleDownloadJobFile(request)
	})
}
