package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	jwt.RegisteredClaims
	IP string `json:"ip"`
}

type JWT interface {
	GenerateToken(sub string, ip string) (string, error)
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

func (t *jwtToken) GenerateToken(sub string, ip string) (string, error) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   sub,
		},
		IP: ip,
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

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.IP != currentIP {
			return nil, fmt.Errorf("token IP mismatch: expected %s, got %s", claims.IP, currentIP)
		}
		return token, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (t *jwtToken) GetClaims(token *jwt.Token) (*Claims, error) {
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims type")
}
