package middleware

import (
	"sync"
	"time"
)

type RateLimiter struct {
	attempts    map[string][]time.Time // Stores attempts per IP
	mu          sync.RWMutex           // Protects the map
	window      time.Duration          // Time window for rate limiting
	maxAttempts int                    // Maximum number of attempts allowed in the window
	cleanup     *time.Ticker           // Periodic cleanup of old entries
}

type Config struct {
	Window          time.Duration // Time window for rate limiting
	MaxAttempts     int           // Maximum number of attempts allowed in the window
	CleanupInterval time.Duration // How often to clean up old entries
}

func NewRateLimiter(cfg Config) *RateLimiter {
	if cfg.Window == 0 {
		cfg.Window = time.Minute
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = 5 * time.Minute
	}

	rl := &RateLimiter{
		attempts:    make(map[string][]time.Time),
		window:      cfg.Window,
		maxAttempts: cfg.MaxAttempts,
		cleanup:     time.NewTicker(cfg.CleanupInterval),
	}

	go rl.startCleanup()

	return rl
}

func (rl *RateLimiter) Stop() {
	rl.cleanup.Stop()
}

func (rl *RateLimiter) startCleanup() {
	for range rl.cleanup.C {
		rl.cleanupOldEntries()
	}
}

func (rl *RateLimiter) cleanupOldEntries() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	for ip, attempts := range rl.attempts {
		valid := make([]time.Time, 0, len(attempts))
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				valid = append(valid, attempt)
			}
		}
		if len(valid) == 0 {
			delete(rl.attempts, ip)
		} else {
			rl.attempts[ip] = valid
		}
	}
}

func (rl *RateLimiter) AllowRequest(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	if attempts, exists := rl.attempts[ip]; exists {
		valid := make([]time.Time, 0, len(attempts))
		for _, attempt := range attempts {
			if attempt.After(cutoff) {
				valid = append(valid, attempt)
			}
		}
		rl.attempts[ip] = valid
	}

	if len(rl.attempts[ip]) >= rl.maxAttempts {
		return false
	}

	rl.attempts[ip] = append(rl.attempts[ip], now)
	return true
}
