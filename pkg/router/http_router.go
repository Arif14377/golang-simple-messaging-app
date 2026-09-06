package router

import (
	"github.com/Arif14377/golang-simple-messaging-app/app/controllers"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/csrf"
)

type HttpRouter struct{}

func (h HttpRouter) InstallRouter(app *fiber.App) {
	group := app.Group("", cors.New(), csrf.New())
	group.Get("/", controllers.RenderHello)
}

func NewHttpRouter() *HttpRouter {
	return &HttpRouter{}
}
