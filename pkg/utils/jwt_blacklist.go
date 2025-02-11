package utils

import (
	"sync"
	"time"
)

type MemoryBlacklist struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

func NewMemoryBlacklist() *MemoryBlacklist {
	return &MemoryBlacklist{
		tokens: make(map[string]time.Time),
		mu:     sync.RWMutex{},
	}
}

func (b *MemoryBlacklist) Add(tokenID string, expiresAt time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[tokenID] = expiresAt
	return nil
}

func (b *MemoryBlacklist) IsBlacklisted(tokenID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	expiry, exists := b.tokens[tokenID]
	return exists && time.Now().Before(expiry)
}

func (b *MemoryBlacklist) Cleanup() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	for id, expiry := range b.tokens {
		if now.After(expiry) {
			delete(b.tokens, id)
		}
	}
	return nil
}
