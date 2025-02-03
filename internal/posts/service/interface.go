package service

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/posts/domain"
)

type PostsService interface {
	CreatePost(ctx context.Context, post domain.Posts) (string, error)
}
