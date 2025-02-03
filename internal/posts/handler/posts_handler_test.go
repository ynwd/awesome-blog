package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/internal/posts/dto"
	"github.com/ynwd/awesome-blog/pkg/res"
	"github.com/ynwd/awesome-blog/tests/helper"
)

type mockPostsService struct {
	createPostFunc func(ctx context.Context, post domain.Posts) (string, error)
}

func (m *mockPostsService) CreatePost(ctx context.Context, post domain.Posts) (string, error) {
	return m.createPostFunc(ctx, post)
}

func TestPostsHandler_CreatePost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		reqBody    interface{}
		mockSvcFn  func(ctx context.Context, post domain.Posts) (string, error)
		wantStatus int
		wantResp   *res.Response
	}{
		{
			name: "success",
			reqBody: dto.CreatePostRequest{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
			},
			mockSvcFn: func(ctx context.Context, post domain.Posts) (string, error) {
				return "post-123", nil
			},
			wantStatus: http.StatusCreated,
			wantResp: &res.Response{
				Status:  "success",
				Message: "Post created successfully",
				Data: dto.PostResponse{
					ID:          "post-123",
					Username:    "testuser",
					Title:       "Test Post",
					Description: "Test Description",
				},
			},
		},
		{
			name:       "invalid request body",
			reqBody:    "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			reqBody: dto.CreatePostRequest{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
			},
			mockSvcFn: func(ctx context.Context, post domain.Posts) (string, error) {
				return "", errors.New("service error")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostsService{
				createPostFunc: tt.mockSvcFn,
			}
			mockPubSub := &helper.MockPubSub{}
			handler := NewPostsHandler(mockSvc, mockPubSub)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreatePost(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantResp != nil {
				var got res.Response
				err := json.Unmarshal(w.Body.Bytes(), &got)
				assert.NoError(t, err)

				// Convert response data to PostResponse
				dataBytes, err := json.Marshal(got.Data)
				assert.NoError(t, err)

				var postResp dto.PostResponse
				err = json.Unmarshal(dataBytes, &postResp)
				assert.NoError(t, err)

				assert.Equal(t, tt.wantResp.Status, got.Status)
				assert.Equal(t, tt.wantResp.Message, got.Message)
				assert.Equal(t, tt.wantResp.Data.(dto.PostResponse), postResp)
			}
		})
	}
}

func TestPostsHandler_PublishPost(t *testing.T) {
	gin.SetMode(gin.TestMode)

	currentTime := time.Now()
	tests := []struct {
		name       string
		reqBody    interface{}
		mockPubFn  func(ctx context.Context, event interface{}) error
		wantStatus int
		wantResp   interface{}
	}{
		{
			name: "success",
			reqBody: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
				CreatedAt:   currentTime,
			},
			mockPubFn: func(ctx context.Context, event interface{}) error {
				return nil
			},
			wantStatus: http.StatusCreated,
			wantResp: res.Response{
				Status:  "success",
				Message: "posts event published successfully",
			},
		},
		{
			name:       "invalid request body",
			reqBody:    "invalid json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "pubsub error",
			reqBody: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
				CreatedAt:   currentTime,
			},
			mockPubFn: func(ctx context.Context, event interface{}) error {
				return errors.New("pubsub error")
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostsService{}
			mockPubSub := &helper.MockPubSub{
				PublishFunc: tt.mockPubFn,
			}
			handler := NewPostsHandler(mockSvc, mockPubSub)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBody, _ := json.Marshal(tt.reqBody)
			c.Request = httptest.NewRequest(http.MethodPost, "/posts/publish", bytes.NewBuffer(jsonBody))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.PublishPost(c)

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
