package network

import (
	"fmt"
	"time"
)

// Cache key helpers
const (
	cachePrefixList      = "list"
	cachePrefixItem      = "item"
	cacheResourceNetwork = "network"
)

// Default TTL for Network cache
const defaultNetworkTTL = 5 * time.Minute

// buildNetworkVPCListKey builds a cache key for VPC lists
// Format: list:network:vpc:{provider}:{credential_id}:{region}
func buildNetworkVPCListKey(provider, credentialID, region string) string {
	return fmt.Sprintf("%s:%s:vpc:%s:%s:%s", cachePrefixList, cacheResourceNetwork, provider, credentialID, region)
}

// buildNetworkVPCItemKey builds a cache key for individual VPCs
// Format: item:network:vpc:{provider}:{credential_id}:{vpc_id}
func buildNetworkVPCItemKey(provider, credentialID, vpcID string) string {
	return fmt.Sprintf("%s:%s:vpc:%s:%s:%s", cachePrefixItem, cacheResourceNetwork, provider, credentialID, vpcID)
}

// buildNetworkSubnetListKey builds a cache key for subnet lists
// Format: list:network:subnet:{provider}:{credential_id}:{vpc_id}
func buildNetworkSubnetListKey(provider, credentialID, vpcID string) string {
	return fmt.Sprintf("%s:%s:subnet:%s:%s:%s", cachePrefixList, cacheResourceNetwork, provider, credentialID, vpcID)
}

// buildNetworkSubnetItemKey builds a cache key for individual subnets
// Format: item:network:subnet:{provider}:{credential_id}:{subnet_id}
func buildNetworkSubnetItemKey(provider, credentialID, subnetID string) string {
	return fmt.Sprintf("%s:%s:subnet:%s:%s:%s", cachePrefixItem, cacheResourceNetwork, provider, credentialID, subnetID)
}

// buildNetworkSecurityGroupListKey builds a cache key for security group lists
// Format: list:network:security-group:{provider}:{credential_id}:{region}
func buildNetworkSecurityGroupListKey(provider, credentialID, region string) string {
	return fmt.Sprintf("%s:%s:security-group:%s:%s:%s", cachePrefixList, cacheResourceNetwork, provider, credentialID, region)
}

// buildNetworkSecurityGroupItemKey builds a cache key for individual security groups
// Format: item:network:security-group:{provider}:{credential_id}:{security_group_id}
func buildNetworkSecurityGroupItemKey(provider, credentialID, securityGroupID string) string {
	return fmt.Sprintf("%s:%s:security-group:%s:%s:%s", cachePrefixItem, cacheResourceNetwork, provider, credentialID, securityGroupID)
}
