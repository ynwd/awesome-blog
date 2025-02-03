package service

import (
	"context"
	"errors"
	"time"

	"github.com/ynwd/awesome-blog/internal/comments/domain"
	"github.com/ynwd/awesome-blog/internal/comments/repo"
)

var (
	ErrInvalidComment  = errors.New("invalid comment: username, postID and comment are required")
	ErrInvalidPostID   = errors.New("invalid post ID: cannot be empty")
	ErrInvalidUsername = errors.New("invalid username: cannot be empty")
)

type commentsService struct {
	commentsRepo repo.CommentsRepository
}

func NewCommentsService(commentsRepo repo.CommentsRepository) CommentsService {
	return &commentsService{
		commentsRepo: commentsRepo,
	}
}

func (s *commentsService) CreateComment(ctx context.Context, comment domain.Comments) error {
	if comment.Username == "" || comment.PostID == "" || comment.Comment == "" {
		return ErrInvalidComment
	}

	comment.CreatedAt = time.Now()
	return s.commentsRepo.Create(ctx, comment)
}
