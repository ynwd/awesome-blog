package users

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.POST("/register", m.h.Register)
	r.POST("/login", m.h.Login)
}
