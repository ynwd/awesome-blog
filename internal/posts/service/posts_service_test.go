package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/ynwd/awesome-blog/internal/posts/domain"
)

type mockPostsRepository struct {
	mock.Mock
}

func (m *mockPostsRepository) Create(ctx context.Context, post domain.Posts) (string, error) {
	args := m.Called(ctx, post)
	return args.String(0), args.Error(1)
}

func TestPostsService_CreatePost(t *testing.T) {
	tests := []struct {
		name    string
		post    domain.Posts
		mockFn  func(*mockPostsRepository)
		wantID  string
		wantErr error
	}{
		{
			name: "successful post creation",
			post: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
				CreatedAt:   time.Now(),
			},
			mockFn: func(m *mockPostsRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(post domain.Posts) bool {
					return post.Username == "testuser" &&
						post.Title == "Test Post" &&
						post.Description == "Test Description" &&
						!post.CreatedAt.IsZero()
				})).Return("post-123", nil)
			},
			wantID:  "post-123",
			wantErr: nil,
		},
		{
			name: "empty title",
			post: domain.Posts{
				Username:    "testuser",
				Title:       "",
				Description: "Test Description",
			},
			mockFn:  func(m *mockPostsRepository) {},
			wantID:  "",
			wantErr: ErrInvalidPost,
		},
		{
			name: "empty description",
			post: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "",
			},
			mockFn:  func(m *mockPostsRepository) {},
			wantID:  "",
			wantErr: ErrInvalidPost,
		},
		{
			name: "empty username",
			post: domain.Posts{
				Username:    "",
				Title:       "Test Post",
				Description: "Test Description",
			},
			mockFn:  func(m *mockPostsRepository) {},
			wantID:  "",
			wantErr: ErrInvalidPost,
		},
		{
			name: "repository error",
			post: domain.Posts{
				Username:    "testuser",
				Title:       "Test Post",
				Description: "Test Description",
			},
			mockFn: func(m *mockPostsRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return("", errors.New("repository error"))
			},
			wantID:  "",
			wantErr: errors.New("repository error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mockPostsRepository)
			if tt.mockFn != nil {
				tt.mockFn(mockRepo)
			}

			service := NewPostsService(mockRepo)
			gotID, err := service.CreatePost(context.Background(), tt.post)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				assert.Empty(t, gotID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
