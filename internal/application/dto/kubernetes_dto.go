package dto

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
	// GCP-specific configuration (for backward compatibility)
	GKEConfig *GKEClusterConfigRequest `json:"gke_config,omitempty"`
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
	Endpoint  string            `json:"endpoint,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
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
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Version       string                 `json:"version"`
	Status        string                 `json:"status"`
	ClusterName   string                 `json:"cluster_name"`
	Region        string                 `json:"region"`
	InstanceTypes []string               `json:"instance_types"`
	ScalingConfig NodeGroupScalingConfig `json:"scaling_config"`
	CapacityType  string                 `json:"capacity_type"`
	DiskSize      int32                  `json:"disk_size"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	UpdatedAt     string                 `json:"updated_at,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
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
