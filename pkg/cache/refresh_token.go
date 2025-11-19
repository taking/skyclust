package cache

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"skyclust/pkg/logger"
)

// RefreshTokenData represents the data stored with a refresh token
type RefreshTokenData struct {
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	ClientIP  string    `json:"client_ip,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// RefreshTokenStore manages refresh tokens using Redis
type RefreshTokenStore struct {
	redis *redis.Client
}

// NewRefreshTokenStore creates a new refresh token store service
func NewRefreshTokenStore(redis *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{
		redis: redis,
	}
}

// GenerateRefreshToken generates a secure random refresh token
func (rts *RefreshTokenStore) GenerateRefreshToken() (string, error) {
	// Generate 64 random bytes (128 hex characters)
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(tokenBytes), nil
}

// hashToken creates a secure hash of the token for storage
func (rts *RefreshTokenStore) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// StoreRefreshToken stores a refresh token in Redis with associated data
func (rts *RefreshTokenStore) StoreRefreshToken(ctx context.Context, token string, data *RefreshTokenData, expiry time.Duration) error {
	// If Redis is not initialized, return error (refresh tokens require Redis)
	if rts.redis == nil {
		return fmt.Errorf("Redis not initialized, refresh token storage requires Redis")
	}

	// Hash the token for security (don't store raw tokens)
	tokenHash := rts.hashToken(token)
	key := fmt.Sprintf("refresh_token:%s", tokenHash)

	// Serialize the data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token data: %w", err)
	}

	// Store in Redis with expiry
	err = rts.redis.Set(ctx, key, dataJSON, expiry).Err()
	if err != nil {
		logger.Errorf("Failed to store refresh token: %v", err)
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	logger.Debugf("Refresh token stored: %s (expiry: %v)", tokenHash, expiry)
	return nil
}

// GetRefreshToken retrieves refresh token data from Redis
func (rts *RefreshTokenStore) GetRefreshToken(ctx context.Context, token string) (*RefreshTokenData, error) {
	// If Redis is not initialized, return error
	if rts.redis == nil {
		return nil, fmt.Errorf("Redis not initialized")
	}

	tokenHash := rts.hashToken(token)
	key := fmt.Sprintf("refresh_token:%s", tokenHash)

	// Get the data from Redis
	dataJSON, err := rts.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		logger.Errorf("Failed to get refresh token: %v", err)
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Deserialize the data
	var data RefreshTokenData
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal refresh token data: %w", err)
	}

	// Check if token is expired (additional check)
	if time.Now().After(data.ExpiresAt) {
		// Token expired, remove it
		rts.redis.Del(ctx, key)
		return nil, fmt.Errorf("refresh token expired")
	}

	return &data, nil
}

// DeleteRefreshToken removes a refresh token from Redis
func (rts *RefreshTokenStore) DeleteRefreshToken(ctx context.Context, token string) error {
	if rts.redis == nil {
		return nil
	}

	tokenHash := rts.hashToken(token)
	key := fmt.Sprintf("refresh_token:%s", tokenHash)

	err := rts.redis.Del(ctx, key).Err()
	if err != nil {
		logger.Errorf("Failed to delete refresh token: %v", err)
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	logger.Debugf("Refresh token deleted: %s", tokenHash)
	return nil
}

// DeleteUserRefreshTokens removes all refresh tokens for a specific user
func (rts *RefreshTokenStore) DeleteUserRefreshTokens(ctx context.Context, userID string) error {
	if rts.redis == nil {
		return nil
	}

	// Scan for all refresh tokens
	pattern := "refresh_token:*"
	var cursor uint64
	var deletedCount int

	for {
		keys, nextCursor, err := rts.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			logger.Errorf("Failed to scan refresh token keys: %v", err)
			return fmt.Errorf("failed to scan refresh token keys: %w", err)
		}

		// Check each key to see if it belongs to the user
		for _, key := range keys {
			dataJSON, err := rts.redis.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var data RefreshTokenData
			if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
				continue
			}

			// If this token belongs to the user, delete it
			if data.UserID == userID {
				if err := rts.redis.Del(ctx, key).Err(); err == nil {
					deletedCount++
				}
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	logger.Infof("Deleted %d refresh tokens for user %s", deletedCount, userID)
	return nil
}

// RotateRefreshToken replaces an old refresh token with a new one
func (rts *RefreshTokenStore) RotateRefreshToken(ctx context.Context, oldToken, newToken string, data *RefreshTokenData, expiry time.Duration) error {
	// Delete old token
	if err := rts.DeleteRefreshToken(ctx, oldToken); err != nil {
		logger.Warnf("Failed to delete old refresh token during rotation: %v", err)
		// Continue anyway, as the old token might have already been deleted
	}

	// Store new token
	return rts.StoreRefreshToken(ctx, newToken, data, expiry)
}

// GetUserRefreshTokenCount returns the number of active refresh tokens for a user
func (rts *RefreshTokenStore) GetUserRefreshTokenCount(ctx context.Context, userID string) (int, error) {
	if rts.redis == nil {
		return 0, nil
	}

	pattern := "refresh_token:*"
	var cursor uint64
	var count int

	for {
		keys, nextCursor, err := rts.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan refresh token keys: %w", err)
		}

		// Check each key to see if it belongs to the user
		for _, key := range keys {
			dataJSON, err := rts.redis.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var data RefreshTokenData
			if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
				continue
			}

			if data.UserID == userID {
				count++
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

