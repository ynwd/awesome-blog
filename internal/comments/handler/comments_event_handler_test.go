package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/comments/domain"
	"github.com/ynwd/awesome-blog/pkg/module"
)

func TestCommentsEventHandler_Handle(t *testing.T) {
	tests := []struct {
		name    string
		event   module.BaseEvent
		mockFn  func(*mockCommentsService)
		wantErr bool
	}{
		{
			name: "different event type returns nil",
			event: module.BaseEvent{
				Type: "different_type",
			},
			mockFn:  func(m *mockCommentsService) {},
			wantErr: false,
		},
		{
			name: "invalid json payload returns error",
			event: module.BaseEvent{
				Type:    module.CommentEvent,
				Payload: []byte("invalid json"),
			},
			mockFn:  func(m *mockCommentsService) {},
			wantErr: true,
		},
		{
			name: "service error returns error",
			event: module.BaseEvent{
				Type: module.CommentEvent,
				Payload: func() json.RawMessage {
					b, _ := json.Marshal(domain.Comments{})
					return b
				}(),
			},
			mockFn: func(m *mockCommentsService) {
				m.createCommentFunc = func(ctx context.Context, comment domain.Comments) error {
					return errors.New("service error")
				}
			},
			wantErr: true,
		},
		{
			name: "successful handling returns nil",
			event: module.BaseEvent{
				Type: module.CommentEvent,
				Payload: func() json.RawMessage {
					b, _ := json.Marshal(domain.Comments{})
					return b
				}(),
			},
			mockFn: func(m *mockCommentsService) {
				m.createCommentFunc = func(ctx context.Context, comment domain.Comments) error {
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCommentsService{}
			tt.mockFn(mockService)
			handler := NewCommentsEventHandler(mockService)

			err := handler.Handle(context.Background(), tt.event)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
