package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/pkg/module"
)

type mockLikesService struct {
	createLikeFunc func(ctx context.Context, like domain.Likes) error
}

func (m *mockLikesService) CreateLike(ctx context.Context, like domain.Likes) error {
	if m.createLikeFunc != nil {
		return m.createLikeFunc(ctx, like)
	}
	return nil
}

func TestLikesEventHandler_Handle(t *testing.T) {
	tests := []struct {
		name      string
		event     module.BaseEvent
		setupMock func(*mockLikesService)
		wantErr   bool
	}{
		{
			name: "different event type returns nil",
			event: module.BaseEvent{
				Type: "different_type",
			},
			setupMock: func(m *mockLikesService) {},
			wantErr:   false,
		},
		{
			name: "invalid json payload returns error",
			event: module.BaseEvent{
				Type:    module.LikeEvent,
				Payload: []byte("invalid json"),
			},
			setupMock: func(m *mockLikesService) {},
			wantErr:   true,
		},
		{
			name: "service error returns error",
			event: module.BaseEvent{
				Type: module.LikeEvent,
				Payload: func() json.RawMessage {
					b, _ := json.Marshal(domain.Likes{})
					return b
				}(),
			},
			setupMock: func(m *mockLikesService) {
				m.createLikeFunc = func(ctx context.Context, like domain.Likes) error {
					return errors.New("service error")
				}
			},
			wantErr: true,
		},
		{
			name: "successful handling returns nil",
			event: module.BaseEvent{
				Type: module.LikeEvent,
				Payload: func() json.RawMessage {
					b, _ := json.Marshal(domain.Likes{
						UsernameFrom: "user1",
						PostID:       "post1",
					})
					return b
				}(),
			},
			setupMock: func(m *mockLikesService) {
				m.createLikeFunc = func(ctx context.Context, like domain.Likes) error {
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockLikesService{}
			tt.setupMock(mockService)

			handler := NewLikeEventHandler(mockService)
			err := handler.Handle(context.Background(), tt.event)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
