package controllers

import (
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/gofiber/fiber/v3"
)

func RenderUI(c fiber.Ctx) error {
	defer tracing.Span(c, "RenderUI", "controller")()

	return c.Render("index", nil)
}
