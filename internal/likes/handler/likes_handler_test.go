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
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/internal/likes/dto"
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/tests/helper"
)

func TestLikesHandler_CreateLike(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		payload    interface{}
		setupMocks func(*mockLikesService, *helper.MockPubSub)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name:       "invalid json payload",
			payload:    "invalid",
			setupMocks: func(s *mockLikesService, p *helper.MockPubSub) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "json: cannot unmarshal string into Go value of type dto.CreateLikeRequest",
			},
		},
		{
			name: "service error",
			payload: dto.CreateLikeRequest{
				PostID:       "post1",
				UsernameFrom: "user1",
			},
			setupMocks: func(s *mockLikesService, p *helper.MockPubSub) {
				s.createLikeFunc = func(ctx context.Context, like domain.Likes) error {
					return assert.AnError
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "assert.AnError general error for testing",
			},
		},
		{
			name: "success",
			payload: dto.CreateLikeRequest{
				PostID:       "post1",
				UsernameFrom: "user1",
			},
			setupMocks: func(s *mockLikesService, p *helper.MockPubSub) {
				s.createLikeFunc = func(ctx context.Context, like domain.Likes) error {
					return nil
				}
			},
			wantStatus: http.StatusCreated,
			wantRes: res.Response{
				Status:  "success",
				Message: "Like created successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			mockService := &mockLikesService{}
			mockPubsub := &helper.MockPubSub{}
			tt.setupMocks(mockService, mockPubsub)

			handler := NewLikesHandler(mockService, mockPubsub)
			handler.RegisterRoutes(router)

			payloadBytes, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPost, "/likes", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			var got res.Response
			err := json.Unmarshal(w.Body.Bytes(), &got)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantRes, got)
		})
	}
}

func TestPostsHandler_PublishPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		reqBody    interface{}
		mockPubFn  func(ctx context.Context, event interface{}) error
		wantStatus int
		wantResp   interface{}
	}{
		{
			name: "success",
			reqBody: domain.Likes{
				PostID:       "post1",
				UsernameFrom: "user1",
			},
			mockPubFn: func(ctx context.Context, event interface{}) error {
				return nil
			},
			wantStatus: http.StatusCreated,
			wantResp: res.Response{
				Status:  "success",
				Message: "likes event published successfully",
			},
		},
		{
			name:       "invalid request body",
			reqBody:    "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "pubsub error",
			reqBody: domain.Likes{
				PostID:       "post1",
				UsernameFrom: "user1",
			},
			mockPubFn: func(ctx context.Context, event interface{}) error {
				return errors.New("pubsub error")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockLikesService{}
			mockPubSub := &helper.MockPubSub{
				PublishFunc: tt.mockPubFn,
			}
			handler := NewLikesHandler(mockSvc, mockPubSub)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/likes/pubsub", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.PublishLike(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantResp != nil {
				var got res.Response
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp, got)
			}
		})
	}
}
