package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrExpiredToken          = errors.New("token has expired")
	ErrInvalidIP             = errors.New("invalid IP address")
	ErrTokenUsedBeforeIssued = errors.New("token used before issued")
)

type Claims struct {
	jwt.RegisteredClaims
	IP       string `json:"ip"`
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
}

type JWT interface {
	GenerateToken(username string, ip string) (string, error)
	ValidateToken(tokenString string, ip string) (*jwt.Token, error)
	GetClaims(token *jwt.Token) (*Claims, error)
}

type jwtToken struct {
	secret string
}

func NewJWT() JWT {
	return &jwtToken{
		secret: os.Getenv("JWT_SECRET"),
	}
}

func (t *jwtToken) GenerateToken(username string, ip string) (string, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Subject:   username,
			Issuer:    "awesome-blog",
		},
		IP:       ip,
		Username: username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secret))
}

func (t *jwtToken) ValidateToken(tokenString string, currentIP string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate IP
	if claims.IP != currentIP {
		return nil, ErrInvalidIP
	}

	// Check if token is used before it was issued
	now := time.Now()
	if claims.IssuedAt != nil && claims.IssuedAt.After(now) {
		return nil, ErrTokenUsedBeforeIssued
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(now) {
		return nil, ErrExpiredToken
	}

	return token, nil
}

func (t *jwtToken) GetClaims(token *jwt.Token) (*Claims, error) {
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
