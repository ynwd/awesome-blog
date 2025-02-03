package summary

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/ynwd/awesome-blog/internal/summary/handler"
	"github.com/ynwd/awesome-blog/internal/summary/repo"
	"github.com/ynwd/awesome-blog/internal/summary/service"
	"github.com/ynwd/awesome-blog/pkg/module"
)

type Module struct {
	handler *handler.SummaryHandler
}

func NewModule(firestoreClient *firestore.Client) *Module {
	// Initialize repository
	summaryRepo := repo.NewSummaryRepository(firestoreClient)

	// Initialize service
	summaryService := service.NewSummaryService(summaryRepo)

	// Initialize handler
	summaryHandler := handler.NewSummaryHandler(summaryService)

	return &Module{
		handler: summaryHandler,
	}
}

func (m *Module) RegisterEventHandlers(ctx context.Context, event module.BaseEvent) {}
