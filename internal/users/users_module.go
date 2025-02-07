package users

import (
	"context"

	"cloud.google.com/go/firestore"

	"github.com/ynwd/awesome-blog/internal/users/handler"
	"github.com/ynwd/awesome-blog/internal/users/repo"
	"github.com/ynwd/awesome-blog/internal/users/service"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type Module struct {
	h *handler.UserHandler
}

func NewModule(firestoreClient *firestore.Client) *Module {
	// Initialize repository
	userRepo := repo.NewFirestoreUserRepository(firestoreClient)

	// Initialize service with repository
	userService := service.NewUserService(userRepo)

	// Initialize handler with service
	jwt, err := utils.NewJWT()
	if err != nil {
		panic(err)
	}

	userHandler := handler.NewUserHandler(userService, jwt)

	return &Module{
		h: userHandler,
	}
}

func (m *Module) RegisterEventHandlers(ctx context.Context, event module.BaseEvent) {}
