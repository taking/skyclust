package messaging

// NATS topic structure constants
// Format: {resource}.{provider}.{credential_id}.{region}.{action}

// Resource types
const (
	ResourceKubernetes = "kubernetes"
	ResourceNetwork    = "network"
	ResourceVM         = "vm"
	ResourceCost       = "cost"
	ResourceWorkspace  = "workspace"
	ResourceCredential = "credential"
)

// Event actions
const (
	ActionCreated = "created"
	ActionUpdated = "updated"
	ActionDeleted = "deleted"
	ActionList    = "list"
)

// Kubernetes topic builders
// Format: kubernetes.{provider}.{credential_id}.{region}.{resource}.{action}
// Example: kubernetes.gcp.cred-123.asia-northeast3.clusters.created

// BuildKubernetesTopic builds a NATS topic for Kubernetes events
func BuildKubernetesTopic(provider, credentialID, region, resource, action string) string {
	if region == "" {
		return ResourceKubernetes + "." + provider + "." + credentialID + "." + resource + "." + action
	}
	return ResourceKubernetes + "." + provider + "." + credentialID + "." + region + "." + resource + "." + action
}

// BuildKubernetesClusterTopic builds a topic for Kubernetes cluster events
func BuildKubernetesClusterTopic(provider, credentialID, region, action string) string {
	return BuildKubernetesTopic(provider, credentialID, region, "clusters", action)
}

// BuildKubernetesNodePoolTopic builds a topic for Kubernetes node pool events
func BuildKubernetesNodePoolTopic(provider, credentialID, clusterID, action string) string {
	return ResourceKubernetes + "." + provider + "." + credentialID + ".clusters." + clusterID + ".nodepools." + action
}

// BuildKubernetesNodeTopic builds a topic for Kubernetes node events
func BuildKubernetesNodeTopic(provider, credentialID, clusterID, action string) string {
	return ResourceKubernetes + "." + provider + "." + credentialID + ".clusters." + clusterID + ".nodes." + action
}

// Network topic builders
// Format: network.{provider}.{credential_id}.{region}.{resource}.{action}
// Example: network.aws.cred-456.ap-northeast-2.vpcs.created

// BuildNetworkTopic builds a NATS topic for Network events
func BuildNetworkTopic(provider, credentialID, region, resource, action string) string {
	if region == "" {
		return ResourceNetwork + "." + provider + "." + credentialID + "." + resource + "." + action
	}
	return ResourceNetwork + "." + provider + "." + credentialID + "." + region + "." + resource + "." + action
}

// BuildNetworkVPCTopic builds a topic for VPC events
func BuildNetworkVPCTopic(provider, credentialID, region, action string) string {
	return BuildNetworkTopic(provider, credentialID, region, "vpcs", action)
}

// BuildNetworkSubnetTopic builds a topic for Subnet events
func BuildNetworkSubnetTopic(provider, credentialID, vpcID, action string) string {
	return ResourceNetwork + "." + provider + "." + credentialID + ".vpcs." + vpcID + ".subnets." + action
}

// BuildNetworkSecurityGroupTopic builds a topic for Security Group events
func BuildNetworkSecurityGroupTopic(provider, credentialID, region, action string) string {
	return BuildNetworkTopic(provider, credentialID, region, "security-groups", action)
}

// VM topic builders (for backward compatibility)
// Format: vm.{provider}.{credential_id}.{region}.{action}

// BuildVMTopic builds a NATS topic for VM events
func BuildVMTopic(provider, credentialID, region, action string) string {
	if region == "" {
		return ResourceVM + "." + provider + "." + credentialID + "." + action
	}
	return ResourceVM + "." + provider + "." + credentialID + "." + region + "." + action
}

// Legacy topic patterns (for backward compatibility with existing SSE handler)
const (
	TopicVMStatusUpdate       = "vm.status.update"
	TopicVMResourceUpdate     = "vm.resource.update"
	TopicProviderStatusUpdate = "provider.status.update"
	TopicProviderInstanceUpdate = "provider.instance.update"
	TopicSystemNotification  = "system.notification"
	TopicSystemAlert          = "system.alert"
)

// Workspace topic builders
// Format: workspace.{workspace_id}.{action}
// Example: workspace.ws-123.created

// BuildWorkspaceTopic builds a NATS topic for Workspace events
func BuildWorkspaceTopic(workspaceID, action string) string {
	return ResourceWorkspace + "." + workspaceID + "." + action
}

// Credential topic builders
// Format: credential.{workspace_id}.{provider}.{action}
// Example: credential.ws-123.aws.created

// BuildCredentialTopic builds a NATS topic for Credential events
func BuildCredentialTopic(workspaceID, provider, action string) string {
	return ResourceCredential + "." + workspaceID + "." + provider + "." + action
}

// Topic patterns for wildcard subscriptions
const (
	PatternKubernetesAll = "kubernetes.*"
	PatternKubernetesProvider = "kubernetes.%s.*"
	PatternKubernetesCredential = "kubernetes.%s.%s.*"
	PatternNetworkAll = "network.*"
	PatternNetworkProvider = "network.%s.*"
	PatternVMAll = "vm.*"
	PatternWorkspaceAll = "workspace.*"
	PatternCredentialAll = "credential.*"
)

