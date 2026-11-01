package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/bootstrap"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
)

func main() {
	app := bootstrap.NewApplication()

	// Shutdown rapi agar koneksi MongoDB ditutup dan request berjalan tidak
	// terputus saat aplikasi dihentikan (Ctrl+C / SIGTERM).
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit

		log.Println("Shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Failed to shutdown server: %v\n", err)
		}
		if err := database.DisconnectMongo(ctx); err != nil {
			log.Printf("Failed to disconnect MongoDB: %v\n", err)
		}
	}()

	address := fmt.Sprintf("%s:%s", env.GetEnv("APP_HOST", "localhost"), env.GetEnv("APP_PORT", "8888"))
	if err := app.Listen(address); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
