package comments

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/comments/handler"
	"github.com/ynwd/awesome-blog/internal/comments/repo"
	"github.com/ynwd/awesome-blog/internal/comments/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
)

type Module struct {
	handler      *handler.CommentsHandler
	pubsub       pubsub.PubSubClient
	eventHandler *handler.CommentsEventHandler
}

func NewModule(firestoreClient *firestore.Client, pubsubClient pubsub.PubSubClient) *Module {
	// Initialize repository
	commentsRepo := repo.NewCommentsRepository(firestoreClient)

	// Initialize service
	commentsService := service.NewCommentsService(commentsRepo)

	// Initialize handler
	commentsHandler := handler.NewCommentsHandler(commentsService, pubsubClient)

	// Initialize event handler
	eventHandler := handler.NewCommentsEventHandler(commentsService)

	return &Module{
		handler:      commentsHandler,
		pubsub:       pubsubClient,
		eventHandler: eventHandler,
	}
}

func (m *Module) RegisterEventHandlers(ctx context.Context, event module.BaseEvent) {
	m.eventHandler.Handle(ctx, event)
}
