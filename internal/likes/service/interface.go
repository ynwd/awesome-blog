package service

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/likes/domain"
)

type LikesService interface {
	CreateLike(ctx context.Context, like domain.Likes) error
}
