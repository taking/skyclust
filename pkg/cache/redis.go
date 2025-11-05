package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"skyclust/pkg/logger"
	"strings"
	"sync"
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
	stats  *CacheStats
	mu     sync.RWMutex
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
		stats:  &CacheStats{},
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
			r.recordMiss()
			return ErrRedisCacheMiss
		}
		r.recordMiss()
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		r.recordMiss()
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	r.recordHit()
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

	// Use SCAN instead of KEYS to avoid blocking Redis
	// SCAN iterates through keys incrementally without blocking
	cursor := uint64(0)
	pattern := "*"
	scanCount := int64(100) // Process 100 keys per iteration

	for {
		var keys []string
		var err error
		keys, cursor, err = r.client.Scan(ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
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

		// If cursor is 0, we've iterated through all keys
		if cursor == 0 {
			break
		}
	}

	return nil
}

// GetPerformanceStats returns cache performance statistics
func (r *RedisService) GetPerformanceStats(ctx context.Context) (*CacheStats, error) {
	// Use DBSIZE instead of KEYS to get key count efficiently (O(1) operation)
	// This avoids blocking Redis by scanning the entire keyspace
	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	r.mu.RLock()
	stats := &CacheStats{
		Hits:   r.stats.Hits,
		Misses: r.stats.Misses,
		Keys:   dbSize,
	}
	r.mu.RUnlock()

	return stats, nil
}

// GetMemoryUsage returns Redis memory usage in bytes
func (r *RedisService) GetMemoryUsage(ctx context.Context) (int64, error) {
	info, err := r.client.Info(ctx, "memory").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get memory info: %w", err)
	}

	var usedMemory int64
	// Parse used_memory from INFO output
	// Simplified parsing - in production, use proper parsing
	if _, err := fmt.Sscanf(info, "used_memory:%d", &usedMemory); err != nil {
		// Try alternative parsing
		lines := strings.Split(info, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory:") {
				fmt.Sscanf(line, "used_memory:%d", &usedMemory)
				break
			}
		}
	}

	return usedMemory, nil
}

// ResetStats resets cache statistics
func (r *RedisService) ResetStats() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Hits = 0
	r.stats.Misses = 0
}

// recordHit records a cache hit
func (r *RedisService) recordHit() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Hits++
}

// recordMiss records a cache miss
func (r *RedisService) recordMiss() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Misses++
}

// Pipeline represents a Redis pipeline for batch operations
type Pipeline struct {
	pipe redis.Pipeliner
	ctx  context.Context
}

// Exec executes all commands in the pipeline
func (p *Pipeline) Exec() ([]redis.Cmder, error) {
	return p.pipe.Exec(p.ctx)
}

// Pipeline creates a new Redis pipeline for batch operations
func (r *RedisService) Pipeline(ctx context.Context) *Pipeline {
	return &Pipeline{
		pipe: r.client.Pipeline(),
		ctx:  ctx,
	}
}

// MGet retrieves multiple values from cache in a single operation
func (r *RedisService) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return []interface{}{}, nil
	}

	results, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to MGet values: %w", err)
	}

	values := make([]interface{}, len(results))
	for i, result := range results {
		if result == nil {
			values[i] = nil
			continue
		}

		var value interface{}
		if err := json.Unmarshal([]byte(result.(string)), &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal value at index %d: %w", i, err)
		}
		values[i] = value
	}

	return values, nil
}

// MSet stores multiple key-value pairs in cache in a single operation
func (r *RedisService) MSet(ctx context.Context, kvPairs map[string]interface{}, expiration time.Duration) error {
	if len(kvPairs) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()
	for key, value := range kvPairs {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		pipe.Set(ctx, key, data, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to MSet values: %w", err)
	}

	return nil
}

// Eval executes a Lua script with the given keys and arguments
func (r *RedisService) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.Eval(ctx, script, keys, args...).Result()
}

// EvalSha executes a Lua script by its SHA1 hash
func (r *RedisService) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return r.client.EvalSha(ctx, sha1, keys, args...).Result()
}

// ScriptLoad loads a Lua script into Redis and returns its SHA1 hash
func (r *RedisService) ScriptLoad(ctx context.Context, script string) (string, error) {
	return r.client.ScriptLoad(ctx, script).Result()
}

// Common Lua scripts for atomic operations
const (
	// ScriptCompareAndSet compares current value and sets if matches (CAS operation)
	ScriptCompareAndSet = `
		local current = redis.call('GET', KEYS[1])
		if current == ARGV[1] then
			redis.call('SET', KEYS[1], ARGV[2], 'EX', ARGV[3])
			return 1
		end
		return 0
	`

	// ScriptIncrementWithLimit increments a counter with a maximum limit
	ScriptIncrementWithLimit = `
		local current = redis.call('GET', KEYS[1])
		local limit = tonumber(ARGV[1])
		local increment = tonumber(ARGV[2])
		local ttl = tonumber(ARGV[3])
		
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end
		
		if current + increment <= limit then
			redis.call('SET', KEYS[1], current + increment, 'EX', ttl)
			return current + increment
		end
		return current
	`

	// ScriptDeleteIfMatches deletes a key only if its value matches
	ScriptDeleteIfMatches = `
		local current = redis.call('GET', KEYS[1])
		if current == ARGV[1] then
			return redis.call('DEL', KEYS[1])
		end
		return 0
	`
)

// CompareAndSet performs an atomic compare-and-set operation
func (r *RedisService) CompareAndSet(ctx context.Context, key string, expectedValue, newValue string, expiration time.Duration) (bool, error) {
	result, err := r.Eval(ctx, ScriptCompareAndSet, []string{key}, expectedValue, newValue, int(expiration.Seconds()))
	if err != nil {
		return false, fmt.Errorf("failed to execute compare-and-set: %w", err)
	}

	matched, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}

	return matched == 1, nil
}

// IncrementWithLimit increments a counter with a maximum limit atomically
func (r *RedisService) IncrementWithLimit(ctx context.Context, key string, limit int64, increment int64, expiration time.Duration) (int64, error) {
	result, err := r.Eval(ctx, ScriptIncrementWithLimit, []string{key}, limit, increment, int(expiration.Seconds()))
	if err != nil {
		return 0, fmt.Errorf("failed to execute increment-with-limit: %w", err)
	}

	value, ok := result.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected result type: %T", result)
	}

	return value, nil
}

// DeleteIfMatches deletes a key only if its value matches
func (r *RedisService) DeleteIfMatches(ctx context.Context, key string, expectedValue string) (bool, error) {
	result, err := r.Eval(ctx, ScriptDeleteIfMatches, []string{key}, expectedValue)
	if err != nil {
		return false, fmt.Errorf("failed to execute delete-if-matches: %w", err)
	}

	deleted, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result type: %T", result)
	}

	return deleted == 1, nil
}

// Redis-specific cache errors
var (
	ErrRedisCacheMiss = fmt.Errorf("redis cache miss")
)
