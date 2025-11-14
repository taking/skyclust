package providers

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// NCPHandler handles NCP NKS-related HTTP requests
type NCPHandler struct {
	*BaseHandler
}

// NewNCPHandler creates a new NCP NKS handler
func NewNCPHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *NCPHandler {
	return &NCPHandler{
		BaseHandler: NewBaseHandler(k8sService, credentialService, domain.ProviderNCP, "ncp-kubernetes"),
	}
}

// CreateCluster handles NKS cluster creation
func (h *NCPHandler) CreateCluster(c *gin.Context) {
	h.NotImplemented(c, "create_cluster")
}

// ListClusters handles listing NKS clusters
func (h *NCPHandler) ListClusters(c *gin.Context) {
	h.NotImplemented(c, "list_clusters")
}

// GetCluster handles getting NKS cluster details
func (h *NCPHandler) GetCluster(c *gin.Context) {
	h.NotImplemented(c, "get_cluster")
}

// DeleteCluster handles NKS cluster deletion
func (h *NCPHandler) DeleteCluster(c *gin.Context) {
	h.NotImplemented(c, "delete_cluster")
}

// GetKubeconfig handles getting kubeconfig for NKS cluster
func (h *NCPHandler) GetKubeconfig(c *gin.Context) {
	h.NotImplemented(c, "get_kubeconfig")
}

// CreateNodePool handles creating a node pool
func (h *NCPHandler) CreateNodePool(c *gin.Context) {
	h.NotImplemented(c, "create_node_pool")
}

// ListNodePools handles listing node pools
func (h *NCPHandler) ListNodePools(c *gin.Context) {
	h.NotImplemented(c, "list_node_pools")
}

// GetNodePool handles getting node pool details
func (h *NCPHandler) GetNodePool(c *gin.Context) {
	h.NotImplemented(c, "get_node_pool")
}

// DeleteNodePool handles deleting a node pool
func (h *NCPHandler) DeleteNodePool(c *gin.Context) {
	h.NotImplemented(c, "delete_node_pool")
}

// ScaleNodePool handles scaling a node pool
func (h *NCPHandler) ScaleNodePool(c *gin.Context) {
	h.NotImplemented(c, "scale_node_pool")
}

// CreateNodeGroup handles creating a node group
func (h *NCPHandler) CreateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "create_node_group")
}

// ListNodeGroups handles listing node groups
func (h *NCPHandler) ListNodeGroups(c *gin.Context) {
	h.NotImplemented(c, "list_node_groups")
}

// GetNodeGroup handles getting node group details
func (h *NCPHandler) GetNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "get_node_group")
}

// UpdateNodeGroup handles updating a node group
func (h *NCPHandler) UpdateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "update_node_group")
}

// DeleteNodeGroup handles deleting a node group
func (h *NCPHandler) DeleteNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_node_group")
}

// UpgradeCluster handles cluster upgrade
func (h *NCPHandler) UpgradeCluster(c *gin.Context) {
	h.NotImplemented(c, "upgrade_cluster")
}

// GetUpgradeStatus handles getting cluster upgrade status
func (h *NCPHandler) GetUpgradeStatus(c *gin.Context) {
	h.NotImplemented(c, "get_upgrade_status")
}

// ListNodes handles listing cluster nodes
func (h *NCPHandler) ListNodes(c *gin.Context) {
	h.NotImplemented(c, "list_nodes")
}

// GetNode handles getting node details
func (h *NCPHandler) GetNode(c *gin.Context) {
	h.NotImplemented(c, "get_node")
}

// DrainNode handles draining a node
func (h *NCPHandler) DrainNode(c *gin.Context) {
	h.NotImplemented(c, "drain_node")
}

// CordonNode handles cordoning a node
func (h *NCPHandler) CordonNode(c *gin.Context) {
	h.NotImplemented(c, "cordon_node")
}

// UncordonNode handles uncordoning a node
func (h *NCPHandler) UncordonNode(c *gin.Context) {
	h.NotImplemented(c, "uncordon_node")
}

// GetNodeSSHConfig handles getting SSH config for a node
func (h *NCPHandler) GetNodeSSHConfig(c *gin.Context) {
	h.NotImplemented(c, "get_node_ssh_config")
}

// ExecuteNodeCommand handles executing a command on a node
func (h *NCPHandler) ExecuteNodeCommand(c *gin.Context) {
	h.NotImplemented(c, "execute_node_command")
}

// GetEKSVersions handles EKS versions listing (AWS only)
func (h *NCPHandler) GetEKSVersions(c *gin.Context) {
	h.NotImplemented(c, "get_eks_versions")
}

// GetAWSRegions handles AWS regions listing (AWS only)
func (h *NCPHandler) GetAWSRegions(c *gin.Context) {
	h.NotImplemented(c, "get_aws_regions")
}

// GetAvailabilityZones handles availability zones listing (AWS only)
func (h *NCPHandler) GetAvailabilityZones(c *gin.Context) {
	h.NotImplemented(c, "get_availability_zones")
}

// GetInstanceTypes handles instance types listing (AWS only)
func (h *NCPHandler) GetInstanceTypes(c *gin.Context) {
	h.NotImplemented(c, "get_instance_types")
}

// GetEKSAmitTypes handles EKS AMI types listing (AWS only)
func (h *NCPHandler) GetEKSAmitTypes(c *gin.Context) {
	h.NotImplemented(c, "get_eks_ami_types")
}

// CheckGPUQuota handles GPU quota checking (AWS only)
func (h *NCPHandler) CheckGPUQuota(c *gin.Context) {
	h.NotImplemented(c, "check_gpu_quota")
}
