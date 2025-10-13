package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Cache interface defines the essential cache operations (simplified)
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// Cache errors
var (
	ErrCacheMiss = errors.New("cache miss")
	ErrCacheFull = errors.New("cache full")
)

// CacheEntry represents a cache entry
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// IsExpired checks if the cache entry is expired
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// CacheStats holds basic cache statistics (simplified)
type CacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Keys   int64 `json:"keys"`
}

// HitRate calculates the cache hit rate
func (s *CacheStats) HitRate() float64 {
	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}

// MemoryCache implements an in-memory cache
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
	stats CacheStats
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheEntry),
		stats: CacheStats{},
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set stores a value in the cache (simplified)
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	m.items[key] = entry
	return nil
}

// Get retrieves a value from the cache (simplified)
func (m *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	entry, exists := m.items[key]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		m.stats.Misses++
		m.mu.Unlock()
		return ErrCacheMiss
	}

	if entry.IsExpired() {
		m.mu.Lock()
		delete(m.items, key)
		m.stats.Misses++
		m.mu.Unlock()
		return ErrCacheMiss
	}

	// Copy value to dest (simplified)
	if destPtr, ok := dest.(*interface{}); ok {
		*destPtr = entry.Value
	}

	m.mu.Lock()
	m.stats.Hits++
	m.mu.Unlock()

	return nil
}

// Delete removes a value from the cache (simplified)
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

// Exists checks if a key exists in the cache
func (m *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	entry, exists := m.items[key]
	m.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if entry.IsExpired() {
		m.mu.Lock()
		delete(m.items, key)
		m.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Close closes the cache (simplified)
func (m *MemoryCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear all items
	m.items = make(map[string]*CacheEntry)
	return nil
}

// cleanup removes expired entries periodically
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, entry := range m.items {
			if now.After(entry.ExpiresAt) {
				delete(m.items, key)
			}
		}
		m.mu.Unlock()
	}
}
