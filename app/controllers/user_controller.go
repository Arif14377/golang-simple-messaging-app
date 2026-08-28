package controllers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/kooroshh/fiber-boostrap/app/models"
)

func Register(ctx fiber.Ctx) error {
	user := new(models.User)
	err := ctx.Bind().Body(user)
	if err != nil {
		errResponse := fmt.Errorf("failed to parse request: %v", err)
		log.Println(errResponse)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, errResponse.Error(), nil)
	}

	return ctx.SendStatus(fiber.StatusOK)
}
