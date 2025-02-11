package users

import "github.com/gin-gonic/gin"

func (m *Module) RegisterRoutes(r *gin.Engine) {
	r.POST("/api/v1/auth/register", m.h.Register)
	r.POST("/api/v1/auth/login", m.h.Login)
}
