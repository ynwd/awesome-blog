package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type RateLimitConfig struct {
	AuthedRequests   Config // Rate limit for authenticated requests
	UnauthedRequests Config // Rate limit for unauthenticated requests
}

type AuthConfig struct {
	JWT            utils.JWT
	RateLimiter    *RateLimiter
	MaxTokenAge    time.Duration
	AllowedIssuers []string
	SecureCookie   bool
	CookieDomain   string
	// TrustedProxies []string
	// AllowedOrigins  []string
	RateLimits      RateLimitConfig
	rateLimitAuthed *RateLimiter
	rateLimitUnauth *RateLimiter
}

func NewAuthConfig() AuthConfig {
	return AuthConfig{
		MaxTokenAge:    15 * time.Minute,
		AllowedIssuers: []string{os.Getenv("APPLICATION_NAME")},
		SecureCookie:   true,
		// AllowedOrigins: []string{"https://awesome-blog.com"},
		// TrustedProxies: []string{"127.0.0.1"},
		RateLimits: RateLimitConfig{
			AuthedRequests: Config{
				Window:          time.Minute,
				MaxAttempts:     100, // More lenient for authenticated users
				CleanupInterval: 5 * time.Minute,
			},
			UnauthedRequests: Config{
				Window:          time.Minute,
				MaxAttempts:     20, // Stricter for unauthenticated requests
				CleanupInterval: 5 * time.Minute,
			},
		},
	}
}

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
	Wait      int    `json:"wait_seconds,omitempty"`
}

func AuthMiddleware(config AuthConfig) gin.HandlerFunc {
	// Initialize rate limiters
	config.rateLimitAuthed = NewRateLimiter(config.RateLimits.AuthedRequests)
	config.rateLimitUnauth = NewRateLimiter(config.RateLimits.UnauthedRequests)

	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Generate request ID for tracking
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		clientIP := c.ClientIP()
		path := c.Request.URL.Path

		// Apply rate limiting for public paths
		if isPublicPath(path) {
			if !config.rateLimitUnauth.AllowRequest(clientIP) {
				sendRateLimitError(c)
				return
			}
			c.Next()
			return
		}

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			if !config.rateLimitUnauth.AllowRequest(clientIP) {
				sendRateLimitError(c)
				return
			}
			sendError(c, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			if !config.rateLimitUnauth.AllowRequest(clientIP) {
				sendRateLimitError(c)
				return
			}
			sendError(c, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		token := parts[1]

		// Apply authenticated rate limiting
		if !config.rateLimitAuthed.AllowRequest(clientIP) {
			sendRateLimitError(c)
			return
		}

		// Create fingerprint for validation
		fingerprint := &utils.TokenFingerprint{
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			DeviceID:  c.GetHeader("X-Device-ID"),
		}

		// Validate token
		validToken, err := config.JWT.ValidateToken(token, fingerprint)
		if err != nil {
			sendError(c, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Get claims
		claims, err := config.JWT.GetClaims(validToken)
		if err != nil {
			sendError(c, http.StatusUnauthorized, "Invalid claims")
			return
		}

		// Validate token age
		tokenAge := time.Since(claims.IssuedAt.Time)
		if tokenAge > config.MaxTokenAge {
			sendError(c, http.StatusUnauthorized, "Token expired")
			return
		}

		// Validate issuer
		validIssuer := false
		for _, issuer := range config.AllowedIssuers {
			if claims.Issuer == issuer {
				validIssuer = true
				break
			}
		}
		if !validIssuer {
			sendError(c, http.StatusUnauthorized, "Invalid token issuer")
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID)
		c.Set("auth_time", time.Now().UTC())

		// If token is approaching expiry, send new token in response header
		if tokenAge > (config.MaxTokenAge / 2) {
			newToken, err := config.JWT.GenerateToken(claims.UserID, fingerprint)
			if err == nil {
				c.Header("X-New-Token", newToken)
			}
		}

		c.Next()
	}
}

// Update isPublicPath to match test paths
func isPublicPath(path string) bool {
	switch path {
	case "/api/v1/auth/login",
		"/api/v1/auth/register",
		"/login",
		"/register":
		return true
	default:
		return false
	}
}

func sendError(c *gin.Context, status int, message string) {
	requestID, _ := c.Get("request_id")
	response := ErrorResponse{
		Error:     message,
		RequestID: requestID.(string),
	}
	c.JSON(status, response)
	c.Abort()
}

func sendRateLimitError(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	response := ErrorResponse{
		Error:     "Rate limit exceeded",
		RequestID: requestID.(string),
		Wait:      60, // Suggest waiting for 1 minute
	}
	c.Header("Retry-After", "60")
	c.JSON(http.StatusTooManyRequests, response)
	c.Abort()
}
