package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/pkg/module"
)

func TestPostsEventHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		event   module.BaseEvent
		mockFn  func(*mockPostsService)
		wantErr bool
	}{
		{
			name: "different event type returns nil",
			event: module.BaseEvent{
				Type: "different_type",
			},
			mockFn:  func(m *mockPostsService) {},
			wantErr: false,
		},
		{
			name: "invalid json payload returns error",
			event: module.BaseEvent{
				Type:    module.PostEvent,
				Payload: json.RawMessage(`invalid json`),
			},
			mockFn:  func(m *mockPostsService) {},
			wantErr: true,
		},
		{
			name: "service error returns error",
			event: module.BaseEvent{
				Type: module.PostEvent,
				Payload: func() json.RawMessage {
					post := domain.Posts{
						Username:    "testuser",
						Title:       "Test Post",
						Description: "Test Description",
						CreatedAt:   time.Now(),
					}
					b, _ := json.Marshal(post)
					return b
				}(),
			},
			mockFn: func(m *mockPostsService) {
				m.createPostFunc = func(ctx context.Context, post domain.Posts) (string, error) {
					return "", errors.New("repository error")
				}
			},
			wantErr: true,
		},
		{
			name: "successful handling returns nil",
			event: module.BaseEvent{
				Type: module.PostEvent,
				Payload: func() json.RawMessage {
					post := domain.Posts{
						Username:    "testuser",
						Title:       "Test Post",
						Description: "Test Description",
						CreatedAt:   time.Now(),
					}
					b, _ := json.Marshal(post)
					return b
				}(),
			},
			mockFn: func(m *mockPostsService) {
				m.createPostFunc = func(ctx context.Context, post domain.Posts) (string, error) {
					return "post-123", nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockPostsService{}
			tt.mockFn(mockService)
			handler := NewPostEventHandler(mockService)

			err := handler.Handle(context.Background(), tt.event)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
