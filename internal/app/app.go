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

	"ims-pocketbase-baas-starter/internal"
	"ims-pocketbase-baas-starter/internal/crons"
	_ "ims-pocketbase-baas-starter/internal/database/migrations" //side effect migration load(from pocketbase)
	"ims-pocketbase-baas-starter/internal/jobs"
	"ims-pocketbase-baas-starter/internal/middlewares"
	"ims-pocketbase-baas-starter/internal/routes"
)

// NewApp creates and configures a new PocketBase app instance
// This is useful for testing and for the main application
func NewApp() *pocketbase.PocketBase {
	app := pocketbase.New()

	// v0.29: register the official migratecmd plugin
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate:  isGoRun, // auto-create migration files only in dev
		TemplateLang: migratecmd.TemplateLangGo,
	})

	// Initialize job manager and processors during app startup
	// This must be called after app creation but before OnServe setup
	jobManager := jobs.GetJobManager()
	if err := jobManager.Initialize(app); err != nil {
		log.Fatalf("Failed to initialize job manager: %v", err)
	}

	// Register scheduled cron jobs during app initialization phase
	// This must be called after job manager initialization
	crons.RegisterCrons(app)

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		middleware := middlewares.NewAuthMiddleware()

		// Apply auth to specific PocketBase API endpoints
		se.Router.Bind(&hook.Handler[*core.RequestEvent]{
			Id: "jwtAuth",
			Func: func(e *core.RequestEvent) error {
				path := e.Request.URL.Path

				// Check if path should be excluded
				for _, excludedPath := range internal.ExcludedPaths {
					if strings.HasPrefix(path, excludedPath) {
						return e.Next() // Skip auth for excluded paths
					}
				}

				// Check if it's a protected collection endpoint
				for _, collection := range internal.ProtectedCollections {
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

	return app
}

func Run() {
	app := NewApp()

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
