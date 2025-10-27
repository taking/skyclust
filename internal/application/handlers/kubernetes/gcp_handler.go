package kubernetes

import (
	"net/http"

	"skyclust/internal/application/dto"
	service "skyclust/internal/application/services"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/responses"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Parse request with new sectioned structure
	var req dto.CreateGKEClusterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	// Parse credential ID
	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create GKE cluster with new structure
	cluster, err := h.k8sService.CreateGCPGKECluster(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create GKE cluster")
		responses.InternalServerError(c, "Failed to create cluster: "+err.Error())
		return
	}

	responses.Created(c, cluster, "GKE cluster creation initiated")
}

// ListGKEClusters handles listing GKE clusters
func (h *GCPHandler) ListGKEClusters(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	// Get credential ID from query param
	credentialIDStr := c.Query("credential_id")
	if credentialIDStr == "" {
		responses.BadRequest(c, "credential_id is required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	region := c.Query("region")

	// List GKE clusters
	clusters, err := h.k8sService.ListEKSClusters(c.Request.Context(), credential, region)
	if err != nil {
		h.LogError(c, err, "Failed to list GKE clusters")
		responses.InternalServerError(c, "Failed to list clusters: "+err.Error())
		return
	}

	responses.OK(c, clusters, "GKE clusters retrieved successfully")
}

// GetGKECluster handles getting GKE cluster details
func (h *GCPHandler) GetGKECluster(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		responses.BadRequest(c, "cluster name is required")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		responses.BadRequest(c, "credential_id and region are required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Get GKE cluster details
	cluster, err := h.k8sService.GetEKSCluster(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.LogError(c, err, "Failed to get GKE cluster")
		responses.InternalServerError(c, "Failed to get cluster: "+err.Error())
		return
	}

	responses.OK(c, cluster, "GKE cluster retrieved successfully")
}

// DeleteGKECluster handles GKE cluster deletion
func (h *GCPHandler) DeleteGKECluster(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		responses.BadRequest(c, "cluster name is required")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		responses.BadRequest(c, "credential_id and region are required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Delete GKE cluster
	err = h.k8sService.DeleteEKSCluster(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.LogError(c, err, "Failed to delete GKE cluster")
		responses.InternalServerError(c, "Failed to delete cluster: "+err.Error())
		return
	}

	responses.OK(c, nil, "GKE cluster deletion initiated")
}

// GetGKEKubeconfig handles getting kubeconfig for GKE cluster
func (h *GCPHandler) GetGKEKubeconfig(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		responses.BadRequest(c, "cluster name is required")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		responses.BadRequest(c, "credential_id and region are required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Get kubeconfig
	kubeconfig, err := h.k8sService.GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
	if err != nil {
		h.LogError(c, err, "Failed to get kubeconfig")
		responses.InternalServerError(c, "Failed to get kubeconfig: "+err.Error())
		return
	}

	// Return as downloadable file
	c.Header("Content-Type", "application/x-yaml")
	c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
	c.String(http.StatusOK, kubeconfig)
}

// CreateGKENodePool handles creating a node pool for GKE
func (h *GCPHandler) CreateGKENodePool(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		responses.BadRequest(c, "cluster name is required")
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
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Create node pool
	nodePool, err := h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to create node pool")
		responses.InternalServerError(c, "Failed to create node pool: "+err.Error())
		return
	}

	responses.Created(c, nodePool, "GKE node pool creation initiated")
}

// ListGKENodePools handles listing node pools for a GKE cluster
func (h *GCPHandler) ListGKENodePools(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	if clusterName == "" {
		responses.BadRequest(c, "cluster name is required")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		responses.BadRequest(c, "credential_id and region are required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// List node pools
	req := dto.ListNodeGroupsRequest{
		CredentialID: credentialIDStr,
		ClusterName:  clusterName,
		Region:       region,
	}

	nodePoolsResponse, err := h.k8sService.ListNodeGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to list node pools")
		responses.InternalServerError(c, "Failed to list node pools: "+err.Error())
		return
	}

	responses.OK(c, nodePoolsResponse, "GKE node pools retrieved successfully")
}

// GetGKENodePool handles getting node pool details
func (h *GCPHandler) GetGKENodePool(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		responses.BadRequest(c, "cluster name and node pool name are required")
		return
	}

	// Get credential ID and region from query
	credentialIDStr := c.Query("credential_id")
	region := c.Query("region")

	if credentialIDStr == "" || region == "" {
		responses.BadRequest(c, "credential_id and region are required")
		return
	}

	credentialID, err := uuid.Parse(credentialIDStr)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Verify credential matches GCP provider
	if credential.Provider != "gcp" {
		responses.BadRequest(c, "Credential provider does not match GCP")
		return
	}

	// Get node pool
	req := dto.GetNodeGroupRequest{
		CredentialID:  credentialIDStr,
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        region,
	}

	nodePool, err := h.k8sService.GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.LogError(c, err, "Failed to get node pool")
		responses.InternalServerError(c, "Failed to get node pool: "+err.Error())
		return
	}

	responses.OK(c, nodePool, "GKE node pool retrieved successfully")
}

// DeleteGKENodePool handles deleting a node pool
func (h *GCPHandler) DeleteGKENodePool(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		responses.BadRequest(c, "cluster name and node pool name are required")
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
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Delete node pool
	deleteReq := dto.DeleteNodeGroupRequest{
		CredentialID:  req.CredentialID,
		ClusterName:   clusterName,
		NodeGroupName: nodePoolName,
		Region:        req.Region,
	}

	err = h.k8sService.DeleteNodeGroup(c.Request.Context(), credential, deleteReq)
	if err != nil {
		h.LogError(c, err, "Failed to delete node pool")
		responses.InternalServerError(c, "Failed to delete node pool: "+err.Error())
		return
	}

	responses.OK(c, nil, "GKE node pool deletion initiated")
}

// ScaleGKENodePool handles scaling a node pool
func (h *GCPHandler) ScaleGKENodePool(c *gin.Context) {
	// Get user ID from token
	userID, err := h.GetUserIDFromToken(c)
	if err != nil {
		responses.Unauthorized(c, "Invalid token")
		return
	}

	clusterName := c.Param("name")
	nodePoolName := c.Param("nodepool")

	if clusterName == "" || nodePoolName == "" {
		responses.BadRequest(c, "cluster name and node pool name are required")
		return
	}

	// Parse request
	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
		DesiredSize  int32  `json:"desired_size" validate:"required,min=0"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		return
	}

	credentialID, err := uuid.Parse(req.CredentialID)
	if err != nil {
		responses.BadRequest(c, "Invalid credential ID")
		return
	}

	// Verify credential belongs to user
	credential, err := h.credentialService.GetCredentialByID(c.Request.Context(), userID, credentialID)
	if err != nil {
		responses.NotFound(c, "Credential not found")
		return
	}

	// Scale node pool
	scaleReq := dto.CreateNodePoolRequest{
		CredentialID: req.CredentialID,
		ClusterName:  clusterName,
		NodePoolName: nodePoolName,
		Region:       req.Region,
		DesiredSize:  req.DesiredSize,
	}

	_, err = h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, scaleReq)
	if err != nil {
		h.LogError(c, err, "Failed to scale node pool")
		responses.InternalServerError(c, "Failed to scale node pool: "+err.Error())
		return
	}

	responses.OK(c, gin.H{"desired_size": req.DesiredSize}, "GKE node pool scaling initiated")
}

// UpgradeGKECluster handles GKE cluster upgrade
func (h *GCPHandler) UpgradeGKECluster(c *gin.Context) {
	// TODO: Implement GKE cluster upgrade
	responses.InternalServerError(c, "GKE cluster upgrade not yet implemented")
}

// GetGKEUpgradeStatus handles getting GKE cluster upgrade status
func (h *GCPHandler) GetGKEUpgradeStatus(c *gin.Context) {
	// TODO: Implement GKE cluster upgrade status
	responses.InternalServerError(c, "GKE cluster upgrade status not yet implemented")
}

// ListGKENodes handles listing GKE nodes
func (h *GCPHandler) ListGKENodes(c *gin.Context) {
	// TODO: Implement GKE nodes listing
	responses.InternalServerError(c, "GKE nodes listing not yet implemented")
}

// GetGKENode handles getting GKE node details
func (h *GCPHandler) GetGKENode(c *gin.Context) {
	// TODO: Implement GKE node details
	responses.InternalServerError(c, "GKE node details not yet implemented")
}

// DrainGKENode handles draining a GKE node
func (h *GCPHandler) DrainGKENode(c *gin.Context) {
	// TODO: Implement GKE node draining
	responses.InternalServerError(c, "GKE node draining not yet implemented")
}

// CordonGKENode handles cordoning a GKE node
func (h *GCPHandler) CordonGKENode(c *gin.Context) {
	// TODO: Implement GKE node cordoning
	responses.InternalServerError(c, "GKE node cordoning not yet implemented")
}

// UncordonGKENode handles uncordoning a GKE node
func (h *GCPHandler) UncordonGKENode(c *gin.Context) {
	// TODO: Implement GKE node uncordoning
	responses.InternalServerError(c, "GKE node uncordoning not yet implemented")
}

// GetGKENodeSSHConfig handles getting SSH config for GKE node
func (h *GCPHandler) GetGKENodeSSHConfig(c *gin.Context) {
	// TODO: Implement GKE node SSH config
	responses.InternalServerError(c, "GKE node SSH config not yet implemented")
}

// ExecuteGKENodeCommand handles executing commands on GKE node
func (h *GCPHandler) ExecuteGKENodeCommand(c *gin.Context) {
	// TODO: Implement GKE node command execution
	responses.InternalServerError(c, "GKE node command execution not yet implemented")
}
