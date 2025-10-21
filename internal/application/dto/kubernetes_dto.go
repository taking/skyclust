package dto

// CreateClusterRequest represents a request to create a Kubernetes cluster
type CreateClusterRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=100"`
	Version      string            `json:"version" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	SubnetIDs    []string          `json:"subnet_ids" validate:"required,min=1"`
	VPCID        string            `json:"vpc_id,omitempty"`
	RoleARN      string            `json:"role_arn,omitempty"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// CreateClusterResponse represents the response after creating a cluster
type CreateClusterResponse struct {
	ClusterID string            `json:"cluster_id"`
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Region    string            `json:"region"`
	Status    string            `json:"status"`
	Endpoint  string            `json:"endpoint,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	CreatedAt string            `json:"created_at"`
}

// ListClustersRequest represents a request to list clusters
type ListClustersRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Region       string `json:"region,omitempty"`
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
