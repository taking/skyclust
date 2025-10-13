package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenBlacklist manages JWT token blacklist using Redis
type TokenBlacklist struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewTokenBlacklist creates a new token blacklist service
func NewTokenBlacklist(redis *redis.Client, logger *zap.Logger) *TokenBlacklist {
	return &TokenBlacklist{
		redis:  redis,
		logger: logger,
	}
}

// AddToBlacklist adds a token to the blacklist
func (tb *TokenBlacklist) AddToBlacklist(ctx context.Context, token string, expiry time.Duration) error {
	// Hash the token for security (don't store raw tokens)
	tokenHash := tb.hashToken(token)

	// Store in Redis with expiry
	err := tb.redis.Set(ctx, fmt.Sprintf("blacklist:%s", tokenHash), "1", expiry).Err()
	if err != nil {
		tb.logger.Error("Failed to add token to blacklist",
			zap.Error(err),
			zap.String("token_hash", tokenHash))
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	tb.logger.Info("Token added to blacklist",
		zap.String("token_hash", tokenHash),
		zap.Duration("expiry", expiry))

	return nil
}

// IsBlacklisted checks if a token is in the blacklist
func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) bool {
	tokenHash := tb.hashToken(token)
	key := fmt.Sprintf("blacklist:%s", tokenHash)

	// Use GET instead of EXISTS for better performance
	// GET returns the value if key exists, nil if not
	_, err := tb.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist, token is not blacklisted
		return false
	}
	if err != nil {
		tb.logger.Error("Failed to check token blacklist",
			zap.Error(err),
			zap.String("token_hash", tokenHash))
		// In case of error, assume token is not blacklisted to avoid blocking valid users
		return false
	}

	// Key exists, token is blacklisted
	return true
}

// RemoveFromBlacklist removes a token from the blacklist (for testing purposes)
func (tb *TokenBlacklist) RemoveFromBlacklist(ctx context.Context, token string) error {
	tokenHash := tb.hashToken(token)

	err := tb.redis.Del(ctx, fmt.Sprintf("blacklist:%s", tokenHash)).Err()
	if err != nil {
		tb.logger.Error("Failed to remove token from blacklist",
			zap.Error(err),
			zap.String("token_hash", tokenHash))
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}

	tb.logger.Info("Token removed from blacklist",
		zap.String("token_hash", tokenHash))

	return nil
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (tb *TokenBlacklist) CleanupExpiredTokens(ctx context.Context) error {
	// Redis automatically handles TTL, but we can add manual cleanup if needed
	pattern := "blacklist:*"

	keys, err := tb.redis.Keys(ctx, pattern).Result()
	if err != nil {
		tb.logger.Error("Failed to get blacklist keys", zap.Error(err))
		return fmt.Errorf("failed to get blacklist keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Check TTL for each key and remove expired ones
	var expiredKeys []string
	for _, key := range keys {
		ttl, err := tb.redis.TTL(ctx, key).Result()
		if err != nil {
			tb.logger.Warn("Failed to get TTL for key",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		if ttl == -2 { // Key doesn't exist (already expired)
			expiredKeys = append(expiredKeys, key)
		}
	}

	if len(expiredKeys) > 0 {
		err = tb.redis.Del(ctx, expiredKeys...).Err()
		if err != nil {
			tb.logger.Error("Failed to cleanup expired tokens", zap.Error(err))
			return fmt.Errorf("failed to cleanup expired tokens: %w", err)
		}

		tb.logger.Info("Cleaned up expired tokens",
			zap.Int("count", len(expiredKeys)))
	}

	return nil
}

// GetBlacklistStats returns statistics about the blacklist
func (tb *TokenBlacklist) GetBlacklistStats(ctx context.Context) (map[string]interface{}, error) {
	pattern := "blacklist:*"

	// Use SCAN instead of KEYS for better performance in production
	var cursor uint64
	var allKeys []string

	for {
		keys, nextCursor, err := tb.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan blacklist keys: %w", err)
		}

		allKeys = append(allKeys, keys...)
		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	stats := map[string]interface{}{
		"total_blacklisted_tokens": len(allKeys),
		"timestamp":                time.Now().Unix(),
		"scan_method":              "optimized", // Indicates we're using SCAN instead of KEYS
	}

	return stats, nil
}

// hashToken creates a secure hash of the token
func (tb *TokenBlacklist) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
