package providers

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// AzureHandler handles Azure AKS-related HTTP requests
type AzureHandler struct {
	*BaseHandler
}

// NewAzureHandler creates a new Azure AKS handler
func NewAzureHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *AzureHandler {
	return &AzureHandler{
		BaseHandler: NewBaseHandler(k8sService, credentialService, domain.ProviderAzure, "azure-kubernetes"),
	}
}

// CreateCluster handles AKS cluster creation
func (h *AzureHandler) CreateCluster(c *gin.Context) {
	h.NotImplemented(c, "create_cluster")
}

// ListClusters handles listing AKS clusters
func (h *AzureHandler) ListClusters(c *gin.Context) {
	h.NotImplemented(c, "list_clusters")
}

// GetCluster handles getting AKS cluster details
func (h *AzureHandler) GetCluster(c *gin.Context) {
	h.NotImplemented(c, "get_cluster")
}

// DeleteCluster handles AKS cluster deletion
func (h *AzureHandler) DeleteCluster(c *gin.Context) {
	h.NotImplemented(c, "delete_cluster")
}

// GetKubeconfig handles getting kubeconfig for AKS cluster
func (h *AzureHandler) GetKubeconfig(c *gin.Context) {
	h.NotImplemented(c, "get_kubeconfig")
}

// CreateNodePool handles creating a node pool
func (h *AzureHandler) CreateNodePool(c *gin.Context) {
	h.NotImplemented(c, "create_node_pool")
}

// ListNodePools handles listing node pools
func (h *AzureHandler) ListNodePools(c *gin.Context) {
	h.NotImplemented(c, "list_node_pools")
}

// GetNodePool handles getting node pool details
func (h *AzureHandler) GetNodePool(c *gin.Context) {
	h.NotImplemented(c, "get_node_pool")
}

// DeleteNodePool handles deleting a node pool
func (h *AzureHandler) DeleteNodePool(c *gin.Context) {
	h.NotImplemented(c, "delete_node_pool")
}

// ScaleNodePool handles scaling a node pool
func (h *AzureHandler) ScaleNodePool(c *gin.Context) {
	h.NotImplemented(c, "scale_node_pool")
}

// CreateNodeGroup handles creating a node group
func (h *AzureHandler) CreateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "create_node_group")
}

// ListNodeGroups handles listing node groups
func (h *AzureHandler) ListNodeGroups(c *gin.Context) {
	h.NotImplemented(c, "list_node_groups")
}

// GetNodeGroup handles getting node group details
func (h *AzureHandler) GetNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "get_node_group")
}

// DeleteNodeGroup handles deleting a node group
func (h *AzureHandler) DeleteNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_node_group")
}

// UpgradeCluster handles cluster upgrade
func (h *AzureHandler) UpgradeCluster(c *gin.Context) {
	h.NotImplemented(c, "upgrade_cluster")
}

// GetUpgradeStatus handles getting cluster upgrade status
func (h *AzureHandler) GetUpgradeStatus(c *gin.Context) {
	h.NotImplemented(c, "get_upgrade_status")
}

// ListNodes handles listing cluster nodes
func (h *AzureHandler) ListNodes(c *gin.Context) {
	h.NotImplemented(c, "list_nodes")
}

// GetNode handles getting node details
func (h *AzureHandler) GetNode(c *gin.Context) {
	h.NotImplemented(c, "get_node")
}

// DrainNode handles draining a node
func (h *AzureHandler) DrainNode(c *gin.Context) {
	h.NotImplemented(c, "drain_node")
}

// CordonNode handles cordoning a node
func (h *AzureHandler) CordonNode(c *gin.Context) {
	h.NotImplemented(c, "cordon_node")
}

// UncordonNode handles uncordoning a node
func (h *AzureHandler) UncordonNode(c *gin.Context) {
	h.NotImplemented(c, "uncordon_node")
}

// GetNodeSSHConfig handles getting SSH config for a node
func (h *AzureHandler) GetNodeSSHConfig(c *gin.Context) {
	h.NotImplemented(c, "get_node_ssh_config")
}

// ExecuteNodeCommand handles executing a command on a node
func (h *AzureHandler) ExecuteNodeCommand(c *gin.Context) {
	h.NotImplemented(c, "execute_node_command")
}

// GetEKSVersions handles EKS versions listing (AWS only)
func (h *AzureHandler) GetEKSVersions(c *gin.Context) {
	h.NotImplemented(c, "get_eks_versions")
}

// GetAWSRegions handles AWS regions listing (AWS only)
func (h *AzureHandler) GetAWSRegions(c *gin.Context) {
	h.NotImplemented(c, "get_aws_regions")
}

// GetAvailabilityZones handles availability zones listing (AWS only)
func (h *AzureHandler) GetAvailabilityZones(c *gin.Context) {
	h.NotImplemented(c, "get_availability_zones")
}

