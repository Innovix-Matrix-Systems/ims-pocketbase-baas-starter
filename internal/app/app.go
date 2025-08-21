package app

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	"ims-pocketbase-baas-starter/internal/crons"
	_ "ims-pocketbase-baas-starter/internal/database/migrations" //side effect migration load(from pocketbase)
	"ims-pocketbase-baas-starter/internal/hooks"
	"ims-pocketbase-baas-starter/internal/jobs"
	"ims-pocketbase-baas-starter/internal/middlewares"
	"ims-pocketbase-baas-starter/internal/routes"
	"ims-pocketbase-baas-starter/internal/swagger"
	"ims-pocketbase-baas-starter/pkg/logger"
	"ims-pocketbase-baas-starter/pkg/metrics"
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

	// Initialize metrics provider early in startup sequence
	// This must be called before hooks and middleware registration
	metricsProvider := metrics.GetInstance()

	// Initialize our logger
	logger := logger.GetLogger(app)
	logger.Info("Metrics provider initialized", "provider", metricsProvider != nil)

	// Initialize job manager and processors during app startup
	// This must be called after app creation but before OnServe setup
	jobManager := jobs.GetJobManager()
	if err := jobManager.Initialize(app); err != nil {
		logger.Error("Failed to initialize job manager", "error", err)
		log.Fatalf("Failed to initialize job manager: %v", err)
	}

	// Register scheduled cron jobs during app initialization phase
	// This must be called after job manager initialization
	logger.Info("Registering scheduled cron jobs")
	crons.RegisterCrons(app)

	// Register custom event hooks
	// This should be called after job manager and crons initialization
	logger.Info("Registering custom event hooks")
	hooks.RegisterHooks(app)

	// Register shutdown hook for metrics provider cleanup
	app.OnTerminate().BindFunc(func(te *core.TerminateEvent) error {
		if metricsProvider != nil {
			logger.Info("Shutting down metrics provider")
			if err := metricsProvider.Shutdown(context.Background()); err != nil {
				logger.Error("Failed to shutdown metrics provider", "error", err)
			}
		}
		return te.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		// Register Prometheus metrics endpoint if provider supports it
		metricsProvider := metrics.GetInstance()
		if handler := metricsProvider.GetHandler(); handler != nil {
			se.Router.GET("/metrics", func(e *core.RequestEvent) error {
				handler.ServeHTTP(e.Response, e.Request)
				return nil
			})
			logger.Info("Metrics endpoint registered", "path", "/metrics")
		}

		// Initialize Swagger generator using singleton pattern
		generator := swagger.InitializeGenerator(app)

		// Register all application middlewares
		logger.Info("Registering middlewares")
		middlewares.RegisterMiddlewares(se)

		// static files
		se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))

		// custom business routes
		routes.RegisterCustom(se)

		// Register Swagger endpoints
		swagger.RegisterEndpoints(se, generator)

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
