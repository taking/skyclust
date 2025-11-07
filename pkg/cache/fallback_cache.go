package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"skyclust/pkg/logger"
)

// FallbackCache wraps a primary cache with a fallback cache
// When the primary cache fails, it automatically falls back to the secondary cache
type FallbackCache struct {
	primary   Cache
	fallback  Cache
	logger    *zap.Logger
	mu        sync.RWMutex
	primaryOK bool
}

// NewFallbackCache creates a new fallback cache
func NewFallbackCache(primary, fallback Cache) *FallbackCache {
	return &FallbackCache{
		primary:   primary,
		fallback:  fallback,
		logger:    logger.DefaultLogger.GetLogger(),
		primaryOK: true,
	}
}

// checkPrimaryHealth checks if primary cache is healthy
func (f *FallbackCache) checkPrimaryHealth(ctx context.Context) bool {
	f.mu.RLock()
	wasOK := f.primaryOK
	f.mu.RUnlock()

	if !wasOK {
		// Try to reconnect periodically
		if _, err := f.primary.Exists(ctx, "health_check"); err == nil {
			f.mu.Lock()
			f.primaryOK = true
			f.mu.Unlock()
			f.logger.Info("Primary cache recovered")
			return true
		}
		return false
	}

	// Check health periodically
	if _, err := f.primary.Exists(ctx, "health_check"); err != nil {
		f.mu.Lock()
		f.primaryOK = false
		f.mu.Unlock()
		f.logger.Warn("Primary cache health check failed, falling back to secondary cache",
			zap.Error(err))
		return false
	}

	return true
}

// Set stores a value in cache
func (f *FallbackCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	// Try primary first
	if f.checkPrimaryHealth(ctx) {
		if err := f.primary.Set(ctx, key, value, expiration); err == nil {
			// Also set in fallback for redundancy
			_ = f.fallback.Set(ctx, key, value, expiration)
			return nil
		}
		f.logger.Warn("Primary cache Set failed, falling back",
			zap.String("key", key))
	}

	// Fallback to secondary
	return f.fallback.Set(ctx, key, value, expiration)
}

// Get retrieves a value from cache
func (f *FallbackCache) Get(ctx context.Context, key string, dest interface{}) error {
	// Try primary first
	if f.checkPrimaryHealth(ctx) {
		err := f.primary.Get(ctx, key, dest)
		if err == nil {
			return nil
		}
		// Check if it's a cache miss (expected) or actual error
		if err != ErrRedisCacheMiss && err != ErrCacheMiss {
			f.logger.Warn("Primary cache Get failed, trying fallback",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	// Fallback to secondary
	return f.fallback.Get(ctx, key, dest)
}

// Delete removes a value from cache
func (f *FallbackCache) Delete(ctx context.Context, key string) error {
	// Delete from both caches
	var primaryErr, fallbackErr error

	if f.checkPrimaryHealth(ctx) {
		primaryErr = f.primary.Delete(ctx, key)
	} else {
		primaryErr = fmt.Errorf("primary cache unavailable")
	}

	fallbackErr = f.fallback.Delete(ctx, key)

	// Return error only if both fail
	if primaryErr != nil && fallbackErr != nil {
		return fmt.Errorf("failed to delete from both caches: primary=%v, fallback=%v", primaryErr, fallbackErr)
	}

	return nil
}

// Exists checks if a key exists in cache
func (f *FallbackCache) Exists(ctx context.Context, key string) (bool, error) {
	// Try primary first
	if f.checkPrimaryHealth(ctx) {
		if exists, err := f.primary.Exists(ctx, key); err == nil {
			return exists, nil
		}
	}

	// Fallback to secondary
	return f.fallback.Exists(ctx, key)
}

// Close closes the cache connections
func (f *FallbackCache) Close() error {
	var errs []error

	if err := f.primary.Close(); err != nil {
		errs = append(errs, fmt.Errorf("primary: %w", err))
	}

	if err := f.fallback.Close(); err != nil {
		errs = append(errs, fmt.Errorf("fallback: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing caches: %v", errs)
	}

	return nil
}
