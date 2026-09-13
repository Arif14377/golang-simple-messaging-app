package token

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func GenerateRefreshToken() (string, error) {
	return uuid.NewString(), nil
}
