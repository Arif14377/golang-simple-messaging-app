package middleware

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/cookie"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/jwt"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/response"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/token"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func Auth(ctx fiber.Ctx) error {
	// ambil access token dari cookie (fallback: header "Authorization: Bearer ...")
	tokenStr := cookie.AccessTokenFromRequest(ctx)
	if tokenStr == "" {
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Missing authentication token")
	}

	// verify/validasi token.
	claims, err := jwt.VerifyToken(tokenStr)
	if err != nil {
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Invalid or expired token.")
	}

	// cek session
	tokenHash := token.HashToken(tokenStr)

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	if _, err := repository.FindSessionByToken(cctx, tokenHash); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Session revoked.")
		}
		log.Printf("Failed to find session: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Internal server error.")
	}

	// set context
	ctx.Locals("userId", claims.UserId)
	ctx.Locals("username", claims.Username)

	// return ctx.Next()
	return ctx.Next()
}
