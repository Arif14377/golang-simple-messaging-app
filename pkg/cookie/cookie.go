package cookie

import (
	"strconv"
	"strings"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"github.com/gofiber/fiber/v3"
)

const (
	// AccessTokenKey adalah nama cookie penyimpan access token (JWT).
	AccessTokenKey = "access_token"
	// RefreshTokenKey adalah nama cookie penyimpan refresh token.
	RefreshTokenKey = "refresh_token"
	// CookiePath menentukan path yang berhak menerima cookie.
	CookiePath = "/"
)

// isSecure menentukan atribut "Secure" pada cookie.
// Set COOKIE_SECURE=true saat aplikasi sudah berjalan di HTTPS.
func isSecure() bool {
	value, err := strconv.ParseBool(strings.TrimSpace(env.GetEnv("COOKIE_SECURE", "false")))
	if err != nil {
		return false // nilai tidak dikenal → anggap false (default aman untuk dev)
	}
	return value
}

// newAuthCookie membuat cookie yang aman untuk menyimpan token.
func newAuthCookie(key, value string, ttl time.Duration) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     key,
		Value:    value,
		Path:     CookiePath,
		Domain:   env.GetEnv("COOKIE_DOMAIN", ""),
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
		Secure:   isSecure(),
		HTTPOnly: true,                        // tidak bisa dibaca JavaScript (anti XSS)
		SameSite: fiber.CookieSameSiteLaxMode, // mitigasi CSRF
	}
}

// SetAuth menyimpan access token & refresh token ke cookie HttpOnly.
// Browser akan otomatis mengirimkannya pada setiap request berikutnya.
func SetAuth(ctx fiber.Ctx, accessToken, refreshToken string, accessTTL, refreshTTL time.Duration) {
	ctx.Cookie(newAuthCookie(AccessTokenKey, accessToken, accessTTL))
	ctx.Cookie(newAuthCookie(RefreshTokenKey, refreshToken, refreshTTL))
}

// expiredCookie membuat cookie kosong yang langsung kedaluwarsa. Browser akan
// menghapus cookie dengan nama + domain + path yang sama persis.
func expiredCookie(key string) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     key,
		Value:    "",
		Path:     CookiePath,
		Domain:   env.GetEnv("COOKIE_DOMAIN", ""),
		MaxAge:   -1, // fasthttp menuliskannya sebagai "Max-Age=0"
		Expires:  time.Now().Add(-time.Hour),
		Secure:   isSecure(),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	}
}

// ClearAuth menghapus kedua cookie token di sisi browser.
//
// Jangan pakai ctx.ClearCookie() di sini: fiber v3 meneruskannya ke fasthttp
// DelClientCookie(), yang TIDAK mengirim atribut Path/Domain (lihat catatan di
// fasthttp header.go: "This doesn't work for a cookie with specific domain or
// path"). Cookie kita di-set dengan Path=/, sedangkan cookie penghapus tanpa
// Path akan memakai default-path dari URL request (mis. /user/v1), sehingga
// cookie lama tidak ikut terhapus.
func ClearAuth(ctx fiber.Ctx) {
	ctx.Cookie(expiredCookie(AccessTokenKey))
	ctx.Cookie(expiredCookie(RefreshTokenKey))
}

// AccessToken mengambil access token dari cookie.
func AccessToken(ctx fiber.Ctx) string {
	return ctx.Cookies(AccessTokenKey)
}

// RefreshToken mengambil refresh token dari cookie.
func RefreshToken(ctx fiber.Ctx) string {
	return ctx.Cookies(RefreshTokenKey)
}

// AccessTokenFromRequest mengambil access token dari cookie dan, jika tidak ada,
// fallback ke header "Authorization: Bearer <token>" untuk client non-browser
// (mis. curl, Postman, aplikasi mobile).
func AccessTokenFromRequest(ctx fiber.Ctx) string {
	if value := AccessToken(ctx); value != "" {
		return value
	}

	parts := strings.SplitN(ctx.Get(fiber.HeaderAuthorization), " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return ""
}
