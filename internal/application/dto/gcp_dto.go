package dto

// GCP-specific DTOs for GKE and Network resources

// GCP GKE Cluster DTOs - 섹션별 그룹핑 구조

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

// Legacy DTOs for backward compatibility (deprecated)
// TODO: Remove these after migration to new structure

// CreateGKEClusterRequestLegacy represents the legacy request structure
type CreateGKEClusterRequestLegacy struct {
	CredentialID string                   `json:"credential_id" validate:"required,uuid"`
	Name         string                   `json:"name" validate:"required,min=1,max=100"`
	Version      string                   `json:"version" validate:"required"`
	Region       string                   `json:"region" validate:"required"`
	Zone         string                   `json:"zone,omitempty"`
	SubnetIDs    []string                 `json:"subnet_ids" validate:"required,min=1"`
	VPCID        string                   `json:"vpc_id,omitempty"`
	ProjectID    string                   `json:"project_id" validate:"required"`
	Tags         map[string]string        `json:"tags,omitempty"`
	GKEConfig    *GKEClusterConfigRequest `json:"gke_config,omitempty"`
}

// GKEClusterConfigRequest represents legacy GKE-specific cluster configuration
type GKEClusterConfigRequest struct {
	ClusterType     string                     `json:"cluster_type,omitempty"`
	NetworkConfig   *GKENetworkConfigRequest   `json:"network_config,omitempty"`
	NodePoolConfig  *GKENodePoolConfigRequest  `json:"node_pool_config,omitempty"`
	AutopilotConfig *GKEAutopilotConfigRequest `json:"autopilot_config,omitempty"`
	SecurityConfig  *GKESecurityConfigRequest  `json:"security_config,omitempty"`
}

// GKENetworkConfigRequest represents legacy GKE network configuration
type GKENetworkConfigRequest struct {
	NetworkName              string   `json:"network_name,omitempty"`
	SubnetName               string   `json:"subnet_name,omitempty"`
	PrivateNodes             bool     `json:"private_nodes,omitempty"`
	PrivateEndpoint          bool     `json:"private_endpoint,omitempty"`
	MasterAuthorizedNetworks []string `json:"master_authorized_networks,omitempty"`
	PodCIDR                  string   `json:"pod_cidr,omitempty"`
	ServiceCIDR              string   `json:"service_cidr,omitempty"`
}

// GKENodePoolConfigRequest represents legacy GKE node pool configuration
type GKENodePoolConfigRequest struct {
	Name        string                `json:"name" validate:"required"`
	MachineType string                `json:"machine_type" validate:"required"`
	DiskSizeGB  int32                 `json:"disk_size_gb,omitempty"`
	DiskType    string                `json:"disk_type,omitempty"`
	NodeCount   int32                 `json:"node_count" validate:"required,min=1"`
	AutoScaling *GKEAutoScalingConfig `json:"auto_scaling,omitempty"`
	Labels      map[string]string     `json:"labels,omitempty"`
	Taints      []string              `json:"taints,omitempty"`
	Preemptible bool                  `json:"preemptible,omitempty"`
	Spot        bool                  `json:"spot,omitempty"`
}

// GKEAutopilotConfigRequest represents legacy GKE Autopilot configuration
type GKEAutopilotConfigRequest struct {
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// GKESecurityConfigRequest represents legacy GKE security configuration
type GKESecurityConfigRequest struct {
	WorkloadIdentity         bool     `json:"workload_identity,omitempty"`
	BinaryAuthorization      bool     `json:"binary_authorization,omitempty"`
	NetworkPolicy            bool     `json:"network_policy,omitempty"`
	PodSecurityPolicy        bool     `json:"pod_security_policy,omitempty"`
	MasterAuthorizedNetworks []string `json:"master_authorized_networks,omitempty"`
}

// GCP Network DTOs

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

// GCP Network List DTOs

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

// GCP Network CRUD DTOs

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

// GCP Security Group Rule Management DTOs

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
