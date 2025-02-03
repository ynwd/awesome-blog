package likes

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/likes/handler"
	"github.com/ynwd/awesome-blog/internal/likes/repo"
	"github.com/ynwd/awesome-blog/internal/likes/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
)

type Module struct {
	handler *handler.LikesHandler
	// publishHandler *handler.LikesPublishHandler
	eventHandler *handler.LikeEventHandler
	pubsub       pubsub.PubSubClient
}

func NewModule(firestoreClient *firestore.Client, pubsubClient pubsub.PubSubClient) *Module {
	// Initialize repository
	likesRepo := repo.NewLikesRepository(firestoreClient)

	// Initialize service
	likesService := service.NewLikesService(likesRepo)

	// Initialize handler
	likesHandler := handler.NewLikesHandler(likesService, pubsubClient)

	// Initialize event handler
	eventHandler := handler.NewLikeEventHandler(likesService)

	return &Module{
		handler:      likesHandler,
		pubsub:       pubsubClient,
		eventHandler: eventHandler,
	}
}

func (m *Module) RegisterEventHandlers(ctx context.Context, event module.BaseEvent) {
	m.eventHandler.Handle(ctx, event)
}
