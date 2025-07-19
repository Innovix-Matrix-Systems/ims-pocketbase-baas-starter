package app

import (
	"log"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	// registers all Go migrations
	_ "ims-pocketbase-baas-starter/internal/database/migrations"
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
