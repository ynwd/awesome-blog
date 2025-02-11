package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockBlacklist implements TokenBlacklist for testing
type MockBlacklist struct {
	blacklistedTokens map[string]time.Time
}

func NewMockBlacklist() *MockBlacklist {
	return &MockBlacklist{
		blacklistedTokens: make(map[string]time.Time),
	}
}

func (m *MockBlacklist) Add(tokenID string, expiresAt time.Time) error {
	m.blacklistedTokens[tokenID] = expiresAt
	return nil
}

func (m *MockBlacklist) IsBlacklisted(tokenID string) bool {
	_, exists := m.blacklistedTokens[tokenID]
	return exists
}

func (m *MockBlacklist) Cleanup() error {
	now := time.Now()
	for id, expiry := range m.blacklistedTokens {
		if expiry.Before(now) {
			delete(m.blacklistedTokens, id)
		}
	}
	return nil
}

func setupTestEnv(t *testing.T) (JWT, *MockBlacklist) {
	t.Setenv("JWT_SECRET", "your-32-character-test-secret-key!")
	t.Setenv("APPLICATION_NAME", "test-app")

	blacklist := NewMockBlacklist()
	jwt, err := NewJWT(blacklist)
	require.NoError(t, err)
	return jwt, blacklist
}

func TestCompleteJWTScenario(t *testing.T) {
	jwtInstance, blacklist := setupTestEnv(t)

	// Test case 1: Generate and validate a valid token
	t.Run("Valid token lifecycle", func(t *testing.T) {
		fingerprint := &TokenFingerprint{
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		// Generate token
		token, err := jwtInstance.GenerateToken("user123", fingerprint)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Validate token with same fingerprint
		validatedToken, err := jwtInstance.ValidateToken(token, fingerprint)
		require.NoError(t, err)
		require.NotNil(t, validatedToken)

		// Get and verify claims
		claims, err := jwtInstance.GetClaims(validatedToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, fingerprint.IP, claims.IP)
		assert.Equal(t, fingerprint.UserAgent, claims.UserAgent)
		assert.Equal(t, fingerprint.DeviceID, claims.DeviceID)
	})

	// Test case 2: Token with different fingerprint
	t.Run("Token with different fingerprint", func(t *testing.T) {
		origFingerprint := &TokenFingerprint{
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		token, err := jwtInstance.GenerateToken("user123", origFingerprint)
		require.NoError(t, err)

		newFingerprint := &TokenFingerprint{
			IP:        "192.168.1.2", // Different IP
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		_, err = jwtInstance.ValidateToken(token, newFingerprint)
		assert.ErrorIs(t, err, ErrInvalidIP)
	})

	// Test case 3: Token revocation
	t.Run("Token revocation", func(t *testing.T) {
		fingerprint := &TokenFingerprint{
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		token, err := jwtInstance.GenerateToken("user123", fingerprint)
		require.NoError(t, err)

		// First validation should succeed
		validatedToken, err := jwtInstance.ValidateToken(token, fingerprint)
		require.NoError(t, err)

		// Get claims to get token ID
		claims, err := jwtInstance.GetClaims(validatedToken)
		require.NoError(t, err)

		// Revoke token
		err = jwtInstance.RevokeToken(claims.ID)
		require.NoError(t, err)

		// Second validation should fail
		_, err = jwtInstance.ValidateToken(token, fingerprint)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	// Test case 4: Expired token
	t.Run("Expired token", func(t *testing.T) {
		fingerprint := &TokenFingerprint{
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		// Create expired claims
		now := time.Now()
		claims := &Claims{
			UserID:    "user123",
			IP:        fingerprint.IP,
			UserAgent: fingerprint.UserAgent,
			DeviceID:  fingerprint.DeviceID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("your-32-character-test-secret-key!"))
		require.NoError(t, err)

		_, err = jwtInstance.ValidateToken(tokenString, fingerprint)
		assert.ErrorIs(t, err, ErrExpiredToken)
	})

	// Test case 5: Token cleanup
	t.Run("Blacklist cleanup", func(t *testing.T) {
		// Add some expired tokens
		blacklist.Add("expired1", time.Now().Add(-1*time.Hour))
		blacklist.Add("expired2", time.Now().Add(-2*time.Hour))
		blacklist.Add("valid1", time.Now().Add(1*time.Hour))

		err := blacklist.Cleanup()
		require.NoError(t, err)

		assert.False(t, blacklist.IsBlacklisted("expired1"))
		assert.False(t, blacklist.IsBlacklisted("expired2"))
		assert.True(t, blacklist.IsBlacklisted("valid1"))
	})

	// Test case 6: Token with future IssuedAt
	t.Run("Token used before issued", func(t *testing.T) {
		fingerprint := &TokenFingerprint{
			IP:        "192.168.1.1",
			UserAgent: "Mozilla/5.0",
			DeviceID:  "device123",
		}

		claims := &Claims{
			UserID:    "user123",
			IP:        fingerprint.IP,
			UserAgent: fingerprint.UserAgent,
			DeviceID:  fingerprint.DeviceID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("your-32-character-test-secret-key!"))
		require.NoError(t, err)

		_, err = jwtInstance.ValidateToken(tokenString, fingerprint)
		assert.ErrorIs(t, err, ErrTokenUsedBeforeIssued)
	})
}
