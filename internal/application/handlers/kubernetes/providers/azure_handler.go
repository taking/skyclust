package providers

import (
	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// AzureHandler: Azure AKS 관련 HTTP 요청을 처리하는 핸들러
type AzureHandler struct {
	*BaseHandler
}

// NewAzureHandler: 새로운 Azure AKS 핸들러를 생성합니다
func NewAzureHandler(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *AzureHandler {
	return &AzureHandler{
		BaseHandler: NewBaseHandler(k8sService, credentialService, domain.ProviderAzure, "azure-kubernetes"),
	}
}

// CreateCluster: AKS 클러스터 생성을 처리합니다
func (h *AzureHandler) CreateCluster(c *gin.Context) {
	var req kubernetesservice.CreateAKSClusterRequest
	if err := h.ValidateRequest(c, &req); err != nil {
		h.HandleError(c, err, "create_cluster")
		return
	}

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

	h.Created(c, cluster, "AKS cluster creation initiated")
}

// ListClusters: AKS 클러스터 목록 조회를 처리합니다
func (h *AzureHandler) ListClusters(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_clusters")
		return
	}

	location := c.Query("region")
	if location == "" {
		location = c.Query("location") // Azure uses "location" instead of "region"
	}

	clusters, err := h.GetK8sService().ListEKSClusters(c.Request.Context(), credential, location)
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

	h.OK(c, clusters.Clusters, "AKS clusters retrieved successfully")
}

// GetCluster: AKS 클러스터 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetCluster(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_cluster")
		return
	}

	clusterName := h.parseClusterName(c)
	location := h.parseRegion(c)

	if clusterName == "" || location == "" {
		return
	}

	cluster, err := h.GetK8sService().GetEKSCluster(c.Request.Context(), credential, clusterName, location)
	if err != nil {
		h.HandleError(c, err, "get_cluster")
		return
	}

	h.OK(c, cluster, "AKS cluster retrieved successfully")
}

// DeleteCluster: AKS 클러스터 삭제를 처리합니다
func (h *AzureHandler) DeleteCluster(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "delete_cluster")
		return
	}

	clusterName := h.parseClusterName(c)
	location := h.parseRegion(c)

	if clusterName == "" || location == "" {
		return
	}

	err = h.GetK8sService().DeleteEKSCluster(c.Request.Context(), credential, clusterName, location)
	if err != nil {
		h.HandleError(c, err, "delete_cluster")
		return
	}

	h.OK(c, nil, "AKS cluster deletion initiated")
}

// GetKubeconfig: AKS 클러스터의 kubeconfig 조회를 처리합니다
func (h *AzureHandler) GetKubeconfig(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_kubeconfig")
		return
	}

	clusterName := h.parseClusterName(c)
	location := h.parseRegion(c)

	if clusterName == "" || location == "" {
		return
	}

	kubeconfig, err := h.GetK8sService().GetEKSKubeconfig(c.Request.Context(), credential, clusterName, location)
	if err != nil {
		h.HandleError(c, err, "get_kubeconfig")
		return
	}

	h.OK(c, map[string]string{"kubeconfig": kubeconfig}, "Kubeconfig retrieved successfully")
}

// ListNodeGroups: 노드 그룹 목록 조회를 처리합니다
func (h *AzureHandler) ListNodeGroups(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "list_node_groups")
		return
	}

	clusterName := h.parseClusterName(c)
	location := h.parseRegion(c)

	if clusterName == "" || location == "" {
		return
	}

	req := kubernetesservice.ListNodeGroupsRequest{
		ClusterName: clusterName,
		Region:      location,
	}

	nodeGroups, err := h.GetK8sService().ListNodeGroups(c.Request.Context(), credential, req)
	if err != nil {
		h.HandleError(c, err, "list_node_groups")
		return
	}

	h.OK(c, nodeGroups.NodeGroups, "Node groups retrieved successfully")
}

// GetNodeGroup: 노드 그룹 상세 정보 조회를 처리합니다
func (h *AzureHandler) GetNodeGroup(c *gin.Context) {
	credential, err := h.GetCredentialFromRequest(c, h.GetCredentialService(), domain.ProviderAzure)
	if err != nil {
		h.HandleError(c, err, "get_node_group")
		return
	}

	clusterName := h.parseClusterName(c)
	location := h.parseRegion(c)
	nodeGroupName := c.Param("node_group_name")

	if clusterName == "" || location == "" || nodeGroupName == "" {
		h.HandleError(c, domain.NewDomainError(domain.ErrCodeBadRequest, "cluster name, region, and node group name are required", 400), "get_node_group")
		return
	}

	req := kubernetesservice.GetNodeGroupRequest{
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
