package kubernetes

// Kubernetes Service Constants
// These constants are specific to Kubernetes resource operations

// HTTP Status Code Constants
const (
	HTTPStatusBadRequest          = 400
	HTTPStatusUnauthorized        = 401
	HTTPStatusForbidden           = 403
	HTTPStatusNotFound            = 404
	HTTPStatusInternalServerError = 500
	HTTPStatusNotImplemented      = 501
	HTTPStatusBadGateway          = 502
	HTTPStatusServiceUnavailable  = 503
)

// Validation Constants
const (
	MinEKSSubnetAZs = 2 // AWS EKS requires subnets from at least 2 different availability zones
)

// Error message constants
const (
	// Credential errors
	ErrMsgFailedToDecryptCredential     = "failed to decrypt credential: %v"
	ErrMsgFailedToMarshalCredentialData = "failed to marshal credential data: %v"
	ErrMsgAccessKeyNotFound             = "access_key not found in credential"
	ErrMsgSecretKeyNotFound             = "secret_key not found in credential"
	ErrMsgProjectIDNotFound             = "project_id not found in credential"

	// Provider errors
	ErrMsgFailedToCreateGCPContainerService = "failed to create GCP container service: %v"
	ErrMsgFailedToCreateGCPGKECluster      = "failed to create GCP GKE cluster: %v"
	ErrMsgFailedToGetGCPGKECluster         = "failed to get GKE cluster: %v"
	ErrMsgFailedToDeleteGCPGKECluster      = "failed to delete GKE cluster: %v"
	ErrMsgFailedToListGCPGKEClusters       = "failed to list GKE clusters: %v"
	ErrMsgFailedToListGCPGKENodePools     = "failed to list GKE node pools for cluster %s in location %s: %v"
	ErrMsgFailedToFindGCPGKECluster        = "failed to find GKE cluster %s in region %s or any of its zones"
	ErrMsgFailedToFindGCPGKENodePool      = "failed to find GKE node pool %s in cluster %s in region %s or any of its zones"

	ErrMsgFailedToLoadAWSConfig      = "failed to load AWS config: %v"
	ErrMsgFailedToCreateEKSCluster   = "failed to create EKS cluster: %v"
	ErrMsgFailedToDescribeEKSCluster = "failed to describe EKS cluster: %v"
	ErrMsgFailedToDeleteEKSCluster   = "failed to delete EKS cluster: %v"
	ErrMsgFailedToListEKSClusters    = "failed to list EKS clusters: %v"
	ErrMsgFailedToCreateNodeGroup    = "failed to create node group: %v"
	ErrMsgFailedToDescribeNodeGroup  = "failed to describe node group: %v"
	ErrMsgFailedToDeleteNodeGroup    = "failed to delete node group: %v"
	ErrMsgFailedToDescribeSubnets     = "failed to describe subnets for validation: %v"

	// Validation errors
	ErrMsgUnsupportedProvider        = "unsupported provider: %s"
	ErrMsgNetworkConfigRequired       = "network configuration is required"
	ErrMsgNodePoolConfigRequired      = "node pool configuration is required for standard clusters"
	ErrMsgResourceGroupRequired       = "resource_group is required for Azure AKS cluster"
	ErrMsgAKSClusterNotFound          = "AKS cluster %s not found"

	// Not implemented errors
	ErrMsgNCPNotImplemented           = "NCP NKS %s not implemented yet"
	ErrMsgNCPNodeGroupsNotImplemented = "NCP node groups not implemented yet"
)

