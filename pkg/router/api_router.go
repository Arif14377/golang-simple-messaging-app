package router

import (
	"github.com/Arif14377/golang-simple-messaging-app/app/controllers"
	"github.com/Arif14377/golang-simple-messaging-app/app/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

type ApiRouter struct{}

func (h ApiRouter) InstallRouter(app *fiber.App) {
	api := app.Group("/api", limiter.New())
	api.Get("/", func(ctx fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Hello from api",
		})
	})

	userGroup := app.Group("/user")
	userV1Group := userGroup.Group("/v1")
	userV1Group.Post("/register", controllers.Register)
	userV1Group.Post("/login", controllers.Login)
	userV1Group.Delete("/logout", middleware.Auth, controllers.Logout)
	userV1Group.Post("/refresh-token", controllers.RefreshToken)

}

func NewApiRouter() *ApiRouter {
	return &ApiRouter{}
}
