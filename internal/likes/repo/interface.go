package repo

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/likes/domain"
)

type LikesRepository interface {
	Create(ctx context.Context, like domain.Likes) error
}
