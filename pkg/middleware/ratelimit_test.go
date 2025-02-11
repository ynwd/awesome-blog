package middleware

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_AllowRequest(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		setup    func(*RateLimiter)
		ip       string
		wait     time.Duration
		expected bool
	}{
		{
			name:     "should allow first request",
			config:   Config{Window: time.Minute, MaxAttempts: 5},
			setup:    func(rl *RateLimiter) {},
			ip:       "192.168.1.1",
			wait:     0,
			expected: true,
		},
		{
			name:   "should allow requests within limit",
			config: Config{Window: time.Minute, MaxAttempts: 5},
			setup: func(rl *RateLimiter) {
				for i := 0; i < 4; i++ {
					rl.AllowRequest("192.168.1.2")
				}
			},
			ip:       "192.168.1.2",
			wait:     0,
			expected: true,
		},
		{
			name:   "should block requests over limit",
			config: Config{Window: time.Minute, MaxAttempts: 5},
			setup: func(rl *RateLimiter) {
				for i := 0; i < 5; i++ {
					rl.AllowRequest("192.168.1.3")
				}
			},
			ip:       "192.168.1.3",
			wait:     0,
			expected: false,
		},
		{
			name:   "should allow requests after window expiration",
			config: Config{Window: time.Minute, MaxAttempts: 5},
			setup: func(rl *RateLimiter) {
				for i := 0; i < 5; i++ {
					rl.AllowRequest("192.168.1.4")
				}
			},
			ip:       "192.168.1.4",
			wait:     61 * time.Second, // Just over 1 minute
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.config)
			defer rl.Stop()
			tt.setup(rl)

			if tt.wait > 0 {
				time.Sleep(tt.wait)
			}

			if got := rl.AllowRequest(tt.ip); got != tt.expected {
				t.Errorf("AllowRequest() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(Config{Window: time.Minute, MaxAttempts: 5})
	ip := "192.168.1.5"

	var wg sync.WaitGroup
	// Launch 10 concurrent requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl.AllowRequest(ip)
		}()
	}

	wg.Wait()

	rl.mu.RLock()
	attempts := len(rl.attempts[ip])
	rl.mu.RUnlock()

	if attempts > 5 {
		t.Errorf("Rate limiter allowed %d attempts, expected maximum 5", attempts)
	}
}

func TestRateLimiter_CleanupOldAttempts(t *testing.T) {
	rl := NewRateLimiter(Config{Window: time.Minute, MaxAttempts: 5})
	ip := "192.168.1.6"

	// Add some old attempts
	rl.mu.Lock()
	rl.attempts[ip] = []time.Time{
		time.Now().Add(-2 * time.Minute),
		time.Now().Add(-90 * time.Second),
		time.Now().Add(-30 * time.Second),
	}
	rl.mu.Unlock()

	rl.AllowRequest(ip)

	rl.mu.RLock()
	attempts := len(rl.attempts[ip])
	rl.mu.RUnlock()

	if attempts != 2 { // Only the recent attempt and the one from 30 seconds ago should remain
		t.Errorf("Cleanup failed: got %d attempts, want 2", attempts)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	cfg := Config{
		Window:          time.Second,
		MaxAttempts:     5,
		CleanupInterval: time.Second,
	}
	rl := NewRateLimiter(cfg)
	defer rl.Stop()

	// Add test data
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for _, ip := range ips {
		rl.AllowRequest(ip)
	}

	// Wait for cleanup
	time.Sleep(2 * time.Second)

	rl.mu.RLock()
	count := len(rl.attempts)
	rl.mu.RUnlock()

	if count != 0 {
		t.Errorf("Cleanup failed: got %d entries, want 0", count)
	}
}
