package providers

import (
	"net/http"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"
	"skyclust/internal/shared/handlers"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// CreateCluster handles AKS cluster creation using decorator pattern
func (h *AzureHandler) CreateCluster(c *gin.Context) {
	handler := h.Compose(
		h.createClusterHandler(),
		h.StandardCRUDDecorators("create_cluster")...,
	)

	handler(c)
}

func (h *AzureHandler) createClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		var req kubernetesservice.CreateAKSClusterRequest
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

		credential, err := h.GetCredentialFromBody(c, h.GetCredentialService(), req.CredentialID, domain.ProviderAzure)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		ctx := h.EnrichContextWithRequestMetadata(c)
		cluster, err := h.GetK8sService().CreateAKSCluster(ctx, credential, req)
		if err != nil {
			h.HandleError(c, err, "create_cluster")
			return
		}

		h.logClusterCreationSuccess(c, userID, cluster)
		h.Created(c, cluster, "AKS cluster creation initiated")
	}
}

// ListClusters handles listing AKS clusters using decorator pattern
func (h *AzureHandler) ListClusters(c *gin.Context) {
	handler := h.Compose(
		h.listClustersHandler(),
		h.StandardCRUDDecorators("list_clusters")...,
	)

	handler(c)
}

func (h *AzureHandler) listClustersHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		userID, err := h.ExtractUserIDFromContext(c)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		location := c.Query("region")
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}

		resourceGroup := c.Query("resource_group")

		h.logClusterListAttempt(c, userID, credential.ID, location)

		clusters, err := h.GetK8sService().ListEKSClusters(c.Request.Context(), credential, location, resourceGroup)
		if err != nil {
			h.HandleError(c, err, "list_clusters")
			return
		}

		if clusters == nil {
			clusters = &kubernetesservice.ListClustersResponse{Clusters: []kubernetesservice.ClusterInfo{}}
		}
		if clusters.Clusters == nil {
			clusters.Clusters = []kubernetesservice.ClusterInfo{}
		}

		h.logClusterListSuccess(c, userID, len(clusters.Clusters))

		// Always include meta information for consistency (direct array: data[])
		page, limit := h.ParsePageLimitParams(c)
		total := int64(len(clusters.Clusters))
		h.BuildPaginatedResponse(c, clusters.Clusters, page, limit, total, "AKS clusters retrieved successfully")
	}
}

// GetCluster handles getting AKS cluster details using decorator pattern
func (h *AzureHandler) GetCluster(c *gin.Context) {
	handler := h.Compose(
		h.getClusterHandler(),
		h.StandardCRUDDecorators("get_cluster")...,
	)

	handler(c)
}

func (h *AzureHandler) getClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
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
		location := h.parseRegion(c)
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}

		if clusterName == "" || location == "" {
			return
		}

		h.logClusterGetAttempt(c, userID, clusterName, credential.ID, location)

		cluster, err := h.GetK8sService().GetEKSCluster(c.Request.Context(), credential, clusterName, location)
		if err != nil {
			h.HandleError(c, err, "get_cluster")
			return
		}

		h.logClusterGetSuccess(c, userID, clusterName)
		h.OK(c, cluster, "AKS cluster retrieved successfully")
	}
}

// DeleteCluster handles AKS cluster deletion using decorator pattern
func (h *AzureHandler) DeleteCluster(c *gin.Context) {
	handler := h.Compose(
		h.deleteClusterHandler(),
		h.StandardCRUDDecorators("delete_cluster")...,
	)

	handler(c)
}

func (h *AzureHandler) deleteClusterHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
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
		location := h.parseRegion(c)
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}

		if clusterName == "" || location == "" {
			return
		}

		h.logClusterDeletionAttempt(c, userID, clusterName, credential.ID, location)

		ctx := h.EnrichContextWithRequestMetadata(c)
		if err := h.GetK8sService().DeleteEKSCluster(ctx, credential, clusterName, location); err != nil {
			h.HandleError(c, err, "delete_cluster")
			return
		}

		h.logClusterDeletionSuccess(c, userID, clusterName)
		h.OK(c, nil, "AKS cluster deletion initiated")
	}
}

// GetKubeconfig handles getting kubeconfig for AKS cluster using decorator pattern
func (h *AzureHandler) GetKubeconfig(c *gin.Context) {
	handler := h.Compose(
		h.getKubeconfigHandler(),
		h.StandardCRUDDecorators("get_kubeconfig")...,
	)

	handler(c)
}

func (h *AzureHandler) getKubeconfigHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
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
		location := h.parseRegion(c)
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}

		if clusterName == "" || location == "" {
			return
		}

		h.logKubeconfigGetAttempt(c, userID, clusterName, credential.ID, location)

		kubeconfig, err := h.GetK8sService().GetEKSKubeconfig(c.Request.Context(), credential, clusterName, location)
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

// ListNodeGroups handles listing node groups for a cluster using decorator pattern
func (h *AzureHandler) ListNodeGroups(c *gin.Context) {
	handler := h.Compose(
		h.listNodeGroupsHandler(),
		h.StandardCRUDDecorators("list_node_groups")...,
	)

	handler(c)
}

func (h *AzureHandler) listNodeGroupsHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
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
		location := h.parseRegion(c)
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}

		if clusterName == "" || location == "" {
			return
		}

		h.logNodeGroupsListAttempt(c, userID, clusterName, credential.ID, location)

		req := kubernetesservice.ListNodeGroupsRequest{
			CredentialID: credential.ID.String(),
			ClusterName:  clusterName,
			Region:       location,
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

// GetNodeGroup handles getting node group details using decorator pattern
func (h *AzureHandler) GetNodeGroup(c *gin.Context) {
	handler := h.Compose(
		h.getNodeGroupHandler(),
		h.StandardCRUDDecorators("get_node_group")...,
	)

	handler(c)
}

func (h *AzureHandler) getNodeGroupHandler() handlers.HandlerFunc {
	return func(c *gin.Context) {
		credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
		if err != nil {
			h.HandleError(c, err, "get_node_group")
			return
		}

		clusterName := h.parseClusterName(c)
		location := h.parseRegion(c)
		if location == "" {
			location = c.Query("location") // Azure uses "location" instead of "region"
		}
		nodeGroupName := c.Param("nodegroup")
		if nodeGroupName == "" {
			nodeGroupName = c.Param("node_group_name")
		}

		if clusterName == "" || location == "" || nodeGroupName == "" {
			h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name, location, and node group name are required", 400), "get_node_group")
			return
		}

		req := kubernetesservice.GetNodeGroupRequest{
			CredentialID:  credential.ID.String(),
			ClusterName:   clusterName,
			NodeGroupName: nodeGroupName,
			Region:        location,
		}

		nodeGroup, err := h.GetK8sService().GetNodeGroup(c.Request.Context(), credential, req)
		if err != nil {
			h.HandleError(c, err, "get_node_group")
			return
		}

		h.OK(c, nodeGroup, "Node group retrieved successfully")
	}
}

// CreateNodePool: 노드 풀 생성을 처리합니다
func (h *AzureHandler) CreateNodePool(c *gin.Context) {
	h.NotImplemented(c, "create_node_pool")
}

// ListNodePools: 노드 풀 목록 조회를 처리합니다
func (h *AzureHandler) ListNodePools(c *gin.Context) {
	h.ListNodeGroups(c) // Azure uses the same endpoint
}

// GetNodePool: 노드 풀 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetNodePool(c *gin.Context) {
	h.GetNodeGroup(c) // Azure uses the same endpoint
}

// DeleteNodePool: 노드 풀 삭제를 처리합니다
func (h *AzureHandler) DeleteNodePool(c *gin.Context) {
	h.NotImplemented(c, "delete_node_pool")
}

// ScaleNodePool: 노드 풀 스케일링을 처리합니다
func (h *AzureHandler) ScaleNodePool(c *gin.Context) {
	h.NotImplemented(c, "scale_node_pool")
}

// CreateNodeGroup: 노드 그룹 생성을 처리합니다
func (h *AzureHandler) CreateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "create_node_group")
}

// DeleteNodeGroup: 노드 그룹 삭제를 처리합니다
// UpdateNodeGroup handles updating a node group
func (h *AzureHandler) UpdateNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "update_node_group")
}

func (h *AzureHandler) DeleteNodeGroup(c *gin.Context) {
	h.NotImplemented(c, "delete_node_group")
}

// UpgradeCluster: 클러스터 업그레이드를 처리합니다
func (h *AzureHandler) UpgradeCluster(c *gin.Context) {
	h.NotImplemented(c, "upgrade_cluster")
}

// GetUpgradeStatus: 클러스터 업그레이드 상태 조회를 처리합니다
func (h *AzureHandler) GetUpgradeStatus(c *gin.Context) {
	h.NotImplemented(c, "get_upgrade_status")
}

// ListNodes: 클러스터 노드 목록 조회를 처리합니다
func (h *AzureHandler) ListNodes(c *gin.Context) {
	h.NotImplemented(c, "list_nodes")
}

// GetNode: 노드 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetNode(c *gin.Context) {
	h.NotImplemented(c, "get_node")
}

// DrainNode: 노드 드레인을 처리합니다
func (h *AzureHandler) DrainNode(c *gin.Context) {
	h.NotImplemented(c, "drain_node")
}

// CordonNode: 노드 코돈을 처리합니다
func (h *AzureHandler) CordonNode(c *gin.Context) {
	h.NotImplemented(c, "cordon_node")
}

// UncordonNode: 노드 언코돈을 처리합니다
func (h *AzureHandler) UncordonNode(c *gin.Context) {
	h.NotImplemented(c, "uncordon_node")
}

// GetNodeSSHConfig: 노드 SSH 설정 조회를 처리합니다
func (h *AzureHandler) GetNodeSSHConfig(c *gin.Context) {
	h.NotImplemented(c, "get_node_ssh_config")
}

// ExecuteNodeCommand: 노드에서 명령 실행을 처리합니다
func (h *AzureHandler) ExecuteNodeCommand(c *gin.Context) {
	h.NotImplemented(c, "execute_node_command")
}

// GetEKSVersions: EKS 버전 목록 조회를 처리합니다 (AWS 전용)
func (h *AzureHandler) GetEKSVersions(c *gin.Context) {
	h.NotImplemented(c, "get_eks_versions")
}

// GetAWSRegions: AWS 리전 목록 조회를 처리합니다 (AWS 전용)
func (h *AzureHandler) GetAWSRegions(c *gin.Context) {
	h.NotImplemented(c, "get_aws_regions")
}

// GetAvailabilityZones: 가용 영역 목록 조회를 처리합니다 (AWS 전용)
func (h *AzureHandler) GetAvailabilityZones(c *gin.Context) {
	h.NotImplemented(c, "get_availability_zones")
}

// GetInstanceTypes: 인스턴스 유형 목록 조회를 처리합니다 (AWS 전용)
func (h *AzureHandler) GetInstanceTypes(c *gin.Context) {
	h.NotImplemented(c, "get_instance_types")
}

// GetEKSAmitTypes: EKS AMI 유형 목록 조회를 처리합니다 (AWS 전용)
func (h *AzureHandler) GetEKSAmitTypes(c *gin.Context) {
	h.NotImplemented(c, "get_eks_ami_types")
}

// CheckGPUQuota: GPU 할당량 확인을 처리합니다 (AWS 전용)
func (h *AzureHandler) CheckGPUQuota(c *gin.Context) {
	h.NotImplemented(c, "check_gpu_quota")
}

// Logging helper methods

func (h *AzureHandler) logClusterCreationAttempt(c *gin.Context, userID uuid.UUID, req kubernetesservice.CreateAKSClusterRequest) {
	h.LogBusinessEvent(c, "cluster_creation_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "create_cluster",
		"cluster_name":  req.Name,
		"provider":      domain.ProviderAzure,
		"credential_id": req.CredentialID,
	})
}

func (h *AzureHandler) logClusterCreationSuccess(c *gin.Context, userID uuid.UUID, cluster interface{}) {
	h.LogBusinessEvent(c, "cluster_created", userID.String(), "", map[string]interface{}{
		"operation": "create_cluster",
		"provider":  domain.ProviderAzure,
	})
}

func (h *AzureHandler) logClusterListAttempt(c *gin.Context, userID uuid.UUID, credentialID uuid.UUID, location string) {
	h.LogBusinessEvent(c, "cluster_list_attempted", userID.String(), "", map[string]interface{}{
		"operation":     "list_clusters",
		"provider":      domain.ProviderAzure,
		"credential_id": credentialID.String(),
		"location":      location,
	})
}

func (h *AzureHandler) logClusterListSuccess(c *gin.Context, userID uuid.UUID, count int) {
	h.LogBusinessEvent(c, "clusters_listed", userID.String(), "", map[string]interface{}{
		"operation": "list_clusters",
		"provider":  domain.ProviderAzure,
		"count":     count,
	})
}

func (h *AzureHandler) logClusterGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, location string) {
	h.LogBusinessEvent(c, "cluster_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_cluster",
		"provider":      domain.ProviderAzure,
		"credential_id": credentialID.String(),
		"location":      location,
	})
}

func (h *AzureHandler) logClusterGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_cluster",
		"provider":  domain.ProviderAzure,
	})
}

func (h *AzureHandler) logClusterDeletionAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, location string) {
	h.LogBusinessEvent(c, "cluster_deletion_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "delete_cluster",
		"provider":      domain.ProviderAzure,
		"credential_id": credentialID.String(),
		"location":      location,
	})
}

func (h *AzureHandler) logClusterDeletionSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "cluster_deleted", userID.String(), clusterName, map[string]interface{}{
		"operation": "delete_cluster",
		"provider":  domain.ProviderAzure,
	})
}

func (h *AzureHandler) logKubeconfigGetAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, location string) {
	h.LogBusinessEvent(c, "kubeconfig_get_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "get_kubeconfig",
		"provider":      domain.ProviderAzure,
		"credential_id": credentialID.String(),
		"location":      location,
	})
}

func (h *AzureHandler) logKubeconfigGetSuccess(c *gin.Context, userID uuid.UUID, clusterName string) {
	h.LogBusinessEvent(c, "kubeconfig_retrieved", userID.String(), clusterName, map[string]interface{}{
		"operation": "get_kubeconfig",
		"provider":  domain.ProviderAzure,
	})
}

func (h *AzureHandler) logNodeGroupsListAttempt(c *gin.Context, userID uuid.UUID, clusterName string, credentialID uuid.UUID, location string) {
	h.LogBusinessEvent(c, "node_groups_list_attempted", userID.String(), clusterName, map[string]interface{}{
		"operation":     "list_node_groups",
		"provider":      domain.ProviderAzure,
		"credential_id": credentialID.String(),
		"location":      location,
	})
}

func (h *AzureHandler) logNodeGroupsListSuccess(c *gin.Context, userID uuid.UUID, clusterName string, count int) {
	h.LogBusinessEvent(c, "node_groups_listed", userID.String(), clusterName, map[string]interface{}{
		"operation": "list_node_groups",
		"provider":  domain.ProviderAzure,
		"count":     count,
	})
}
