package app

import (
	"log"

	"github.com/ynwd/awesome-blog/internal/comments"
	"github.com/ynwd/awesome-blog/internal/likes"
	"github.com/ynwd/awesome-blog/internal/posts"
	"github.com/ynwd/awesome-blog/internal/summary"
	"github.com/ynwd/awesome-blog/internal/users"
	"github.com/ynwd/awesome-blog/pkg/middleware"
	"github.com/ynwd/awesome-blog/pkg/module"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

func (a *App) setupModules() {

	// Setup middleware
	auth := middleware.AuthMiddleware(utils.NewJWT())
	a.router.Use(auth)

	// Register modules
	client, err := a.firestoreDB.Client()
	if err != nil {
		log.Fatal("Failed to get firestore client:", err)
	}
	modules := []module.Module{
		users.NewModule(client),
		posts.NewModule(client, a.pubsub),
		comments.NewModule(client, a.pubsub),
		likes.NewModule(client, a.pubsub),
		summary.NewModule(client),
	}

	for _, m := range modules {
		m.RegisterRoutes(a.router)
	}

	a.modules = modules
}
