package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/internal/likes/domain"
)

type mockLikesRepository struct {
	createFunc func(ctx context.Context, like domain.Likes) error
}

func (m *mockLikesRepository) Create(ctx context.Context, like domain.Likes) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, like)
	}
	return nil
}

func TestLikesService_CreateLike(t *testing.T) {
	tests := []struct {
		name      string
		like      domain.Likes
		mockFn    func(*mockLikesRepository)
		wantError error
	}{
		{
			name: "empty post id returns error",
			like: domain.Likes{
				UsernameFrom: "user1",
				PostID:       "",
			},
			mockFn:    func(m *mockLikesRepository) {},
			wantError: ErrInvalidLike,
		},
		{
			name: "empty username returns error",
			like: domain.Likes{
				UsernameFrom: "",
				PostID:       "post1",
			},
			mockFn:    func(m *mockLikesRepository) {},
			wantError: ErrInvalidLike,
		},
		{
			name: "repository error returns error",
			like: domain.Likes{
				UsernameFrom: "user1",
				PostID:       "post1",
			},
			mockFn: func(m *mockLikesRepository) {
				m.createFunc = func(ctx context.Context, like domain.Likes) error {
					return assert.AnError
				}
			},
			wantError: assert.AnError,
		},
		{
			name: "success create like",
			like: domain.Likes{
				UsernameFrom: "user1",
				PostID:       "post1",
			},
			mockFn: func(m *mockLikesRepository) {
				m.createFunc = func(ctx context.Context, like domain.Likes) error {
					return nil
				}
			},
			wantError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockLikesRepository{}
			tt.mockFn(mockRepo)

			service := NewLikesService(mockRepo)
			err := service.CreateLike(context.Background(), tt.like)

			assert.Equal(t, tt.wantError, err)
		})
	}
}
