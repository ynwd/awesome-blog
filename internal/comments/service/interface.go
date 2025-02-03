package service

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/comments/domain"
)

type CommentsService interface {
	CreateComment(ctx context.Context, comment domain.Comments) error
}
