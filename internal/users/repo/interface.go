package repo

import (
	"context"

	"github.com/ynwd/awesome-blog/internal/users/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByUsernameAndPassword(ctx context.Context, username string, password string) (domain.User, error)
	IsUsernameExists(ctx context.Context, username string) (bool, error)
}
