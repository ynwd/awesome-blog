package app

import (
	"log"

	"github.com/ynwd/awesome-blog/pkg/middleware"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

// setupMiddleware sets up the middleware for the app
func (a *App) setupMiddleware() {
	blacklist := utils.NewMemoryBlacklist()
	jwt, err := utils.NewJWT(blacklist)
	if err != nil {
		log.Fatal("Failed to create JWT:", err)
	}
	config := middleware.NewAuthConfig()
	config.JWT = jwt
	auth := middleware.AuthMiddleware(config)
	a.router.Use(auth)
}
