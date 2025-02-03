package posts

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/posts/handler"
	"github.com/ynwd/awesome-blog/internal/posts/repo"
	"github.com/ynwd/awesome-blog/internal/posts/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/pubsub"
)

type Module struct {
	handler      *handler.PostsHandler
	pubsub       pubsub.PubSubClient
	eventHandler *handler.PostEventHandler
}

func NewModule(firestoreClient *firestore.Client, pubsubClient pubsub.PubSubClient) *Module {
	// Initialize repository
	postsRepo := repo.NewPostsRepository(firestoreClient)

	// Initialize service with repository
	postsService := service.NewPostsService(postsRepo)

	// Initialize handler with service
	postsHandler := handler.NewPostsHandler(postsService, pubsubClient)

	// Initialize event handler with repository
	eventHandler := handler.NewPostEventHandler(postsService)

	return &Module{
		pubsub:       pubsubClient,
		handler:      postsHandler,
		eventHandler: eventHandler,
	}
}

func (m *Module) RegisterEventHandlers(ctx context.Context, event module.BaseEvent) {
	m.eventHandler.Handle(ctx, event)
}
