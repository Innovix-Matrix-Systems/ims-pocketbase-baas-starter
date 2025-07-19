package routes

import (
	"github.com/pocketbase/pocketbase/core"
)

func RegisterCustom(e *core.ServeEvent) {
	g := e.Router.Group("/api/v1")

	g.GET("/hello", func(e *core.RequestEvent) error {
		return e.JSON(200, map[string]string{"msg": "Hello from custom route"})
	})
}
