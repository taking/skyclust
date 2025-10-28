package kubernetes

import (
	"net/http"

	"skyclust/internal/application/dto"
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles Kubernetes-related HTTP requests
type Handler struct {
	*handlers.BaseHandler
	k8sService        *service.KubernetesService
	credentialService domain.CredentialService
	provider          string // "aws", "gcp", "azure", "ncp"
}

// NewHandler creates a new Kubernetes handler for a specific provider
func NewHandler(k8sService *service.KubernetesService, credentialService domain.CredentialService, provider string) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("kubernetes"),
		k8sService:        k8sService,
		credentialService: credentialService,
		provider:          provider,
	}
}

// CreateCluster handles EKS cluster creation
func (h *Handler) CreateCluster(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	// Parse request
	var req dto.CreateClusterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Parse credential ID
	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Verify credential matches the provider
	if credential.Provider != h.provider {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Credential provider does not match the requested provider", 400), "create_cluster")
		return
	}

	// Create cluster
	cluster, err := h.k8sService.CreateEKSCluster(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create EKS cluster")
		h.HandleError(c, err, "create_cluster")
		return
	}

	h.Created(c, cluster, "Cluster creation initiated")
}

// ListClusters handles listing EKS clusters
func (h *Handler) ListClusters(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	// Get credential ID from query param
	credentialIDStr := c.Query("credential_id")
	if credentialIDStr == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id is required", 400), "list_clusters")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Verify credential matches the provider
	if credential.Provider != h.provider {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Credential provider does not match the requested provider", 400), "create_cluster")
		return
	}

	region := c.Query("region")

	// List clusters
	clusters, err := h.k8sService.ListEKSClusters(c.Request.Context(), credential, region)
	if err != nil {
		h.LogError(c, err, "Failed to list clusters")
		h.HandleError(c, err, "list_clusters")
		return
	}

	h.OK(c, clusters, "Clusters retrieved successfully")
}

// GetCluster handles getting EKS cluster details
func (h *Handler) GetCluster(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id and region are required", 400), "get_cluster")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Get cluster details
	cluster, err := h.k8sService.GetEKSCluster(c.Request.Context(), credential, clusterName, region)

	if err != nil {
		h.LogError(c, err, "Failed to get EKS cluster")
		h.HandleError(c, err, "get_cluster")
		return
	}

	h.OK(c, cluster, "Cluster retrieved successfully")
}

// DeleteCluster handles EKS cluster deletion
func (h *Handler) DeleteCluster(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id and region are required", 400), "get_cluster")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Delete cluster
	err = h.k8sService.DeleteEKSCluster(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.LogError(c, err, "Failed to delete EKS cluster")
		h.HandleError(c, err, "delete_cluster")
		return
	}

	h.OK(c, nil, "Cluster deletion initiated")
}

// GetKubeconfig handles getting kubeconfig for EKS cluster
func (h *Handler) GetKubeconfig(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id and region are required", 400), "get_cluster")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Get kubeconfig
	kubeconfig, err := h.k8sService.GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.LogError(c, err, "Failed to get kubeconfig")
		h.HandleError(c, err, "get_kubeconfig")
		return
	}

	// Return as downloadable file
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
	c.String(http.StatusOK, kubeconfig)
}

// CreateNodePool handles creating a node pool (node group) for EKS
func (h *Handler) CreateNodePool(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Parse request
	var req dto.CreateNodePoolRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	req.ClusterName = clusterName

	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Create node pool
	nodePool, err := h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create node pool")
		h.HandleError(c, err, "create_node_pool")
		return
	}

	h.Created(c, nodePool, "Node pool creation initiated")
}

// CreateNodeGroup handles creating an EKS node group
func (h *Handler) CreateNodeGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Parse request
	var req dto.CreateNodeGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}
	req.ClusterName = clusterName

	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Verify credential matches the provider
	if credential.Provider != h.provider {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Credential provider does not match the requested provider", 400), "create_cluster")
		return
	}

	// Create node group
	nodeGroup, err := h.k8sService.CreateEKSNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create node group")
		h.HandleError(c, err, "create_node_group")
		return
	}

	h.Created(c, nodeGroup, "Node group creation initiated")
}

// ListNodeGroups handles listing node groups for a cluster
func (h *Handler) ListNodeGroups(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "get_cluster")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id and region are required", 400), "get_cluster")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Verify credential matches the provider
	h.LogInfo(c, "Checking credential provider",
		zap.String("credential_provider", credential.Provider),
		zap.String("expected_provider", h.provider))

	if credential.Provider != h.provider {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Credential provider does not match the requested provider", 400), "create_cluster")
		return
	}

	// List node groups
	req := dto.ListNodeGroupsRequest{
		CredentialID: credentialIDStr,
		ClusterName:  clusterName,
		Region:       region,
	}

	nodeGroupsResponse, err := h.k8sService.ListNodeGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list node groups")
		h.HandleError(c, err, "list_node_groups")
		return
	}

	h.OK(c, nodeGroupsResponse, "Node groups retrieved successfully")
}

// GetNodeGroup handles getting node group details
func (h *Handler) GetNodeGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "get_node_group")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "credential_id and region are required", 400), "get_cluster")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Get node group
	req := dto.GetNodeGroupRequest{
		CredentialID:  credentialIDStr,
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        region,
	}

	nodeGroup, err := h.k8sService.GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get node group")
		h.HandleError(c, err, "get_node_group")
		return
	}

	h.OK(c, nodeGroup, "Node group retrieved successfully")
}

// DeleteNodeGroup handles deleting a node group
func (h *Handler) DeleteNodeGroup(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "get_node_group")
		return
	}

	// Parse request
	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid credential ID", 400), "create_cluster")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeNotFound, "Credential not found", 404), "create_cluster")
		return
	}

	// Delete node group
	deleteReq := dto.DeleteNodeGroupRequest{
		CredentialID:  req.CredentialID,
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        req.Region,
	}

	err = h.k8sService.DeleteNodeGroup(c.Request.Context(), credential, deleteReq)
	if err != nil {
		h.LogError(c, err, "Failed to delete node group")
		h.HandleError(c, err, "delete_node_group")
		return
	}

	h.OK(c, nil, "Node group deletion initiated")
}
