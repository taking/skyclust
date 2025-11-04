package kubernetes

import (
	"net/http"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"
	"skyclust/internal/shared/readability"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles Kubernetes-related HTTP requests using improved patterns
type Handler struct {
	*handlers.BaseHandler
	k8sService        *kubernetesservice.Service
	credentialService domain.CredentialService
	provider          string // "aws", "gcp", "azure", "ncp"
	readabilityHelper *readability.ReadabilityHelper
}

// NewHandler creates a new Kubernetes handler for a specific provider
func NewHandler(k8sService *kubernetesservice.Service, credentialService domain.CredentialService, provider string) *Handler {
	return &Handler{
		BaseHandler:       handlers.NewBaseHandler("kubernetes"),
		k8sService:        k8sService,
		credentialService: credentialService,
		provider:          provider,
		readabilityHelper: readability.NewReadabilityHelper(),
	}
}

// CreateCluster handles EKS cluster creation using decorator pattern
func (h *Handler) CreateCluster(c *gin.Context) {
	var req kubernetesservice.CreateClusterRequest

	handler := h.Compose(
		h.createClusterHandler(req),
		h.StandardCRUDDecorators("create_cluster")...,
	)

	handler(c)
}

// createClusterHandler is the core business logic for creating a cluster
func (h *Handler) createClusterHandler(req kubernetesservice.CreateClusterRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateClusterRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		h.logClusterCreationAttempt(c, userID, req)

		// Get and validate credential from request body
		credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		cluster, err := h.k8sService.CreateEKSCluster(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		h.logClusterCreationSuccess(c, userID, cluster)
		h.Created(c, cluster, "Cluster creation initiated")
	}
}

// ListClusters handles listing EKS clusters using decorator pattern
func (h *Handler) ListClusters(c *gin.Context) {
	handler := h.Compose(
		h.listClustersHandler(),
		h.StandardCRUDDecorators("list_clusters")...,
	)

	handler(c)
}

// listClustersHandler is the core business logic for listing clusters
func (h *Handler) listClustersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (with provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		region := c.Query("region")

		h.logClusterListAttempt(c, userID, credential.ID, region)

		clusters, err := h.k8sService.ListEKSClusters(c.Request.Context(), credential, region)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		h.logClusterListSuccess(c, userID, len(clusters.Clusters))
		h.OK(c, clusters, "Clusters retrieved successfully")
	}
}

// GetCluster handles getting EKS cluster details using decorator pattern
func (h *Handler) GetCluster(c *gin.Context) {
	handler := h.Compose(
		h.getClusterHandler(),
		h.StandardCRUDDecorators("get_cluster")...,
	)

	handler(c)
}

// getClusterHandler is the core business logic for getting a cluster
func (h *Handler) getClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (with provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logClusterGetAttempt(c, userID, clusterName, credential.ID, region)

		cluster, err := h.k8sService.GetEKSCluster(c.Request.Context(), credential, clusterName, region)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		h.logClusterGetSuccess(c, userID, clusterName)
		h.OK(c, cluster, "Cluster retrieved successfully")
	}
}

// DeleteCluster handles EKS cluster deletion using decorator pattern
func (h *Handler) DeleteCluster(c *gin.Context) {
	handler := h.Compose(
		h.deleteClusterHandler(),
		h.StandardCRUDDecorators("delete_cluster")...,
	)

	handler(c)
}

// deleteClusterHandler is the core business logic for deleting a cluster
func (h *Handler) deleteClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (with provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logClusterDeletionAttempt(c, userID, clusterName, credential.ID, region)

		if err := h.k8sService.DeleteEKSCluster(c.Request.Context(), credential, clusterName, region); err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		h.logClusterDeletionSuccess(c, userID, clusterName)
		h.OK(c, nil, "Cluster deletion initiated")
	}
}

// GetKubeconfig handles getting kubeconfig for EKS cluster using decorator pattern
func (h *Handler) GetKubeconfig(c *gin.Context) {
	handler := h.Compose(
		h.getKubeconfigHandler(),
		h.StandardCRUDDecorators("get_kubeconfig")...,
	)

	handler(c)
}

// getKubeconfigHandler is the core business logic for getting kubeconfig
func (h *Handler) getKubeconfigHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (with provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "get_kubeconfig")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logKubeconfigGetAttempt(c, userID, clusterName, credential.ID, region)

		kubeconfig, err := h.k8sService.GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
		if err != nil {
			h.HandleError(c, err, "get_kubeconfig")
			return
		}

		h.logKubeconfigGetSuccess(c, userID, clusterName)

		// Return as downloadable file
		c.Header("Content-Type", "application/x-yaml")
		c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
		c.String(http.StatusOK, kubeconfig)
	}
}

// CreateNodePool handles creating a node pool (node group) for EKS using decorator pattern
func (h *Handler) CreateNodePool(c *gin.Context) {
	var req kubernetesservice.CreateNodePoolRequest

	handler := h.Compose(
		h.createNodePoolHandler(req),
		h.StandardCRUDDecorators("create_node_pool")...,
	)

	handler(c)
}

// createNodePoolHandler is the core business logic for creating a node pool
func (h *Handler) createNodePoolHandler(req kubernetesservice.CreateNodePoolRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateNodePoolRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_node_pool")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)

		if clusterName == "" {
			return
		}

		req.ClusterName = clusterName

		h.logNodePoolCreationAttempt(c, userID, clusterName, req)

		// Get and validate credential from request body
		credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_node_pool")
			return
		}

		nodePool, err := h.k8sService.CreateEKSNodePool(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "create_node_pool")
			return
		}

		h.logNodePoolCreationSuccess(c, userID, clusterName, nodePool)
		h.Created(c, nodePool, "Node pool creation initiated")
	}
}

// CreateNodeGroup handles creating an EKS node group using decorator pattern
func (h *Handler) CreateNodeGroup(c *gin.Context) {
	var req kubernetesservice.CreateNodeGroupRequest

	handler := h.Compose(
		h.createNodeGroupHandler(req),
		h.StandardCRUDDecorators("create_node_group")...,
	)

	handler(c)
}

// createNodeGroupHandler is the core business logic for creating a node group
func (h *Handler) createNodeGroupHandler(req kubernetesservice.CreateNodeGroupRequest) handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateNodeGroupRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)

		if clusterName == "" {
			return
		}

		req.ClusterName = clusterName

		h.logNodeGroupCreationAttempt(c, userID, clusterName, req)

		// Get and validate credential from request body
		credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, h.provider)
		if err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		nodeGroup, err := h.k8sService.CreateEKSNodeGroup(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		h.logNodeGroupCreationSuccess(c, userID, clusterName, nodeGroup)
		h.Created(c, nodeGroup, "Node group creation initiated")
	}
}

// ListNodeGroups handles listing node groups for a cluster using decorator pattern
func (h *Handler) ListNodeGroups(c *gin.Context) {
	handler := h.Compose(
		h.listNodeGroupsHandler(),
		h.StandardCRUDDecorators("list_node_groups")...,
	)

	handler(c)
}

// listNodeGroupsHandler is the core business logic for listing node groups
func (h *Handler) listNodeGroupsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		// Get and validate credential using BaseHandler helper (with provider validation)
		credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
		if err != nil {
			h.HandleError(c, err, "list_node_groups")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}
		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logNodeGroupsListAttempt(c, userID, clusterName, credential.ID, region)

		req := kubernetesservice.ListNodeGroupsRequest{
			CredentialID: credential.ID.String(),
			ClusterName:  clusterName,
			Region:       region,
		}

		nodeGroupsResponse, err := h.k8sService.ListNodeGroups(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "list_node_groups")
			return
		}

		h.logNodeGroupsListSuccess(c, userID, clusterName, len(nodeGroupsResponse.NodeGroups))
		h.OK(c, nodeGroupsResponse, "Node groups retrieved successfully")
	}
}

// GetNodeGroup handles getting node group details
func (h *Handler) GetNodeGroup(c *gin.Context) {
	// Get and validate credential using BaseHandler helper (with provider validation)
	credential, err := h.GetCredentialFromRequest(c, h.credentialService, h.provider)
	if err != nil {
		h.HandleError(c, err, "get_node_group")
		return
	}

	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "get_node_group")
		return
	}

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_node_group")
		return
	}

	// Get node group
	req := kubernetesservice.GetNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        region,
	}

	nodeGroup, err := h.k8sService.GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "get_node_group")
		return
	}

	h.OK(c, nodeGroup, "Node group retrieved successfully")
}

// DeleteNodeGroup handles deleting a node group
func (h *Handler) DeleteNodeGroup(c *gin.Context) {
	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "delete_node_group")
		return
	}

	// Parse request
	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	// Get and validate credential from request body
	credential, err := h.GetCredentialFromBody(c, h.credentialService, req.CredentialID, h.provider)
	if err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	// Delete node group
	deleteReq := kubernetesservice.DeleteNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        req.Region,
	}

	if err := h.k8sService.DeleteNodeGroup(c.Request.Context(), credential, deleteReq); err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	h.OK(c, nil, "Node group deletion initiated")
}

// Helper methods for better readability

func (h *Handler) parseClusterName(c *gin.Context) string {
	clusterName := c.Param("name")
	if clusterName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name is required", 400), "parse_cluster_name")
		return ""
	}
	return clusterName
}

func (h *Handler) parseRegion(c *gin.Context) string {
	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "parse_region")
		return ""
	}
	return region
}

// Logging helper methods

func (h *Handler) logClusterCreationAttempt(c *gin.Context, userID uuid.UUID, req kubernetesservice.CreateClusterRequest) {
	h.LogBusinessEvent(c, "cluster_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_cluster",
		"cluster_name":  req.Name,
		"provider":      h.provider,
		"credential_id": req.CredentialID,
	})
}

func (h *Handler) logClusterCreationSuccess(c *gin.Context, userID uuid.UUID, cluster interface{}) {
	h.LogBusinessEvent(c, "cluster_created", userID.String(), "", map[string]interface{}{
		"operation": "create_cluster",
		"provider":  h.provider,
	})
}

func (h *Handler) logClusterListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_list_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "list_clusters",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logClusterListSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "clusters_listed", userID.String(), "", map[string]interface{}{
		"operation": "list_clusters",
		"provider":  h.provider,
		"count":     count,
	})
}

func (h *Handler) logClusterGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_cluster",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logClusterGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_cluster",
		"provider":  h.provider,
	})
}

func (h *Handler) logClusterDeletionAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_deletion_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "delete_cluster",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logClusterDeletionSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_deleted", userID.String(), clusterName, map[string]interface{}{
		"operation": "delete_cluster",
		"provider":  h.provider,
	})
}

func (h *Handler) logKubeconfigGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "kubeconfig_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_kubeconfig",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logKubeconfigGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "kubeconfig_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_kubeconfig",
		"provider":  h.provider,
	})
}

func (h *Handler) logNodePoolCreationAttempt(c *gin.Context, userID uuid.UUID, clusterName string, req kubernetesservice.CreateNodePoolRequest) {
	h.LogBusinessEvent(c, "node_pool_creation_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":      "create_node_pool",
		"provider":       h.provider,
		"node_pool_name": req.NodePoolName,
		"credential_id":  req.CredentialID,
	})
}

func (h *Handler) logNodePoolCreationSuccess(c *gin.Context, userID uuid.UUID, clusterName string, nodePool interface{}) {
	h.LogBusinessEvent(c, "node_pool_created", userID.String(), clusterName, map[string]interface{}{
		"operation": "create_node_pool",
		"provider":  h.provider,
	})
}

func (h *Handler) logNodeGroupCreationAttempt(c *gin.Context, userID uuid.UUID, clusterName string, req kubernetesservice.CreateNodeGroupRequest) {
	h.LogBusinessEvent(c, "node_group_creation_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":       "create_node_group",
		"provider":        h.provider,
		"node_group_name": req.NodeGroupName,
		"credential_id":   req.CredentialID,
	})
}

func (h *Handler) logNodeGroupCreationSuccess(c *gin.Context, userID uuid.UUID, clusterName string, nodeGroup interface{}) {
	h.LogBusinessEvent(c, "node_group_created", userID.String(), clusterName, map[string]interface{}{
		"operation": "create_node_group",
		"provider":  h.provider,
	})
}

func (h *Handler) logNodeGroupsListAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "node_groups_list_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "list_node_groups",
		"provider":      h.provider,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *Handler) logNodeGroupsListSuccess(c *gin.Context, userID uuid.UUID, clusterName string, count int) {
	h.LogBusinessEvent(c, "node_groups_listed", userID.String(), clusterName, map[string]interface{}{
		"operation": "list_node_groups",
		"provider":  h.provider,
		"count":     count,
	})
}
