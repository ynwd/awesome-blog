package service

import (
	"context"
	"errors"

	"github.com/ynwd/awesome-blog/internal/users/domain"
	"github.com/ynwd/awesome-blog/internal/users/repo"
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrNotFound       = errors.New("user not found")
	ErrUsernameExists = errors.New("username already exists")
)

type userService struct {
	repo repo.UserRepository
}

func NewUserService(repo repo.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) CreateUser(ctx context.Context, user domain.User) error {
	if err := validateUser(user); err != nil {
		return err
	}

	exists, err := s.repo.IsUsernameExists(ctx, user.Username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameExists
	}

	return s.repo.Create(ctx, user)
}

func (s *userService) AuthenticateUser(ctx context.Context, username, password string) (domain.User, error) {
	if username == "" || password == "" {
		return domain.User{}, ErrInvalidInput
	}
	return s.repo.GetByUsernameAndPassword(ctx, username, password)
}

func validateUser(user domain.User) error {
	if user.Username == "" || user.Password == "" {
		return ErrInvalidInput
	}
	return nil
}
