package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentBlacklistAccess(t *testing.T) {
	bl := NewMemoryBlacklist()
	var wg sync.WaitGroup

	// Add tokens concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tokenID := fmt.Sprintf("token-%d", id)
			err := bl.Add(tokenID, time.Now().Add(time.Hour))
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all tokens were added
	assert.Equal(t, 100, len(bl.tokens))
}

func TestIsBlacklisted(t *testing.T) {
	bl := NewMemoryBlacklist()

	tests := []struct {
		name      string
		setup     func()
		tokenID   string
		expected  bool
		sleepTime time.Duration
	}{
		{
			name: "Token is blacklisted and not expired",
			setup: func() {
				bl.Add("valid-token", time.Now().Add(1*time.Hour))
			},
			tokenID:  "valid-token",
			expected: true,
		},
		{
			name: "Token is not in blacklist",
			setup: func() {
				bl.Add("other-token", time.Now().Add(1*time.Hour))
			},
			tokenID:  "non-existent-token",
			expected: false,
		},
		{
			name: "Token is blacklisted but expired",
			setup: func() {
				bl.Add("expired-token", time.Now().Add(10*time.Millisecond))
			},
			tokenID:   "expired-token",
			expected:  false,
			sleepTime: 20 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous tokens
			bl.tokens = make(map[string]time.Time)

			// Setup test case
			tt.setup()

			// Wait if needed for expiration
			if tt.sleepTime > 0 {
				time.Sleep(tt.sleepTime)
			}

			// Check blacklist status
			result := bl.IsBlacklisted(tt.tokenID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanup(t *testing.T) {
	bl := NewMemoryBlacklist()

	// Add tokens with different expiry times
	tokens := map[string]time.Duration{
		"expired1":    -1 * time.Hour,
		"expired2":    -30 * time.Minute,
		"notExpired1": 1 * time.Hour,
		"notExpired2": 2 * time.Hour,
	}

	// Add tokens to blacklist
	for id, offset := range tokens {
		bl.Add(id, time.Now().Add(offset))
	}

	// Verify initial state
	assert.Equal(t, 4, len(bl.tokens))

	// Run cleanup
	err := bl.Cleanup()
	assert.NoError(t, err)

	// Verify expired tokens were removed
	assert.Equal(t, 2, len(bl.tokens))
	assert.False(t, bl.IsBlacklisted("expired1"))
	assert.False(t, bl.IsBlacklisted("expired2"))
	assert.True(t, bl.IsBlacklisted("notExpired1"))
	assert.True(t, bl.IsBlacklisted("notExpired2"))
}

func TestConcurrentAccess(t *testing.T) {
	bl := NewMemoryBlacklist()
	tokenID := "test-token"
	expiry := time.Now().Add(1 * time.Hour)

	// Add token
	err := bl.Add(tokenID, expiry)
	assert.NoError(t, err)

	// Test concurrent access
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			bl.IsBlacklisted(tokenID)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			bl.Cleanup()
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

func TestBlacklistEdgeCases(t *testing.T) {
	bl := NewMemoryBlacklist()

	t.Run("Empty token ID", func(t *testing.T) {
		err := bl.Add("", time.Now().Add(time.Hour))
		assert.NoError(t, err)
		assert.True(t, bl.IsBlacklisted(""))
	})

	t.Run("Past expiry time", func(t *testing.T) {
		err := bl.Add("token1", time.Now().Add(-1*time.Hour))
		assert.NoError(t, err)
		assert.False(t, bl.IsBlacklisted("token1"))
	})

	t.Run("Zero expiry time", func(t *testing.T) {
		err := bl.Add("token2", time.Time{})
		assert.NoError(t, err)
		assert.False(t, bl.IsBlacklisted("token2"))
	})
}

func TestCleanupPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	bl := NewMemoryBlacklist()

	// Add 10000 tokens
	for i := 0; i < 10000; i++ {
		bl.Add(fmt.Sprintf("token-%d", i), time.Now().Add(time.Duration(i)*time.Minute))
	}

	start := time.Now()
	err := bl.Cleanup()
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Cleanup took too long")
}
