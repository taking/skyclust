package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockAcquisitionTimeout = errors.New("lock acquisition timeout")
	ErrLockNotOwned           = errors.New("lock not owned by this instance")
)

// DistributedLock represents a distributed lock interface
type DistributedLock interface {
	Lock(ctx context.Context, key string, ttl time.Duration) error
	Unlock(ctx context.Context, key string) error
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Extend(ctx context.Context, key string, ttl time.Duration) error
	IsLocked(ctx context.Context, key string) (bool, error)
}

// RedisDistributedLock implements distributed locking using Redis
type RedisDistributedLock struct {
	client *redis.Client
	id     string
}

// NewRedisDistributedLock creates a new Redis-based distributed lock
func NewRedisDistributedLock(client *redis.Client, instanceID string) *RedisDistributedLock {
	return &RedisDistributedLock{
		client: client,
		id:     instanceID,
	}
}

// Lock acquires a distributed lock with the given key and TTL
// Uses SET with NX (only if not exists) and EX (expiration) for atomic lock acquisition
func (r *RedisDistributedLock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	lockKey := r.getLockKey(key)
	success, err := r.client.SetNX(ctx, lockKey, r.id, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !success {
		return ErrLockAcquisitionTimeout
	}

	return nil
}

// TryLock attempts to acquire a lock without blocking
// Returns true if lock was acquired, false if already locked
func (r *RedisDistributedLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := r.getLockKey(key)
	success, err := r.client.SetNX(ctx, lockKey, r.id, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to try acquire lock: %w", err)
	}

	return success, nil
}

// Unlock releases a distributed lock
// Uses Lua script to ensure atomic unlock (only if lock is owned by this instance)
func (r *RedisDistributedLock) Unlock(ctx context.Context, key string) error {
	lockKey := r.getLockKey(key)

	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := r.client.Eval(ctx, script, []string{lockKey}, r.id).Result()
	if err != nil {
		return fmt.Errorf("failed to unlock: %w", err)
	}

	unlocked, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	if unlocked == 0 {
		return ErrLockNotOwned
	}

	return nil
}

// Extend extends the TTL of an existing lock (only if owned by this instance)
func (r *RedisDistributedLock) Extend(ctx context.Context, key string, ttl time.Duration) error {
	lockKey := r.getLockKey(key)

	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := r.client.Eval(ctx, script, []string{lockKey}, r.id, int(ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	extended, ok := result.(int64)
	if !ok {
		return fmt.Errorf("unexpected result type: %T", result)
	}

	if extended == 0 {
		return ErrLockNotOwned
	}

	return nil
}

// IsLocked checks if a lock exists for the given key
func (r *RedisDistributedLock) IsLocked(ctx context.Context, key string) (bool, error) {
	lockKey := r.getLockKey(key)
	exists, err := r.client.Exists(ctx, lockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	return exists > 0, nil
}

// getLockKey returns the formatted lock key
func (r *RedisDistributedLock) getLockKey(key string) string {
	return fmt.Sprintf("lock:%s", key)
}

// LockWithRetry acquires a lock with retry mechanism
func (r *RedisDistributedLock) LockWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryInterval):
			}
		}

		success, err := r.TryLock(ctx, key, ttl)
		if err != nil {
			return err
		}

		if success {
			return nil
		}
	}

	return ErrLockAcquisitionTimeout
}

// LockWithAutoRenew acquires a lock and automatically renews it until context is cancelled
func (r *RedisDistributedLock) LockWithAutoRenew(ctx context.Context, key string, ttl time.Duration, renewInterval time.Duration) error {
	if err := r.Lock(ctx, key, ttl); err != nil {
		return err
	}

	go r.autoRenew(ctx, key, ttl, renewInterval)

	return nil
}

// autoRenew automatically renews the lock at regular intervals
func (r *RedisDistributedLock) autoRenew(ctx context.Context, key string, ttl time.Duration, renewInterval time.Duration) {
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = r.Unlock(context.Background(), key)
			return
		case <-ticker.C:
			if err := r.Extend(ctx, key, ttl); err != nil {
				if err == ErrLockNotOwned {
					return
				}
			}
		}
	}
}
