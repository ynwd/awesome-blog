package service

import (
	"context"
	"errors"
	"time"

	"github.com/ynwd/awesome-blog/internal/likes/domain"
	"github.com/ynwd/awesome-blog/internal/likes/repo"
)

var (
	ErrInvalidLike     = errors.New("invalid like: postID and username are required")
	ErrInvalidPostID   = errors.New("invalid post ID: cannot be empty")
	ErrInvalidUsername = errors.New("invalid username: cannot be empty")
)

type likesService struct {
	likesRepo repo.LikesRepository
}

func NewLikesService(likesRepo repo.LikesRepository) LikesService {
	return &likesService{
		likesRepo: likesRepo,
	}
}

func (s *likesService) CreateLike(ctx context.Context, like domain.Likes) error {
	if like.PostID == "" || like.UsernameFrom == "" {
		return ErrInvalidLike
	}

	like.CreatedAt = time.Now()
	return s.likesRepo.Create(ctx, like)
}
