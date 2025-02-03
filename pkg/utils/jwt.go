package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWT interface {
	GenerateToken(sub string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetClaims(token *jwt.Token) (jwt.MapClaims, error)
}

type jwtToken struct {
	secret string
}

func NewJWT() JWT {
	return &jwtToken{
		secret: os.Getenv("JWT_SECRET"),
	}
}

func (t *jwtToken) GenerateToken(sub string) (string, error) {
	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp": time.Now().Add(time.Hour * 12).Unix(),
			"sub": sub,
		},
	)

	tokenString, err := jwtToken.SignedString(
		[]byte(t.secret),
	)

	return tokenString, err
}

func (t *jwtToken) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(t.secret), nil
	})
}

func (t *jwtToken) GetClaims(token *jwt.Token) (jwt.MapClaims, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
