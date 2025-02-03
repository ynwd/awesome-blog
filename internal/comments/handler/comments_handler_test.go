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
	"github.com/ynwd/awesome-blog/internal/comments/domain"
	"github.com/ynwd/awesome-blog/internal/comments/dto"
	"github.com/ynwd/awesome-blog/pkg/res"
)

type mockCommentsService struct {
	createCommentFunc func(ctx context.Context, comment domain.Comments) error
}

func (m *mockCommentsService) CreateComment(ctx context.Context, comment domain.Comments) error {
	if m.createCommentFunc != nil {
		return m.createCommentFunc(ctx, comment)
	}
	return nil
}

type mockPubSub struct {
	publishFunc   func(ctx context.Context, event interface{}) error
	subscribeFunc func(ctx context.Context, subscriptionID string, handler func(event interface{})) error
}

func (m *mockPubSub) Publish(ctx context.Context, event interface{}) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, event)
	}
	return nil
}

func (m *mockPubSub) Subscribe(ctx context.Context, subscriptionID string, handler func(event interface{})) error {
	if m.subscribeFunc != nil {
		return m.subscribeFunc(ctx, subscriptionID, handler)
	}
	return nil
}

func (m *mockPubSub) Close() {}

func TestPublishComment(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		setupMocks func(*mockCommentsService, *mockPubSub)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name:       "invalid json payload",
			payload:    "invalid",
			setupMocks: func(s *mockCommentsService, p *mockPubSub) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "Invalid request payload",
			},
		},
		{
			name: "pubsub error",
			payload: domain.Comments{
				PostID:   "post1",
				Username: "user1",
				Comment:  "test comment",
			},
			setupMocks: func(s *mockCommentsService, p *mockPubSub) {
				p.publishFunc = func(ctx context.Context, event interface{}) error {
					return errors.New("pubsub error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "Failed to publish comments event",
			},
		},
		{
			name: "successful publish",
			payload: domain.Comments{
				PostID:   "post1",
				Username: "user1",
				Comment:  "test comment",
			},
			setupMocks: func(s *mockCommentsService, p *mockPubSub) {
				p.publishFunc = func(ctx context.Context, event interface{}) error {
					return nil
				}
			},
			wantStatus: http.StatusCreated,
			wantRes: res.Response{
				Status:  "success",
				Message: "comments event published successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCommentsService{}
			mockPub := &mockPubSub{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockService, mockPub)
			}

			handler := NewCommentsHandler(mockService, mockPub)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			switch p := tt.payload.(type) {
			case string:
				body = []byte(p)
			default:
				body, _ = json.Marshal(p)
			}

			c.Request, _ = http.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.PublishComment(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			var response res.Response
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, tt.wantRes, response)
		})
	}
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		setupMock  func(*mockCommentsService)
		wantStatus int
		wantRes    res.Response
	}{
		{
			name:       "invalid json payload",
			payload:    "invalid-json",
			setupMock:  func(m *mockCommentsService) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "invalid character 'i' looking for beginning of value",
			},
		},
		{
			name:       "missing required fields",
			payload:    dto.CreateCommentRequest{},
			setupMock:  func(m *mockCommentsService) {},
			wantStatus: http.StatusBadRequest,
			wantRes: res.Response{
				Status:  "error",
				Message: "Key: 'CreateCommentRequest.Username' Error:Field validation for 'Username' failed on the 'required' tag\nKey: 'CreateCommentRequest.PostID' Error:Field validation for 'PostID' failed on the 'required' tag\nKey: 'CreateCommentRequest.Comment' Error:Field validation for 'Comment' failed on the 'required' tag",
			},
		},
		{
			name: "service error",
			payload: dto.CreateCommentRequest{
				Username: "user1",
				PostID:   "post1",
				Comment:  "test comment",
			},
			setupMock: func(m *mockCommentsService) {
				m.createCommentFunc = func(ctx context.Context, comment domain.Comments) error {
					return errors.New("service error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantRes: res.Response{
				Status:  "error",
				Message: "service error",
			},
		},
		{
			name: "successful creation",
			payload: dto.CreateCommentRequest{
				Username: "user1",
				PostID:   "post1",
				Comment:  "test comment",
			},
			setupMock: func(m *mockCommentsService) {
				m.createCommentFunc = func(ctx context.Context, comment domain.Comments) error {
					return nil
				}
			},
			wantStatus: http.StatusCreated,
			wantRes: res.Response{
				Status:  "success",
				Message: "Comment created successfully",
				Data: domain.Comments{
					Username: "user1",
					PostID:   "post1",
					Comment:  "test comment",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCommentsService{}
			if tt.setupMock != nil {
				tt.setupMock(mockService)
			}

			handler := NewCommentsHandler(mockService, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var body []byte
			switch p := tt.payload.(type) {
			case string:
				body = []byte(p)
			default:
				body, _ = json.Marshal(p)
			}

			c.Request, _ = http.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			handler.CreateComment(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			var gotRes res.Response
			err := json.Unmarshal(w.Body.Bytes(), &gotRes)
			assert.NoError(t, err)

			if gotRes.Data != nil {
				dataBytes, err := json.Marshal(gotRes.Data)
				assert.NoError(t, err)
				var comment domain.Comments
				err = json.Unmarshal(dataBytes, &comment)
				assert.NoError(t, err)
				gotRes.Data = comment
			}

			assert.Equal(t, tt.wantRes, gotRes)
		})
	}
}
