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
	RoleARN      string            `json:"role_arn,omitempty"`
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
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Region       string `json:"region,omitempty"`
}

// ClusterInfo represents basic cluster information for listing
type ClusterInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Status    string            `json:"status"`
	Region    string            `json:"region"`
	Zone      string            `json:"zone,omitempty"`
	Endpoint  string            `json:"endpoint,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`

	// Network information
	NetworkConfig *NetworkConfigInfo `json:"network_config,omitempty"`

	// Node pool information
	NodePoolInfo *NodePoolSummaryInfo `json:"node_pool_info,omitempty"`

	// Security configuration
	SecurityConfig *SecurityConfigInfo `json:"security_config,omitempty"`
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
type ListClustersResponse struct {
	Clusters []ClusterInfo `json:"clusters"`
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
	CredentialID  string                 `json:"credential_id" validate:"required,uuid"`
	ClusterName   string                 `json:"cluster_name" validate:"required"`
	Region        string                 `json:"region" validate:"required"`
	NodeGroupName string                 `json:"node_group_name" validate:"required"`
	NodeRoleARN   string                 `json:"node_role_arn" validate:"required"`
	SubnetIDs     []string               `json:"subnet_ids" validate:"required,min=1"`
	InstanceTypes []string               `json:"instance_types" validate:"required,min=1"`
	ScalingConfig NodeGroupScalingConfig `json:"scaling_config" validate:"required"`
	DiskSize      int32                  `json:"disk_size,omitempty"`
	AMI           string                 `json:"ami,omitempty"`
	CapacityType  string                 `json:"capacity_type,omitempty"` // ON_DEMAND, SPOT
	Tags          map[string]string      `json:"tags,omitempty"`
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


