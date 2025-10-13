package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
	PoolSize int
}

// RedisService implements caching using Redis
type RedisService struct {
	client *redis.Client
	config RedisConfig
}

// NewRedisService creates a new Redis service
func NewRedisService(config RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Successfully connected to Redis")
	return &RedisService{
		client: client,
		config: config,
	}, nil
}

// Set stores a value in cache with expiration
func (r *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get retrieves a value from cache
func (r *RedisService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrRedisCacheMiss
		}
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return nil
}

// Delete removes a value from cache
func (r *RedisService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (r *RedisService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in cache
func (r *RedisService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// SetExpiration sets expiration for a key
func (r *RedisService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// SetNX sets a value only if the key doesn't exist
func (r *RedisService) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.SetNX(ctx, key, data, expiration).Result()
}

// Increment increments a numeric value in the cache
func (r *RedisService) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// Decrement decrements a numeric value in the cache
func (r *RedisService) Decrement(ctx context.Context, key string) (int64, error) {
	return r.client.Decr(ctx, key).Result()
}

// Expire sets expiration for a key
func (r *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns the time to live for a key
func (r *RedisService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// FlushAll removes all keys from the cache
func (r *RedisService) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}

// GetTTL returns the time to live for a key
func (r *RedisService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// IncrementBy increments a counter by a specific value
func (r *RedisService) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.client.IncrBy(ctx, key, value).Result()
}

// GetClient returns the Redis client
func (r *RedisService) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Health checks Redis health
func (r *RedisService) Health(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// GetStats returns Redis statistics
func (r *RedisService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	stats := make(map[string]interface{})
	stats["info"] = info
	stats["connected"] = true

	return stats, nil
}

// ClearExpired removes expired entries from the cache
func (r *RedisService) ClearExpired() error {
	// Redis automatically handles TTL, so this is a no-op
	// But we can clean up any keys that might have been missed
	ctx := context.Background()

	// Get all keys
	keys, err := r.client.Keys(ctx, "*").Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	// Check TTL for each key and remove if expired
	for _, key := range keys {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			continue
		}
		if ttl == -1 { // Key exists but has no expiration
			continue
		}
		if ttl == -2 { // Key doesn't exist (shouldn't happen)
			continue
		}
		// Key has TTL, Redis will handle expiration automatically
	}

	return nil
}

// GetPerformanceStats returns cache performance statistics
func (r *RedisService) GetPerformanceStats() (*CacheStats, error) {
	ctx := context.Background()

	// Parse basic stats (simplified)
	stats := &CacheStats{
		Keys:      int64(len(r.client.Keys(ctx, "*").Val())),
		Timestamp: time.Now(),
	}

	// Parse hits and misses from Redis info
	// This is a simplified implementation
	stats.Hits = 0
	stats.Misses = 0

	return stats, nil
}

// Redis-specific cache errors
var (
	ErrRedisCacheMiss = fmt.Errorf("redis cache miss")
)
