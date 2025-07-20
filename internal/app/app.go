package app

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	"github.com/pocketbase/pocketbase/tools/hook"

	_ "ims-pocketbase-baas-starter/internal/database/migrations" //side effect migration load(from pocketbase)
	"ims-pocketbase-baas-starter/internal/middlewares"
	"ims-pocketbase-baas-starter/internal/routes"
)

func Run() {
	app := pocketbase.New()

	// v0.29: register the official migratecmd plugin
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate:  isGoRun, // auto-create migration files only in dev
		TemplateLang: migratecmd.TemplateLangGo,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		middleware := middlewares.NewAuthMiddleware()

		// Apply auth to specific PocketBase API endpoints
		se.Router.Bind(&hook.Handler[*core.RequestEvent]{
			Id: "jwtAuth",
			Func: func(e *core.RequestEvent) error {
				path := e.Request.URL.Path

				// Define collections that require authentication
				protectedCollections := []string{"users", "roles", "permissions"} // Add your collections here

				// Define endpoints to exclude from authentication
				excludedPaths := []string{
					"/api/collections/users/auth-with-password",
					"/api/collections/users/auth-refresh",
					"/api/collections/users/request-password-reset",
					"/api/collections/users/confirm-password-reset",
					"/api/collections/users/request-verification",
					"/api/collections/users/confirm-verification",
					"/api/collections/users/request-email-change",
					"/api/collections/users/confirm-email-change",
				}

				// Check if path should be excluded
				for _, excludedPath := range excludedPaths {
					if strings.HasPrefix(path, excludedPath) {
						return e.Next() // Skip auth for excluded paths
					}
				}

				// Check if it's a protected collection endpoint
				for _, collection := range protectedCollections {
					collectionPath := "/api/collections/" + collection
					if strings.HasPrefix(path, collectionPath) {
						authFunc := middleware.RequireAuthFunc()
						if err := authFunc(e); err != nil {
							return err
						}
						break
					}
				}

				return e.Next()
			},
		})

		// static files
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		// custom business routes
		routes.RegisterCustom(se)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
