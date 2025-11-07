package sse

import "time"

// SSE event types
const (
	EventTypeSystemNotification = "system-notification"
	EventTypeSystemAlert        = "system-alert"
	EventTypeVMStatus           = "vm-status"
	EventTypeVMResource         = "vm-resource"
	EventTypeVMError            = "vm-error"
	EventTypeProviderStatus     = "provider-status"
	EventTypeProviderInstance   = "provider-instance"

	// Kubernetes event types
	EventTypeKubernetesClusterCreated  = "kubernetes-cluster-created"
	EventTypeKubernetesClusterUpdated  = "kubernetes-cluster-updated"
	EventTypeKubernetesClusterDeleted  = "kubernetes-cluster-deleted"
	EventTypeKubernetesClusterList     = "kubernetes-cluster-list"
	EventTypeKubernetesNodePoolCreated = "kubernetes-node-pool-created"
	EventTypeKubernetesNodePoolUpdated = "kubernetes-node-pool-updated"
	EventTypeKubernetesNodePoolDeleted = "kubernetes-node-pool-deleted"
	EventTypeKubernetesNodeCreated     = "kubernetes-node-created"
	EventTypeKubernetesNodeUpdated     = "kubernetes-node-updated"
	EventTypeKubernetesNodeDeleted     = "kubernetes-node-deleted"

	// Network event types
	EventTypeNetworkVPCCreated           = "network-vpc-created"
	EventTypeNetworkVPCUpdated           = "network-vpc-updated"
	EventTypeNetworkVPCDeleted           = "network-vpc-deleted"
	EventTypeNetworkVPCList              = "network-vpc-list"
	EventTypeNetworkSubnetCreated        = "network-subnet-created"
	EventTypeNetworkSubnetUpdated        = "network-subnet-updated"
	EventTypeNetworkSubnetDeleted        = "network-subnet-deleted"
	EventTypeNetworkSubnetList           = "network-subnet-list"
	EventTypeNetworkSecurityGroupCreated = "network-security-group-created"
	EventTypeNetworkSecurityGroupUpdated = "network-security-group-updated"
	EventTypeNetworkSecurityGroupDeleted = "network-security-group-deleted"
	EventTypeNetworkSecurityGroupList    = "network-security-group-list"
)

// SSE timing constants
const (
	CleanupInterval    = 30 * time.Second
	ClientTimeout      = 5 * time.Minute
	HeartbeatInterval  = 30 * time.Second
	BatchFlushInterval = 100 * time.Millisecond
	BatchMaxSize       = 10
)

// SSE data field names
const (
	FieldVMID         = "vmId"
	FieldProvider     = "provider"
	FieldCredentialID = "credentialId"
	FieldRegion       = "region"
	FieldClusterID    = "clusterId"
	FieldVPCID        = "vpcId"
	FieldSubnetID     = "subnetId"
)
