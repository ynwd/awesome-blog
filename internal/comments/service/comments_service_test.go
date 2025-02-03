package service

import (
	"context"
	"testing"

	"github.com/ynwd/awesome-blog/internal/comments/domain"
)

type mockCommentsRepo struct {
	createFunc func(ctx context.Context, comment domain.Comments) error
}

func (m *mockCommentsRepo) Create(ctx context.Context, comment domain.Comments) error {
	if m.createFunc == nil {
		return nil
	}
	return m.createFunc(ctx, comment)
}

func TestCreateComment(t *testing.T) {
	tests := []struct {
		name    string
		comment *domain.Comments
		mockFn  func(ctx context.Context, comment domain.Comments) error
		wantErr error
	}{
		{
			name: "empty username should return error",
			comment: &domain.Comments{
				PostID:  "post-1",
				Comment: "test comment",
			},
			wantErr: ErrInvalidComment,
		},
		{
			name: "empty postID should return error",
			comment: &domain.Comments{
				Username: "user1",
				Comment:  "test comment",
			},
			wantErr: ErrInvalidComment,
		},
		{
			name: "empty comment content should return error",
			comment: &domain.Comments{
				Username: "user1",
				PostID:   "post-1",
			},
			wantErr: ErrInvalidComment,
		},
		{
			name: "successful comment creation",
			comment: &domain.Comments{
				Username: "user1",
				PostID:   "post-1",
				Comment:  "test comment",
			},
			mockFn: func(ctx context.Context, comment domain.Comments) error {
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCommentsRepo{
				createFunc: tt.mockFn,
			}
			service := NewCommentsService(mockRepo)

			err := service.CreateComment(context.Background(), *tt.comment)

			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
		})
	}
}
