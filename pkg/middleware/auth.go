package middleware

import (
	"net/http"
	"strings"
	"time"

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
			c.JSON(http.StatusUnauthorized, res.Error("Authorization header required"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, res.Error("Invalid authorization format"))
			c.Abort()
			return
		}

		// Get current client IP
		clientIP := c.ClientIP()

		// Validate token with IP check
		token, err := jwt.ValidateToken(parts[1], clientIP)
		if err != nil {
			c.JSON(http.StatusUnauthorized, res.Error(err.Error()))
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
		c.Set("username", claims.Username) // Changed from claims.Subject to claims.Username

		// Add request timestamp for rate limiting
		c.Set("request_time", time.Now().Unix())

		c.Next()
	}
}
