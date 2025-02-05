package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/ynwd/awesome-blog/config"
	"github.com/ynwd/awesome-blog/internal/app"
)

func main() {
	env := os.Getenv("ENV")
	log.Printf("Starting application in %s mode", env)

	if env != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		} else {
			log.Println("Loaded .env file successfully")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Configuration loaded successfully: %s", cfg.Application.Name)

	application := app.NewApp(cfg)
	defer func() {
		log.Println("Cleaning up application resources...")
		if err := application.Close(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting %s on port %s", cfg.Application.Name, cfg.Application.Ports[0])
		if err := application.Start(); err != nil {
			log.Printf("Server error: %v", err)
			cancel()
		}
	}()

	select {
	case <-sigChan:
		log.Println("Shutdown signal received")
	case <-ctx.Done():
		log.Println("Server error occurred")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Println("Shutdown timed out")
		}
	default:
		log.Println("Server shutdown complete")
	}
}
