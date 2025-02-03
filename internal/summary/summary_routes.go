package summary

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(router *gin.Engine) {
	router.POST("/summary", m.handler.GetYearlySummary)
}
