package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ynwd/awesome-blog/config"
	"github.com/ynwd/awesome-blog/internal/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	application := app.NewApp(cfg)
	defer application.Close()

	log.Printf("Starting %s on port %s", cfg.Application.Name, cfg.Application.Ports[0])
	if err := application.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
