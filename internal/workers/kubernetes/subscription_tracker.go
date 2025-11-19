package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SubscriptionInfo represents subscription information for a credential:region combination
type SubscriptionInfo struct {
	CredentialID string
	Region      string
	Count       int64
}

// SubscriptionPriority represents sync priority based on subscription count
type SubscriptionPriority int

const (
	PriorityHigh   SubscriptionPriority = iota // 1 minute interval (5+ subscriptions)
	PriorityMedium                             // 3 minute interval (1-4 subscriptions)
	PriorityLow                                // 10 minute interval (0 subscriptions)
)

const (
	// Redis key pattern for subscription tracking
	subscriptionKeyPattern = "sse:subscriptions:kubernetes-cluster:%s:%s"
	// Resource type for Kubernetes clusters
	resourceTypeKubernetesCluster = "kubernetes-cluster"
)

// getActiveSubscriptions retrieves active subscription information from Redis
func (w *SyncWorker) getActiveSubscriptions(ctx context.Context) (map[string]map[string]int64, error) {
	if w.redisClient == nil {
		// No Redis available, return empty map (will fall back to default sync)
		return make(map[string]map[string]int64), nil
	}

	// Scan all subscription keys
	pattern := fmt.Sprintf(subscriptionKeyPattern, "*", "*")
	keys, err := w.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		w.logger.Warn("Failed to scan subscription keys",
			zap.Error(err))
		return nil, err
	}

	// Map: credential_id -> region -> subscription count
	subscriptions := make(map[string]map[string]int64)

	for _, key := range keys {
		// Parse key: sse:subscriptions:kubernetes-cluster:{credential_id}:{region}
		parts := strings.Split(key, ":")
		if len(parts) != 5 {
			continue
		}

		credentialID := parts[3]
		region := parts[4]

		// Get subscription count (number of connection IDs in the set)
		count, err := w.redisClient.SCard(ctx, key).Result()
		if err != nil {
			w.logger.Warn("Failed to get subscription count",
				zap.String("key", key),
				zap.Error(err))
			continue
		}

		if count == 0 {
			// Skip empty subscriptions
			continue
		}

		if subscriptions[credentialID] == nil {
			subscriptions[credentialID] = make(map[string]int64)
		}
		subscriptions[credentialID][region] = count
	}

	return subscriptions, nil
}

// getSubscriptionPriority determines sync priority based on subscription count
func (w *SyncWorker) getSubscriptionPriority(count int64) SubscriptionPriority {
	if count >= 5 {
		return PriorityHigh
	} else if count >= 1 {
		return PriorityMedium
	}
	return PriorityLow
}

// getIntervalForPriority returns sync interval for a given priority
func (w *SyncWorker) getIntervalForPriority(priority SubscriptionPriority) time.Duration {
	switch priority {
	case PriorityHigh:
		return 1 * time.Minute
	case PriorityMedium:
		return 3 * time.Minute
	case PriorityLow:
		return 10 * time.Minute
	default:
		return 10 * time.Minute
	}
}

// getSubscribedRegions returns regions that have active subscriptions for a credential
func (w *SyncWorker) getSubscribedRegions(ctx context.Context, credentialID string) ([]string, error) {
	if w.redisClient == nil {
		return []string{}, nil
	}

	// Scan subscription keys for this credential
	pattern := fmt.Sprintf(subscriptionKeyPattern, credentialID, "*")
	keys, err := w.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	regions := make([]string, 0, len(keys))
	for _, key := range keys {
		// Parse key: sse:subscriptions:kubernetes-cluster:{credential_id}:{region}
		parts := strings.Split(key, ":")
		if len(parts) != 5 {
			continue
		}

		region := parts[4]

		// Check if subscription has active connections
		count, err := w.redisClient.SCard(ctx, key).Result()
		if err != nil {
			continue
		}

		if count > 0 {
			regions = append(regions, region)
		}
	}

	return regions, nil
}

// String returns string representation of SubscriptionPriority
func (p SubscriptionPriority) String() string {
	switch p {
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

