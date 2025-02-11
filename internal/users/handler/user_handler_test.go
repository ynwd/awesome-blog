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
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/pkg/utils"
)

// Mock services
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

type MockJWT struct {
	mock.Mock
}

func (m *MockJWT) GenerateToken(userID string, fingerprint *utils.TokenFingerprint) (string, error) {
	args := m.Called(userID, fingerprint)
	return args.String(0), args.Error(1)
}

func (m *MockJWT) ValidateToken(tokenString string, fingerprint *utils.TokenFingerprint) (*jwt.Token, error) {
	args := m.Called(tokenString, fingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJWT) GetClaims(token *jwt.Token) (*utils.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.Claims), args.Error(1)
}

func (m *MockJWT) RevokeToken(tokenID string) error {
	args := m.Called(tokenID)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
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
				m.On("CreateUser", mock.Anything, domain.User{
					Username: "testuser",
					Password: "testpass",
				}).Return(nil)
			},
			wantStatus: http.StatusCreated,
			wantRes: res.Response{
				Status:  "success",
				Message: "User registered successfully",
			},
		},
		{
			name: "Invalid Request",
			reqBody: map[string]string{
				"username": "",
			},
			mockSetup:  func(m *MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "Invalid request format",
			},
		},
		{
			name: "Service Error",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.Anything).
					Return(errors.New("service error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "service error",
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
			c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Register(c)

			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantRes, response)
			mockService.AssertExpectations(t)
		})
	}
}

// Add helper function to set up client IP in gin context
func setupTestContext(c *gin.Context, ip string) {
	// Set the X-Real-IP header
	c.Request.Header.Set("X-Real-IP", ip)
	// Set the X-Forwarded-For header
	c.Request.Header.Set("X-Forwarded-For", ip)
	// Set the RemoteAddr
	c.Request.RemoteAddr = ip + ":12345"
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Define test fingerprint
	testFingerprint := &utils.TokenFingerprint{
		IP:        "192.168.1.1",
		UserAgent: "test-agent",
		DeviceID:  "test-device",
	}

	tests := []struct {
		name       string
		reqBody    map[string]string
		setupMocks func(*MockUserService, *MockJWT)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name: "Success",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			setupMocks: func(ms *MockUserService, mj *MockJWT) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(domain.User{Username: "testuser"}, nil)

				// Update mock expectation with exact fingerprint matching
				mj.On("GenerateToken", "testuser", mock.MatchedBy(func(f *utils.TokenFingerprint) bool {
					return f.IP == testFingerprint.IP &&
						f.UserAgent == testFingerprint.UserAgent &&
						f.DeviceID == testFingerprint.DeviceID
				})).Return("valid.token", nil)
			},
			wantStatus: http.StatusOK,
			wantRes: res.Response{
				Status:  "success",
				Message: "Login successful",
				Data:    "valid.token",
			},
		},
		{
			name: "Invalid Request",
			reqBody: map[string]string{
				"username": "",
			},
			setupMocks: func(ms *MockUserService, mj *MockJWT) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "Invalid request format",
			},
		},
		{
			name: "Authentication Failed",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			setupMocks: func(ms *MockUserService, mj *MockJWT) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "wrongpass").
					Return(domain.User{}, errors.New("invalid credentials"))
			},
			wantStatus: http.StatusUnauthorized,
			wantRes: res.Response{
				Status:  "error",
				Message: "Invalid credentials",
			},
		},
		{
			name: "Token Generation Failed",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			setupMocks: func(ms *MockUserService, mj *MockJWT) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(domain.User{Username: "testuser"}, nil)
				mj.On("GenerateToken", "testuser", mock.Anything).
					Return("", errors.New("token generation failed"))
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "Failed to generate token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			mockJWT := new(MockJWT)
			tt.setupMocks(mockService, mockJWT)
			h := NewUserHandler(mockService, mockJWT)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))

			// Set all required headers
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("User-Agent", testFingerprint.UserAgent)
			req.Header.Set("X-Device-ID", testFingerprint.DeviceID)

			// Set the request in gin context
			c.Request = req

			// Set up client IP properly
			setupTestContext(c, testFingerprint.IP)

			h.Login(c)

			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantRes, response)
			mockService.AssertExpectations(t)
			mockJWT.AssertExpectations(t)
		})
	}
}
