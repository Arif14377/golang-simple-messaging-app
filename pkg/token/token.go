package token

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/google/uuid"
)

func HashToken(ctx context.Context, token string) string {
	_, span := tracing.StartSpan(ctx, "HashToken", "token")
	defer span.End()

	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func GenerateRefreshToken(ctx context.Context) (string, error) {
	_, span := tracing.StartSpan(ctx, "GenerateRefreshToken", "token")
	defer span.End()

	return uuid.NewString(), nil
}
