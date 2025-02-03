package comments

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.POST("/comments", m.handler.CreateComment)
	router.POST("/comments/pubsub", m.handler.PublishComment)
}
