package app

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/config"
	"github.com/ynwd/awesome-blog/pkg/database"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
)

type App struct {
	config      *config.Config
	router      *gin.Engine
	firestoreDB *database.FirestoreDB
	pubsub      pubsub.PubSubClient
	modules     []module.Module
}

func NewApp(cfg *config.Config) *App {
	ctx := context.Background()
	firestoreDB := database.NewFirestore(
		cfg.GoogleCloud.ProjectID,
		cfg.GoogleCloud.FirestoreDB,
	)

	if err := firestoreDB.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Firestore: %v", err)
	}

	// Initialize PubSub
	pubsubClient, err := pubsub.NewPubSubClient(
		cfg.GoogleCloud.ProjectID,
		os.Getenv("GOOGLE_CLOUD_PUBSUB_TOPIC"),
	)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}

	app := &App{
		config:      cfg,
		router:      gin.Default(),
		firestoreDB: firestoreDB,
		pubsub:      pubsubClient,
	}

	app.setupModules()
	app.pubSubSubsribe(ctx)
	return app
}

func (a *App) Router() *gin.Engine {
	return a.router
}

func (a *App) Start() error {
	return a.router.Run(":" + a.config.Application.Ports[0])
}

func (a *App) Close() error {
	return a.firestoreDB.Close()
}
