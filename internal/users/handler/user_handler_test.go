package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/internal/users/service"
	"github.com/ynwd/awesome-blog/pkg/middleware"
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, username, password string) (domain.User, error) {
	args := m.Called(ctx, username, password)
	return args.Get(0).(domain.User), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		reqBody    map[string]string
		mockSetup  func(*MockUserService)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name: "Success",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).Return(nil)
			},
			wantStatus: http.StatusCreated,
			wantRes: res.Response{
				Status:  "success",
				Message: "User registered successfully",
			},
		},
		{
			name: "Empty Username",
			reqBody: map[string]string{
				"username": "",
				"password": "testpass",
			},
			mockSetup:  func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "Invalid request format",
			},
		},
		{
			name: "Username Already Exists",
			reqBody: map[string]string{
				"username": "existing",
				"password": "testpass",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(service.ErrUsernameExists)
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "username already exists",
			},
		},
		{
			name: "Service Error",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("domain.User")).
					Return(errors.New("Internal server error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "Internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			h := NewUserHandler(mockService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Register(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.wantRes, response)
			mockService.AssertExpectations(t)
		})
	}
}

type MockJWTToken struct {
	mock.Mock
}

func (m *MockJWTToken) GenerateToken(username string, ip string) (string, error) {
	args := m.Called(username, ip)
	return args.String(0), args.Error(1)
}

func (m *MockJWTToken) ValidateToken(tokenString string, ip string) (*jwt.Token, error) {
	args := m.Called(tokenString, ip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJWTToken) GetClaims(token *jwt.Token) (*utils.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.Claims), args.Error(1)
}

func setClientIP(c *gin.Context, ip string) {
	// Set multiple headers to ensure IP is detected
	c.Request.Header.Set("X-Real-IP", ip)
	c.Request.Header.Set("X-Forwarded-For", ip)
	c.Request.RemoteAddr = ip + ":12345"
}

func TestLoginHandler(t *testing.T) {
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid"
	tests := []struct {
		name       string
		reqBody    map[string]string
		clientIP   string
		setupMocks func(*MockUserService, *MockJWTToken)
		wantStatus int
		checkRes   func(*testing.T, res.Response)
	}{
		{
			name: "Success with IP Validation",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP: "192.168.1.1",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("GenerateToken", "testuser", "192.168.1.1").
					Return(validToken, nil)
			},
			wantStatus: http.StatusOK,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "success", got.Status)
				assert.Equal(t, "Login successful", got.Message)
				assert.Equal(t, validToken, got.Data)
			},
		},
		{
			name:       "Invalid Request Format",
			reqBody:    map[string]string{},
			clientIP:   "192.168.1.1",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {},
			wantStatus: http.StatusBadRequest,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid request format", got.Message)
			},
		},
		{
			name: "Empty Username",
			reqBody: map[string]string{
				"username": "",
				"password": "testpass",
			},
			clientIP:   "192.168.1.1",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {},
			wantStatus: http.StatusBadRequest,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid request format", got.Message)
			},
		},
		{
			name: "Authentication Failure",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			clientIP: "192.168.1.1",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "wrongpass").
					Return(domain.User{}, errors.New("authentication failed"))
			},
			wantStatus: http.StatusUnauthorized,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid credentials", got.Message)
			},
		},
		{
			name: "Token Generation Error",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP: "192.168.1.1",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("GenerateToken", "testuser", "192.168.1.1").
					Return("", errors.New("token generation failed"))
			},
			wantStatus: http.StatusInternalServerError,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Failed to generate token", got.Message)
			},
		},
		{
			name: "Empty IP Address",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP: "",
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("GenerateToken", "testuser", "").
					Return("", errors.New("empty IP address"))
			},
			wantStatus: http.StatusInternalServerError,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Failed to generate token", got.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			mockJWT := new(MockJWTToken)
			tt.setupMocks(mockService, mockJWT)

			h := NewUserHandler(mockService, mockJWT)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			c.Request, _ = http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")
			setClientIP(c, tt.clientIP)

			h.Login(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			tt.checkRes(t, response)

			mockService.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func setupTestRouter(h *UserHandler, jwt utils.JWT) *gin.Engine {
	r := gin.New()
	r.Use(middleware.AuthMiddleware(jwt))
	r.POST("/login", h.Login)
	return r
}

func TestLoginHandlerWithMiddleware(t *testing.T) {
	validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid"

	// Create test claims
	testClaims := &utils.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "testuser",
		},
		IP: "192.168.1.1",
	}

	// Create test token with claims
	testToken := &jwt.Token{
		Valid:  true,
		Claims: testClaims,
	}

	tests := []struct {
		name       string
		reqBody    map[string]string
		clientIP   string
		authHeader string
		setupMocks func(*MockUserService, *MockJWTToken)
		wantStatus int
		checkRes   func(*testing.T, res.Response)
	}{
		{
			name: "Success with Valid Token and IP",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP:   "192.168.1.1",
			authHeader: "Bearer " + validToken,
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("ValidateToken", validToken, "192.168.1.1").
					Return(testToken, nil)
				jwt.On("GetClaims", testToken).
					Return(testClaims, nil)
				jwt.On("GenerateToken", "testuser", "192.168.1.1").
					Return(validToken, nil)
			},
			wantStatus: http.StatusOK,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "success", got.Status)
				assert.Equal(t, "Login successful", got.Message)
				assert.Equal(t, validToken, got.Data)
			},
		},
		{
			name: "Invalid Token Claims",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP:   "192.168.1.1",
			authHeader: "Bearer " + validToken,
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				jwt.On("ValidateToken", validToken, "192.168.1.1").
					Return(testToken, nil)
				jwt.On("GetClaims", testToken).
					Return(nil, errors.New("invalid token claims type"))
			},
			wantStatus: http.StatusUnauthorized,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid token claims", got.Message)
			},
		},
		{
			name: "IP Mismatch in Claims",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			clientIP:   "10.0.0.2",
			authHeader: "Bearer " + validToken,
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				jwt.On("ValidateToken", validToken, "10.0.0.2").
					Return(nil, errors.New("token IP mismatch: expected 192.168.1.1, got 10.0.0.2"))
			},
			wantStatus: http.StatusUnauthorized,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Contains(t, got.Message, "token IP mismatch")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			mockJWT := new(MockJWTToken)
			tt.setupMocks(mockService, mockJWT)

			h := NewUserHandler(mockService, mockJWT)
			r := setupTestRouter(h, mockJWT)
			w := httptest.NewRecorder()

			jsonBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			setClientIPWithHttp(req, tt.clientIP)

			r.ServeHTTP(w, req)

			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.wantStatus, w.Code)
			tt.checkRes(t, response)

			mockService.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}

func setClientIPWithHttp(req *http.Request, ip string) {
	req.Header.Set("X-Real-IP", ip)
	req.Header.Set("X-Forwarded-For", ip)
	req.RemoteAddr = ip + ":12345"
}
