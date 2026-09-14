package controllers

import "github.com/gofiber/fiber/v3"

func RenderUI(c fiber.Ctx) error {
	return c.Render("index", nil)
}
