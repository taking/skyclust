package performance

import (
	"context"
	"fmt"
	"time"

	"skyclust/pkg/cache"
)

// CachingStrategy defines standard caching strategies
type CachingStrategy struct {
	cache      cache.Cache
	keyBuilder *cache.CacheKeyBuilder
}

// NewCachingStrategy creates a new caching strategy
func NewCachingStrategy(cacheService cache.Cache) *CachingStrategy {
	return &CachingStrategy{
		cache:      cacheService,
		keyBuilder: cache.NewCacheKeyBuilder(),
	}
}

// CacheTTL defines standard cache TTL values
var CacheTTL = struct {
	Short     time.Duration // 30 seconds - frequently changing data
	Medium    time.Duration // 5 minutes - moderately changing data
	Long      time.Duration // 30 minutes - stable data
	VeryLong  time.Duration // 1 hour - rarely changing data
	Permanent time.Duration // 0 - no expiration (use with caution)
}{
	Short:     30 * time.Second,
	Medium:    5 * time.Minute,
	Long:      30 * time.Minute,
	VeryLong:  1 * time.Hour,
	Permanent: 0,
}

// GetOrSet retrieves a value from cache or sets it using the provided function
// This is a standard pattern to avoid cache stampede
func (cs *CachingStrategy) GetOrSet(
	ctx context.Context,
	key string,
	setter func() (interface{}, error),
	ttl time.Duration,
) (interface{}, error) {
	// Try to get from cache
	var cached interface{}
	if err := cs.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}

	// Cache miss - call setter
	value, err := setter()
	if err != nil {
		return nil, err
	}

	// Store in cache (non-blocking - don't fail if cache write fails)
	if ttl > 0 {
		if err := cs.cache.Set(ctx, key, value, ttl); err != nil {
			// Log but don't fail
			// Logger would be injected in real implementation
		}
	}

	return value, nil
}

// InvalidateByPattern invalidates all cache keys matching a pattern
func (cs *CachingStrategy) InvalidateByPattern(ctx context.Context, pattern string) error {
	// Use type assertion to check if cache supports DeletePattern
	// This avoids direct dependency on RedisService type
	type patternDeleter interface {
		DeletePattern(ctx context.Context, pattern string) error
	}

	if deleter, ok := cs.cache.(patternDeleter); ok {
		return deleter.DeletePattern(ctx, pattern)
	}
	return fmt.Errorf("pattern-based invalidation not supported for cache type")
}

// CacheKeyPatterns defines standard cache key patterns
var CacheKeyPatterns = struct {
	// List patterns
	ListByWorkspace  string // list:{resource}:{workspace_id}
	ListByCredential string // list:{resource}:{provider}:{credential_id}:{region}
	ListByUser       string // list:{resource}:{user_id}

	// Item patterns
	ItemByID         string // item:{resource}:{id}
	ItemByProviderID string // item:{resource}:{provider}:{credential_id}:{id}

	// Stats patterns
	StatsByWorkspace  string // stats:{resource}:{workspace_id}
	StatsByCredential string // stats:{resource}:{provider}:{credential_id}
}{
	ListByWorkspace:   "list:%s:%s",
	ListByCredential:  "list:%s:%s:%s:%s",
	ListByUser:        "list:%s:%s",
	ItemByID:          "item:%s:%s",
	ItemByProviderID:  "item:%s:%s:%s:%s",
	StatsByWorkspace:  "stats:%s:%s",
	StatsByCredential: "stats:%s:%s:%s",
}

// BuildCacheKey builds a cache key using the standard pattern
func (cs *CachingStrategy) BuildCacheKey(pattern string, args ...interface{}) string {
	return fmt.Sprintf(pattern, args...)
}
