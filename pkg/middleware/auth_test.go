package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

// Mock JWT implementation
type mockJWT struct {
	generateTokenFunc func(userID string, fingerprint *utils.TokenFingerprint) (string, error)
	validateTokenFunc func(tokenString string, fingerprint *utils.TokenFingerprint) (*jwt.Token, error)
	getClaimsFunc     func(token *jwt.Token) (*utils.Claims, error)
	revokeTokenFunc   func(tokenID string) error
}

func (m *mockJWT) GenerateToken(userID string, fingerprint *utils.TokenFingerprint) (string, error) {
	return m.generateTokenFunc(userID, fingerprint)
}

func (m *mockJWT) ValidateToken(tokenString string, fingerprint *utils.TokenFingerprint) (*jwt.Token, error) {
	return m.validateTokenFunc(tokenString, fingerprint)
}

func (m *mockJWT) GetClaims(token *jwt.Token) (*utils.Claims, error) {
	return m.getClaimsFunc(token)
}

func (m *mockJWT) RevokeToken(tokenID string) error {
	return m.revokeTokenFunc(tokenID)
}

func setupTestRouter(config AuthConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})
	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "success"})
	})
	return r
}

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		method         string
		setupAuth      func(*http.Request)
		setupMockJWT   func(*mockJWT)
		preRequest     int // number of requests to make before the actual test
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "should allow login route without token",
			path:   "/api/v1/auth/login",
			method: "POST",
			setupAuth: func(req *http.Request) {
				// No auth header needed
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:   "should block protected route without token",
			path:   "/test",
			method: "GET",
			setupAuth: func(req *http.Request) {
				// No auth header
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Missing authorization header"`,
		},
		{
			name:   "should validate valid token",
			path:   "/test",
			method: "GET",
			setupAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer valid.token.here")
			},
			setupMockJWT: func(m *mockJWT) {
				m.validateTokenFunc = func(tokenString string, fingerprint *utils.TokenFingerprint) (*jwt.Token, error) {
					return &jwt.Token{Valid: true}, nil
				}
				m.getClaimsFunc = func(token *jwt.Token) (*utils.Claims, error) {
					return &utils.Claims{
						RegisteredClaims: jwt.RegisteredClaims{
							IssuedAt:  jwt.NewNumericDate(time.Now()),
							Issuer:    "awesome-blog",
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
						},
						UserID: "test-user",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"success"}`,
		},
		{
			name:   "should handle rate limit for unauthenticated requests",
			path:   "/test",
			method: "GET",
			setupAuth: func(req *http.Request) {
				// No auth header
			},
			preRequest:     21, // Exceed the default unauth limit (20)
			expectedStatus: http.StatusTooManyRequests,
			expectedBody:   `{"error":"Rate limit exceeded"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockJWT := &mockJWT{}
			if tt.setupMockJWT != nil {
				tt.setupMockJWT(mockJWT)
			}

			config := NewAuthConfig()
			config.JWT = mockJWT
			router := setupTestRouter(config)

			// Make pre-requests to trigger rate limit
			if tt.preRequest > 0 {
				for i := 0; i < tt.preRequest; i++ {
					preReq, _ := http.NewRequest(tt.method, tt.path, nil)
					if tt.setupAuth != nil {
						tt.setupAuth(preReq)
					}
					w := httptest.NewRecorder()
					router.ServeHTTP(w, preReq)
				}
			}

			// Actual test request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Device-ID", "test-device")
			req.Header.Set("User-Agent", "test-agent")

			if tt.setupAuth != nil {
				tt.setupAuth(req)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestRateLimiting(t *testing.T) {
	config := NewAuthConfig()
	config.RateLimits.UnauthedRequests.MaxAttempts = 2
	config.RateLimits.UnauthedRequests.Window = time.Minute

	// Create new router for each test to ensure clean rate limit state
	router := setupTestRouter(config)

	ip := "192.168.1.1"

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/auth/login", nil)
		// Set client IP consistently
		req.Header.Set("X-Real-IP", ip)
		req.Header.Set("X-Forwarded-For", ip)
		req.RemoteAddr = ip + ":12345"

		router.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should be allowed", i+1)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be rate limited", i+1)
			assert.Contains(t, w.Body.String(), "Rate limit exceeded")
		}
	}
}

func TestTokenRenewal(t *testing.T) {
	mockJWT := &mockJWT{
		validateTokenFunc: func(tokenString string, fingerprint *utils.TokenFingerprint) (*jwt.Token, error) {
			return &jwt.Token{Valid: true}, nil
		},
		getClaimsFunc: func(token *jwt.Token) (*utils.Claims, error) {
			// Return claims with an old issueAt time to trigger renewal
			return &utils.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
					Issuer:    "awesome-blog",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
				},
				UserID: "test-user",
			}, nil
		},
		generateTokenFunc: func(userID string, fingerprint *utils.TokenFingerprint) (string, error) {
			return "new.token.here", nil
		},
	}

	config := NewAuthConfig()
	config.JWT = mockJWT

	router := setupTestRouter(config)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer old.token.here")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "new.token.here", w.Header().Get("X-New-Token"))
}
