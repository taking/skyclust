package kubernetes

import (
	"fmt"
	"time"
)

// Cache key helpers
const (
	cachePrefixList  = "list"
	cachePrefixItem  = "item"
	cacheResourceK8s = "kubernetes"
)

// Default TTL for Kubernetes cache
const defaultK8sTTL = 5 * time.Minute

// buildKubernetesClusterListKey builds a cache key for Kubernetes cluster lists
// Format: list:kubernetes:{provider}:{credential_id}:{region}
func buildKubernetesClusterListKey(provider, credentialID, region string) string {
	if region == "" {
		return fmt.Sprintf("%s:%s:%s:%s", cachePrefixList, cacheResourceK8s, provider, credentialID)
	}
	return fmt.Sprintf("%s:%s:%s:%s:%s", cachePrefixList, cacheResourceK8s, provider, credentialID, region)
}

// buildKubernetesClusterItemKey builds a cache key for individual Kubernetes clusters
// Format: item:kubernetes:{provider}:{credential_id}:{cluster_id}
func buildKubernetesClusterItemKey(provider, credentialID, clusterID string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", cachePrefixItem, cacheResourceK8s, provider, credentialID, clusterID)
}

// buildKubernetesNodePoolListKey builds a cache key for node pool lists
// Format: list:kubernetes:nodepool:{provider}:{credential_id}:{cluster_id}
func buildKubernetesNodePoolListKey(provider, credentialID, clusterID string) string {
	return fmt.Sprintf("%s:%s:nodepool:%s:%s:%s", cachePrefixList, cacheResourceK8s, provider, credentialID, clusterID)
}

