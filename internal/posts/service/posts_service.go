package service

import (
	"context"
	"errors"
	"time"

	"github.com/ynwd/awesome-blog/internal/posts/domain"
	"github.com/ynwd/awesome-blog/internal/posts/repo"
)

var (
	ErrInvalidPost     = errors.New("invalid post: title, description and username are required")
	ErrInvalidUsername = errors.New("invalid username: username cannot be empty")
)

type postsService struct {
	postsRepo repo.PostsRepository
}

func NewPostsService(postsRepo repo.PostsRepository) PostsService {
	return &postsService{
		postsRepo: postsRepo,
	}
}

func (s *postsService) CreatePost(ctx context.Context, post domain.Posts) (string, error) {
	if post.Title == "" || post.Description == "" || post.Username == "" {
		return "", ErrInvalidPost
	}

	post.CreatedAt = time.Now()
	return s.postsRepo.Create(ctx, post)
}
