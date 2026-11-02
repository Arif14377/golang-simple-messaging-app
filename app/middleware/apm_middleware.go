package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.elastic.co/apm/v2"
)

// APM mencatat setiap request sebagai transaksi Elastic APM.
//
// skipPaths dipakai untuk route berumur panjang seperti WebSocket, yang tidak
// cocok diukur sebagai satu transaksi HTTP karena handler-nya tetap terbuka
// selama koneksi hidup. Tracer boleh nil (mis. gagal inisialisasi); middleware
// akan menjadi pass-through.
func APM(tracer *apm.Tracer, skipPaths ...string) fiber.Handler {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skip[path] = struct{}{}
	}

	return func(c fiber.Ctx) error {
		if tracer == nil {
			return c.Next()
		}
		if _, ok := skip[c.Path()]; ok {
			return c.Next()
		}

		// Pakai pola route (mis. GET /user/v1/me), bukan c.Path(), supaya nama
		// transaksi tidak meledak menjadi ribuan variasi.
		name := c.Method() + " unmatched"
		if route := c.Endpoint(); route != nil {
			name = c.Method() + " " + route.Path
		}

		tx := tracer.StartTransaction(name, "request")
		c.SetContext(apm.ContextWithTransaction(c.Context(), tx))
		defer tx.End()

		err := c.Next()

		status := c.Response().StatusCode()
		tx.Result = fmt.Sprintf("%d", status)
		tx.Context.SetHTTPStatusCode(status)
		tx.Context.SetFramework("fiber", "v3")

		// Auth middleware menaruh identitas di Locals, jadi baru tersedia
		// setelah c.Next() selesai.
		if userID, ok := c.Locals("userId").(uint); ok && userID != 0 {
			tx.Context.SetUserID(fmt.Sprintf("%d", userID))
		}
		if username, ok := c.Locals("username").(string); ok && username != "" {
			tx.Context.SetUsername(username)
		}

		if err != nil || status >= fiber.StatusInternalServerError {
			tx.Outcome = "failure"
		} else {
			tx.Outcome = "success"
		}

		if err != nil {
			apm.CaptureError(c.Context(), err).Send()
		}

		return err
	}
}
