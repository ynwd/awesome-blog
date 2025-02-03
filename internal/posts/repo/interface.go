package repo

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/posts/domain"
)

type PostsRepository interface {
	Create(ctx context.Context, post domain.Posts) (string, error)
}
