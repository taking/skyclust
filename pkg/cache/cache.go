package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Cache interface defines the cache operations
type Cache interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	FlushAll(ctx context.Context) error
	Close() error
	Health(ctx context.Context) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
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

// MemoryCache implements an in-memory cache
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*CacheEntry
	stats CacheStats
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits   int64
	Misses int64
	Sets   int64
	Dels   int64
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

// Set stores a value in the cache
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	m.items[key] = entry
	m.stats.Sets++

	return nil
}

// Get retrieves a value from the cache
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

// Delete removes a value from the cache
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[key]; exists {
		delete(m.items, key)
		m.stats.Dels++
	}

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

// SetNX sets a value only if the key doesn't exist
func (m *MemoryCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.items[key]; exists {
		return false, nil
	}

	entry := &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(expiration),
	}

	m.items[key] = entry
	m.stats.Sets++

	return true, nil
}

// Increment increments a numeric value in the cache
func (m *MemoryCache) Increment(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.items[key]
	if !exists {
		entry = &CacheEntry{
			Value:     int64(0),
			ExpiresAt: time.Now().Add(time.Hour), // Default expiration
		}
		m.items[key] = entry
	}

	if entry.IsExpired() {
		entry.Value = int64(1)
		return 1, nil
	}

	if val, ok := entry.Value.(int64); ok {
		entry.Value = val + 1
		return entry.Value.(int64), nil
	}

	// If not a number, start from 1
	entry.Value = int64(1)
	return 1, nil
}

// Decrement decrements a numeric value in the cache
func (m *MemoryCache) Decrement(ctx context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.items[key]
	if !exists {
		entry = &CacheEntry{
			Value:     int64(0),
			ExpiresAt: time.Now().Add(time.Hour), // Default expiration
		}
		m.items[key] = entry
	}

	if entry.IsExpired() {
		entry.Value = int64(-1)
		return -1, nil
	}

	if val, ok := entry.Value.(int64); ok {
		entry.Value = val - 1
		return entry.Value.(int64), nil
	}

	// If not a number, start from -1
	entry.Value = int64(-1)
	return -1, nil
}

// Expire sets expiration for a key
func (m *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.items[key]
	if !exists {
		return ErrCacheMiss
	}

	entry.ExpiresAt = time.Now().Add(expiration)
	return nil
}

// TTL returns the time to live for a key
func (m *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	entry, exists := m.items[key]
	m.mu.RUnlock()

	if !exists {
		return 0, ErrCacheMiss
	}

	if entry.IsExpired() {
		return 0, ErrCacheMiss
	}

	return time.Until(entry.ExpiresAt), nil
}

// FlushAll removes all keys from the cache
func (m *MemoryCache) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*CacheEntry)
	return nil
}

// Close closes the cache (no-op for memory cache)
func (m *MemoryCache) Close() error {
	return nil
}

// Health checks the health of the cache
func (m *MemoryCache) Health(ctx context.Context) error {
	return nil
}

// GetStats returns cache statistics
func (m *MemoryCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["hits"] = m.stats.Hits
	stats["misses"] = m.stats.Misses
	stats["sets"] = m.stats.Sets
	stats["dels"] = m.stats.Dels
	stats["items"] = len(m.items)

	return stats, nil
}

// cleanup removes expired entries periodically
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for key, entry := range m.items {
			if entry.IsExpired() {
				delete(m.items, key)
			}
		}
		m.mu.Unlock()
	}
}
