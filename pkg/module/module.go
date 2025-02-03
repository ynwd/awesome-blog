package module

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Module interface {
	RegisterRoutes(router *gin.Engine)
	RegisterEventHandlers(ctx context.Context, baseEvent BaseEvent)
}
