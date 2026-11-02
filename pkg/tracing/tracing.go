package tracing

import (
	"context"
	"os"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"github.com/gofiber/fiber/v3"
	"go.elastic.co/apm/v2"
)

// tracer disimpan agar bisa di-flush saat aplikasi berhenti.
var tracer *apm.Tracer

// Setup membuat tracer Elastic APM dari konfigurasi .env.
//
// Tracer default dari package apm dibentuk saat package init, yaitu sebelum
// env.SetupEnvFile() membaca file .env. Akibatnya service name/environment
// tidak akan terbaca dari .env. Karena itu tracer dibuat eksplisit di sini
// setelah .env tersedia.
func Setup() error {
	applyEnvDefaults()

	t, err := apm.NewTracerOptions(apm.TracerOptions{})
	if err != nil {
		return err
	}

	tracer = t
	return nil
}

// Tracer mengembalikan tracer yang sudah dibuat, atau nil jika Setup gagal.
func Tracer() *apm.Tracer {
	return tracer
}

// Span memulai child span dari transaksi/span yang aktif di Fiber context, lalu
// mengembalikan fungsi untuk mengakhirinya. Pemakaian:
//
//	defer tracing.Span(ctx, "Register", "controller")()
//
// Context Fiber ikut diperbarui (ctx.SetContext) agar span berikutnya berantai
// ke span ini. Saat span berakhir, context dikembalikan ke induknya; tanpa ini
// span yang dibuat setelah helper ini selesai (mis. repository setelah
// HashPassword) akan salah menempel ke span helper yang sudah berakhir.
func Span(ctx fiber.Ctx, name, spanType string) func() {
	prev := ctx.Context()
	span, spanCtx := apm.StartSpan(prev, name, spanType)
	if span == nil {
		return func() {}
	}
	ctx.SetContext(spanCtx)

	return func() {
		span.End()
		ctx.SetContext(prev)
	}
}

// StartSpan sama seperti Span, tetapi untuk kode yang tidak memakai Fiber
// (repository, worker, handler WebSocket). Caller wajib memakai ctx hasilnya.
// Span yang dikembalikan selalu non-nil, jadi aman untuk defer span.End().
func StartSpan(ctx context.Context, name, spanType string) (context.Context, *apm.Span) {
	span, spanCtx := apm.StartSpan(ctx, name, spanType)

	return spanCtx, span
}

// StartTransaction memulai transaksi baru untuk alur non-HTTP, misalnya satu
// pesan WebSocket. Jika tracer belum siap, tx bernilai nil dan ctx tidak
// berubah, sehingga caller cukup memeriksa tx != nil.
func StartTransaction(ctx context.Context, name, txType string) (context.Context, *apm.Transaction) {
	if tracer == nil {
		return ctx, nil
	}

	tx := tracer.StartTransaction(name, txType)

	return apm.ContextWithTransaction(ctx, tx), tx
}

// SetDatabase menambahkan detail database pada span agar muncul di tab
// "Databases" APM (bukan sekadar span generik).
func SetDatabase(span *apm.Span, dbType, instance, action string) {
	if span == nil {
		return
	}

	span.Subtype = dbType
	span.Action = action
	span.Context.SetDatabase(apm.DatabaseSpanContext{
		Type:     dbType,
		Instance: instance,
	})
}

// Close mengirim sisa event lalu menutup tracer. Dibatasi 3 detik supaya
// shutdown tidak menggantung ketika APM Server tidak bisa dihubungi.
func Close() {
	if tracer == nil {
		return
	}

	abort := make(chan struct{})
	go func() {
		time.Sleep(3 * time.Second)
		close(abort)
	}()

	tracer.Flush(abort)
	tracer.Close()
}

// applyEnvDefaults menyalin nilai .env ke environment proses. Nilai yang sudah
// di-export di shell tidak ditimpa, sehingga deployment tetap bisa override.
func applyEnvDefaults() {
	defaults := map[string]string{
		"ELASTIC_APM_SERVICE_NAME": envOrDefault("ELASTIC_APM_SERVICE_NAME", "simple-messaging-app"),
		"ELASTIC_APM_SERVER_URL":   envOrDefault("ELASTIC_APM_SERVER_URL", "http://localhost:8200"),
		"ELASTIC_APM_ENVIRONMENT":  envOrDefault("ELASTIC_APM_ENVIRONMENT", "development"),
	}

	for key, value := range defaults {
		if value == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

// envOrDefault membedakan "tidak diisi" dari "diisi kosong" di .env.
func envOrDefault(key, def string) string {
	if value := env.GetEnv(key, ""); value != "" {
		return value
	}

	return def
}
