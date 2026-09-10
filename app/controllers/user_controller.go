package controllers

import (
	"context"
	"log"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/dto"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/jwt"
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

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	if err := repository.InsertNewUser(cctx, user); err != nil {
		log.Printf("failed to insert new user: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "failed to register user")
	}

	userRegistered := dto.NewRegisterResponse(*user)

	return response.SendSuccessResponse(ctx, fiber.StatusCreated, "user registered successfully", userRegistered)
}

func Login(ctx fiber.Ctx) error {
	// ambil body request
	var req dto.LoginRequest
	if err := ctx.Bind().Body(&req); err != nil {
		log.Printf("Failed to bind body request: %v", err)
		return response.SendFailureResponse(ctx, fiber.ErrBadRequest.Code, "Failed to bind body request")
	}

	if err := req.Validate(); err != nil {
		log.Printf("Failed to validate request: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	// cek username di database, ambil data user jika ada
	user, err := repository.FindUserByUsername(ctx, req.Username)
	if err != nil {
		log.Printf("User not found: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusNotFound, "User not found.")
	}

	// compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Failed to compare hash and password: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "Wrong password")
	}

	// buat token jwt
	token, err := jwt.GenerateToken(user)
	if err != nil {
		log.Printf("Failed to generate access token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	refreshToken, err := jwt.GenerateRefreshToken()
	if err != nil {
		log.Printf("Failed to generate refresh token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	// simpan session
	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	session := models.UserSession{
		UserID:              user.ID,
		Token:               token,
		RefreshToken:        refreshToken,
		TokenExpired:        time.Now().Add(jwt.AccessTokenDuration),
		RefreshTokenExpired: time.Now().Add(jwt.RefreshTokenDuration),
	}

	if err := repository.InsertUserSession(cctx, &session); err != nil {
		log.Printf("Failed to insert user session: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	// kirim response ke client
	loginResponse := dto.LoginResponse{
		Username:     user.Username,
		FullName:     user.FullName,
		Token:        token,
		RefreshToken: refreshToken,
	}

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "Login successful", loginResponse)
}
