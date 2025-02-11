package utils

import (
	"crypto/rand"
	"encoding/hex"
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
	ErrInvalidUserAgent      = errors.New("invalid user agent")
	ErrInvalidDevice         = errors.New("invalid device")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"userId"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	DeviceID  string `json:"deviceId"`
}

type TokenFingerprint struct {
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	DeviceID  string `json:"deviceId"`
}

type JWT interface {
	GenerateToken(userID string, fingerprint *TokenFingerprint) (string, error)
	ValidateToken(tokenString string, fingerprint *TokenFingerprint) (*jwt.Token, error)
	RevokeToken(tokenID string) error
	GetClaims(token *jwt.Token) (*Claims, error)
}

// Add TokenBlacklist interface
type TokenBlacklist interface {
	Add(tokenID string, expiresAt time.Time) error
	IsBlacklisted(tokenID string) bool
	Cleanup() error
}

// Update jwtToken struct to include blacklist
type jwtToken struct {
	secret    string
	blacklist TokenBlacklist
}

// Update NewJWT to accept blacklist
func NewJWT(blacklist TokenBlacklist) (JWT, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must be at least 32 characters")
	}
	return &jwtToken{
		secret:    secret,
		blacklist: blacklist,
	}, nil
}

func (t *jwtToken) GenerateToken(userID string, fingerprint *TokenFingerprint) (string, error) {
	now := time.Now()
	expirationTime := time.Now().Add(15 * time.Minute)

	tokenID, err := generateTokenID()
	if err != nil {
		return "", fmt.Errorf("failed to generate token ID: %w", err)
	}

	claims := &Claims{
		UserID:    userID,
		IP:        fingerprint.IP,
		UserAgent: fingerprint.UserAgent,
		DeviceID:  fingerprint.DeviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    os.Getenv("APPLICATION_NAME"),
			Subject:   userID,
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.secret))
}

// Add RevokeToken implementation
func (t *jwtToken) RevokeToken(tokenID string) error {
	// Get current time for expiry
	now := time.Now()

	// Add token to blacklist
	if err := t.blacklist.Add(tokenID, now.Add(24*time.Hour)); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// Update ValidateToken to check blacklist
func (t *jwtToken) ValidateToken(tokenString string, fingerprint *TokenFingerprint) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrExpiredToken
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, ErrTokenUsedBeforeIssued
		default:
			return nil, ErrInvalidToken
		}
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Check if token is blacklisted
	if t.blacklist.IsBlacklisted(claims.ID) {
		return nil, ErrInvalidToken
	}

	// Check if token is used before it was issued
	now := time.Now()
	if claims.IssuedAt != nil && claims.IssuedAt.After(now) {
		return nil, ErrTokenUsedBeforeIssued
	}

	// Validate fingerprint
	if claims.IP != fingerprint.IP {
		return nil, ErrInvalidIP
	}
	if claims.UserAgent != fingerprint.UserAgent {
		return nil, ErrInvalidUserAgent
	}
	if claims.DeviceID != fingerprint.DeviceID {
		return nil, ErrInvalidDevice
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

// generateTokenID creates a cryptographically secure random token ID
func generateTokenID() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
