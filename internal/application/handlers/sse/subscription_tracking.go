package sse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Redis key patterns for subscription tracking
const (
	// sse:subscriptions:{resource_type}:{credential_id}:{region} - Set of connection IDs
	subscriptionKeyPattern = "sse:subscriptions:%s:%s:%s"
	// TTL for subscription tracking (2 hours - longer than connection TTL)
	subscriptionTTL = 2 * time.Hour
)

// Resource type constants for subscription tracking
const (
	ResourceTypeKubernetesCluster = "kubernetes-cluster"
	ResourceTypeNetworkVPC        = "network-vpc"
)

// updateSubscriptionTracking updates Redis subscription tracking when a client subscribes/unsubscribes
func (h *SSEHandler) updateSubscriptionTracking(ctx context.Context, clientID string, eventType string, filters map[string]interface{}, isSubscribe bool) {
	if h.redisClient == nil {
		return
	}

	// Only track Kubernetes and Network resource events
	resourceType := h.getResourceTypeFromEventType(eventType)
	if resourceType == "" {
		return
	}

	// Extract credential_id and region from filters
	credentialIDs := h.extractCredentialIDs(filters)
	regions := h.extractRegions(filters)

	// If no filters, we can't track specific subscriptions
	if len(credentialIDs) == 0 || len(regions) == 0 {
		return
	}

	// Update Redis for each credential_id:region combination
	for _, credentialID := range credentialIDs {
		for _, region := range regions {
			key := fmt.Sprintf(subscriptionKeyPattern, resourceType, credentialID, region)

			if isSubscribe {
				// Add connection ID to set
				if err := h.redisClient.SAdd(ctx, key, clientID).Err(); err != nil {
					h.logger.Warn("Failed to add subscription tracking",
						zap.String("key", key),
						zap.String("client_id", clientID),
						zap.Error(err))
					continue
				}
				// Set TTL
				h.redisClient.Expire(ctx, key, subscriptionTTL)
			} else {
				// Remove connection ID from set
				if err := h.redisClient.SRem(ctx, key, clientID).Err(); err != nil {
					h.logger.Warn("Failed to remove subscription tracking",
						zap.String("key", key),
						zap.String("client_id", clientID),
						zap.Error(err))
					continue
				}
			}
		}
	}
}

// getResourceTypeFromEventType extracts resource type from event type
func (h *SSEHandler) getResourceTypeFromEventType(eventType string) string {
	if h.isKubernetesEvent(eventType) {
		// Extract resource type from event type
		// e.g., "kubernetes-cluster-created" -> "kubernetes-cluster"
		if strings.HasPrefix(eventType, "kubernetes-cluster") {
			return ResourceTypeKubernetesCluster
		}
		// For other Kubernetes events (node-pool, node), we don't track them separately
		// as they are tied to clusters
		return ""
	}

	if h.isNetworkEvent(eventType) {
		// Extract resource type from event type
		// e.g., "network-vpc-created" -> "network-vpc"
		if strings.HasPrefix(eventType, "network-vpc") {
			return ResourceTypeNetworkVPC
		}
		// For other Network events (subnet, security-group), we don't track them separately
		// as they are tied to VPCs
		return ""
	}

	return ""
}

// extractCredentialIDs extracts credential IDs from filters
func (h *SSEHandler) extractCredentialIDs(filters map[string]interface{}) []string {
	var credentialIDs []string

	// Check "credential_ids" (backend standard)
	if credentialIDsInterface, ok := filters["credential_ids"].([]interface{}); ok {
		for _, cid := range credentialIDsInterface {
			if credentialID, ok := cid.(string); ok && credentialID != "" {
				credentialIDs = append(credentialIDs, credentialID)
			}
		}
	}

	// Check "credential_id" (singular, also supported)
	if credentialIDInterface, ok := filters["credential_id"]; ok {
		if credentialID, ok := credentialIDInterface.(string); ok && credentialID != "" {
			credentialIDs = append(credentialIDs, credentialID)
		}
	}

	return credentialIDs
}

// extractRegions extracts regions from filters
func (h *SSEHandler) extractRegions(filters map[string]interface{}) []string {
	var regions []string

	// Check "regions" (plural)
	if regionsInterface, ok := filters["regions"].([]interface{}); ok {
		for _, r := range regionsInterface {
			if region, ok := r.(string); ok && region != "" {
				regions = append(regions, region)
			}
		}
	}

	// Check "region" (singular, also supported)
	if regionInterface, ok := filters["region"]; ok {
		if region, ok := regionInterface.(string); ok && region != "" {
			regions = append(regions, region)
		}
	}

	return regions
}

// cleanupSubscriptionTracking removes subscription tracking for a disconnected client
func (h *SSEHandler) cleanupSubscriptionTracking(ctx context.Context, clientID string) {
	if h.redisClient == nil {
		return
	}

	// Find all subscription keys that contain this client ID
	// We need to scan all subscription keys and remove the client ID
	pattern := fmt.Sprintf(subscriptionKeyPattern, "*", "*", "*")
	keys, err := h.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		h.logger.Warn("Failed to scan subscription keys for cleanup",
			zap.String("client_id", clientID),
			zap.Error(err))
		return
	}

	for _, key := range keys {
		if err := h.redisClient.SRem(ctx, key, clientID).Err(); err != nil {
			h.logger.Warn("Failed to remove client from subscription key",
				zap.String("key", key),
				zap.String("client_id", clientID),
				zap.Error(err))
		}
	}
}

