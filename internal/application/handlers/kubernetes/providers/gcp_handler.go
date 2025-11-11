package providers

import (
	"net/http"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// GCPHandler handles GCP GKE-related HTTP requests
type GCPHandler struct {
	*BaseHandler
}

// NewGCPHandler creates a new GCP GKE handler
func NewGCPHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *GCPHandler {
	return &GCPHandler{
		BaseHandler: NewBaseHandler(k8sService, credentialService, domain.ProviderGCP, "gcp-kubernetes"),
	}
}

// CreateCluster handles GKE cluster creation
func (h *GCPHandler) CreateCluster(c *gin.Context) {
	var req kubernetesservice.CreateGKEClusterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	cluster, err := h.GetK8sService().CreateGCPGKECluster(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	h.Created(c, cluster, "GKE cluster creation initiated")
}

// ListClusters handles listing GKE clusters
func (h *GCPHandler) ListClusters(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "list_clusters")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "list_clusters")
		return
	}

	clusters, err := h.GetK8sService().ListEKSClusters(c.Request.Context(), credential, region)
	if err != nil {
		h.HandleError(c, err, "list_clusters")
		return
	}

	// 응답 구조 확인: clusters가 nil이 아닌지 확인하고, Clusters 필드도 확인
	if clusters == nil {
		clusters = &kubernetesservice.ListClustersResponse{Clusters: []kubernetesservice.ClusterInfo{}}
	}
	// Clusters 필드가 nil인 경우 빈 배열로 초기화
	if clusters.Clusters == nil {
		clusters.Clusters = []kubernetesservice.ClusterInfo{}
	}

	h.OK(c, clusters.Clusters, "GKE clusters retrieved successfully")
}

// GetCluster handles getting GKE cluster details
func (h *GCPHandler) GetCluster(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_cluster")
		return
	}

	clusterName := h.parseClusterName(c)
	region := h.parseRegion(c)

	if clusterName == "" || region == "" {
		return
	}

	cluster, err := h.GetK8sService().GetEKSCluster(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.HandleError(c, err, "get_cluster")
		return
	}

	h.OK(c, cluster, "GKE cluster retrieved successfully")
}

// DeleteCluster handles GKE cluster deletion
func (h *GCPHandler) DeleteCluster(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "delete_cluster")
		return
	}

	clusterName := h.parseClusterName(c)
	region := h.parseRegion(c)

	if clusterName == "" || region == "" {
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	if err := h.GetK8sService().DeleteEKSCluster(ctx, credential, clusterName, region); err != nil {
		h.HandleError(c, err, "delete_cluster")
		return
	}

	h.OK(c, nil, "GKE cluster deletion initiated")
}

// GetKubeconfig handles getting kubeconfig for GKE cluster
func (h *GCPHandler) GetKubeconfig(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_kubeconfig")
		return
	}

	clusterName := h.parseClusterName(c)
	region := h.parseRegion(c)

	if clusterName == "" || region == "" {
		return
	}

	kubeconfig, err := h.GetK8sService().GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.HandleError(c, err, "get_kubeconfig")
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
	c.String(http.StatusOK, kubeconfig)
}

// CreateNodePool handles creating a node pool for GKE
func (h *GCPHandler) CreateNodePool(c *gin.Context) {
	clusterName := h.parseClusterName(c)
	if clusterName == "" {
		return
	}

	var req kubernetesservice.CreateNodePoolRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_node_pool")
		return
	}
	req.ClusterName = clusterName

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "create_node_pool")
		return
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	nodePool, err := h.GetK8sService().CreateEKSNodePool(ctx, credential, req)
	if err != nil {
		h.HandleError(c, err, "create_node_pool")
		return
	}

	h.Created(c, nodePool, "GKE node pool creation initiated")
}

// ListNodePools handles listing node pools for a GKE cluster
func (h *GCPHandler) ListNodePools(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "list_node_pools")
		return
	}

	clusterName := h.parseClusterName(c)
	region := h.parseRegion(c)

	if clusterName == "" || region == "" {
		return
	}

	req := kubernetesservice.ListNodeGroupsRequest{
		CredentialID: credential.ID.String(),
		ClusterName:  clusterName,
		Region:       region,
	}

	nodePoolsResponse, err := h.GetK8sService().ListNodeGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "list_node_pools")
		return
	}

	h.OK(c, nodePoolsResponse.NodeGroups, "GKE node pools retrieved successfully")
}

// GetNodePool handles getting node pool details
func (h *GCPHandler) GetNodePool(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "get_node_pool")
		return
	}

	clusterName := h.parseClusterName(c)
	nodePoolName := c.Param("nodepool")
	region := h.parseRegion(c)

	if clusterName == "" || nodePoolName == "" || region == "" {
		return
	}

	req := kubernetesservice.GetNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        region,
	}

	nodePool, err := h.GetK8sService().GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "get_node_pool")
		return
	}

	h.OK(c, nodePool, "GKE node pool retrieved successfully")
}

// DeleteNodePool handles deleting a node pool
func (h *GCPHandler) DeleteNodePool(c *gin.Context) {
	clusterName := h.parseClusterName(c)
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		return
	}

	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "delete_node_pool")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "delete_node_pool")
		return
	}

	deleteReq := kubernetesservice.DeleteNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        req.Region,
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	if err := h.GetK8sService().DeleteNodeGroup(ctx, credential, deleteReq); err != nil {
		h.HandleError(c, err, "delete_node_pool")
		return
	}

	h.OK(c, nil, "GKE node pool deletion initiated")
}

// ScaleNodePool handles scaling a node pool
func (h *GCPHandler) ScaleNodePool(c *gin.Context) {
	clusterName := h.parseClusterName(c)
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		return
	}

	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
		DesiredSize  int32  `json:"desired_size" validate:"required,min=0"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "scale_node_pool")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderGCP)
	if err != nil {
		h.HandleError(c, err, "scale_node_pool")
		return
	}

	scaleReq := kubernetesservice.CreateNodePoolRequest{
		CredentialID: credential.ID.String(),
		ClusterName:  clusterName,
		NodePoolName: nodePoolName,
		Region:       req.Region,
		DesiredSize:  req.DesiredSize,
	}

	_, err = h.GetK8sService().CreateEKSNodePool(c.Request.Context(), credential, scaleReq)
	if err != nil {
		h.HandleError(c, err, "scale_node_pool")
		return
	}

	h.OK(c, gin.H{"desired_size": req.DesiredSize}, "GKE node pool scaling initiated")
}

// CreateNodeGroup handles creating a node group (not used for GCP, but required by interface)
func (h *GCPHandler) CreateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "create_node_group")
}

// ListNodeGroups handles listing node groups (not used for GCP, but required by interface)
func (h *GCPHandler) ListNodeGroups(c *gin.Context) {
	h.NotImplemented(c, "list_node_groups")
}

// GetNodeGroup handles getting node group details (not used for GCP, but required by interface)
func (h *GCPHandler) GetNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "get_node_group")
}

// DeleteNodeGroup handles deleting a node group (not used for GCP, but required by interface)
func (h *GCPHandler) DeleteNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_node_group")
}

// UpgradeCluster handles GKE cluster upgrade
func (h *GCPHandler) UpgradeCluster(c *gin.Context) {
	h.NotImplemented(c, "upgrade_cluster")
}

// GetUpgradeStatus handles getting GKE cluster upgrade status
func (h *GCPHandler) GetUpgradeStatus(c *gin.Context) {
	h.NotImplemented(c, "get_upgrade_status")
}

// ListNodes handles listing GKE nodes
func (h *GCPHandler) ListNodes(c *gin.Context) {
	h.NotImplemented(c, "list_nodes")
}

// GetNode handles getting GKE node details
func (h *GCPHandler) GetNode(c *gin.Context) {
	h.NotImplemented(c, "get_node")
}

// DrainNode handles draining a GKE node
func (h *GCPHandler) DrainNode(c *gin.Context) {
	h.NotImplemented(c, "drain_node")
}

// CordonNode handles cordoning a GKE node
func (h *GCPHandler) CordonNode(c *gin.Context) {
	h.NotImplemented(c, "cordon_node")
}

// UncordonNode handles uncordoning a GKE node
func (h *GCPHandler) UncordonNode(c *gin.Context) {
	h.NotImplemented(c, "uncordon_node")
}

// GetNodeSSHConfig handles getting SSH config for GKE node
func (h *GCPHandler) GetNodeSSHConfig(c *gin.Context) {
	h.NotImplemented(c, "get_node_ssh_config")
}

// ExecuteNodeCommand handles executing commands on GKE node
func (h *GCPHandler) ExecuteNodeCommand(c *gin.Context) {
	h.NotImplemented(c, "execute_node_command")
}

// GetEKSVersions handles EKS versions listing (AWS only)
func (h *GCPHandler) GetEKSVersions(c *gin.Context) {
	h.NotImplemented(c, "get_eks_versions")
}

// GetAWSRegions handles AWS regions listing (AWS only)
func (h *GCPHandler) GetAWSRegions(c *gin.Context) {
	h.NotImplemented(c, "get_aws_regions")
}

// GetAvailabilityZones handles availability zones listing (AWS only)
func (h *GCPHandler) GetAvailabilityZones(c *gin.Context) {
	h.NotImplemented(c, "get_availability_zones")
}
