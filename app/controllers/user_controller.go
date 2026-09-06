package controllers

import (
	"log"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/dto"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/response"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
)

func Register(ctx fiber.Ctx) error {
	user := new(models.User)
	if err := ctx.Bind().Body(user); err != nil {
		log.Printf("failed to parse request: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "failed to parse request")
	}

	if err := user.Validate(); err != nil {
		log.Printf("failed to validate request: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to encrypt password: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "failed to register user")
	}

	user.Password = string(hashPassword)

	if err := repository.InsertNewUser(ctx, user); err != nil {
		log.Printf("failed to insert new user: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "failed to register user")
	}

	userRegistered := dto.NewUserRegister(*user)

	return response.SendSuccessResponse(ctx, fiber.StatusCreated, "user registered successfully", userRegistered)
}
