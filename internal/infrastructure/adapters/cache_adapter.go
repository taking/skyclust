package adapters

import (
	"context"
	"time"

	"skyclust/internal/domain"
	"skyclust/pkg/cache"
)

// CacheAdapter adapts pkg/cache.Cache to domain.CacheService
type CacheAdapter struct {
	cache cache.Cache
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(cacheService cache.Cache) domain.CacheService {
	return &CacheAdapter{cache: cacheService}
}

// Get retrieves a value from cache
func (a *CacheAdapter) Get(ctx context.Context, key string) (interface{}, error) {
	var value interface{}
	err := a.cache.Get(ctx, key, &value)
	if err == cache.ErrCacheMiss {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, "cache miss", 404)
	}
	return value, err
}

// Set stores a value in cache
func (a *CacheAdapter) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return a.cache.Set(ctx, key, value, ttl)
}

// Delete removes a value from cache
func (a *CacheAdapter) Delete(ctx context.Context, key string) error {
	return a.cache.Delete(ctx, key)
}

// Clear clears all cache entries
func (a *CacheAdapter) Clear(ctx context.Context) error {
	// Note: pkg/cache.Cache doesn't have Clear method
	// This is a limitation - we might need to extend the interface
	return nil
}

// GetOrSet retrieves a value from cache or sets it using the setter function
func (a *CacheAdapter) GetOrSet(ctx context.Context, key string, setter func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	value, err := a.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// If cache miss, call setter
	value, err = setter()
	if err != nil {
		return nil, err
	}

	// Set in cache
	if err := a.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail
		return value, nil
	}

	return value, nil
}

// InvalidatePattern invalidates all cache keys matching a pattern
func (a *CacheAdapter) InvalidatePattern(ctx context.Context, pattern string) error {
	// Note: pkg/cache.Cache doesn't have pattern-based invalidation
	// This would need to be implemented at the infrastructure level
	// For now, return nil (no-op)
	return nil
}

// GetStats returns cache statistics
func (a *CacheAdapter) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Note: pkg/cache.Cache doesn't expose stats
	// Return empty stats for now
	return map[string]interface{}{
		"hits":   0,
		"misses": 0,
		"keys":   0,
	}, nil
}

// Health checks cache health
func (a *CacheAdapter) Health(ctx context.Context) error {
	// Check if cache is accessible by trying to check existence of a test key
	_, err := a.cache.Exists(ctx, "health_check")
	return err
}
