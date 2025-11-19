package providers

import (
	"fmt"
	"net/http"
	"strings"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AWSHandler handles AWS EKS-related HTTP requests
type AWSHandler struct {
	*BaseHandler
}

// NewAWSHandler creates a new AWS EKS handler
func NewAWSHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *AWSHandler {
	return &AWSHandler{
		BaseHandler: NewBaseHandler(k8sService, credentialService, domain.ProviderAWS, "aws-kubernetes"),
	}
}

// CreateCluster handles EKS cluster creation using decorator pattern
func (h *AWSHandler) CreateCluster(c *gin.Context) {
	handler := h.Compose(
		h.createClusterHandler(),
		h.StandardCRUDDecorators("create_cluster")...,
	)

	handler(c)
}

func (h *AWSHandler) createClusterHandler() handlers.HandlerFunc {
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

		credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		ctx := h.EnrichContextWithRequestMetadata(c)
		cluster, err := h.GetK8sService().CreateEKSCluster(ctx, credential, req)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		h.logClusterCreationSuccess(c, userID, cluster)
		h.Created(c, cluster, "Cluster creation initiated")
	}
}

// ListClusters handles listing EKS clusters using decorator pattern
func (h *AWSHandler) ListClusters(c *gin.Context) {
	handler := h.Compose(
		h.listClustersHandler(),
		h.StandardCRUDDecorators("list_clusters")...,
	)

	handler(c)
}

func (h *AWSHandler) listClustersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		region := c.Query("region")

		h.logClusterListAttempt(c, userID, credential.ID, region)

		clusters, err := h.GetK8sService().ListEKSClusters(c.Request.Context(), credential, region)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		h.logClusterListSuccess(c, userID, len(clusters.Clusters))
		
		// Always include meta information for consistency (direct array: data[])
		page, limit := h.ParsePageLimitParams(c)
		total := int64(len(clusters.Clusters))
		h.BuildPaginatedResponse(c, clusters.Clusters, page, limit, total, "Clusters retrieved successfully")
	}
}

// BatchListClusters handles batch listing clusters from multiple credentials and regions
func (h *AWSHandler) BatchListClusters(c *gin.Context) {
	handler := h.Compose(
		h.batchListClustersHandler(),
		h.StandardCRUDDecorators("batch_list_clusters")...,
	)

	handler(c)
}

func (h *AWSHandler) batchListClustersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.BatchListClustersRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "batch_list_clusters")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "batch_list_clusters")
			return
		}

		h.LogInfo(c, "Batch cluster listing requested",
			zap.String("user_id", userID.String()),
			zap.Int("query_count", len(req.Queries)))

		result, err := h.GetK8sService().BatchListClusters(c.Request.Context(), req, h.GetCredentialService())
		if err != nil {
			h.HandleError(c, err, "batch_list_clusters")
			return
		}

		h.LogInfo(c, "Batch cluster listing completed",
			zap.String("user_id", userID.String()),
			zap.Int("total_clusters", result.Total))

		h.Success(c, http.StatusOK, result, "Batch clusters retrieved successfully")
	}
}

// GetCluster handles getting EKS cluster details using decorator pattern
func (h *AWSHandler) GetCluster(c *gin.Context) {
	handler := h.Compose(
		h.getClusterHandler(),
		h.StandardCRUDDecorators("get_cluster")...,
	)

	handler(c)
}

func (h *AWSHandler) getClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logClusterGetAttempt(c, userID, clusterName, credential.ID, region)

		cluster, err := h.GetK8sService().GetEKSCluster(c.Request.Context(), credential, clusterName, region)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		h.logClusterGetSuccess(c, userID, clusterName)
		h.OK(c, cluster, "Cluster retrieved successfully")
	}
}

// DeleteCluster handles EKS cluster deletion using decorator pattern
func (h *AWSHandler) DeleteCluster(c *gin.Context) {
	handler := h.Compose(
		h.deleteClusterHandler(),
		h.StandardCRUDDecorators("delete_cluster")...,
	)

	handler(c)
}

func (h *AWSHandler) deleteClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logClusterDeletionAttempt(c, userID, clusterName, credential.ID, region)

		ctx := h.EnrichContextWithRequestMetadata(c)
		if err := h.GetK8sService().DeleteEKSCluster(ctx, credential, clusterName, region); err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		h.logClusterDeletionSuccess(c, userID, clusterName)
		h.OK(c, nil, "Cluster deletion initiated")
	}
}

// GetKubeconfig handles getting kubeconfig for EKS cluster using decorator pattern
func (h *AWSHandler) GetKubeconfig(c *gin.Context) {
	handler := h.Compose(
		h.getKubeconfigHandler(),
		h.StandardCRUDDecorators("get_kubeconfig")...,
	)

	handler(c)
}

func (h *AWSHandler) getKubeconfigHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_kubeconfig")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_kubeconfig")
			return
		}

		clusterName := h.parseClusterName(c)
		region := h.parseRegion(c)

		if clusterName == "" || region == "" {
			return
		}

		h.logKubeconfigGetAttempt(c, userID, clusterName, credential.ID, region)

		kubeconfig, err := h.GetK8sService().GetEKSKubeconfig(c.Request.Context(), credential, clusterName, region)
		if err != nil {
			h.HandleError(c, err, "get_kubeconfig")
			return
		}

		h.logKubeconfigGetSuccess(c, userID, clusterName)

		c.Header("Content-Type", "application/x-yaml")
		c.Header("Content-Disposition", "attachment; filename=kubeconfig-"+clusterName+".yaml")
		c.String(http.StatusOK, kubeconfig)
	}
}

// CreateNodePool handles creating a node pool (node group) for EKS using decorator pattern
func (h *AWSHandler) CreateNodePool(c *gin.Context) {
	handler := h.Compose(
		h.createNodePoolHandler(),
		h.StandardCRUDDecorators("create_node_pool")...,
	)

	handler(c)
}

func (h *AWSHandler) createNodePoolHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateNodePoolRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_node_pool")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_node_pool")
			return
		}

		clusterName := h.parseClusterName(c)
		if clusterName == "" {
			return
		}

		req.ClusterName = clusterName

		h.logNodePoolCreationAttempt(c, userID, clusterName, req)

		credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderAWS)
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

		h.logNodePoolCreationSuccess(c, userID, clusterName, nodePool)
		h.Created(c, nodePool, "Node pool creation initiated")
	}
}

// ListNodePools handles listing node pools
func (h *AWSHandler) ListNodePools(c *gin.Context) {
	h.NotImplemented(c, "list_node_pools")
}

// GetNodePool handles getting node pool details
func (h *AWSHandler) GetNodePool(c *gin.Context) {
	h.NotImplemented(c, "get_node_pool")
}

// DeleteNodePool handles deleting a node pool
func (h *AWSHandler) DeleteNodePool(c *gin.Context) {
	h.NotImplemented(c, "delete_node_pool")
}

// ScaleNodePool handles scaling a node pool
func (h *AWSHandler) ScaleNodePool(c *gin.Context) {
	h.NotImplemented(c, "scale_node_pool")
}

// CreateNodeGroup handles creating an EKS node group using decorator pattern
func (h *AWSHandler) CreateNodeGroup(c *gin.Context) {
	handler := h.Compose(
		h.createNodeGroupHandler(),
		h.StandardCRUDDecorators("create_node_group")...,
	)

	handler(c)
}

func (h *AWSHandler) createNodeGroupHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateNodeGroupRequest
		if err := h.ValidateRequest(c, &req); err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		clusterName := h.parseClusterName(c)
		if clusterName == "" {
			return
		}

		req.ClusterName = clusterName

		h.logNodeGroupCreationAttempt(c, userID, clusterName, req)

		credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		ctx := h.EnrichContextWithRequestMetadata(c)
		nodeGroup, err := h.GetK8sService().CreateEKSNodeGroup(ctx, credential, req)
		if err != nil {
			h.HandleError(c, err, "create_node_group")
			return
		}

		h.logNodeGroupCreationSuccess(c, userID, clusterName, nodeGroup)
		h.Created(c, nodeGroup, "Node group creation initiated")
	}
}

// ListNodeGroups handles listing node groups for a cluster using decorator pattern
func (h *AWSHandler) ListNodeGroups(c *gin.Context) {
	handler := h.Compose(
		h.listNodeGroupsHandler(),
		h.StandardCRUDDecorators("list_node_groups")...,
	)

	handler(c)
}

func (h *AWSHandler) listNodeGroupsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "list_node_groups")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_node_groups")
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

		nodeGroupsResponse, err := h.GetK8sService().ListNodeGroups(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "list_node_groups")
			return
		}

		h.logNodeGroupsListSuccess(c, userID, clusterName, len(nodeGroupsResponse.NodeGroups))
		// Always include meta information for consistency (direct array: data[])
		page, limit := h.ParsePageLimitParams(c)
		total := int64(len(nodeGroupsResponse.NodeGroups))
		h.BuildPaginatedResponse(c, nodeGroupsResponse.NodeGroups, page, limit, total, "Node groups retrieved successfully")
	}
}

// GetNodeGroup handles getting node group details
func (h *AWSHandler) GetNodeGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
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

	req := kubernetesservice.GetNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        region,
	}

	nodeGroup, err := h.GetK8sService().GetNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "get_node_group")
		return
	}

	h.OK(c, nodeGroup, "Node group retrieved successfully")
}

// UpdateNodeGroup handles updating a node group
func (h *AWSHandler) UpdateNodeGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "update_node_group")
		return
	}

	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "update_node_group")
		return
	}

	var req kubernetesservice.UpdateNodeGroupRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "update_node_group")
		return
	}

	req.CredentialID = credential.ID.String()
	req.ClusterName = clusterName
	req.NodeGroupName = nodeGroupName

	region := c.Query("region")
	if region == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "update_node_group")
		return
	}
	req.Region = region

	nodeGroup, err := h.GetK8sService().UpdateNodeGroup(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "update_node_group")
		return
	}

	h.OK(c, nodeGroup, "Node group update initiated")
}

// DeleteNodeGroup handles deleting a node group
func (h *AWSHandler) DeleteNodeGroup(c *gin.Context) {
	clusterName := c.Param("name")
	nodeGroupName := c.Param("nodegroup")

	if clusterName == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name and node group name are required", 400), "delete_node_group")
		return
	}

	var req struct {
		CredentialID string `json:"credential_id" validate:"required,uuid"`
		Region       string `json:"region" validate:"required"`
	}

	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderAWS)
	if err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	deleteReq := kubernetesservice.DeleteNodeGroupRequest{
		CredentialID:  credential.ID.String(),
		ClusterName:   clusterName,
		NodeGroupName: nodeGroupName,
		Region:        req.Region,
	}

	ctx := h.EnrichContextWithRequestMetadata(c)
	if err := h.GetK8sService().DeleteNodeGroup(ctx, credential, deleteReq); err != nil {
		h.HandleError(c, err, "delete_node_group")
		return
	}

	h.OK(c, nil, "Node group deletion initiated")
}

// UpgradeCluster handles cluster upgrade
func (h *AWSHandler) UpgradeCluster(c *gin.Context) {
	h.NotImplemented(c, "upgrade_cluster")
}

// GetUpgradeStatus handles getting cluster upgrade status
func (h *AWSHandler) GetUpgradeStatus(c *gin.Context) {
	h.NotImplemented(c, "get_upgrade_status")
}

// ListNodes handles listing cluster nodes
func (h *AWSHandler) ListNodes(c *gin.Context) {
	h.NotImplemented(c, "list_nodes")
}

// GetNode handles getting node details
func (h *AWSHandler) GetNode(c *gin.Context) {
	h.NotImplemented(c, "get_node")
}

// DrainNode handles draining a node
func (h *AWSHandler) DrainNode(c *gin.Context) {
	h.NotImplemented(c, "drain_node")
}

// CordonNode handles cordoning a node
func (h *AWSHandler) CordonNode(c *gin.Context) {
	h.NotImplemented(c, "cordon_node")
}

// UncordonNode handles uncordoning a node
func (h *AWSHandler) UncordonNode(c *gin.Context) {
	h.NotImplemented(c, "uncordon_node")
}

// GetNodeSSHConfig handles getting SSH config for a node
func (h *AWSHandler) GetNodeSSHConfig(c *gin.Context) {
	h.NotImplemented(c, "get_node_ssh_config")
}

// ExecuteNodeCommand handles executing a command on a node
func (h *AWSHandler) ExecuteNodeCommand(c *gin.Context) {
	h.NotImplemented(c, "execute_node_command")
}

// Logging helper methods

func (h *AWSHandler) logClusterCreationAttempt(c *gin.Context, userID uuid.UUID, req kubernetesservice.CreateClusterRequest) {
	h.LogBusinessEvent(c, "cluster_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_cluster",
		"cluster_name":  req.Name,
		"provider":      domain.ProviderAWS,
		"credential_id": req.CredentialID,
	})
}

func (h *AWSHandler) logClusterCreationSuccess(c *gin.Context, userID uuid.UUID, cluster interface{}) {
	h.LogBusinessEvent(c, "cluster_created", userID.String(), "", map[string]interface{}{
		"operation": "create_cluster",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logClusterListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_list_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "list_clusters",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logClusterListSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "clusters_listed", userID.String(), "", map[string]interface{}{
		"operation": "list_clusters",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

func (h *AWSHandler) logClusterGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_cluster",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logClusterGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_cluster",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logClusterDeletionAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "cluster_deletion_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "delete_cluster",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logClusterDeletionSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_deleted", userID.String(), clusterName, map[string]interface{}{
		"operation": "delete_cluster",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logKubeconfigGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "kubeconfig_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_kubeconfig",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logKubeconfigGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "kubeconfig_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_kubeconfig",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logNodePoolCreationAttempt(c *gin.Context, userID uuid.UUID, clusterName string, req kubernetesservice.CreateNodePoolRequest) {
	h.LogBusinessEvent(c, "node_pool_creation_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":      "create_node_pool",
		"provider":       domain.ProviderAWS,
		"node_pool_name": req.NodePoolName,
		"credential_id":  req.CredentialID,
	})
}

func (h *AWSHandler) logNodePoolCreationSuccess(c *gin.Context, userID uuid.UUID, clusterName string, nodePool interface{}) {
	h.LogBusinessEvent(c, "node_pool_created", userID.String(), clusterName, map[string]interface{}{
		"operation": "create_node_pool",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logNodeGroupCreationAttempt(c *gin.Context, userID uuid.UUID, clusterName string, req kubernetesservice.CreateNodeGroupRequest) {
	h.LogBusinessEvent(c, "node_group_creation_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":       "create_node_group",
		"provider":        domain.ProviderAWS,
		"node_group_name": req.NodeGroupName,
		"credential_id":   req.CredentialID,
	})
}

func (h *AWSHandler) logNodeGroupCreationSuccess(c *gin.Context, userID uuid.UUID, clusterName string, nodeGroup interface{}) {
	h.LogBusinessEvent(c, "node_group_created", userID.String(), clusterName, map[string]interface{}{
		"operation": "create_node_group",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logNodeGroupsListAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "node_groups_list_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "list_node_groups",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logNodeGroupsListSuccess(c *gin.Context, userID uuid.UUID, clusterName string, count int) {
	h.LogBusinessEvent(c, "node_groups_listed", userID.String(), clusterName, map[string]interface{}{
		"operation": "list_node_groups",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

// GetEKSVersions handles EKS versions listing using decorator pattern
func (h *AWSHandler) GetEKSVersions(c *gin.Context) {
	handler := h.Compose(
		h.getEKSVersionsHandler(),
		h.StandardCRUDDecorators("get_eks_versions")...,
	)
	handler(c)
}

func (h *AWSHandler) getEKSVersionsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_eks_versions")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_eks_versions")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_eks_versions")
			return
		}

		h.logEKSVersionsGetAttempt(c, userID, credential.ID, region)

		versions, err := h.GetK8sService().GetEKSVersions(c.Request.Context(), credential, region)
		if err != nil {
			h.HandleError(c, err, "get_eks_versions")
			return
		}

		h.logEKSVersionsGetSuccess(c, userID, len(versions))
		h.OK(c, gin.H{
			"versions": versions,
		}, "EKS versions retrieved successfully")
	}
}

// GetGKEVersions handles GKE versions listing (GCP only)
func (h *AWSHandler) GetGKEVersions(c *gin.Context) {
	h.NotImplemented(c, "get_gke_versions")
}

// GetAKSVersions handles AKS versions listing (Azure only)
func (h *AWSHandler) GetAKSVersions(c *gin.Context) {
	h.NotImplemented(c, "get_aks_versions")
}

// GetAWSRegions handles AWS regions listing using decorator pattern
func (h *AWSHandler) GetAWSRegions(c *gin.Context) {
	handler := h.Compose(
		h.getAWSRegionsHandler(),
		h.StandardCRUDDecorators("get_aws_regions")...,
	)
	handler(c)
}

func (h *AWSHandler) getAWSRegionsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_aws_regions")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_aws_regions")
			return
		}

		h.logAWSRegionsGetAttempt(c, userID, credential.ID)

		regions, err := h.GetK8sService().GetAWSRegions(c.Request.Context(), credential)
		if err != nil {
			h.HandleError(c, err, "get_aws_regions")
			return
		}

		h.logAWSRegionsGetSuccess(c, userID, len(regions))
		h.OK(c, gin.H{
			"regions": regions,
		}, "AWS regions retrieved successfully")
	}
}

// GetAvailabilityZones handles availability zones listing using decorator pattern
func (h *AWSHandler) GetAvailabilityZones(c *gin.Context) {
	handler := h.Compose(
		h.getAvailabilityZonesHandler(),
		h.StandardCRUDDecorators("get_availability_zones")...,
	)
	handler(c)
}

func (h *AWSHandler) getAvailabilityZonesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_availability_zones")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_availability_zones")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_availability_zones")
			return
		}

		h.logAvailabilityZonesGetAttempt(c, userID, credential.ID, region)

		zones, err := h.GetK8sService().GetAvailabilityZones(c.Request.Context(), credential, region)
		if err != nil {
			h.HandleError(c, err, "get_availability_zones")
			return
		}

		h.logAvailabilityZonesGetSuccess(c, userID, len(zones))
		h.OK(c, gin.H{
			"zones": zones,
		}, "Availability zones retrieved successfully")
	}
}

func (h *AWSHandler) logEKSVersionsGetAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "eks_versions_get_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "get_eks_versions",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logEKSVersionsGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "eks_versions_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_eks_versions",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

func (h *AWSHandler) logAWSRegionsGetAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID) {
	h.LogBusinessEvent(c, "aws_regions_get_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "get_aws_regions",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
	})
}

func (h *AWSHandler) logAWSRegionsGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "aws_regions_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_aws_regions",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

func (h *AWSHandler) logAvailabilityZonesGetAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "availability_zones_get_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "get_availability_zones",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logAvailabilityZonesGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "availability_zones_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_availability_zones",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

// GetInstanceTypes handles EC2 instance types listing using decorator pattern
func (h *AWSHandler) GetInstanceTypes(c *gin.Context) {
	handler := h.Compose(
		h.getInstanceTypesHandler(),
		h.StandardCRUDDecorators("get_instance_types")...,
	)
	handler(c)
}

func (h *AWSHandler) getInstanceTypesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_instance_types")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_instance_types")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_instance_types")
			return
		}

		h.logInstanceTypesGetAttempt(c, userID, credential.ID, region)

		instanceTypes, err := h.GetK8sService().GetInstanceTypes(c.Request.Context(), credential, region)
		if err != nil {
			h.HandleError(c, err, "get_instance_types")
			return
		}

		h.logInstanceTypesGetSuccess(c, userID, len(instanceTypes))
		h.OK(c, gin.H{
			"instance_types": instanceTypes,
		}, "Instance types retrieved successfully")
	}
}

// GetEKSAmitTypes handles EKS AMI types listing using decorator pattern
func (h *AWSHandler) GetEKSAmitTypes(c *gin.Context) {
	handler := h.Compose(
		h.getEKSAmitTypesHandler(),
		h.StandardCRUDDecorators("get_eks_ami_types")...,
	)
	handler(c)
}

func (h *AWSHandler) getEKSAmitTypesHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_eks_ami_types")
			return
		}

		h.logEKSAmitTypesGetAttempt(c, userID)

		amiTypes, err := h.GetK8sService().GetEKSAmitTypes(c.Request.Context())
		if err != nil {
			h.HandleError(c, err, "get_eks_ami_types")
			return
		}

		h.logEKSAmitTypesGetSuccess(c, userID, len(amiTypes))
		h.OK(c, gin.H{
			"ami_types": amiTypes,
		}, "EKS AMI types retrieved successfully")
	}
}

// CheckGPUQuota handles GPU quota checking using decorator pattern
func (h *AWSHandler) CheckGPUQuota(c *gin.Context) {
	handler := h.Compose(
		h.checkGPUQuotaHandler(),
		h.StandardCRUDDecorators("check_gpu_quota")...,
	)
	handler(c)
}

func (h *AWSHandler) checkGPUQuotaHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "check_gpu_quota")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "check_gpu_quota")
			return
		}

		instanceType := c.Query("instance_type")
		if instanceType == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "instance_type is required", 400), "check_gpu_quota")
			return
		}

		// required_count는 선택적 (기본값: 1)
		requiredCount := int32(1)
		if reqCountStr := c.Query("required_count"); reqCountStr != "" {
			var reqCount int32
			if _, err := fmt.Sscanf(reqCountStr, "%d", &reqCount); err == nil && reqCount > 0 {
				requiredCount = reqCount
			}
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "check_gpu_quota")
			return
		}

		h.logGPUQuotaCheckAttempt(c, userID, credential.ID, region, instanceType, requiredCount)

		availability, err := h.GetK8sService().CheckGPUQuotaAvailability(c.Request.Context(), credential, region, instanceType, requiredCount)
		if err != nil {
			h.HandleError(c, err, "check_gpu_quota")
			return
		}

		h.logGPUQuotaCheckSuccess(c, userID, availability.Available)
		h.OK(c, gin.H{
			"availability": availability,
		}, "GPU quota checked successfully")
	}
}

func (h *AWSHandler) logGPUQuotaCheckAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region, instanceType string, requiredCount int32) {
	h.LogBusinessEvent(c, "gpu_quota_check_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "check_gpu_quota",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
		"instance_type": instanceType,
		"required_count": requiredCount,
	})
}

// GetVMSizes handles VM sizes listing (Azure only)
func (h *AWSHandler) GetVMSizes(c *gin.Context) {
	h.NotImplemented(c, "get_vm_sizes")
}

// CheckCPUQuota handles CPU quota checking using decorator pattern
func (h *AWSHandler) CheckCPUQuota(c *gin.Context) {
	handler := h.Compose(
		h.checkCPUQuotaHandler(),
		h.StandardCRUDDecorators("check_cpu_quota")...,
	)

	handler(c)
}

// checkCPUQuotaHandler: CPU quota 확인 핵심 비즈니스 로직을 처리합니다
func (h *AWSHandler) checkCPUQuotaHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "check_cpu_quota")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "check_cpu_quota")
			return
		}

		// instance_types는 필수 (쉼표로 구분된 목록)
		instanceTypesStr := c.Query("instance_types")
		if instanceTypesStr == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "instance_types is required (comma-separated)", 400), "check_cpu_quota")
			return
		}

		instanceTypes := strings.Split(instanceTypesStr, ",")
		for i := range instanceTypes {
			instanceTypes[i] = strings.TrimSpace(instanceTypes[i])
		}

		// required_count는 선택적 (기본값: 1)
		requiredCount := int32(1)
		if reqCountStr := c.Query("required_count"); reqCountStr != "" {
			var reqCount int32
			if _, err := fmt.Sscanf(reqCountStr, "%d", &reqCount); err == nil && reqCount > 0 {
				requiredCount = reqCount
			}
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "check_cpu_quota")
			return
		}

		h.logCPUQuotaCheckAttempt(c, userID, credential.ID, region, instanceTypes, requiredCount)

		availability, err := h.GetK8sService().CheckCPUQuotaAvailability(c.Request.Context(), credential, region, instanceTypes, requiredCount)
		if err != nil {
			h.HandleError(c, err, "check_cpu_quota")
			return
		}

		h.logCPUQuotaCheckSuccess(c, userID, availability.Available)
		h.OK(c, gin.H{
			"availability": availability,
		}, "CPU quota checked successfully")
	}
}

func (h *AWSHandler) logCPUQuotaCheckAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string, instanceTypes []string, requiredCount int32) {
	h.LogBusinessEvent(c, "cpu_quota_check_attempted", userID.String(), "", map[string]interface{}{
		"operation":      "check_cpu_quota",
		"provider":       domain.ProviderAWS,
		"credential_id":   credentialID.String(),
		"region":         region,
		"instance_types": instanceTypes,
		"required_count": requiredCount,
	})
}

func (h *AWSHandler) logCPUQuotaCheckSuccess(c *gin.Context, userID uuid.UUID, available bool) {
	h.LogBusinessEvent(c, "cpu_quota_checked", userID.String(), "", map[string]interface{}{
		"available": available,
	})
}

func (h *AWSHandler) logGPUQuotaCheckSuccess(c *gin.Context, userID uuid.UUID, available bool) {
	h.LogBusinessEvent(c, "gpu_quota_checked", userID.String(), "", map[string]interface{}{
		"operation": "check_gpu_quota",
		"provider":  domain.ProviderAWS,
		"available": available,
	})
}

func (h *AWSHandler) logInstanceTypesGetAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region string) {
	h.LogBusinessEvent(c, "instance_types_get_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "get_instance_types",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
	})
}

func (h *AWSHandler) logInstanceTypesGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "instance_types_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_instance_types",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

func (h *AWSHandler) logEKSAmitTypesGetAttempt(c *gin.Context, userID uuid.UUID) {
	h.LogBusinessEvent(c, "eks_ami_types_get_attempted", userID.String(), "", map[string]interface{}{
		"operation": "get_eks_ami_types",
		"provider":  domain.ProviderAWS,
	})
}

func (h *AWSHandler) logEKSAmitTypesGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "eks_ami_types_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_eks_ami_types",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}

// GetInstanceTypeOfferings handles instance type offerings listing using decorator pattern
func (h *AWSHandler) GetInstanceTypeOfferings(c *gin.Context) {
	handler := h.Compose(
		h.getInstanceTypeOfferingsHandler(),
		h.StandardCRUDDecorators("get_instance_type_offerings")...,
	)
	handler(c)
}

func (h *AWSHandler) getInstanceTypeOfferingsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAWS)
		if err != nil {
			h.HandleError(c, err, "get_instance_type_offerings")
			return
		}

		region := c.Query("region")
		if region == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "region is required", 400), "get_instance_type_offerings")
			return
		}

		instanceType := c.Query("instance_type")
		if instanceType == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "instance_type is required", 400), "get_instance_type_offerings")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "get_instance_type_offerings")
			return
		}

		h.logInstanceTypeOfferingsGetAttempt(c, userID, credential.ID, region, instanceType)

		offerings, err := h.GetK8sService().GetInstanceTypeOfferings(c.Request.Context(), credential, region, instanceType)
		if err != nil {
			h.HandleError(c, err, "get_instance_type_offerings")
			return
		}

		h.logInstanceTypeOfferingsGetSuccess(c, userID, len(offerings))
		h.OK(c, gin.H{
			"offerings": offerings,
		}, "Instance type offerings retrieved successfully")
	}
}

func (h *AWSHandler) logInstanceTypeOfferingsGetAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, region, instanceType string) {
	h.LogBusinessEvent(c, "instance_type_offerings_get_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "get_instance_type_offerings",
		"provider":      domain.ProviderAWS,
		"credential_id": credentialID.String(),
		"region":        region,
		"instance_type": instanceType,
	})
}

func (h *AWSHandler) logInstanceTypeOfferingsGetSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "instance_type_offerings_retrieved", userID.String(), "", map[string]interface{}{
		"operation": "get_instance_type_offerings",
		"provider":  domain.ProviderAWS,
		"count":     count,
	})
}
