package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/internal/users/service"
	"github.com/ynwd/awesome-blog/pkg/res"
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

type MockJWTToken struct {
	mock.Mock
}

func (m *MockJWTToken) GenerateToken(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockJWTToken) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func (m *MockJWTToken) GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
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

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    map[string]string
		setupMocks func(*MockUserService, *MockJWTToken)
		wantStatus int
		checkRes   func(*testing.T, res.Response)
	}{
		{
			name: "Success",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("GenerateToken", "testuser").
					Return("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0dXNlciJ9.signature", nil)
			},
			wantStatus: http.StatusOK,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "success", got.Status)
				assert.Equal(t, "Login successful", got.Message)
				assert.True(t, strings.HasPrefix(got.Data.(string), "eyJ"))
			},
		},
		{
			name:       "Invalid Request Format",
			reqBody:    map[string]string{},
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
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {},
			wantStatus: http.StatusBadRequest,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid request format", got.Message)
			},
		},
		{
			name: "Empty Password",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "",
			},
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {},
			wantStatus: http.StatusBadRequest,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid request format", got.Message)
			},
		},
		{
			name: "Invalid Credentials",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "wrongpass",
			},
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "wrongpass").
					Return(domain.User{}, service.ErrNotFound)
			},
			wantStatus: http.StatusUnauthorized,
			checkRes: func(t *testing.T, got res.Response) {
				assert.Equal(t, "error", got.Status)
				assert.Equal(t, "Invalid credentials", got.Message)
			},
		},
		{
			name: "AuthenticateUser Error",
			reqBody: map[string]string{
				"username": "testuser",
				"password": "testpass",
			},
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(domain.User{}, errors.New("internal error"))
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
			setupMocks: func(ms *MockUserService, jwt *MockJWTToken) {
				expectedUser := domain.User{Id: "123", Username: "testuser"}
				ms.On("AuthenticateUser", mock.Anything, "testuser", "testpass").
					Return(expectedUser, nil)
				jwt.On("GenerateToken", "testuser").
					Return("", errors.New("token generation failed"))
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
