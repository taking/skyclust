package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// SSEConnectionState represents the state of an SSE connection stored in Redis
type SSEConnectionState struct {
	ConnectionID  string                 `json:"connection_id"`
	UserID        string                 `json:"user_id"`
	WorkspaceID   string                 `json:"workspace_id"`
	LastEventID   string                 `json:"last_event_id"`
	Subscriptions []string               `json:"subscriptions"`
	Filters       map[string]interface{} `json:"filters"`
	LastSeen      time.Time              `json:"last_seen"`
	CreatedAt     time.Time              `json:"created_at"`
}

// Redis key patterns for connection state
const (
	// sse:connection:{connection_id} - Connection state
	connectionStateKeyPattern = "sse:connection:%s"
	// sse:user:{user_id}:connections - Set of connection IDs for a user
	userConnectionsKeyPattern = "sse:user:%s:connections"
	// TTL for connection state (1 hour)
	connectionStateTTL = 1 * time.Hour
)

// saveConnectionState saves SSE connection state to Redis
func (h *SSEHandler) saveConnectionState(ctx context.Context, client *SSEClient) {
	if h.redisClient == nil {
		return
	}

	state := SSEConnectionState{
		ConnectionID:  client.ID,
		UserID:        client.UserID,
		WorkspaceID:   "", // WorkspaceID는 SSEClient에 없으므로 빈 문자열
		LastEventID:   "", // 마지막 이벤트 ID는 별도로 추적
		Subscriptions: h.getSubscribedEventTypes(client),
		Filters:       h.getClientFilters(client),
		LastSeen:      client.LastSeen,
		CreatedAt:     time.Now(),
	}

	// Marshal state
	stateJSON, err := json.Marshal(state)
	if err != nil {
		h.logger.Warn("Failed to marshal connection state",
			zap.Error(err),
			zap.String("connection_id", client.ID))
		return
	}

	// Save connection state
	connectionKey := fmt.Sprintf(connectionStateKeyPattern, client.ID)
	if err := h.redisClient.Set(ctx, connectionKey, stateJSON, connectionStateTTL).Err(); err != nil {
		h.logger.Warn("Failed to save connection state",
			zap.Error(err),
			zap.String("connection_id", client.ID))
		return
	}

	// Add connection ID to user's connection set
	userConnectionsKey := fmt.Sprintf(userConnectionsKeyPattern, client.UserID)
	if err := h.redisClient.SAdd(ctx, userConnectionsKey, client.ID).Err(); err != nil {
		h.logger.Warn("Failed to add connection to user set",
			zap.Error(err),
			zap.String("connection_id", client.ID),
			zap.String("user_id", client.UserID))
		return
	}

	// Set TTL on user connections set
	h.redisClient.Expire(ctx, userConnectionsKey, connectionStateTTL)

	h.logger.Debug("Connection state saved",
		zap.String("connection_id", client.ID),
		zap.String("user_id", client.UserID))
}

// loadConnectionState loads SSE connection state from Redis
func (h *SSEHandler) loadConnectionState(ctx context.Context, connectionID string) (*SSEConnectionState, error) {
	if h.redisClient == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	connectionKey := fmt.Sprintf(connectionStateKeyPattern, connectionID)
	stateJSON, err := h.redisClient.Get(ctx, connectionKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("connection state not found: %s", connectionID)
		}
		return nil, fmt.Errorf("failed to get connection state: %w", err)
	}

	var state SSEConnectionState
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal connection state: %w", err)
	}

	return &state, nil
}

// deleteConnectionState deletes SSE connection state from Redis
func (h *SSEHandler) deleteConnectionState(ctx context.Context, client *SSEClient) {
	if h.redisClient == nil {
		return
	}

	// Delete connection state
	connectionKey := fmt.Sprintf(connectionStateKeyPattern, client.ID)
	if err := h.redisClient.Del(ctx, connectionKey).Err(); err != nil {
		h.logger.Warn("Failed to delete connection state",
			zap.Error(err),
			zap.String("connection_id", client.ID))
	}

	// Remove connection ID from user's connection set
	userConnectionsKey := fmt.Sprintf(userConnectionsKeyPattern, client.UserID)
	if err := h.redisClient.SRem(ctx, userConnectionsKey, client.ID).Err(); err != nil {
		h.logger.Warn("Failed to remove connection from user set",
			zap.Error(err),
			zap.String("connection_id", client.ID),
			zap.String("user_id", client.UserID))
	}

	h.logger.Debug("Connection state deleted",
		zap.String("connection_id", client.ID),
		zap.String("user_id", client.UserID))
}

// updateConnectionLastSeen updates the LastSeen timestamp for a connection
func (h *SSEHandler) updateConnectionLastSeen(ctx context.Context, client *SSEClient) {
	if h.redisClient == nil {
		return
	}

	// Load existing state
	state, err := h.loadConnectionState(ctx, client.ID)
	if err != nil {
		// State doesn't exist, create new one
		h.saveConnectionState(ctx, client)
		return
	}

	// Update LastSeen
	state.LastSeen = client.LastSeen

	// Save updated state
	stateJSON, err := json.Marshal(state)
	if err != nil {
		h.logger.Warn("Failed to marshal updated connection state",
			zap.Error(err),
			zap.String("connection_id", client.ID))
		return
	}

	connectionKey := fmt.Sprintf(connectionStateKeyPattern, client.ID)
	if err := h.redisClient.Set(ctx, connectionKey, stateJSON, connectionStateTTL).Err(); err != nil {
		h.logger.Warn("Failed to update connection state",
			zap.Error(err),
			zap.String("connection_id", client.ID))
	}
}

// getUserConnections returns all connection IDs for a user
func (h *SSEHandler) getUserConnections(ctx context.Context, userID string) ([]string, error) {
	if h.redisClient == nil {
		return nil, fmt.Errorf("Redis client not available")
	}

	userConnectionsKey := fmt.Sprintf(userConnectionsKeyPattern, userID)
	connectionIDs, err := h.redisClient.SMembers(ctx, userConnectionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user connections: %w", err)
	}

	return connectionIDs, nil
}

// getSubscribedEventTypes extracts subscribed event types from client
func (h *SSEHandler) getSubscribedEventTypes(client *SSEClient) []string {
	var eventTypes []string
	for eventType := range client.SubscribedEvents {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// getClientFilters extracts filters from client
func (h *SSEHandler) getClientFilters(client *SSEClient) map[string]interface{} {
	filters := make(map[string]interface{})

	// Providers
	if len(client.Filters.Providers) > 0 {
		providers := make([]string, 0, len(client.Filters.Providers))
		for provider := range client.Filters.Providers {
			providers = append(providers, provider)
		}
		filters["providers"] = providers
	}

	// Credential IDs
	if len(client.Filters.CredentialIDs) > 0 {
		credentialIDs := make([]string, 0, len(client.Filters.CredentialIDs))
		for credentialID := range client.Filters.CredentialIDs {
			credentialIDs = append(credentialIDs, credentialID)
		}
		filters["credential_ids"] = credentialIDs
	}

	// Regions
	if len(client.Filters.Regions) > 0 {
		regions := make([]string, 0, len(client.Filters.Regions))
		for region := range client.Filters.Regions {
			regions = append(regions, region)
		}
		filters["regions"] = regions
	}

	// Resource Types
	if len(client.Filters.ResourceTypes) > 0 {
		resourceTypes := make([]string, 0, len(client.Filters.ResourceTypes))
		for resourceType := range client.Filters.ResourceTypes {
			resourceTypes = append(resourceTypes, resourceType)
		}
		filters["resource_types"] = resourceTypes
	}

	return filters
}
