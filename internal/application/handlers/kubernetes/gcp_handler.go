package kubernetes

import (
	"net/http"

	"skyclust/internal/application/dto"
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
)

// GCPHandler handles GCP GKE-related HTTP requests
type GCPHandler struct {
	*handlers.BaseHandler
	k8sService        *service.KubernetesService
	credentialService domain.CredentialService
}

// NewGCPHandler creates a new GCP GKE handler
func NewGCPHandler(k8sService *service.KubernetesService, credentialService domain.CredentialService) *GCPHandler {
	return &GCPHandler{
		BaseHandler:       handlers.NewBaseHandler("gcp-kubernetes"),
		k8sService:        k8sService,
		credentialService: credentialService,
	}
}

// CreateGKECluster handles GKE cluster creation
func (h *GCPHandler) CreateGKECluster(c *gin.Context) {
	// Parse request
	var req dto.CreateGKEClusterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_gke_cluster")
		return
	}

	// Get and validate credential from request body
	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, "gcp")
	if err != nil {
		h.HandleError(c, err, "create_gke_cluster")
		return
	}

	// Create GKE cluster with new structure
	cluster, err := h.k8sService.CreateGCPGKECluster(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_gke_cluster")
		return
	}

	h.Created(c, cluster, "GKE cluster creation initiated")
}

// ListGKEClusters handles listing GKE clusters
func (h *GCPHandler) ListGKEClusters(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "list_gke_clusters")
		return
	}

	region := c.Query("region")

	// List GKE clusters
	clusters, err := h.k8sService.ListEKSClusters(c.Request.Context(), credential, region)
	if err != nil {
		h.HandleError(c, err, "list_gke_clusters")
		return
	}

	h.OK(c, clusters, "GKE clusters retrieved successfully")
}

// GetGKECluster handles getting GKE cluster details
func (h *GCPHandler) GetGKECluster(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_gke_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_gke_cluster")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_gke_cluster")
		return
	}

	// Get GKE cluster details
	cluster, err := h.k8sService.GetEKSCluster(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.HandleError(c, err, "get_gke_cluster")
		return
	}

	h.OK(c, cluster, "GKE cluster retrieved successfully")
}

// DeleteGKECluster handles GKE cluster deletion
func (h *GCPHandler) DeleteGKECluster(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "delete_gke_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "delete_gke_cluster")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "delete_gke_cluster")
		return
	}

	// Delete GKE cluster
	if err := h.k8sService.DeleteEKSCluster(c.Request.Context(), credential, clusterName, region); err != nil {
		h.HandleError(c, err, "delete_gke_cluster")
		return
	}

	h.OK(c, nil, "GKE cluster deletion initiated")
}

// GetGKEKubeconfig handles getting kubeconfig for GKE cluster
func (h *GCPHandler) GetGKEKubeconfig(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_gke_kubeconfig")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_gke_kubeconfig")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_gke_kubeconfig")
		return
	}

	// Get kubeconfig
	kubeconfig, err := h.k8sService.GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.HandleError(c, err, "get_gke_kubeconfig")
		return
	}

	// Return as downloadable file
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
	c.String(http.StatusOK, kubeconfig)
}

// CreateGKENodePool handles creating a node pool for GKE
func (h *GCPHandler) CreateGKENodePool(c *gin.Context) {
	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "create_gke_node_pool")
		return
	}

	// Parse request
	var req dto.CreateNodePoolRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_gke_node_pool")
		return
	}
	req.ClusterName = clusterName

	// Get and validate credential from request body
	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, "gcp")
	if err != nil {
		h.HandleError(c, err, "create_gke_node_pool")
		return
	}

	// Create node pool
	nodePool, err := h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "create_gke_node_pool")
		return
	}

	h.Created(c, nodePool, "GKE node pool creation initiated")
}

// ListGKENodePools handles listing node pools for a GKE cluster
func (h *GCPHandler) ListGKENodePools(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "list_gke_node_pools")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "list_gke_node_pools")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "list_gke_node_pools")
		return
	}

	// List node pools
	req := dto.ListNodeGroupsRequest{
		CredentialID: credential.ID.String(),
		ClusterName:  clusterName,
		Region:       region,
	}

	nodePoolsResponse, err := h.k8sService.ListNodeGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "list_gke_node_pools")
		return
	}

	h.OK(c, nodePoolsResponse, "GKE node pools retrieved successfully")
}

// GetGKENodePool handles getting node pool details
func (h *GCPHandler) GetGKENodePool(c *gin.Context) {
	// Get and validate credential using BaseHandler helper
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, "gcp")
	if err != nil {
		h.HandleError(c, err, "get_gke_node_pool")
		return
	}

	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node pool name are required", 400), "get_gke_node_pool")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_gke_node_pool")
		return
	}

	// Get node pool
	req := dto.GetNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        region,
	}

	nodePool, err := h.k8sService.GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "get_gke_node_pool")
		return
	}

	h.OK(c, nodePool, "GKE node pool retrieved successfully")
}

// DeleteGKENodePool handles deleting a node pool
func (h *GCPHandler) DeleteGKENodePool(c *gin.Context) {
	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node pool name are required", 400), "delete_gke_node_pool")
		return
	}

	// Parse request
	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "delete_gke_node_pool")
		return
	}

	// Get and validate credential from request body
	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, "gcp")
	if err != nil {
		h.HandleError(c, err, "delete_gke_node_pool")
		return
	}

	// Delete node pool
	deleteReq := dto.DeleteNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        req.Region,
	}

	if err := h.k8sService.DeleteNodeGroup(c.Request.Context(), credential, deleteReq); err != nil {
		h.HandleError(c, err, "delete_gke_node_pool")
		return
	}

	h.OK(c, nil, "GKE node pool deletion initiated")
}

// ScaleGKENodePool handles scaling a node pool
func (h *GCPHandler) ScaleGKENodePool(c *gin.Context) {
	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node pool name are required", 400), "scale_gke_node_pool")
		return
	}

	// Parse request
	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
		DesiredSize  int32  `json:"desired_size" validate:"required,min=0"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "scale_gke_node_pool")
		return
	}

	// Get and validate credential from request body
	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, "gcp")
	if err != nil {
		h.HandleError(c, err, "scale_gke_node_pool")
		return
	}

	// Scale node pool
	scaleReq := dto.CreateNodePoolRequest{
		CredentialID: credential.ID.String(),
		ClusterName:  clusterName,
		NodePoolName: nodePoolName,
		Region:       req.Region,
		DesiredSize:  req.DesiredSize,
	}

	_, err = h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, scaleReq)
	if err != nil {
		h.HandleError(c, err, "scale_gke_node_pool")
		return
	}

	h.OK(c, gin.H{"desired_size": req.DesiredSize}, "GKE node pool scaling initiated")
}

// UpgradeGKECluster handles GKE cluster upgrade
func (h *GCPHandler) UpgradeGKECluster(c *gin.Context) {
	h.InternalServerError(c, "GKE cluster upgrade not yet implemented")
}

// GetGKEUpgradeStatus handles getting GKE cluster upgrade status
func (h *GCPHandler) GetGKEUpgradeStatus(c *gin.Context) {
	h.InternalServerError(c, "GKE cluster upgrade status not yet implemented")
}

// ListGKENodes handles listing GKE nodes
func (h *GCPHandler) ListGKENodes(c *gin.Context) {
	h.InternalServerError(c, "GKE nodes listing not yet implemented")
}

// GetGKENode handles getting GKE node details
func (h *GCPHandler) GetGKENode(c *gin.Context) {
	h.InternalServerError(c, "GKE node details not yet implemented")
}

// DrainGKENode handles draining a GKE node
func (h *GCPHandler) DrainGKENode(c *gin.Context) {
	h.InternalServerError(c, "GKE node draining not yet implemented")
}

// CordonGKENode handles cordoning a GKE node
func (h *GCPHandler) CordonGKENode(c *gin.Context) {
	h.InternalServerError(c, "GKE node cordoning not yet implemented")
}

// UncordonGKENode handles uncordoning a GKE node
func (h *GCPHandler) UncordonGKENode(c *gin.Context) {
	h.InternalServerError(c, "GKE node uncordoning not yet implemented")
}

// GetGKENodeSSHConfig handles getting SSH config for GKE node
func (h *GCPHandler) GetGKENodeSSHConfig(c *gin.Context) {
	h.InternalServerError(c, "GKE node SSH config not yet implemented")
}

// ExecuteGKENodeCommand handles executing commands on GKE node
func (h *GCPHandler) ExecuteGKENodeCommand(c *gin.Context) {
	h.InternalServerError(c, "GKE node command execution not yet implemented")
}
