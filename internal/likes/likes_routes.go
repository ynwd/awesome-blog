package likes

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.POST("/likes", m.handler.CreateLike)
	router.POST("/likes/pubsub", m.handler.PublishLike)
}
