package cache

import (
	"fmt"
	"time"
)

// Cache key prefixes for different resource types
const (
	PrefixList     = "list"
	PrefixItem     = "item"
	PrefixMetadata = "metadata"
	PrefixLock     = "lock"
	PrefixCounter  = "counter"
)

// Resource types
const (
	ResourceKubernetes = "kubernetes"
	ResourceNetwork    = "network"
	ResourceVM         = "vm"
	ResourceCost       = "cost"
)

// CacheKeyBuilder builds cache keys following the naming convention
type CacheKeyBuilder struct{}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{}
}

// BuildListKey builds a cache key for resource lists
// Format: list:{resource}:{provider}:{credential_id}:{region}
func (b *CacheKeyBuilder) BuildListKey(resource, provider, credentialID, region string) string {
	if region == "" {
		return fmt.Sprintf("%s:%s:%s:%s", PrefixList, resource, provider, credentialID)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", PrefixList, resource, provider, credentialID, region)
}

// BuildItemKey builds a cache key for individual resource items
// Format: item:{resource}:{provider}:{credential_id}:{id}
func (b *CacheKeyBuilder) BuildItemKey(resource, provider, credentialID, id string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", PrefixItem, resource, provider, credentialID, id)
}

// BuildMetadataKey builds a cache key for resource metadata
// Format: metadata:{resource}:{provider}:{credential_id}
func (b *CacheKeyBuilder) BuildMetadataKey(resource, provider, credentialID string) string {
	return fmt.Sprintf("%s:%s:%s:%s", PrefixMetadata, resource, provider, credentialID)
}

// BuildLockKey builds a cache key for distributed locks
// Format: lock:{resource}:{provider}:{credential_id}:{region}
func (b *CacheKeyBuilder) BuildLockKey(resource, provider, credentialID, region string) string {
	if region == "" {
		return fmt.Sprintf("%s:%s:%s:%s", PrefixLock, resource, provider, credentialID)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", PrefixLock, resource, provider, credentialID, region)
}

// BuildCounterKey builds a cache key for counters
// Format: counter:{resource}:{provider}:{credential_id}:{name}
func (b *CacheKeyBuilder) BuildCounterKey(resource, provider, credentialID, name string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", PrefixCounter, resource, provider, credentialID, name)
}

// Kubernetes-specific cache key builders

// BuildKubernetesClusterListKey builds a cache key for Kubernetes cluster lists
func (b *CacheKeyBuilder) BuildKubernetesClusterListKey(provider, credentialID, region string) string {
	return b.BuildListKey(ResourceKubernetes, provider, credentialID, region)
}

// BuildKubernetesClusterItemKey builds a cache key for individual Kubernetes clusters
func (b *CacheKeyBuilder) BuildKubernetesClusterItemKey(provider, credentialID, clusterID string) string {
	return b.BuildItemKey(ResourceKubernetes, provider, credentialID, clusterID)
}

// BuildKubernetesNodePoolListKey builds a cache key for node pool lists
func (b *CacheKeyBuilder) BuildKubernetesNodePoolListKey(provider, credentialID, clusterID string) string {
	return fmt.Sprintf("%s:%s:nodepool:%s:%s:%s", PrefixList, ResourceKubernetes, provider, credentialID, clusterID)
}

// Network-specific cache key builders

// BuildNetworkVPCListKey builds a cache key for VPC lists
func (b *CacheKeyBuilder) BuildNetworkVPCListKey(provider, credentialID, region string) string {
	return fmt.Sprintf("%s:%s:vpc:%s:%s:%s", PrefixList, ResourceNetwork, provider, credentialID, region)
}

// BuildNetworkVPCItemKey builds a cache key for individual VPCs
func (b *CacheKeyBuilder) BuildNetworkVPCItemKey(provider, credentialID, vpcID string) string {
	return fmt.Sprintf("%s:%s:vpc:%s:%s:%s", PrefixItem, ResourceNetwork, provider, credentialID, vpcID)
}

// BuildNetworkSubnetListKey builds a cache key for subnet lists
func (b *CacheKeyBuilder) BuildNetworkSubnetListKey(provider, credentialID, vpcID string) string {
	return fmt.Sprintf("%s:%s:subnet:%s:%s:%s", PrefixList, ResourceNetwork, provider, credentialID, vpcID)
}

// BuildNetworkSubnetItemKey builds a cache key for individual subnets
func (b *CacheKeyBuilder) BuildNetworkSubnetItemKey(provider, credentialID, subnetID string) string {
	return fmt.Sprintf("%s:%s:subnet:%s:%s:%s", PrefixItem, ResourceNetwork, provider, credentialID, subnetID)
}

// BuildNetworkSecurityGroupListKey builds a cache key for security group lists
func (b *CacheKeyBuilder) BuildNetworkSecurityGroupListKey(provider, credentialID, region string) string {
	return fmt.Sprintf("%s:%s:security-group:%s:%s:%s", PrefixList, ResourceNetwork, provider, credentialID, region)
}

// BuildNetworkSecurityGroupItemKey builds a cache key for individual security groups
func (b *CacheKeyBuilder) BuildNetworkSecurityGroupItemKey(provider, credentialID, securityGroupID string) string {
	return fmt.Sprintf("%s:%s:security-group:%s:%s:%s", PrefixItem, ResourceNetwork, provider, credentialID, securityGroupID)
}

// VM-specific cache key builders

// BuildVMListKey builds a cache key for VM lists
// Format: list:vm:{workspace_id}
func (b *CacheKeyBuilder) BuildVMListKey(workspaceID string) string {
	return fmt.Sprintf("%s:%s:%s", PrefixList, ResourceVM, workspaceID)
}

// BuildVMItemKey builds a cache key for individual VMs
// Format: item:vm:{vm_id}
func (b *CacheKeyBuilder) BuildVMItemKey(vmID string) string {
	return fmt.Sprintf("%s:%s:%s", PrefixItem, ResourceVM, vmID)
}

// Cache TTL constants
const (
	TTLShort     = 30 * time.Second // For frequently changing data
	TTLMedium    = 5 * time.Minute  // For moderately changing data
	TTLLong      = 30 * time.Minute // For stable data
	TTLVeryLong  = 1 * time.Hour    // For rarely changing data
	TTLLockShort = 10 * time.Second // For short-lived locks
	TTLLockLong  = 5 * time.Minute  // For long-lived locks
)

// GetDefaultTTL returns the default TTL for a resource type
func GetDefaultTTL(resource string) time.Duration {
	switch resource {
	case ResourceKubernetes:
		return TTLShort
	case ResourceNetwork:
		return TTLMedium
	case ResourceVM:
		return TTLShort
	case ResourceCost:
		return TTLLong
	default:
		return TTLMedium
	}
}
