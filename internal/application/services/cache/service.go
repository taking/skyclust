package cache

import (
	"context"
	"go.uber.org/zap"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/logger"
	"time"
)

// cacheService: 캐시 비즈니스 로직 구현체
type cacheService struct {
	cache cache.Cache
}

// NewService: 새로운 캐시 서비스를 생성합니다
func NewService(cache cache.Cache) domain.CacheService {
	return &cacheService{
		cache: cache,
	}
}

// Get: 캐시에서 값을 조회합니다
func (s *cacheService) Get(ctx context.Context, key string) (interface{}, error) {
	var result interface{}
	err := s.cache.Get(ctx, key, &result)
	return result, err
}

// Set: 캐시에 값을 저장합니다
func (s *cacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.cache.Set(ctx, key, value, ttl)
}

// Delete: 캐시에서 값을 제거합니다
func (s *cacheService) Delete(ctx context.Context, key string) error {
	return s.cache.Delete(ctx, key)
}

// Clear: 캐시의 모든 값을 제거합니다
func (s *cacheService) Clear(ctx context.Context) error {
	// Redis doesn't have a direct clear method, we'll return an error
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "clear operation not supported", 501)
}

// GetOrSet: 캐시에서 값을 조회하거나 없으면 설정합니다
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

// InvalidatePattern: 패턴과 일치하는 모든 키를 제거합니다
func (s *cacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	// This would need to be implemented in the cache interface
	// For now, we'll just return nil
	return nil
}

// GetStats: 캐시 통계를 반환합니다
func (s *cacheService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// Return basic stats since GetStats was removed from cache interface
	return map[string]interface{}{
		"status": "available",
		"type":   "memory",
	}, nil
}

// Health: 캐시 상태를 확인합니다
func (s *cacheService) Health(ctx context.Context) error {
	// Simple health check by trying to get a non-existent key
	err := s.cache.Get(ctx, "health_check", nil)
	if err != nil && err.Error() != "key not found" {
		return err
	}
	return nil
}
