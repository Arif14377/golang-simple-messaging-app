package main

import (
	"fmt"
	"log"

	"github.com/Arif14377/golang-simple-messaging-app/bootstrap"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/env"
)

func main() {
	app := bootstrap.NewApplication()
	log.Fatal(app.Listen(fmt.Sprintf("%s:%s", env.GetEnv("APP_HOST", "localhost"), env.GetEnv("APP_PORT", "4000"))))
}
