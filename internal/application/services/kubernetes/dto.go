package kubernetes

// Kubernetes Cluster DTOs

// CreateClusterRequest represents a request to create a Kubernetes cluster
type CreateClusterRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=100"`
	Version      string            `json:"version" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	Zone         string            `json:"zone,omitempty"` // GCP zone (optional)
	SubnetIDs    []string          `json:"subnet_ids" validate:"required,min=1"`
	VPCID        string            `json:"vpc_id,omitempty"`
	RoleARN      string            `json:"role_arn,omitempty"` // Optional: 없으면 자동 생성 (arn:aws:iam::{accountId}:role/EKSClusterRole)
	Tags         map[string]string `json:"tags,omitempty"`
	// Access Entry configuration
	AccessConfig *AccessConfigRequest `json:"access_config,omitempty"`
}

// AccessConfigRequest represents access configuration for EKS cluster
type AccessConfigRequest struct {
	// Authentication mode for the cluster
	// Values: "API", "CONFIG_MAP", "API_AND_CONFIG_MAP"
	AuthenticationMode string `json:"authentication_mode,omitempty"`
	// Whether to bootstrap cluster creator as admin
	BootstrapClusterCreatorAdminPermissions *bool `json:"bootstrap_cluster_creator_admin_permissions,omitempty"`
}

// CreateClusterResponse represents the response after creating a cluster
type CreateClusterResponse struct {
	ClusterID string            `json:"cluster_id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Region    string            `json:"region"`
	Zone      string            `json:"zone,omitempty"` // GCP zone
	Status    string            `json:"status"`
	Endpoint  string            `json:"endpoint,omitempty"`
	ProjectID string            `json:"project_id,omitempty"` // GCP project ID
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ListClustersRequest represents a request to list clusters
type ListClustersRequest struct {
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	Region        string `json:"region,omitempty"`
	ResourceGroup string `json:"resource_group,omitempty" form:"resource_group"` // Azure-specific: Resource Group filter
}

// BaseClusterInfo represents common cluster information shared across all providers
// Used for listing clusters and as base for provider-specific cluster info
type BaseClusterInfo struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	Status        string            `json:"status"`
	Region        string            `json:"region"`
	Zone          string            `json:"zone,omitempty"` // GCP zone
	Endpoint      string            `json:"endpoint,omitempty"`
	ResourceGroup string            `json:"resource_group,omitempty"` // Azure-specific, but included in base for listing compatibility
	ProjectID     string            `json:"project_id,omitempty"`     // GCP-specific, but included in base for listing compatibility
	CreatedAt     string            `json:"created_at,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
	Tags          map[string]string `json:"tags,omitempty"`

	// Common network information (basic, provider-agnostic)
	NetworkConfig *NetworkConfigInfo `json:"network_config,omitempty"`

	// Node pool information
	NodePoolInfo *NodePoolSummaryInfo `json:"node_pool_info,omitempty"`

	// Security configuration (common)
	SecurityConfig *SecurityConfigInfo `json:"security_config,omitempty"`
}

// ClusterInfo is an alias for BaseClusterInfo for backward compatibility
// In list operations, this is used. For detail operations, provider-specific types are used.
type ClusterInfo = BaseClusterInfo

// AWSClusterInfo represents AWS EKS cluster information with AWS-specific fields
type AWSClusterInfo struct {
	BaseClusterInfo
	// AWS EKS specific fields
	ResourcesVPCConfig      *AWSResourcesVPCConfig      `json:"resources_vpc_config,omitempty"`
	KubernetesNetworkConfig *AWSKubernetesNetworkConfig `json:"kubernetes_network_config,omitempty"`
	AccessConfig            *AWSAccessConfig            `json:"access_config,omitempty"`
	UpgradePolicy           *AWSUpgradePolicy           `json:"upgrade_policy,omitempty"`
	RoleARN                 string                      `json:"role_arn,omitempty"`
	PlatformVersion         string                      `json:"platform_version,omitempty"`
	DeletionProtection      bool                        `json:"deletion_protection,omitempty"`
}

// AWSResourcesVPCConfig represents AWS EKS VPC configuration
type AWSResourcesVPCConfig struct {
	SubnetIDs              []string `json:"subnet_ids"`
	SecurityGroupIDs       []string `json:"security_group_ids"`
	ClusterSecurityGroupID string   `json:"cluster_security_group_id,omitempty"`
	VPCID                  string   `json:"vpc_id"`
	EndpointPublicAccess   bool     `json:"endpoint_public_access"`
	EndpointPrivateAccess  bool     `json:"endpoint_private_access"`
	PublicAccessCIDRs      []string `json:"public_access_cidrs"`
}

// AWSKubernetesNetworkConfig represents AWS EKS Kubernetes network configuration
type AWSKubernetesNetworkConfig struct {
	ServiceIPv4CIDR      string                   `json:"service_ipv4_cidr,omitempty"`
	ServiceIPv6CIDR      string                   `json:"service_ipv6_cidr,omitempty"`
	IPFamily             string                   `json:"ip_family,omitempty"` // "ipv4" or "ipv6"
	ElasticLoadBalancing *AWSElasticLoadBalancing `json:"elastic_load_balancing,omitempty"`
}

// AWSElasticLoadBalancing represents AWS EKS Elastic Load Balancing configuration
type AWSElasticLoadBalancing struct {
	Enabled bool `json:"enabled"`
}

// AWSAccessConfig represents AWS EKS access configuration
type AWSAccessConfig struct {
	BootstrapClusterCreatorAdminPermissions *bool  `json:"bootstrap_cluster_creator_admin_permissions,omitempty"`
	AuthenticationMode                      string `json:"authentication_mode,omitempty"` // "API", "CONFIG_MAP", "API_AND_CONFIG_MAP"
}

// AWSUpgradePolicy represents AWS EKS upgrade policy
type AWSUpgradePolicy struct {
	SupportType string `json:"support_type,omitempty"` // "EXTENDED", "STANDARD"
}

// GCPClusterInfo represents GCP GKE cluster information with GCP-specific fields
type GCPClusterInfo struct {
	BaseClusterInfo
	// GCP GKE specific fields
	ProjectID                      string                             `json:"project_id,omitempty"`
	NetworkConfig                  *GCPNetworkConfig                  `json:"network_config,omitempty"`
	SecurityConfig                 *GCPSecurityConfig                 `json:"security_config,omitempty"`
	WorkloadIdentityConfig         *GCPWorkloadIdentityConfig         `json:"workload_identity_config,omitempty"`
	PrivateClusterConfig           *GCPPrivateClusterConfig           `json:"private_cluster_config,omitempty"`
	MasterAuthorizedNetworksConfig *GCPMasterAuthorizedNetworksConfig `json:"master_authorized_networks_config,omitempty"`
}

// GCPNetworkConfig represents GCP GKE network configuration
type GCPNetworkConfig struct {
	Network         string `json:"network,omitempty"`      // VPC network name
	Subnetwork      string `json:"subnetwork,omitempty"`   // Subnet name
	PodCIDR         string `json:"pod_cidr,omitempty"`     // Pod CIDR range
	ServiceCIDR     string `json:"service_cidr,omitempty"` // Service CIDR range
	PrivateNodes    bool   `json:"private_nodes,omitempty"`
	PrivateEndpoint bool   `json:"private_endpoint,omitempty"`
}

// GCPSecurityConfig represents GCP GKE security configuration
type GCPSecurityConfig struct {
	WorkloadIdentity    bool `json:"workload_identity,omitempty"`
	BinaryAuthorization bool `json:"binary_authorization,omitempty"`
	NetworkPolicy       bool `json:"network_policy,omitempty"`
	PodSecurityPolicy   bool `json:"pod_security_policy,omitempty"`
}

// GCPWorkloadIdentityConfig represents GCP GKE workload identity configuration
type GCPWorkloadIdentityConfig struct {
	WorkloadPool string `json:"workload_pool,omitempty"`
}

// GCPPrivateClusterConfig represents GCP GKE private cluster configuration
type GCPPrivateClusterConfig struct {
	EnablePrivateNodes    bool   `json:"enable_private_nodes"`
	EnablePrivateEndpoint bool   `json:"enable_private_endpoint"`
	MasterIPv4CIDR        string `json:"master_ipv4_cidr,omitempty"`
}

// GCPMasterAuthorizedNetworksConfig represents GCP GKE master authorized networks configuration
type GCPMasterAuthorizedNetworksConfig struct {
	Enabled    bool     `json:"enabled"`
	CIDRBlocks []string `json:"cidr_blocks,omitempty"`
}

// AzureClusterInfo represents Azure AKS cluster information with Azure-specific fields
type AzureClusterInfo struct {
	BaseClusterInfo
	// Azure AKS specific fields
	ResourceGroup               string                 `json:"resource_group,omitempty"`
	NetworkProfile              *AzureNetworkProfile   `json:"network_profile,omitempty"`
	ServicePrincipal            *AzureServicePrincipal `json:"service_principal,omitempty"`
	AddonProfiles               map[string]interface{} `json:"addon_profiles,omitempty"`
	EnableRBAC                  bool                   `json:"enable_rbac,omitempty"`
	EnablePodSecurityPolicy     bool                   `json:"enable_pod_security_policy,omitempty"`
	APIServerAuthorizedIPRanges []string               `json:"api_server_authorized_ip_ranges,omitempty"`
}

// AzureNetworkProfile represents Azure AKS network profile
type AzureNetworkProfile struct {
	NetworkPlugin    string `json:"network_plugin,omitempty"` // "azure" or "kubenet"
	NetworkPolicy    string `json:"network_policy,omitempty"` // "azure" or "calico"
	PodCIDR          string `json:"pod_cidr,omitempty"`
	ServiceCIDR      string `json:"service_cidr,omitempty"`
	DNSServiceIP     string `json:"dns_service_ip,omitempty"`
	DockerBridgeCIDR string `json:"docker_bridge_cidr,omitempty"`
	LoadBalancerSku  string `json:"load_balancer_sku,omitempty"`
	NetworkMode      string `json:"network_mode,omitempty"`
}

// AzureServicePrincipal represents Azure AKS service principal information
type AzureServicePrincipal struct {
	ClientID string `json:"client_id,omitempty"`
}

// NetworkConfigInfo represents network configuration for a cluster
type NetworkConfigInfo struct {
	VPCID           string `json:"vpc_id,omitempty"`
	SubnetID        string `json:"subnet_id,omitempty"`
	PodCIDR         string `json:"pod_cidr,omitempty"`
	ServiceCIDR     string `json:"service_cidr,omitempty"`
	PrivateNodes    bool   `json:"private_nodes,omitempty"`
	PrivateEndpoint bool   `json:"private_endpoint,omitempty"`
}

// NodePoolSummaryInfo represents summary information about node pools
type NodePoolSummaryInfo struct {
	TotalNodePools int32 `json:"total_node_pools"`
	TotalNodes     int32 `json:"total_nodes"`
	MinNodes       int32 `json:"min_nodes"`
	MaxNodes       int32 `json:"max_nodes"`
}

// SecurityConfigInfo represents security configuration for a cluster
type SecurityConfigInfo struct {
	WorkloadIdentity    bool `json:"workload_identity,omitempty"`
	BinaryAuthorization bool `json:"binary_authorization,omitempty"`
	NetworkPolicy       bool `json:"network_policy,omitempty"`
	PodSecurityPolicy   bool `json:"pod_security_policy,omitempty"`
}

// ListClustersResponse represents the response after listing clusters
// Uses BaseClusterInfo for listing (common fields only)
type ListClustersResponse struct {
	Clusters []BaseClusterInfo `json:"clusters"`
}

// BatchListClustersRequest represents a request to list clusters from multiple credentials and regions
type BatchListClustersRequest struct {
	Queries []BatchClusterQuery `json:"queries" validate:"required,min=1,dive"`
}

// BatchClusterQuery represents a single query in a batch request
type BatchClusterQuery struct {
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	Region        string `json:"region" validate:"required"`
	ResourceGroup string `json:"resource_group,omitempty"` // Azure-specific
}

// BatchListClustersResponse represents the response from a batch cluster listing request
type BatchListClustersResponse struct {
	Results []BatchClusterResult `json:"results"`
	Total   int                  `json:"total"`
}

// BatchClusterResult represents the result of a single query in a batch request
type BatchClusterResult struct {
	CredentialID string            `json:"credential_id"`
	Region       string            `json:"region"`
	Provider     string            `json:"provider"`
	Clusters     []BaseClusterInfo `json:"clusters"`
	Error        *BatchError       `json:"error,omitempty"`
}

// BatchError represents an error for a single query in a batch request
type BatchError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetClusterRequest represents a request to get cluster details
type GetClusterRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	ClusterName  string `json:"cluster_name" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// DeleteClusterRequest represents a request to delete a cluster
type DeleteClusterRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	ClusterName  string `json:"cluster_name" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// GetKubeconfigRequest represents a request to get kubeconfig
type GetKubeconfigRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	ClusterName  string `json:"cluster_name" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// CreateNodePoolRequest represents a request to create a node pool
type CreateNodePoolRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	ClusterName  string            `json:"cluster_name" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	NodePoolName string            `json:"node_pool_name" validate:"required"`
	InstanceType string            `json:"instance_type" validate:"required"`
	MinSize      int32             `json:"min_size" validate:"required,min=1"`
	MaxSize      int32             `json:"max_size" validate:"required,min=1"`
	DesiredSize  int32             `json:"desired_size" validate:"required,min=1"`
	DiskSize     int32             `json:"disk_size,omitempty"`
	SubnetIDs    []string          `json:"subnet_ids" validate:"required,min=1"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// CreateNodeGroupRequest represents a request to create an EKS node group
type CreateNodeGroupRequest struct {
	CredentialID      string                 `json:"credential_id" validate:"required,uuid"`
	ClusterName       string                 `json:"cluster_name" validate:"required"`
	Region            string                 `json:"region" validate:"required"`
	NodeGroupName     string                 `json:"name" validate:"required"`     // Changed from node_group_name to name for frontend consistency
	NodeRoleARN       string                 `json:"node_role_arn,omitempty"`      // Optional: 없으면 자동 생성 (arn:aws:iam::{accountId}:role/EKSNodeRole)
	AvailabilityZones []string               `json:"availability_zones,omitempty"` // Optional: 선택한 AZ 목록 (GPU instance type이 사용 가능한 AZ만)
	SubnetIDs         []string               `json:"subnet_ids" validate:"required,min=1"`
	InstanceTypes     []string               `json:"instance_types" validate:"required,min=1"`
	ScalingConfig     NodeGroupScalingConfig `json:"scaling_config" validate:"required"`
	DiskSize          int32                  `json:"disk_size,omitempty"`
	AMIType           string                 `json:"ami_type,omitempty"`      // EKS AMI Type (AL2023_x86_64_STANDARD, AL2023_x86_64_NVIDIA, etc.)
	AMI               string                 `json:"ami,omitempty"`           // Deprecated: Use ami_type instead
	CapacityType      string                 `json:"capacity_type,omitempty"` // ON_DEMAND, SPOT
	Tags              map[string]string      `json:"tags,omitempty"`
}

// NodeGroupScalingConfig represents scaling configuration for node group
type NodeGroupScalingConfig struct {
	MinSize     int32 `json:"min_size" validate:"required,min=1"`
	MaxSize     int32 `json:"max_size" validate:"required,min=1"`
	DesiredSize int32 `json:"desired_size" validate:"required,min=1"`
}

// CreateNodeGroupResponse represents the response after creating a node group
type CreateNodeGroupResponse struct {
	NodeGroupName string                 `json:"node_group_name"`
	ClusterName   string                 `json:"cluster_name"`
	Status        string                 `json:"status"`
	InstanceTypes []string               `json:"instance_types"`
	ScalingConfig NodeGroupScalingConfig `json:"scaling_config"`
	Tags          map[string]string      `json:"tags,omitempty"`
	CreatedAt     string                 `json:"created_at"`
}

// ListNodeGroupsRequest represents a request to list node groups
type ListNodeGroupsRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	ClusterName  string `json:"cluster_name" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// NodeGroupInfo represents basic node group information for listing
// This is a base struct that can be extended for provider-specific details
type NodeGroupInfo struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Status          string                 `json:"status"`
	ClusterName     string                 `json:"cluster_name"`
	Region          string                 `json:"region"`
	InstanceTypes   []string               `json:"instance_types"`
	ScalingConfig   NodeGroupScalingConfig `json:"scaling_config"`
	CapacityType    string                 `json:"capacity_type"`
	DiskSize        int32                  `json:"disk_size"`
	DiskType        string                 `json:"disk_type,omitempty"`
	ImageType       string                 `json:"image_type,omitempty"`
	Preemptible     bool                   `json:"preemptible,omitempty"`
	Spot            bool                   `json:"spot,omitempty"`
	ServiceAccount  string                 `json:"service_account,omitempty"`
	OAuthScopes     []string               `json:"oauth_scopes,omitempty"`
	Tags            map[string]string      `json:"tags,omitempty"`
	Taints          []NodeTaint            `json:"taints,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	NetworkConfig   *NodeNetworkConfig     `json:"network_config,omitempty"`
	Management      *NodeManagement        `json:"management,omitempty"`
	UpgradeSettings *UpgradeSettings       `json:"upgrade_settings,omitempty"`
	CreatedAt       string                 `json:"created_at,omitempty"`
	UpdatedAt       string                 `json:"updated_at,omitempty"`
}

// AWSNodeGroupInfo represents AWS EKS node group detailed information
// Extends NodeGroupInfo with AWS-specific fields
type AWSNodeGroupInfo struct {
	NodeGroupInfo
	// AWS EKS specific fields
	NodeRoleARN        string                 `json:"node_role_arn,omitempty"`
	AMIType            string                 `json:"ami_type,omitempty"`
	ReleaseVersion     string                 `json:"release_version,omitempty"`
	Subnets            []string               `json:"subnets,omitempty"`
	RemoteAccessConfig *AWSRemoteAccessConfig `json:"remote_access_config,omitempty"`
	Resources          *AWSNodeGroupResources `json:"resources,omitempty"`
	Health             *AWSNodeGroupHealth    `json:"health,omitempty"`
	LaunchTemplate     *AWSLaunchTemplateSpec `json:"launch_template,omitempty"`
	UpdateConfig       *AWSUpdateConfig       `json:"update_config,omitempty"`
}

// AWSRemoteAccessConfig represents AWS EKS node group remote access configuration
type AWSRemoteAccessConfig struct {
	EC2SSHKey            string   `json:"ec2_ssh_key,omitempty"`
	SourceSecurityGroups []string `json:"source_security_groups,omitempty"`
}

// AWSNodeGroupResources represents AWS EKS node group resources
type AWSNodeGroupResources struct {
	AutoScalingGroups         []AWSAutoScalingGroup `json:"auto_scaling_groups,omitempty"`
	RemoteAccessSecurityGroup string                `json:"remote_access_security_group,omitempty"`
}

// AWSAutoScalingGroup represents an Auto Scaling Group
type AWSAutoScalingGroup struct {
	Name string `json:"name"`
}

// AWSNodeGroupHealth represents AWS EKS node group health information
type AWSNodeGroupHealth struct {
	Issues []AWSNodeGroupHealthIssue `json:"issues,omitempty"`
}

// AWSNodeGroupHealthIssue represents a health issue
type AWSNodeGroupHealthIssue struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	ResourceIDs []string `json:"resource_ids,omitempty"`
}

// AWSLaunchTemplateSpec represents AWS Launch Template specification
type AWSLaunchTemplateSpec struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

// AWSUpdateConfig represents AWS EKS node group update configuration
type AWSUpdateConfig struct {
	MaxUnavailable           int32 `json:"max_unavailable,omitempty"`
	MaxUnavailablePercentage int32 `json:"max_unavailable_percentage,omitempty"`
}

// ListNodeGroupsResponse represents the response after listing node groups
type ListNodeGroupsResponse struct {
	NodeGroups []NodeGroupInfo `json:"node_groups"`
	Total      int             `json:"total"`
}

// GetNodeGroupRequest represents a request to get node group details
type GetNodeGroupRequest struct {
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	ClusterName   string `json:"cluster_name" validate:"required"`
	NodeGroupName string `json:"node_group_name" validate:"required"`
	Region        string `json:"region" validate:"required"`
}

// UpdateNodeGroupRequest represents a request to update a node group
type UpdateNodeGroupRequest struct {
	CredentialID  string                  `json:"credential_id" validate:"required,uuid"`
	ClusterName   string                  `json:"cluster_name" validate:"required"`
	NodeGroupName string                  `json:"node_group_name" validate:"required"`
	Region        string                  `json:"region" validate:"required"`
	ScalingConfig *NodeGroupScalingConfig `json:"scaling_config,omitempty"`
	UpdateConfig  *UpdateConfig           `json:"update_config,omitempty"`
	Labels        map[string]string       `json:"labels,omitempty"`
	Taints        []NodeTaint             `json:"taints,omitempty"`
}

// UpdateConfig represents update configuration for node group
type UpdateConfig struct {
	MaxUnavailable           *int32 `json:"max_unavailable,omitempty"`
	MaxUnavailablePercentage *int32 `json:"max_unavailable_percentage,omitempty"`
}

// DeleteNodeGroupRequest represents a request to delete a node group
type DeleteNodeGroupRequest struct {
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	ClusterName   string `json:"cluster_name" validate:"required"`
	NodeGroupName string `json:"node_group_name" validate:"required"`
	Region        string `json:"region" validate:"required"`
}

// AWS Resource DTOs for EKS cluster creation

// IAMRoleInfo represents IAM role information
type IAMRoleInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ARN          string `json:"arn"`
	Path         string `json:"path"`
	Description  string `json:"description"`
	IsEKSRelated bool   `json:"is_eks_related"`
	CreatedAt    string `json:"created_at"`
	Region       string `json:"region"`
}

// ListIAMRolesRequest represents a request to list IAM roles
type ListIAMRolesRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Region       string `json:"region" validate:"required"`
}

// ListIAMRolesResponse represents the response after listing IAM roles
type ListIAMRolesResponse struct {
	Roles []IAMRoleInfo `json:"roles"`
}

// NodeTaint represents a node taint
type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value,omitempty"`
	Effect string `json:"effect"` // NoSchedule, PreferNoSchedule, NoExecute
}

// NodeNetworkConfig represents node network configuration
type NodeNetworkConfig struct {
	CreatePodRange     bool   `json:"create_pod_range,omitempty"`
	PodRange           string `json:"pod_range,omitempty"`
	PodRangeName       string `json:"pod_range_name,omitempty"`
	EnablePrivateNodes bool   `json:"enable_private_nodes,omitempty"`
}

// NodeManagement represents node management configuration
type NodeManagement struct {
	AutoRepair  bool `json:"auto_repair,omitempty"`
	AutoUpgrade bool `json:"auto_upgrade,omitempty"`
}

// UpgradeSettings represents upgrade settings for node pool
type UpgradeSettings struct {
	MaxSurge       int32  `json:"max_surge,omitempty"`
	MaxUnavailable int32  `json:"max_unavailable,omitempty"`
	Strategy       string `json:"strategy,omitempty"` // SURGE, BLUE_GREEN
}

// GCP GKE Cluster DTOs

// CreateGKEClusterRequest represents a request to create a GCP GKE cluster with sectioned grouping
type CreateGKEClusterRequest struct {
	// 기본 정보 (필수)
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Name         string `json:"name" validate:"required,min=1,max=100"`
	Version      string `json:"version" validate:"required"`
	Region       string `json:"region" validate:"required"`
	Zone         string `json:"zone,omitempty"`
	ProjectID    string `json:"project_id" validate:"required"`

	// 네트워크 설정 (필수)
	Network *GKENetworkConfig `json:"network" validate:"required"`

	// 노드풀 설정 (GKE Standard 모드 필수)
	NodePool *GKENodePoolConfig `json:"node_pool,omitempty"`

	// 보안 설정 (선택)
	Security *GKESecurityConfig `json:"security,omitempty"`

	// 클러스터 모드 설정 (선택)
	ClusterMode *GKEClusterModeConfig `json:"cluster_mode,omitempty"`

	// 태그
	Tags map[string]string `json:"tags,omitempty"`
}

// GKENetworkConfig represents GKE network configuration
type GKENetworkConfig struct {
	VPCID    string `json:"vpc_id" validate:"required"`
	SubnetID string `json:"subnet_id" validate:"required"`

	// 고급 네트워크 설정
	PrivateNodes             bool     `json:"private_nodes,omitempty"`
	PrivateEndpoint          bool     `json:"private_endpoint,omitempty"`
	MasterAuthorizedNetworks []string `json:"master_authorized_networks,omitempty"`
	PodCIDR                  string   `json:"pod_cidr,omitempty"`
	ServiceCIDR              string   `json:"service_cidr,omitempty"`
}

// GKENodePoolConfig represents GKE node pool configuration
type GKENodePoolConfig struct {
	Name        string `json:"name" validate:"required"`
	MachineType string `json:"machine_type" validate:"required"`
	DiskSizeGB  int32  `json:"disk_size_gb,omitempty"`
	DiskType    string `json:"disk_type,omitempty"`
	NodeCount   int32  `json:"node_count" validate:"required,min=1"`

	// 오토스케일링
	AutoScaling *GKEAutoScalingConfig `json:"auto_scaling,omitempty"`

	// 노드 설정
	Labels      map[string]string `json:"labels,omitempty"`
	Taints      []string          `json:"taints,omitempty"`
	Preemptible bool              `json:"preemptible,omitempty"`
	Spot        bool              `json:"spot,omitempty"`
}

// GKEAutoScalingConfig represents GKE auto scaling configuration
type GKEAutoScalingConfig struct {
	Enabled      bool  `json:"enabled"`
	MinNodeCount int32 `json:"min_node_count" validate:"min=0"`
	MaxNodeCount int32 `json:"max_node_count" validate:"min=1"`
}

// GKESecurityConfig represents GKE security configuration
type GKESecurityConfig struct {
	WorkloadIdentity    bool `json:"workload_identity,omitempty"`
	BinaryAuthorization bool `json:"binary_authorization,omitempty"`
	NetworkPolicy       bool `json:"network_policy,omitempty"`
	PodSecurityPolicy   bool `json:"pod_security_policy,omitempty"`
}

// GKEClusterModeConfig represents GKE cluster mode configuration
type GKEClusterModeConfig struct {
	Type                  string `json:"type" validate:"oneof=standard autopilot"` // "standard" or "autopilot"
	RemoveDefaultNodePool bool   `json:"remove_default_node_pool,omitempty"`       // 기본 노드풀 제거 여부
}

// CreateGKEClusterResponse represents the response after creating a GKE cluster
type CreateGKEClusterResponse struct {
	ClusterID string            `json:"cluster_id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Region    string            `json:"region"`
	Zone      string            `json:"zone,omitempty"`
	Status    string            `json:"status"`
	Endpoint  string            `json:"endpoint,omitempty"`
	ProjectID string            `json:"project_id"`
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// GCP Network DTOs (GCP-specific network resources, related to GKE)

// GCPVPCInfo represents GCP VPC information
type GCPVPCInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CIDRBlock string            `json:"cidr_block"`
	State     string            `json:"state"`
	IsDefault bool              `json:"is_default"`
	Region    string            `json:"region"`
	ProjectID string            `json:"project_id"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// GCPSubnetInfo represents GCP subnet information
type GCPSubnetInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	VPCID     string            `json:"vpc_id"`
	CIDRBlock string            `json:"cidr_block"`
	Region    string            `json:"region"`
	Zone      string            `json:"zone,omitempty"`
	State     string            `json:"state"`
	IsPublic  bool              `json:"is_public"`
	ProjectID string            `json:"project_id"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// GCPSecurityGroupInfo represents GCP firewall rule information
type GCPSecurityGroupInfo struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	VPCID       string                     `json:"vpc_id"`
	Region      string                     `json:"region"`
	ProjectID   string                     `json:"project_id"`
	Rules       []GCPSecurityGroupRuleInfo `json:"rules,omitempty"`
	Tags        map[string]string          `json:"tags,omitempty"`
}

// GCPSecurityGroupRuleInfo represents GCP firewall rule information
type GCPSecurityGroupRuleInfo struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"` // ingress or egress
	Protocol    string   `json:"protocol"`
	FromPort    int32    `json:"from_port,omitempty"`
	ToPort      int32    `json:"to_port,omitempty"`
	CIDRBlocks  []string `json:"cidr_blocks,omitempty"`
	SourceTags  []string `json:"source_tags,omitempty"`
	TargetTags  []string `json:"target_tags,omitempty"`
	Description string   `json:"description,omitempty"`
	Priority    int32    `json:"priority,omitempty"`
	Action      string   `json:"action,omitempty"` // "allow" or "deny"
}

// ListGCPVPCsRequest represents a request to list GCP VPCs
type ListGCPVPCsRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Region       string `json:"region,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	VPCID        string `json:"vpc_id,omitempty"`
}

// ListGCPVPCsResponse represents the response after listing GCP VPCs
type ListGCPVPCsResponse struct {
	VPCs []GCPVPCInfo `json:"vpcs"`
}

// ListGCPSubnetsRequest represents a request to list GCP subnets
type ListGCPSubnetsRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id,omitempty"`
	Region       string `json:"region,omitempty"`
	ProjectID    string `json:"project_id,omitempty"`
	SubnetID     string `json:"subnet_id,omitempty"`
}

// ListGCPSubnetsResponse represents the response after listing GCP subnets
type ListGCPSubnetsResponse struct {
	Subnets []GCPSubnetInfo `json:"subnets"`
}

// ListGCPSecurityGroupsRequest represents a request to list GCP security groups
type ListGCPSecurityGroupsRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	VPCID           string `json:"vpc_id,omitempty"`
	Region          string `json:"region,omitempty"`
	ProjectID       string `json:"project_id,omitempty"`
	SecurityGroupID string `json:"security_group_id,omitempty"`
}

// ListGCPSecurityGroupsResponse represents the response after listing GCP security groups
type ListGCPSecurityGroupsResponse struct {
	SecurityGroups []GCPSecurityGroupInfo `json:"security_groups"`
}

// CreateGCPVPCRequest represents a request to create a GCP VPC
type CreateGCPVPCRequest struct {
	CredentialID      string            `json:"credential_id" validate:"required,uuid"`
	Name              string            `json:"name" validate:"required,min=1,max=255"`
	Description       string            `json:"description,omitempty"`
	CIDRBlock         string            `json:"cidr_block" validate:"required,cidr"`
	Region            string            `json:"region,omitempty"` // Optional for VPC (Global resource)
	ProjectID         string            `json:"project_id" validate:"required"`
	AutoCreateSubnets bool              `json:"auto_create_subnets,omitempty"`
	RoutingMode       string            `json:"routing_mode,omitempty"` // "REGIONAL" or "GLOBAL"
	MTU               int64             `json:"mtu,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// UpdateGCPVPCRequest represents a request to update a GCP VPC
type UpdateGCPVPCRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateGCPSubnetRequest represents a request to create a GCP subnet
type CreateGCPSubnetRequest struct {
	CredentialID          string            `json:"credential_id" validate:"required,uuid"`
	Name                  string            `json:"name" validate:"required,min=1,max=255"`
	Description           string            `json:"description,omitempty"`
	VPCID                 string            `json:"vpc_id" validate:"required"`
	CIDRBlock             string            `json:"cidr_block" validate:"required,cidr"`
	Region                string            `json:"region" validate:"required"`
	Zone                  string            `json:"zone,omitempty"`
	ProjectID             string            `json:"project_id" validate:"required"`
	PrivateIPGoogleAccess bool              `json:"private_ip_google_access,omitempty"`
	FlowLogs              bool              `json:"flow_logs,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// UpdateGCPSubnetRequest represents a request to update a GCP subnet
type UpdateGCPSubnetRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateGCPSecurityGroupRequest represents a request to create a GCP security group (firewall rule)
type CreateGCPSecurityGroupRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=255"`
	Description  string            `json:"description" validate:"required,min=1,max=255"`
	VPCID        string            `json:"vpc_id" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	ProjectID    string            `json:"project_id" validate:"required"`
	Priority     int64             `json:"priority,omitempty"`
	Direction    string            `json:"direction,omitempty"` // "INGRESS" or "EGRESS"
	Action       string            `json:"action,omitempty"`    // "ALLOW" or "DENY"
	SourceRanges []string          `json:"source_ranges,omitempty"`
	TargetTags   []string          `json:"target_tags,omitempty"`
	Allowed      []GCPFirewallRule `json:"allowed,omitempty"`
	Denied       []GCPFirewallRule `json:"denied,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// GCPFirewallRule represents a firewall rule for GCP
type GCPFirewallRule struct {
	Protocol string   `json:"protocol" validate:"required"`
	Ports    []string `json:"ports,omitempty"`
}

// UpdateGCPSecurityGroupRequest represents a request to update a GCP security group
type UpdateGCPSecurityGroupRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description string            `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// GetGCPVPCRequest represents a request to get a specific GCP VPC
type GetGCPVPCRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
	ProjectID    string `json:"project_id" validate:"required"`
}

// GetGCPSubnetRequest represents a request to get a specific GCP subnet
type GetGCPSubnetRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	SubnetID     string `json:"subnet_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
	ProjectID    string `json:"project_id" validate:"required"`
}

// GetGCPSecurityGroupRequest represents a request to get a specific GCP security group
type GetGCPSecurityGroupRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	Region          string `json:"region" validate:"required"`
	ProjectID       string `json:"project_id" validate:"required"`
}

// DeleteGCPVPCRequest represents a request to delete a GCP VPC
type DeleteGCPVPCRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
	ProjectID    string `json:"project_id" validate:"required"`
}

// DeleteGCPSubnetRequest represents a request to delete a GCP subnet
type DeleteGCPSubnetRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	SubnetID     string `json:"subnet_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
	ProjectID    string `json:"project_id" validate:"required"`
}

// DeleteGCPSecurityGroupRequest represents a request to delete a GCP security group
type DeleteGCPSecurityGroupRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	Region          string `json:"region" validate:"required"`
	ProjectID       string `json:"project_id" validate:"required"`
}

// AddGCPSecurityGroupRuleRequest represents a request to add a GCP security group rule
type AddGCPSecurityGroupRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	ProjectID       string   `json:"project_id" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=ingress egress"`
	Protocol        string   `json:"protocol" validate:"required"`
	FromPort        int32    `json:"from_port"`
	ToPort          int32    `json:"to_port"`
	CIDRBlocks      []string `json:"cidr_blocks,omitempty"`
	SourceTags      []string `json:"source_tags,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
	Description     string   `json:"description,omitempty"`
	Priority        int32    `json:"priority,omitempty"`
	Action          string   `json:"action,omitempty"` // "allow" or "deny"
}

// RemoveGCPSecurityGroupRuleRequest represents a request to remove a GCP security group rule
type RemoveGCPSecurityGroupRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	ProjectID       string   `json:"project_id" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=ingress egress"`
	Protocol        string   `json:"protocol" validate:"required"`
	FromPort        int32    `json:"from_port"`
	ToPort          int32    `json:"to_port"`
	CIDRBlocks      []string `json:"cidr_blocks,omitempty"`
	SourceTags      []string `json:"source_tags,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
}

// UpdateGCPSecurityGroupRulesRequest represents a request to update GCP security group rules
type UpdateGCPSecurityGroupRulesRequest struct {
	CredentialID    string                     `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string                     `json:"security_group_id" validate:"required"`
	Region          string                     `json:"region" validate:"required"`
	ProjectID       string                     `json:"project_id" validate:"required"`
	IngressRules    []GCPSecurityGroupRuleInfo `json:"ingress_rules,omitempty"`
	EgressRules     []GCPSecurityGroupRuleInfo `json:"egress_rules,omitempty"`
}

// GCPSecurityGroupRuleResponse represents a response for GCP security group rule operations
type GCPSecurityGroupRuleResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Data    *GCPSecurityGroupInfo `json:"data,omitempty"`
}

// Azure AKS Cluster DTOs

// CreateAKSClusterRequest represents a request to create an Azure AKS cluster
type CreateAKSClusterRequest struct {
	// 기본 정보 (필수)
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	Name          string `json:"name" validate:"required,min=1,max=63"`
	Version       string `json:"version,omitempty"`            // 선택사항: 지정하지 않으면 지원되는 최신 버전 자동 선택
	Location      string `json:"location" validate:"required"` // Azure location (e.g., "eastus")
	ResourceGroup string `json:"resource_group" validate:"required"`

	// 네트워크 설정 (선택)
	// nil이면 Kubenet 모드로 자동 생성 (VNet/Subnet 자동 생성)
	Network *AKSNetworkConfig `json:"network,omitempty"`

	// 노드 풀 설정 (필수)
	NodePool *AKSNodePoolConfig `json:"node_pool" validate:"required"`

	// 보안 설정 (선택)
	Security *AKSSecurityConfig `json:"security,omitempty"`

	// 태그
	Tags map[string]*string `json:"tags,omitempty"`
}

// AKSNetworkConfig represents AKS network configuration
type AKSNetworkConfig struct {
	// VirtualNetworkID: Azure Virtual Network ID (Azure CNI 모드일 때만 필수)
	VirtualNetworkID string `json:"virtual_network_id,omitempty"`
	// SubnetID: Azure Subnet ID (Azure CNI 모드일 때만 필수)
	SubnetID string `json:"subnet_id,omitempty"`
	// NetworkPlugin: "azure" (기존 VNet 사용) or "kubenet" (자동 생성, 기본값)
	NetworkPlugin string `json:"network_plugin,omitempty"`
	// NetworkPolicy: "azure" or "calico" (선택)
	NetworkPolicy string `json:"network_policy,omitempty"`
	// PodCIDR: Pod CIDR 블록 (Kubenet 모드일 때 선택)
	PodCIDR string `json:"pod_cidr,omitempty"`
	// ServiceCIDR: Service CIDR 블록 (선택)
	ServiceCIDR string `json:"service_cidr,omitempty"`
	// DNSServiceIP: DNS 서비스 IP (선택)
	DNSServiceIP string `json:"dns_service_ip,omitempty"`
	// DockerBridgeCIDR: Docker 브리지 CIDR (선택)
	DockerBridgeCIDR string `json:"docker_bridge_cidr,omitempty"`
}

// AKSNodePoolConfig represents AKS node pool configuration
type AKSNodePoolConfig struct {
	Name              string            `json:"name" validate:"required"`
	VMSize            string            `json:"vm_size" validate:"required"` // e.g., "Standard_D2s_v3"
	OSDiskSizeGB      int32             `json:"os_disk_size_gb,omitempty"`
	OSDiskType        string            `json:"os_disk_type,omitempty"` // "Managed" or "Ephemeral"
	OSType            string            `json:"os_type,omitempty"`      // "Linux" or "Windows"
	OSSKU             string            `json:"os_sku,omitempty"`       // "Ubuntu" or "CBLMariner"
	NodeCount         int32             `json:"node_count" validate:"required,min=1"`
	MinCount          int32             `json:"min_count,omitempty"`
	MaxCount          int32             `json:"max_count,omitempty"`
	EnableAutoScaling bool              `json:"enable_auto_scaling,omitempty"`
	MaxPods           int32             `json:"max_pods,omitempty"`
	VnetSubnetID      string            `json:"vnet_subnet_id,omitempty"`
	AvailabilityZones []string          `json:"availability_zones,omitempty"`
	Labels            map[string]string `json:"labels,omitempty"`
	Taints            []string          `json:"taints,omitempty"`
	Mode              string            `json:"mode,omitempty"` // "System" or "User"
}

// AKSSecurityConfig represents AKS security configuration
type AKSSecurityConfig struct {
	EnableRBAC                  bool     `json:"enable_rbac,omitempty"`
	EnablePodSecurityPolicy     bool     `json:"enable_pod_security_policy,omitempty"`
	EnablePrivateCluster        bool     `json:"enable_private_cluster,omitempty"`
	APIServerAuthorizedIPRanges []string `json:"api_server_authorized_ip_ranges,omitempty"`
	EnableAzurePolicy           bool     `json:"enable_azure_policy,omitempty"`
	EnableWorkloadIdentity      bool     `json:"enable_workload_identity,omitempty"`
}

// CreateAKSClusterResponse represents the response after creating an AKS cluster
type CreateAKSClusterResponse struct {
	ClusterID     string             `json:"cluster_id"`
	Name          string             `json:"name"`
	Version       string             `json:"version"`
	Location      string             `json:"location"`
	ResourceGroup string             `json:"resource_group"`
	Status        string             `json:"status"`
	Endpoint      string             `json:"endpoint,omitempty"`
	Tags          map[string]*string `json:"tags,omitempty"`
	CreatedAt     string             `json:"created_at"`
}
