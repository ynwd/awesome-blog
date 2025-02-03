package posts

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.POST("/post", m.handler.CreatePost)
	router.POST("/post/pubsub", m.handler.PublishPost)
}
