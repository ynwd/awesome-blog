package repo

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/comments/domain"
)

type CommentsRepository interface {
	Create(ctx context.Context, comment domain.Comments) error
}
