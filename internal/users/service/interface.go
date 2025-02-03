package service

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/users/domain"
)

type UserService interface {
	CreateUser(ctx context.Context, user domain.User) error
	AuthenticateUser(ctx context.Context, username, password string) (domain.User, error)
}
