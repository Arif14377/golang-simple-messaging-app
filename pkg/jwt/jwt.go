package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
)

type CustomClaims struct {
	UserId   uint   `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(ctx context.Context, user models.User) (token string, err error) {
	_, span := tracing.StartSpan(ctx, "GenerateToken", "jwt")
	defer span.End()

	secret := env.GetEnv("JWT_SECRET", "secret")

	claims := CustomClaims{
		UserId:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "Login",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenDuration)),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token, err
}

func VerifyToken(ctx context.Context, tokenString string) (*CustomClaims, error) {
	_, span := tracing.StartSpan(ctx, "VerifyToken", "jwt")
	defer span.End()

	secret := env.GetEnv("JWT_SECRET", "secret")

	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return claims, nil
}
