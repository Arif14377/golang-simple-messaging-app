package bootstrap

import (
	"io"
	"log"
	"os"
	"runtime/debug"

	"github.com/Arif14377/golang-simple-messaging-app/app/middleware"
	"github.com/Arif14377/golang-simple-messaging-app/app/ws"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/router"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/template/html/v3"
)

func NewApplication() *fiber.App {
	env.SetupEnvFile()
	logWriter := SetupLogFile()
	if err := tracing.Setup(); err != nil {
		log.Printf("Failed to initialize APM tracer: %v\n", err)
	}
	database.SetupDatabase()
	database.SetupMongo()
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})
	app.Use(middleware.APM(tracing.Tracer(), ws.WSMessagePath))
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e any) {
			log.Printf("panic: %v\n%s", e, debug.Stack())
		},
	}))
	app.Use(logger.New(logger.Config{Stream: logWriter}))
	app.Get("/dashboard", monitor.New())

	// Chat WebSocket berbagi app & origin dengan halaman utama, sehingga cookie
	// sesi (HttpOnly) otomatis ikut saat handshake.
	ws.RegisterWSRoutes(app)

	router.InstallRouter(app)

	return app
}

func SetupLogFile() io.Writer {
	logFile, err := os.OpenFile("./logs/simple_messaging_app.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	return mw
}
