package cache

import (
	"context"
	"fmt"
)

// EventPublisher interface for publishing cache invalidation events
type EventPublisher interface {
	PublishKubernetesClusterEvent(ctx context.Context, provider, credentialID, region, action string, data map[string]interface{}) error
	PublishVPCEvent(ctx context.Context, provider, credentialID, region, action string, data map[string]interface{}) error
	PublishVMEvent(ctx context.Context, provider, workspaceID, vmID, action string, data map[string]interface{}) error
}

// Invalidator handles cache invalidation with event publishing support
type Invalidator struct {
	cache          Cache
	keyBuilder     *CacheKeyBuilder
	eventPublisher EventPublisher
}

// NewInvalidator creates a new cache invalidator
func NewInvalidator(cache Cache) *Invalidator {
	return &Invalidator{
		cache:      cache,
		keyBuilder: NewCacheKeyBuilder(),
	}
}

// NewInvalidatorWithEvents creates a new cache invalidator with event publishing
func NewInvalidatorWithEvents(cache Cache, eventPublisher EventPublisher) *Invalidator {
	return &Invalidator{
		cache:          cache,
		keyBuilder:     NewCacheKeyBuilder(),
		eventPublisher: eventPublisher,
	}
}

// InvalidateByKey invalidates a specific cache key
func (i *Invalidator) InvalidateByKey(ctx context.Context, key string) error {
	if err := i.cache.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to invalidate cache key %s: %w", key, err)
	}
	return nil
}

// InvalidateByPattern invalidates all cache keys matching a pattern
func (i *Invalidator) InvalidateByPattern(ctx context.Context, pattern string) error {
	if redisService, ok := i.cache.(*RedisService); ok {
		if err := redisService.DeletePattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate cache pattern %s: %w", pattern, err)
		}
		return nil
	}

	return fmt.Errorf("pattern-based invalidation not supported for cache type")
}

// InvalidateKubernetesClusterList invalidates Kubernetes cluster list cache
func (i *Invalidator) InvalidateKubernetesClusterList(ctx context.Context, provider, credentialID, region string) error {
	key := i.keyBuilder.BuildKubernetesClusterListKey(provider, credentialID, region)
	err := i.InvalidateByKey(ctx, key)

	// 자동 이벤트 발행: list 업데이트 이벤트
	if err == nil && i.eventPublisher != nil {
		eventData := map[string]interface{}{
			"provider":      provider,
			"credential_id": credentialID,
			"region":        region,
		}
		_ = i.eventPublisher.PublishKubernetesClusterEvent(ctx, provider, credentialID, region, "list", eventData)
	}

	return err
}

// InvalidateKubernetesClusterItem invalidates a specific Kubernetes cluster cache
func (i *Invalidator) InvalidateKubernetesClusterItem(ctx context.Context, provider, credentialID, clusterID string) error {
	key := i.keyBuilder.BuildKubernetesClusterItemKey(provider, credentialID, clusterID)
	err := i.InvalidateByKey(ctx, key)

	// 자동 이벤트 발행: 개별 클러스터 업데이트 이벤트 (region 정보가 없으므로 기본 이벤트만)
	if err == nil && i.eventPublisher != nil {
		eventData := map[string]interface{}{
			"cluster_id": clusterID,
		}
		_ = i.eventPublisher.PublishKubernetesClusterEvent(ctx, provider, credentialID, "", "updated", eventData)
	}

	return err
}

// InvalidateNetworkVPCList invalidates VPC list cache
func (i *Invalidator) InvalidateNetworkVPCList(ctx context.Context, provider, credentialID, region string) error {
	key := i.keyBuilder.BuildNetworkVPCListKey(provider, credentialID, region)
	err := i.InvalidateByKey(ctx, key)

	// 자동 이벤트 발행: VPC 목록 업데이트 이벤트
	if err == nil && i.eventPublisher != nil {
		eventData := map[string]interface{}{
			"provider":      provider,
			"credential_id": credentialID,
			"region":        region,
		}
		_ = i.eventPublisher.PublishVPCEvent(ctx, provider, credentialID, region, "list", eventData)
	}

	return err
}

// InvalidateNetworkVPCItem invalidates a specific VPC cache
func (i *Invalidator) InvalidateNetworkVPCItem(ctx context.Context, provider, credentialID, vpcID string) error {
	key := i.keyBuilder.BuildNetworkVPCItemKey(provider, credentialID, vpcID)
	return i.InvalidateByKey(ctx, key)
}

// InvalidateAllKubernetes invalidates all Kubernetes-related cache for a credential
func (i *Invalidator) InvalidateAllKubernetes(ctx context.Context, provider, credentialID string) error {
	pattern := fmt.Sprintf("list:%s:%s:%s:*", ResourceKubernetes, provider, credentialID)
	return i.InvalidateByPattern(ctx, pattern)
}

// InvalidateAllNetwork invalidates all Network-related cache for a credential
func (i *Invalidator) InvalidateAllNetwork(ctx context.Context, provider, credentialID string) error {
	pattern := fmt.Sprintf("list:%s:%s:%s:*", ResourceNetwork, provider, credentialID)
	return i.InvalidateByPattern(ctx, pattern)
}

// InvalidateVMList invalidates VM list cache for a workspace
func (i *Invalidator) InvalidateVMList(ctx context.Context, workspaceID string) error {
	key := i.keyBuilder.BuildVMListKey(workspaceID)
	err := i.InvalidateByKey(ctx, key)

	// 자동 이벤트 발행: VM 목록 업데이트 이벤트 (provider 정보가 없으므로 기본 이벤트만)
	if err == nil && i.eventPublisher != nil {
		eventData := map[string]interface{}{
			"workspace_id": workspaceID,
		}
		_ = i.eventPublisher.PublishVMEvent(ctx, "", workspaceID, "", "list", eventData)
	}

	return err
}

// InvalidateVMItem invalidates a specific VM cache
func (i *Invalidator) InvalidateVMItem(ctx context.Context, vmID string) error {
	key := i.keyBuilder.BuildVMItemKey(vmID)
	return i.InvalidateByKey(ctx, key)
}

// InvalidateAllVM invalidates all VM-related cache for a workspace
func (i *Invalidator) InvalidateAllVM(ctx context.Context, workspaceID string) error {
	pattern := fmt.Sprintf("%s:%s:%s:*", PrefixList, ResourceVM, workspaceID)
	return i.InvalidateByPattern(ctx, pattern)
}
