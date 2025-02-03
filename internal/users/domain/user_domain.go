package domain

import (
	"errors"
)

var (
	ErrUsernameRequired = errors.New("username is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrUsernameTaken    = errors.New("username is already taken")
	ErrInvalidPassword  = errors.New("password must be at least 6 characters")
)

type User struct {
	Id       string `firestore:"id,omitempty"`
	Username string `firestore:"username"`
	Password string `firestore:"password"`
}

func (u *User) Validate() error {
	if u.Username == "" {
		return ErrUsernameRequired
	}
	if u.Password == "" {
		return ErrPasswordRequired
	}
	if len(u.Password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}
