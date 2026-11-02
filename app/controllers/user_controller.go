package controllers

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/cookie"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/dto"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/jwt"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/response"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/token"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Register(ctx fiber.Ctx) error {
	defer tracing.Span(ctx, "Register", "controller")()

	user := new(models.User)
	if err := ctx.Bind().Body(user); err != nil {
		log.Printf("failed to parse request: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "failed to parse request")
	}

	if err := user.Validate(ctx.Context()); err != nil {
		log.Printf("failed to validate request: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, err.Error())
	}

	hashedPassword, err := hashPassword(ctx, user.Password)
	if err != nil {
		log.Printf("failed to encrypt password: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "failed to register user")
	}

	user.Password = hashedPassword

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
	defer tracing.Span(ctx, "Login", "controller")()

	// ambil body request
	var req dto.LoginRequest
	if err := ctx.Bind().Body(&req); err != nil {
		log.Printf("Failed to bind body request: %v", err)
		return response.SendFailureResponse(ctx, fiber.ErrBadRequest.Code, "Failed to bind body request")
	}

	if err := req.Validate(ctx.Context()); err != nil {
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
	if err := comparePassword(ctx, user.Password, req.Password); err != nil {
		log.Printf("Failed to compare hash and password: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "Wrong password")
	}

	// buat token jwt
	tokenStr, err := jwt.GenerateToken(ctx.Context(), user)
	if err != nil {
		log.Printf("Failed to generate access token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	refreshToken, err := token.GenerateRefreshToken(ctx.Context())
	if err != nil {
		log.Printf("Failed to generate refresh token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to login")
	}

	// simpan session
	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	session := models.UserSession{
		UserID:              user.ID,
		TokenHash:           token.HashToken(ctx.Context(), tokenStr),
		RefreshToken:        token.HashToken(ctx.Context(), refreshToken),
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
		Token:        tokenStr,
		RefreshToken: refreshToken,
	}

	// simpan token ke cookie HttpOnly (dikelola server, tidak bisa dibaca JavaScript)
	cookie.SetAuth(ctx, tokenStr, refreshToken, jwt.AccessTokenDuration, jwt.RefreshTokenDuration)

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "Login successful", loginResponse)
}

func Logout(ctx fiber.Ctx) error {
	defer tracing.Span(ctx, "Logout", "controller")()

	// ambil token dari cookie, fallback ke header "Authorization: Bearer ..."
	tokenStr := cookie.AccessTokenFromRequest(ctx)
	if tokenStr == "" {
		cookie.ClearAuth(ctx)
		return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "Empty token.")
	}

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	if err := repository.DeleteSessionByToken(cctx, token.HashToken(ctx.Context(), tokenStr)); err != nil {
		log.Printf("Failed to delete session: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to logout")
	}

	// hapus cookie agar browser berhenti mengirim token
	cookie.ClearAuth(ctx)

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "Logout successful", nil)
}

// Profile mengembalikan data user yang sedang login. Dipakai frontend untuk
// memastikan sesi (cookie) masih valid saat halaman dibuka/di-refresh.
func Profile(ctx fiber.Ctx) error {
	defer tracing.Span(ctx, "Profile", "controller")()

	userID, ok := ctx.Locals("userId").(uint)
	if !ok {
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Unauthorized.")
	}

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	user, err := repository.FindUserByID(cctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "User not found.")
		}
		log.Printf("Failed to find user: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to get profile")
	}

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "User profile", dto.NewRegisterResponse(user))
}

func RefreshToken(ctx fiber.Ctx) error {
	defer tracing.Span(ctx, "RefreshToken", "controller")()

	// refresh token dibaca dari cookie; fallback ke body untuk client non-browser
	refreshToken := cookie.RefreshToken(ctx)

	if refreshToken == "" {
		var req dto.RefreshTokenRequest
		if err := ctx.Bind().Body(&req); err != nil {
			log.Printf("Failed to bind body request: %v\n", err)
		} else if err := req.Validate(ctx.Context()); err != nil {
			log.Printf("No refresh token in cookie/body: %v\n", err)
		} else {
			refreshToken = req.RefreshToken
		}
	}

	// tidak ada token di cookie maupun body → client dianggap belum login
	if refreshToken == "" {
		cookie.ClearAuth(ctx)
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Missing refresh token")
	}

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	// cari session berdasarkan refresh token yang dikirim client
	session, err := repository.FindSessionByRefreshToken(cctx, token.HashToken(ctx.Context(), refreshToken))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Refresh token not found: %v\n", err)
			// session sudah tidak ada (mis. sudah logout) → bersihkan cookie
			cookie.ClearAuth(ctx)
			return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Invalid refresh token")
		}
		log.Printf("Failed to find session: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to refresh token")
	}

	// cek masa berlaku refresh token
	if time.Now().After(session.RefreshTokenExpired) {
		if err := repository.DeleteSessionByID(cctx, session.ID); err != nil {
			log.Printf("Failed to delete expired session: %v\n", err)
		}
		cookie.ClearAuth(ctx)
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Refresh token expired")
	}

	// ambil data user untuk membuat access token baru
	user, err := repository.FindUserByID(cctx, session.UserID)
	if err != nil {
		log.Printf("User not found: %v\n", err)
		cookie.ClearAuth(ctx)
		return response.SendFailureResponse(ctx, fiber.StatusUnauthorized, "Invalid refresh token")
	}

	// buat access token dan refresh token baru
	tokenStr, err := jwt.GenerateToken(ctx.Context(), user)
	if err != nil {
		log.Printf("Failed to generate access token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to refresh token")
	}

	newRefreshToken, err := token.GenerateRefreshToken(ctx.Context())
	if err != nil {
		log.Printf("Failed to generate refresh token: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to refresh token")
	}

	// rotasi session dengan token terbaru
	session.TokenHash = token.HashToken(ctx.Context(), tokenStr)
	session.RefreshToken = token.HashToken(ctx.Context(), newRefreshToken)
	session.RefreshTokenExpired = time.Now().Add(jwt.RefreshTokenDuration)

	if err := repository.UpdateUserSession(cctx, &session); err != nil {
		log.Printf("Failed to update user session: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to refresh token")
	}

	// rotasi cookie dengan token terbaru
	cookie.SetAuth(ctx, tokenStr, newRefreshToken, jwt.AccessTokenDuration, jwt.RefreshTokenDuration)

	refreshResponse := dto.RefreshTokenResponse{
		Token:        tokenStr,
		RefreshToken: newRefreshToken,
	}

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "Token refreshed successfully", refreshResponse)
}

// hashPassword membungkus bcrypt sekaligus mencatat latensinya. Cost bcrypt
// cukup mahal (~50-100 ms), jadi bagian ini layak terlihat terpisah dari
// durasi handler.
func hashPassword(ctx fiber.Ctx, password string) (string, error) {
	defer tracing.Span(ctx, "HashPassword", "crypto")()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// comparePassword membungkus bcrypt.CompareHashAndPassword dengan span.
func comparePassword(ctx fiber.Ctx, hashedPassword, password string) error {
	defer tracing.Span(ctx, "ComparePassword", "crypto")()

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
