package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

func AuthMiddleware(jwt utils.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for login and register paths
		path := c.Request.URL.Path
		if path == "/login" || path == "/register" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, res.Error("Authorization header is required"))
			c.Abort()
			return
		}

		// Bearer token-string
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, res.Error("Invalid authorization header format"))
			c.Abort()
			return
		}

		token, err := jwt.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, res.Error("Invalid token"))
			c.Abort()
			return
		}

		claims, err := jwt.GetClaims(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, res.Error("Invalid token claims"))
			c.Abort()
			return
		}

		// Store username from token in context
		c.Set("username", claims["sub"])
		c.Next()
	}
}
