package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"

	"google.golang.org/api/container/v1"
)

// GCP Kubernetes Functions
// All GCP-specific Kubernetes operations are implemented in this file

func (s *Service) createGCPGKEClusterWithAdvanced(ctx context.Context, credential *domain.Credential, req CreateGKEClusterRequest) (*CreateClusterResponse, error) {
	// Create GCP Container service
	containerService, projectID, err := s.setupGCPContainerService(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Use projectID from credential if req.ProjectID is not provided
	if req.ProjectID == "" {
		req.ProjectID = projectID
	}

	// Determine cluster type
	clusterType := "standard" // Default to standard
	if req.ClusterMode != nil && req.ClusterMode.Type == "autopilot" {
		clusterType = "autopilot"
	}

	// Build cluster configuration
	clusterConfig := &container.Cluster{
		Name:                  req.Name,
		InitialClusterVersion: req.Version,
	}

	// Network configuration
	if req.Network != nil {
		clusterConfig.Network = req.Network.VPCID
		clusterConfig.Subnetwork = req.Network.SubnetID
		// Note: Network configuration details will be set through IP allocation policy
		if req.Network.PodCIDR != "" || req.Network.ServiceCIDR != "" {
			clusterConfig.IpAllocationPolicy = &container.IPAllocationPolicy{
				ClusterIpv4CidrBlock:  req.Network.PodCIDR,
				ServicesIpv4CidrBlock: req.Network.ServiceCIDR,
				UseIpAliases:          true,
			}
		}
		if req.Network.PrivateNodes || req.Network.PrivateEndpoint {
			clusterConfig.PrivateClusterConfig = &container.PrivateClusterConfig{
				EnablePrivateNodes:    req.Network.PrivateNodes,
				EnablePrivateEndpoint: req.Network.PrivateEndpoint,
			}
		}
		if len(req.Network.MasterAuthorizedNetworks) > 0 {
			var cidrBlocks []*container.CidrBlock
			for _, network := range req.Network.MasterAuthorizedNetworks {
				cidrBlocks = append(cidrBlocks, &container.CidrBlock{
					CidrBlock: network,
				})
			}
			clusterConfig.MasterAuthorizedNetworksConfig = &container.MasterAuthorizedNetworksConfig{
				Enabled:    true,
				CidrBlocks: cidrBlocks,
			}
		}
	} else {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "network configuration is required", 400)
	}

	// Node pool configuration for standard clusters
	if clusterType == "standard" {
		if req.NodePool == nil {
			return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "node pool configuration is required for standard clusters", 400)
		}
		clusterConfig.NodePools = []*container.NodePool{
			s.buildNodePoolConfig(req.NodePool),
		}
	}

	// Security configuration
	if req.Security != nil {
		clusterConfig.WorkloadIdentityConfig = &container.WorkloadIdentityConfig{
			WorkloadPool: fmt.Sprintf("%s.svc.id.goog", projectID),
		}
		clusterConfig.NetworkPolicy = &container.NetworkPolicy{
			Enabled: req.Security.NetworkPolicy,
		}
		clusterConfig.BinaryAuthorization = &container.BinaryAuthorization{
			Enabled: req.Security.BinaryAuthorization,
		}
	}

	// Add tags (convert to GCP format)
	if len(req.Tags) > 0 {
		gcpTags := make(map[string]string)
		for key, value := range req.Tags {
			gcpKey := convertToGCPTagKey(key)
			gcpTags[gcpKey] = value
		}
		clusterConfig.ResourceLabels = gcpTags
	}

	// Create cluster request
	createRequest := &container.CreateClusterRequest{
		Cluster: clusterConfig,
	}

	// Determine location (zone or region)
	location := req.Region
	if req.Zone != "" {
		location = req.Zone
	}

	// Create cluster
	_, err = containerService.Projects.Locations.Clusters.Create(
		fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		createRequest,
	).Context(ctx).Do()

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP GKE cluster: %v", err), 502)
	}

	s.logger.Info(ctx, "GCP GKE cluster creation initiated",
		domain.NewLogField("cluster_name", req.Name),
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("location", location),
		domain.NewLogField("cluster_type", clusterType))

	// Build response
	response := &CreateClusterResponse{
		ClusterID: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, req.Name),
		Name:      req.Name,
		Version:   req.Version,
		Region:    req.Region,
		Zone:      req.Zone,
		Status:    "creating",
		ProjectID: projectID,
		Tags:      req.Tags,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 캐시 무효화: 클러스터 목록 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		cacheKey := buildKubernetesClusterListKey(credential.Provider, credentialID, req.Region)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate Kubernetes cluster list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", req.Region),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: 클러스터 생성 이벤트
	if s.eventService != nil {
		clusterData := map[string]interface{}{
			"cluster_id":    response.ClusterID,
			"name":          response.Name,
			"version":       response.Version,
			"status":        response.Status,
			"region":        response.Region,
			"provider":      credential.Provider,
			"credential_id": credentialID,
		}
		eventType := fmt.Sprintf("kubernetes.cluster.%s.created", credential.Provider)
		if err := s.eventService.Publish(ctx, eventType, clusterData); err != nil {
			s.logger.Warn(ctx, "Failed to publish Kubernetes cluster created event",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("cluster_name", req.Name),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionKubernetesClusterCreate,
		fmt.Sprintf("POST /api/v1/%s/kubernetes/clusters", credential.Provider),
		map[string]interface{}{
			"cluster_id":    response.ClusterID,
			"cluster_name":  response.Name,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        response.Region,
			"version":       response.Version,
		},
	)

	return response, nil
}

func (s *Service) buildNodePoolConfig(nodePool *GKENodePoolConfig) *container.NodePool {
	if nodePool == nil {
		return nil
	}

	config := &container.NodePool{
		Name: nodePool.Name,
		Config: &container.NodeConfig{
			MachineType: nodePool.MachineType,
			DiskSizeGb:  int64(nodePool.DiskSizeGB),
			DiskType:    nodePool.DiskType,
			Labels:      nodePool.Labels,
		},
		InitialNodeCount: int64(nodePool.NodeCount),
	}

	// Add auto scaling configuration
	if nodePool.AutoScaling != nil && nodePool.AutoScaling.Enabled {
		config.Autoscaling = &container.NodePoolAutoscaling{
			Enabled:      true,
			MinNodeCount: int64(nodePool.AutoScaling.MinNodeCount),
			MaxNodeCount: int64(nodePool.AutoScaling.MaxNodeCount),
		}
	}

	// Add preemptible/spot configuration
	if nodePool.Preemptible {
		config.Config.Preemptible = true
	}
	if nodePool.Spot {
		config.Config.Spot = true
	}

	return config
}

func convertToGCPTagKey(key string) string {
	if key == "" {
		return key
	}

	// Convert to lowercase and replace invalid characters
	result := strings.ToLower(key)

	// Replace spaces and other invalid characters with underscores
	result = strings.ReplaceAll(result, " ", "_")
	result = strings.ReplaceAll(result, ".", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	result = strings.ReplaceAll(result, ":", "_")
	result = strings.ReplaceAll(result, ";", "_")
	result = strings.ReplaceAll(result, "=", "_")
	result = strings.ReplaceAll(result, "+", "_")
	result = strings.ReplaceAll(result, "*", "_")
	result = strings.ReplaceAll(result, "?", "_")
	result = strings.ReplaceAll(result, "!", "_")
	result = strings.ReplaceAll(result, "@", "_")
	result = strings.ReplaceAll(result, "#", "_")
	result = strings.ReplaceAll(result, "$", "_")
	result = strings.ReplaceAll(result, "%", "_")
	result = strings.ReplaceAll(result, "^", "_")
	result = strings.ReplaceAll(result, "&", "_")
	result = strings.ReplaceAll(result, "(", "_")
	result = strings.ReplaceAll(result, ")", "_")
	result = strings.ReplaceAll(result, "[", "_")
	result = strings.ReplaceAll(result, "]", "_")
	result = strings.ReplaceAll(result, "{", "_")
	result = strings.ReplaceAll(result, "}", "_")
	result = strings.ReplaceAll(result, "|", "_")
	result = strings.ReplaceAll(result, "~", "_")
	result = strings.ReplaceAll(result, "`", "_")

	// Ensure it starts with a letter
	if len(result) > 0 && !isLetter(result[0]) {
		result = "tag_" + result
	}

	// Remove consecutive underscores
	for strings.Contains(result, "__") {
		result = strings.ReplaceAll(result, "__", "_")
	}

	// Remove leading/trailing underscores
	result = strings.Trim(result, "_")

	// Ensure it's not empty
	if result == "" {
		result = "tag"
	}

	return result
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func (s *Service) listGCPGKENodePools(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Find the cluster first to determine its actual location
	locations := s.getGCPLocations(req.Region)

	var nodePools *container.ListNodePoolsResponse
	var clusterLocation string

	// Search for the cluster in all possible locations
	for _, location := range locations {
		clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, req.ClusterName)
		_, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Context(ctx).Do()
		if err != nil {
			// Log debug but continue with other locations
			s.logger.Debug(ctx, "Failed to find cluster in location",
				domain.NewLogField("location", location),
				domain.NewLogField("cluster_name", req.ClusterName),
				domain.NewLogField("error", err))
			continue
		}

		// Found the cluster, now get its node pools
		clusterLocation = location
		nodePools, err = containerService.Projects.Locations.Clusters.NodePools.List(clusterPath).Context(ctx).Do()
		if err != nil {
			return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list GKE node pools for cluster %s in location %s: %v", req.ClusterName, location, err), 502)
		}
		break
	}

	if nodePools == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("failed to find GKE cluster %s in region %s or any of its zones", req.ClusterName, req.Region), 404)
	}

	// Convert to NodeGroupInfo format
	var nodeGroups []NodeGroupInfo
	for _, nodePool := range nodePools.NodePools {
		nodeGroup := s.convertGCPNodePoolToNodeGroupInfo(nodePool, req.ClusterName, req.Region)
		nodeGroups = append(nodeGroups, nodeGroup)
	}

	s.logger.Info(ctx, "GKE node pools listed successfully",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("cluster_location", clusterLocation),
		domain.NewLogField("node_pool_count", len(nodeGroups)))

	return &ListNodeGroupsResponse{
		NodeGroups: nodeGroups,
		Total:      len(nodeGroups),
	}, nil
}

func (s *Service) getGCPGKENodePool(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Find the cluster first to determine its actual location
	locations := s.getGCPLocations(req.Region)

	var nodePool *container.NodePool
	var clusterLocation string

	// Search for the cluster and specific node pool in all possible locations
	for _, location := range locations {
		nodePoolPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", projectID, location, req.ClusterName, req.NodeGroupName)

		// Try to get the specific node pool
		nodePoolResp, err := containerService.Projects.Locations.Clusters.NodePools.Get(nodePoolPath).Context(ctx).Do()
		if err != nil {
			// Log debug but continue with other locations
			s.logger.Debug(ctx, "Failed to find node pool in location",
				domain.NewLogField("location", location),
				domain.NewLogField("cluster_name", req.ClusterName),
				domain.NewLogField("node_pool_name", req.NodeGroupName),
				domain.NewLogField("error", err))
			continue
		}

		nodePool = nodePoolResp
		clusterLocation = location
		break
	}

	if nodePool == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("failed to find GKE node pool %s in cluster %s in region %s or any of its zones", req.NodeGroupName, req.ClusterName, req.Region), 404)
	}

	// Convert to NodeGroupInfo format (reusing common conversion function)
	nodeGroup := s.convertGCPNodePoolToNodeGroupInfo(nodePool, req.ClusterName, req.Region)

	s.logger.Info(ctx, "GKE node pool retrieved successfully",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("node_pool_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("cluster_location", clusterLocation))

	return &nodeGroup, nil
}

func (s *Service) listGCPGKEClusters(ctx context.Context, credential *domain.Credential, region string) (*ListClustersResponse, error) {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return nil, err
	}

	// List clusters in the region and all zones of the specified region
	locations := s.getGCPLocations(region)

	var allClusters []ClusterInfo

	for _, location := range locations {
		// List clusters in each location (region or zone)
		clustersResp, err := containerService.Projects.Locations.Clusters.List(
			fmt.Sprintf("projects/%s/locations/%s", projectID, location),
		).Context(ctx).Do()
		if err != nil {
			// Log warning but continue with other locations
			s.logger.Warn(ctx, "Failed to list clusters in location",
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
			continue
		}

		// Convert to ClusterInfo
		// clustersResp.Clusters가 nil인 경우 처리
		if clustersResp.Clusters == nil {
			s.logger.Debug(ctx, "No clusters found in location",
				domain.NewLogField("location", location))
			continue
		}

		for _, cluster := range clustersResp.Clusters {
			// Determine if this is a region or zone
			var clusterZone string
			if location == region {
				// Region level cluster - extract zone from cluster location
				clusterZone = extractZoneFromLocation(cluster.Location)
			} else {
				// Zone level cluster - use the location as zone
				clusterZone = location
			}

			// Debug logging
			s.logger.Debug(ctx, "Processing GKE cluster",
				domain.NewLogField("cluster_name", cluster.Name),
				domain.NewLogField("cluster_location", cluster.Location),
				domain.NewLogField("query_location", location),
				domain.NewLogField("zone", clusterZone))

			// Build detailed cluster information
			clusterInfo := ClusterInfo{
				ID:        cluster.Name,
				Name:      cluster.Name,
				Version:   cluster.CurrentMasterVersion,
				Status:    cluster.Status,
				Region:    region,
				Zone:      clusterZone,
				Endpoint:  cluster.Endpoint,
				CreatedAt: cluster.CreateTime,
				UpdatedAt: "", // UpdateTime field doesn't exist in GCP SDK
				Tags:      cluster.ResourceLabels,
			}

			// Add network configuration (simplified)
			if cluster.NetworkConfig != nil {
				clusterInfo.NetworkConfig = &NetworkConfigInfo{
					VPCID:           cluster.NetworkConfig.Network,
					SubnetID:        cluster.NetworkConfig.Subnetwork,
					PodCIDR:         "",    // Will be populated if available
					ServiceCIDR:     "",    // Will be populated if available
					PrivateNodes:    false, // Will be populated if available
					PrivateEndpoint: false, // Will be populated if available
				}
			}

			// Add node pool summary
			if len(cluster.NodePools) > 0 {
				var totalNodes, minNodes, maxNodes int32
				for _, nodePool := range cluster.NodePools {
					totalNodes += int32(nodePool.InitialNodeCount)
					if nodePool.Autoscaling != nil {
						minNodes += int32(nodePool.Autoscaling.MinNodeCount)
						maxNodes += int32(nodePool.Autoscaling.MaxNodeCount)
					}
				}

				clusterInfo.NodePoolInfo = &NodePoolSummaryInfo{
					TotalNodePools: int32(len(cluster.NodePools)),
					TotalNodes:     totalNodes,
					MinNodes:       minNodes,
					MaxNodes:       maxNodes,
				}
			}

			// Add security configuration
			if cluster.WorkloadIdentityConfig != nil || cluster.BinaryAuthorization != nil || cluster.NetworkPolicy != nil {
				clusterInfo.SecurityConfig = &SecurityConfigInfo{
					WorkloadIdentity:    cluster.WorkloadIdentityConfig != nil,
					BinaryAuthorization: cluster.BinaryAuthorization != nil,
					NetworkPolicy:       cluster.NetworkPolicy != nil,
					PodSecurityPolicy:   false, // Deprecated in newer versions
				}
			}

			allClusters = append(allClusters, clusterInfo)
		}
	}

	s.logger.Info(ctx, "GCP GKE clusters listed successfully",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("region", region),
		domain.NewLogField("count", len(allClusters)))

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환 보장
	if allClusters == nil {
		allClusters = []ClusterInfo{}
	}

	return &ListClustersResponse{Clusters: allClusters}, nil
}

func (s *Service) getGCPGKECluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*ClusterInfo, error) {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Get cluster details
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, region, clusterName)
	cluster, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Context(ctx).Do()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to get GKE cluster: %v", err), 502)
	}

	// Convert to ClusterInfo
	clusterZone := extractZoneFromLocation(cluster.Location)

	// Build detailed cluster information
	clusterInfo := ClusterInfo{
		ID:        cluster.Name,
		Name:      cluster.Name,
		Version:   cluster.CurrentMasterVersion,
		Status:    cluster.Status,
		Region:    region,
		Zone:      clusterZone,
		Endpoint:  cluster.Endpoint,
		CreatedAt: cluster.CreateTime,
		UpdatedAt: "", // UpdateTime field doesn't exist in GCP SDK
		Tags:      cluster.ResourceLabels,
	}

	// Add network configuration (simplified)
	if cluster.NetworkConfig != nil {
		clusterInfo.NetworkConfig = &NetworkConfigInfo{
			VPCID:           cluster.NetworkConfig.Network,
			SubnetID:        cluster.NetworkConfig.Subnetwork,
			PodCIDR:         "",    // Will be populated if available
			ServiceCIDR:     "",    // Will be populated if available
			PrivateNodes:    false, // Will be populated if available
			PrivateEndpoint: false, // Will be populated if available
		}
	}

	// Add node pool summary
	if len(cluster.NodePools) > 0 {
		var totalNodes, minNodes, maxNodes int32
		for _, nodePool := range cluster.NodePools {
			totalNodes += int32(nodePool.InitialNodeCount)
			if nodePool.Autoscaling != nil {
				minNodes += int32(nodePool.Autoscaling.MinNodeCount)
				maxNodes += int32(nodePool.Autoscaling.MaxNodeCount)
			}
		}

		clusterInfo.NodePoolInfo = &NodePoolSummaryInfo{
			TotalNodePools: int32(len(cluster.NodePools)),
			TotalNodes:     totalNodes,
			MinNodes:       minNodes,
			MaxNodes:       maxNodes,
		}
	}

	// Add security configuration
	if cluster.WorkloadIdentityConfig != nil || cluster.BinaryAuthorization != nil || cluster.NetworkPolicy != nil {
		clusterInfo.SecurityConfig = &SecurityConfigInfo{
			WorkloadIdentity:    cluster.WorkloadIdentityConfig != nil,
			BinaryAuthorization: cluster.BinaryAuthorization != nil,
			NetworkPolicy:       cluster.NetworkPolicy != nil,
			PodSecurityPolicy:   false, // Deprecated in newer versions
		}
	}

	s.logger.Info(ctx, "GCP GKE cluster retrieved successfully",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("cluster_name", clusterName),
		domain.NewLogField("region", region))

	return &clusterInfo, nil
}

func extractZoneFromLocation(location string) string {
	if location == "" {
		return ""
	}

	// Split by "/" and get the last part
	parts := strings.Split(location, "/")
	if len(parts) < 2 {
		return ""
	}

	locationPart := parts[len(parts)-1]

	// If it's a zone (contains a letter at the end), return it
	// If it's a region (no letter at the end), return empty string
	if len(locationPart) > 0 && isLetter(locationPart[len(locationPart)-1]) {
		return locationPart
	}

	// If it's a region, we need to find the actual zone from the cluster
	// For now, return empty string for regions
	return ""
}

func (s *Service) getGCPLocations(region string) []string {
	return []string{
		region,                      // Region level (e.g., asia-northeast3)
		fmt.Sprintf("%s-a", region), // Zone level (e.g., asia-northeast3-a)
		fmt.Sprintf("%s-b", region), // Zone level (e.g., asia-northeast3-b)
		fmt.Sprintf("%s-c", region), // Zone level (e.g., asia-northeast3-c)
	}
}

func (s *Service) convertGCPNodePoolToNodeGroupInfo(nodePool *container.NodePool, clusterName, region string) NodeGroupInfo {
	s.logger.Debug(context.Background(), "Processing GCP NodePool",
		domain.NewLogField("name", nodePool.Name),
		domain.NewLogField("status", nodePool.Status),
		domain.NewLogField("version", nodePool.Version),
		domain.NewLogField("has_config", nodePool.Config != nil),
		domain.NewLogField("has_autoscaling", nodePool.Autoscaling != nil),
		domain.NewLogField("has_management", nodePool.Management != nil),
		domain.NewLogField("has_upgrade_settings", nodePool.UpgradeSettings != nil))

	nodeGroup := NodeGroupInfo{
		ID:          nodePool.Name,
		Name:        nodePool.Name,
		Status:      nodePool.Status,
		ClusterName: clusterName,
		Region:      region,
	}

	if nodePool.Version != "" {
		nodeGroup.Version = nodePool.Version
	}

	if nodePool.Config != nil {
		s.logger.Info(context.Background(), "NodePool Config details",
			domain.NewLogField("name", nodePool.Name),
			domain.NewLogField("machine_type", nodePool.Config.MachineType),
			domain.NewLogField("disk_size_gb", nodePool.Config.DiskSizeGb),
			domain.NewLogField("disk_type", nodePool.Config.DiskType),
			domain.NewLogField("image_type", nodePool.Config.ImageType),
			domain.NewLogField("preemptible", nodePool.Config.Preemptible),
			domain.NewLogField("spot", nodePool.Config.Spot),
			domain.NewLogField("service_account", nodePool.Config.ServiceAccount),
			domain.NewLogField("oauth_scopes_count", len(nodePool.Config.OauthScopes)),
			domain.NewLogField("tags_count", len(nodePool.Config.Tags)),
			domain.NewLogField("labels_count", len(nodePool.Config.Labels)),
			domain.NewLogField("taints_count", len(nodePool.Config.Taints)))

		s.populateNodeGroupFromConfig(&nodeGroup, nodePool.Config)
	}

	if nodePool.Autoscaling != nil {
		nodeGroup.ScalingConfig = NodeGroupScalingConfig{
			MinSize:     int32(nodePool.Autoscaling.MinNodeCount),
			MaxSize:     int32(nodePool.Autoscaling.MaxNodeCount),
			DesiredSize: int32(nodePool.InitialNodeCount),
		}
	} else {
		nodeGroup.ScalingConfig = NodeGroupScalingConfig{
			DesiredSize: int32(nodePool.InitialNodeCount),
		}
	}

	if nodePool.Config != nil && nodePool.Config.WorkloadMetadataConfig != nil {
		nodeGroup.NetworkConfig = &NodeNetworkConfig{
			EnablePrivateNodes: nodePool.Config.WorkloadMetadataConfig.Mode == "GKE_METADATA",
		}
	}

	if nodePool.Management != nil {
		nodeGroup.Management = &NodeManagement{
			AutoRepair:  nodePool.Management.AutoRepair,
			AutoUpgrade: nodePool.Management.AutoUpgrade,
		}
	}

	if nodePool.UpgradeSettings != nil {
		nodeGroup.UpgradeSettings = &UpgradeSettings{
			MaxSurge:       int32(nodePool.UpgradeSettings.MaxSurge),
			MaxUnavailable: int32(nodePool.UpgradeSettings.MaxUnavailable),
			Strategy:       nodePool.UpgradeSettings.Strategy,
		}
	}

	nodeGroup.CreatedAt = ""
	nodeGroup.UpdatedAt = ""

	return nodeGroup
}

func (s *Service) populateNodeGroupFromConfig(nodeGroup *NodeGroupInfo, config *container.NodeConfig) {
	if config.MachineType != "" {
		nodeGroup.InstanceTypes = []string{config.MachineType}
	}
	if config.DiskSizeGb > 0 {
		nodeGroup.DiskSize = int32(config.DiskSizeGb)
	}
	if config.DiskType != "" {
		nodeGroup.DiskType = config.DiskType
	}
	if config.ImageType != "" {
		nodeGroup.ImageType = config.ImageType
	}
	if config.Preemptible {
		nodeGroup.Preemptible = true
	}
	if config.Spot {
		nodeGroup.Spot = true
	}
	if config.ServiceAccount != "" {
		nodeGroup.ServiceAccount = config.ServiceAccount
	}
	if len(config.OauthScopes) > 0 {
		nodeGroup.OAuthScopes = config.OauthScopes
	}
	if len(config.Tags) > 0 {
		tags := make(map[string]string)
		for _, tag := range config.Tags {
			tags[tag] = ""
		}
		nodeGroup.Tags = tags
	}
	if len(config.Labels) > 0 {
		nodeGroup.Labels = config.Labels
	}
	if len(config.Taints) > 0 {
		var taints []NodeTaint
		for _, taint := range config.Taints {
			taints = append(taints, NodeTaint{
				Key:    taint.Key,
				Value:  taint.Value,
				Effect: taint.Effect,
			})
		}
		nodeGroup.Taints = taints
	}
}

func (s *Service) getGCPGKEKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return "", err
	}

	// Get cluster details - search in region and all zones
	locations := s.getGCPLocations(region)

	var cluster *container.Cluster

	for _, location := range locations {
		clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, clusterName)
		clusterResp, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Context(ctx).Do()
		if err != nil {
			// Log warning but continue with other locations
			s.logger.Debug(ctx, "Failed to get cluster in location",
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
			continue
		}

		// Found the cluster
		cluster = clusterResp
		break
	}

	if cluster == nil {
		return "", domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("failed to find GKE cluster %s in region %s or any of its zones", clusterName, region), 404)
	}

	// Generate kubeconfig for GCP GKE using Google's standard format
	// Format: gke_PROJECT_ID_LOCATION_CLUSTER_NAME
	clusterContextName := fmt.Sprintf("gke_%s_%s_%s", projectID, region, clusterName)
	kubeconfig := fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: gke-gcloud-auth-plugin
      installHint: Install gke-gcloud-auth-plugin for use with kubectl by following
        https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
      provideClusterInfo: true
`,
		cluster.MasterAuth.ClusterCaCertificate,
		cluster.Endpoint,
		clusterContextName,
		clusterContextName,
		clusterContextName,
		clusterContextName,
		clusterContextName,
		clusterContextName,
	)

	s.logger.Info(ctx, "GCP GKE kubeconfig generated successfully",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("cluster_name", clusterName),
		domain.NewLogField("region", region))

	return kubeconfig, nil
}

func (s *Service) deleteGCPGKECluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	containerService, projectID, err := s.getGCPContainerServiceAndProjectID(ctx, credential)
	if err != nil {
		return err
	}

	// Find cluster location - search in region and all zones
	locations := s.getGCPLocations(region)

	var clusterLocation string

	for _, location := range locations {
		clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, clusterName)
		_, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Context(ctx).Do()
		if err != nil {
			// Log warning but continue with other locations
			s.logger.Debug(ctx, "Failed to find cluster in location",
				domain.NewLogField("location", location),
				domain.NewLogField("error", err))
			continue
		}

		// Found the cluster
		clusterLocation = location
		break
	}

	if clusterLocation == "" {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("failed to find GKE cluster %s in region %s or any of its zones", clusterName, region), 404)
	}

	// Delete cluster
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, clusterLocation, clusterName)
	operation, err := containerService.Projects.Locations.Clusters.Delete(clusterPath).Context(ctx).Do()
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete GKE cluster: %v", err), 502)
	}

	s.logger.Info(ctx, "GCP GKE cluster deletion initiated",
		domain.NewLogField("project_id", projectID),
		domain.NewLogField("cluster_name", clusterName),
		domain.NewLogField("location", clusterLocation),
		domain.NewLogField("operation_name", operation.Name))

	// 캐시 무효화: 클러스터 목록 및 개별 클러스터 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildKubernetesClusterListKey(credential.Provider, credentialID, region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate Kubernetes cluster list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
		itemKey := buildKubernetesClusterItemKey(credential.Provider, credentialID, clusterName)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate Kubernetes cluster item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("cluster_name", clusterName),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: 클러스터 삭제 이벤트
	if s.eventService != nil {
		clusterData := map[string]interface{}{
			"cluster_name":  clusterName,
			"region":        region,
			"provider":      credential.Provider,
			"credential_id": credentialID,
		}
		eventType := fmt.Sprintf("kubernetes.cluster.%s.deleted", credential.Provider)
		if err := s.eventService.Publish(ctx, eventType, clusterData); err != nil {
			s.logger.Warn(ctx, "Failed to publish Kubernetes cluster deleted event",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("cluster_name", clusterName),
				domain.NewLogField("error", err))
		}
	}

	return nil
}
