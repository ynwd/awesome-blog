package utils

import (
	"os"
	"testing"
	"time"

	jwt5 "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestNewJWT(t *testing.T) {
	tests := []struct {
		name        string
		secret      string
		expectError bool
	}{
		{
			name:        "Valid secret",
			secret:      "this-is-a-valid-secret-key-with-32-chars",
			expectError: false,
		},
		{
			name:        "Invalid short secret",
			secret:      "short",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("JWT_SECRET", tt.secret)
			defer os.Unsetenv("JWT_SECRET")

			jwt, err := NewJWT()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, jwt)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, jwt)
			}
		})
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt, err := NewJWT()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		username    string
		ip          string
		validateIP  string
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid token",
			username:    "testuser",
			ip:          "192.168.1.1",
			validateIP:  "192.168.1.1",
			expectError: false,
		},
		{
			name:        "Invalid IP",
			username:    "testuser",
			ip:          "192.168.1.1",
			validateIP:  "192.168.1.2",
			expectError: true,
			errorType:   ErrInvalidIP,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate token
			token, err := jwt.GenerateToken(tt.username, tt.ip)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Validate token
			validatedToken, err := jwt.ValidateToken(token, tt.validateIP)
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorType, err)
				assert.Nil(t, validatedToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validatedToken)

				// Verify claims
				claims, err := jwt.GetClaims(validatedToken)
				assert.NoError(t, err)
				assert.Equal(t, tt.username, claims.Username)
				assert.Equal(t, tt.ip, claims.IP)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &jwtToken{
		secret: "this-is-a-valid-secret-key-with-32-chars",
	}

	// Create claims with very short expiration
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(now.Add(1 * time.Millisecond)),
			IssuedAt:  jwt5.NewNumericDate(now),
			NotBefore: jwt5.NewNumericDate(now),
			Subject:   "testuser",
			Issuer:    "awesome-blog",
		},
		IP:       "192.168.1.1",
		Username: "testuser",
	}

	// Generate token with custom claims
	token := jwt5.NewWithClaims(jwt5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwt.secret))
	assert.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	_, err = jwt.ValidateToken(tokenString, "192.168.1.1")
	assert.Error(t, err)
	assert.Equal(t, ErrExpiredToken, err)
}

func TestInvalidTokenFormat(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt, err := NewJWT()
	assert.NoError(t, err)

	// Test invalid token format
	_, err = jwt.ValidateToken("invalid-token-format", "192.168.1.1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestInvalidSigningMethod(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	// Create token with wrong signing method
	claims := &Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   "testuser",
		},
		IP:       "192.168.1.1",
		Username: "testuser",
	}

	token := jwt5.NewWithClaims(jwt5.SigningMethodRS256, claims)
	tokenString, _ := token.SignedString([]byte("invalid-key"))

	jwt, _ := NewJWT()
	_, err := jwt.ValidateToken(tokenString, "192.168.1.1")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestNotBeforeToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &jwtToken{
		secret: "this-is-a-valid-secret-key-with-32-chars",
	}

	// Create future token
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt5.NewNumericDate(now),
			NotBefore: jwt5.NewNumericDate(now.Add(time.Hour)), // Future time
			Subject:   "testuser",
			Issuer:    "awesome-blog",
		},
		IP:       "192.168.1.1",
		Username: "testuser",
	}

	token := jwt5.NewWithClaims(jwt5.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwt.secret))

	_, err := jwt.ValidateToken(tokenString, "192.168.1.1")
	assert.Error(t, err)
	assert.Equal(t, ErrTokenUsedBeforeIssued, err)
}

func TestGetClaimsWithInvalidToken(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &jwtToken{
		secret: "this-is-a-valid-secret-key-with-32-chars",
	}

	// Create token with wrong type of claims
	wrongClaims := jwt5.MapClaims{
		"sub": "testuser",
		"ip":  "192.168.1.1",
	}
	token := jwt5.NewWithClaims(jwt5.SigningMethodHS256, wrongClaims)

	// Sign the token
	tokenString, err := token.SignedString([]byte(jwt.secret))
	assert.NoError(t, err)

	// Parse the token to get a valid token object
	parsedToken, err := jwt5.Parse(tokenString, func(token *jwt5.Token) (interface{}, error) {
		return []byte(jwt.secret), nil
	})
	assert.NoError(t, err)

	// Test GetClaims with wrong claims type
	_, err = jwt.GetClaims(parsedToken)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidToken, err)
}

func TestTokenWithRole(t *testing.T) {
	// Setup
	os.Setenv("JWT_SECRET", "this-is-a-valid-secret-key-with-32-chars")
	defer os.Unsetenv("JWT_SECRET")

	jwt := &jwtToken{
		secret: "this-is-a-valid-secret-key-with-32-chars",
	}

	// Create claims with role
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt5.RegisteredClaims{
			ExpiresAt: jwt5.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt5.NewNumericDate(now),
			NotBefore: jwt5.NewNumericDate(now),
			Subject:   "testuser",
			Issuer:    "awesome-blog",
		},
		IP:       "192.168.1.1",
		Username: "testuser",
		Role:     "admin",
	}

	token := jwt5.NewWithClaims(jwt5.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwt.secret))
	assert.NoError(t, err)

	// Validate token and check role
	validatedToken, err := jwt.ValidateToken(tokenString, "192.168.1.1")
	assert.NoError(t, err)

	validatedClaims, err := jwt.GetClaims(validatedToken)
	assert.NoError(t, err)
	assert.Equal(t, "admin", validatedClaims.Role)
}
