package cache

import (
	"context"
	"go.uber.org/zap"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"time"
)

// cacheService implements the cache business logic
type cacheService struct {
	cache cache.Cache
}

// NewService creates a new cache service
func NewService(cache cache.Cache) domain.CacheService {
	return &cacheService{
		cache: cache,
	}
}

// Get retrieves a value from cache
func (s *cacheService) Get(ctx context.Context, key string) (interface{}, error) {
	var result interface{}
	err := s.cache.Get(ctx, key, &result)
	return result, err
}

// Set stores a value in cache
func (s *cacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// Delete removes a value from cache
func (s *cacheService) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Clear removes all values from cache
func (s *cacheService) Clear(ctx context.Context) error {
	// Redis doesn't have a direct clear method, we'll return an error
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "clear operation not supported", 501)
}

// GetOrSet retrieves a value from cache or sets it if not found
func (s *cacheService) GetOrSet(ctx context.Context, key string, setter func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache first
	value, err := s.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// If not found, call setter function
	value, err = setter()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := s.Set(ctx, key, value, ttl); err != nil {
		// Log error but don't fail the operation
		logger.DefaultLogger.GetLogger().Warn("Failed to cache value",
			zap.String("key", key),
			zap.Error(err),
		)
	}

	return value, nil
}

// InvalidatePattern removes all keys matching a pattern
func (s *cacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// This would need to be implemented in the cache interface
	// For now, we'll just return nil
	return nil
}

// GetStats returns cache statistics
func (s *cacheService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Return basic stats since GetStats was removed from cache interface
	return map[string]interface{}{
		"status": "available",
		"type":   "memory",
	}, nil
}

// Health checks cache health
func (s *cacheService) Health(ctx context.Context) error {
	// Simple health check by trying to get a non-existent key
	err := s.cache.Get(ctx, "health_check", nil)
	if err != nil && err.Error() != "key not found" {
		return err
	}
	return nil
}
