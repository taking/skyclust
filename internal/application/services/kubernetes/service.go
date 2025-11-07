package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/smithy-go"
	"go.uber.org/zap"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
)

// handleAWSError: AWS SDK 에러를 적절한 도메인 에러로 변환합니다
func (s *Service) handleAWSError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// Check if it's an AWS API error
	var apiErr *smithy.OperationError
	if errors.As(err, &apiErr) {
		// Check for specific AWS error codes
		switch {
		case strings.Contains(err.Error(), "UnrecognizedClientException"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid AWS credentials", 400)
		case strings.Contains(err.Error(), "InvalidUserID.NotFound"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "AWS user not found", 400)
		case strings.Contains(err.Error(), "AccessDenied") || strings.Contains(err.Error(), "UnauthorizedOperation"):
			// Extract the specific action and user from the error message for better context
			errorMsg := err.Error()
			if strings.Contains(errorMsg, "not authorized to perform") {
				// Extract action name if available
				if strings.Contains(errorMsg, "ec2:DescribeRegions") {
					return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have permission to list AWS regions. Please add 'ec2:DescribeRegions' permission to the IAM user or role.", 403)
				}
				if strings.Contains(errorMsg, "eks:") {
					return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have permission to access EKS. Please add the required EKS permissions to the IAM user or role.", 403)
				}
				return domain.NewDomainError(domain.ErrCodeForbidden, "AWS IAM permission required: The credential does not have the required permissions. Please check your IAM policy.", 403)
			}
			return domain.NewDomainError(domain.ErrCodeForbidden, "Access denied to AWS resources", 403)
		case strings.Contains(err.Error(), "NoSuchEntity"):
			return domain.NewDomainError(domain.ErrCodeNotFound, "AWS resource not found", 404)
		case strings.Contains(err.Error(), "InvalidParameterValue"):
			return domain.NewDomainError(domain.ErrCodeBadRequest, "Invalid AWS parameter", 400)
		case strings.Contains(err.Error(), "ThrottlingException"):
			return domain.NewDomainError(domain.ErrCodeProviderQuota, "AWS API rate limit exceeded", 429)
		default:
			return domain.NewDomainError(domain.ErrCodeBadRequest, fmt.Sprintf("AWS API error: %s", err.Error()), 400)
		}
	}

	// For other errors, return as internal server error
	return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("Failed to %s: %v", operation, err), 500)
}

// Service: Kubernetes 서비스 구현체
type Service struct {
	credentialService domain.CredentialService
	cache             cache.Cache
	keyBuilder        *cache.CacheKeyBuilder
	invalidator       *cache.Invalidator
	eventPublisher    *messaging.Publisher
	auditLogRepo      domain.AuditLogRepository
	logger            *zap.Logger
}

// NewService: 새로운 Kubernetes 서비스를 생성합니다
func NewService(credentialService domain.CredentialService, cacheService cache.Cache, eventBus messaging.Bus, auditLogRepo domain.AuditLogRepository, logger *zap.Logger) *Service {
	eventPublisher := messaging.NewPublisher(eventBus, logger)
	return &Service{
		credentialService: credentialService,
		cache:             cacheService,
		keyBuilder:        cache.NewCacheKeyBuilder(),
		invalidator:       cache.NewInvalidatorWithEvents(cacheService, eventPublisher),
		eventPublisher:    eventPublisher,
		auditLogRepo:      auditLogRepo,
		logger:            logger,
	}
}

// CreateEKSCluster: AWS EKS 클러스터를 생성합니다 (하위 호환성을 위해)
func (s *Service) CreateEKSCluster(ctx context.Context, credential *domain.Credential, req CreateClusterRequest) (*CreateClusterResponse, error) {
	return s.createAWSEKSCluster(ctx, credential, req)
}

// CreateGCPGKECluster: 새로운 섹션 구조로 GCP GKE 클러스터를 생성합니다
func (s *Service) CreateGCPGKECluster(ctx context.Context, credential *domain.Credential, req CreateGKEClusterRequest) (*CreateClusterResponse, error) {
	return s.createGCPGKEClusterWithAdvanced(ctx, credential, req)
}

// createGCPGKEClusterWithAdvanced: 고급 설정으로 GCP GKE 클러스터를 생성합니다
func (s *Service) createGCPGKEClusterWithAdvanced(ctx context.Context, credential *domain.Credential, req CreateGKEClusterRequest) (*CreateClusterResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	// Create GCP Container service
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP container service: %v", err), 502)
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
			WorkloadPool: fmt.Sprintf("%s.svc.id.goog", req.ProjectID),
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
		fmt.Sprintf("projects/%s/locations/%s", req.ProjectID, location),
		createRequest,
	).Context(ctx).Do()

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP GKE cluster: %v", err), 502)
	}

	s.logger.Info("GCP GKE cluster creation initiated",
		zap.String("cluster_name", req.Name),
		zap.String("project_id", req.ProjectID),
		zap.String("location", location),
		zap.String("cluster_type", clusterType))

	// Build response
	response := &CreateClusterResponse{
		ClusterID: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", req.ProjectID, location, req.Name),
		Name:      req.Name,
		Version:   req.Version,
		Region:    req.Region,
		Zone:      req.Zone,
		Status:    "creating",
		ProjectID: req.ProjectID,
		Tags:      req.Tags,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 캐시 무효화: 클러스터 목록 캐시 삭제
	credentialID := credential.ID.String()
	if err := s.invalidator.InvalidateKubernetesClusterList(ctx, credential.Provider, credentialID, req.Region); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", req.Region),
			zap.Error(err))
	}

	// 이벤트 발행: 클러스터 생성 이벤트
	clusterData := map[string]interface{}{
		"cluster_id": response.ClusterID,
		"name":       response.Name,
		"version":    response.Version,
		"status":     response.Status,
		"region":     response.Region,
	}
	if err := s.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, req.Region, "created", clusterData); err != nil {
		s.logger.Warn("Failed to publish Kubernetes cluster created event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("cluster_name", req.Name),
			zap.Error(err))
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

// buildNodePoolConfig: GCP 노드 풀 구성을 생성합니다
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

// convertToGCPTagKey: 태그 키를 GCP 호환 형식으로 변환합니다
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

// isLetter: 문자가 알파벳인지 확인합니다
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// createAWSEKSCluster: AWS EKS 클러스터를 생성합니다
func (s *Service) createAWSEKSCluster(ctx context.Context, credential *domain.Credential, req CreateClusterRequest) (*CreateClusterResponse, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, req.Region)
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Prepare cluster input
	input := &eks.CreateClusterInput{
		Name:    aws.String(req.Name),
		Version: aws.String(req.Version),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: req.SubnetIDs,
		},
	}

	// Add optional fields
	if req.RoleARN != "" {
		input.RoleArn = aws.String(req.RoleARN)
	}

	if len(req.Tags) > 0 {
		input.Tags = req.Tags
	}

	// Add Access Entry configuration
	if req.AccessConfig != nil {
		accessConfig := &types.CreateAccessConfigRequest{}

		// Set authentication mode
		if req.AccessConfig.AuthenticationMode != "" {
			accessConfig.AuthenticationMode = types.AuthenticationMode(req.AccessConfig.AuthenticationMode)
		} else {
			// Default to API mode for Access Entries
			accessConfig.AuthenticationMode = types.AuthenticationModeApi
		}

		// Set bootstrap cluster creator admin permissions
		if req.AccessConfig.BootstrapClusterCreatorAdminPermissions != nil {
			accessConfig.BootstrapClusterCreatorAdminPermissions = req.AccessConfig.BootstrapClusterCreatorAdminPermissions
		} else {
			// Default to true for automatic admin access
			accessConfig.BootstrapClusterCreatorAdminPermissions = aws.Bool(true)
		}

		input.AccessConfig = accessConfig
	} else {
		// Default Access Entry configuration
		input.AccessConfig = &types.CreateAccessConfigRequest{
			AuthenticationMode:                      types.AuthenticationModeApi,
			BootstrapClusterCreatorAdminPermissions: aws.Bool(true),
		}
	}

	// Create cluster
	output, err := eksClient.CreateCluster(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EKS cluster: %v", err), 502)
	}

	// Convert to response
	response := &CreateClusterResponse{
		ClusterID: aws.ToString(output.Cluster.Arn),
		Name:      aws.ToString(output.Cluster.Name),
		Version:   aws.ToString(output.Cluster.Version),
		Region:    req.Region,
		Status:    string(output.Cluster.Status),
		Tags:      output.Cluster.Tags,
	}

	if output.Cluster.Endpoint != nil {
		response.Endpoint = *output.Cluster.Endpoint
	}

	if output.Cluster.CreatedAt != nil {
		response.CreatedAt = output.Cluster.CreatedAt.String()
	}

	s.logger.Info("EKS cluster creation initiated",
		zap.String("cluster_name", req.Name),
		zap.String("region", req.Region),
		zap.String("version", req.Version))

	// 캐시 무효화: 클러스터 목록 캐시 삭제
	credentialID := credential.ID.String()
	if err := s.invalidator.InvalidateKubernetesClusterList(ctx, credential.Provider, credentialID, req.Region); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", req.Region),
			zap.Error(err))
	}

	return response, nil
}

// ListEKSClusters: 모든 Kubernetes 클러스터 목록을 조회합니다 (다중 프로바이더 지원)
func (s *Service) ListEKSClusters(ctx context.Context, credential *domain.Credential, region string) (*ListClustersResponse, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := s.keyBuilder.BuildKubernetesClusterListKey(credential.Provider, credentialID, region)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedResponse ListClustersResponse
		if err := s.cache.Get(ctx, cacheKey, &cachedResponse); err == nil {
			s.logger.Debug("Kubernetes clusters retrieved from cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region))
			return &cachedResponse, nil
		}
	}

	// 캐시 미스 시 실제 API 호출
	var response *ListClustersResponse
	var err error

	switch credential.Provider {
	case "aws":
		response, err = s.listAWSEKSClusters(ctx, credential, region)
	case "gcp":
		response, err = s.listGCPGKEClusters(ctx, credential, region)
	case "azure":
		// TODO: Implement Azure AKS cluster listing
		response = &ListClustersResponse{Clusters: []ClusterInfo{}}
	case "ncp":
		// TODO: Implement NCP NKS cluster listing
		response = &ListClustersResponse{Clusters: []ClusterInfo{}}
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cache != nil && response != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceKubernetes)
		if err := s.cache.Set(ctx, cacheKey, response, ttl); err != nil {
			s.logger.Warn("Failed to cache Kubernetes clusters, continuing without cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("region", region),
				zap.Error(err))
			// 캐시 실패는 치명적이지 않으므로 계속 진행
		}
	}

	return response, nil
}

// listAWSEKSClusters: AWS EKS 클러스터 목록을 조회합니다
func (s *Service) listAWSEKSClusters(ctx context.Context, credential *domain.Credential, region string) (*ListClustersResponse, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("AWS credentials extracted",
		zap.String("access_key_prefix", creds.AccessKey[:min(10, len(creds.AccessKey))]),
		zap.Int("access_key_length", len(creds.AccessKey)),
		zap.Int("secret_key_length", len(creds.SecretKey)),
		zap.String("region", creds.Region))

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List clusters
	output, err := eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list EKS clusters: %v", err), 502)
	}

	// Get detailed information for each cluster
	var clusters []ClusterInfo
	for _, clusterName := range output.Clusters {
		// Describe each cluster to get detailed information
		describeOutput, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			s.logger.Warn("Failed to describe cluster",
				zap.String("cluster_name", clusterName),
				zap.Error(err))
			continue
		}

		cluster := ClusterInfo{
			ID:      aws.ToString(describeOutput.Cluster.Arn),
			Name:    aws.ToString(describeOutput.Cluster.Name),
			Version: aws.ToString(describeOutput.Cluster.Version),
			Status:  string(describeOutput.Cluster.Status),
			Region:  region,
			Tags:    describeOutput.Cluster.Tags,
		}

		if describeOutput.Cluster.Endpoint != nil {
			cluster.Endpoint = *describeOutput.Cluster.Endpoint
		}

		if describeOutput.Cluster.CreatedAt != nil {
			cluster.CreatedAt = describeOutput.Cluster.CreatedAt.String()
		}

		clusters = append(clusters, cluster)
	}

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환 보장
	if clusters == nil {
		clusters = []ClusterInfo{}
	}

	return &ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// GetEKSCluster: 이름으로 EKS 클러스터 상세 정보를 조회합니다
func (s *Service) GetEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*ClusterInfo, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := s.keyBuilder.BuildKubernetesClusterItemKey(credential.Provider, credentialID, clusterName)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedCluster ClusterInfo
		if err := s.cache.Get(ctx, cacheKey, &cachedCluster); err == nil {
			s.logger.Debug("Kubernetes cluster retrieved from cache",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("cluster_name", clusterName))
			return &cachedCluster, nil
		}
	}

	// 캐시 미스 시 실제 API 호출
	var cluster *ClusterInfo
	var err error

	switch credential.Provider {
	case "aws":
		cluster, err = s.getAWSEKSCluster(ctx, credential, clusterName, region)
	case "gcp":
		cluster, err = s.getGCPGKECluster(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS cluster retrieval
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure AKS cluster retrieval not implemented yet", 501)
	case "ncp":
		// TODO: Implement NCP NKS cluster retrieval
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS cluster retrieval not implemented yet", 501)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장
	if s.cache != nil && cluster != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceKubernetes)
		if err := s.cache.Set(ctx, cacheKey, cluster, ttl); err != nil {
			s.logger.Warn("Failed to cache Kubernetes cluster",
				zap.String("provider", credential.Provider),
				zap.String("credential_id", credentialID),
				zap.String("cluster_name", clusterName),
				zap.Error(err))
		}
	}

	return cluster, nil
}

// getAWSEKSCluster: AWS EKS 클러스터 상세 정보를 조회합니다
func (s *Service) getAWSEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*ClusterInfo, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe cluster
	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe EKS cluster: %v", err), 502)
	}

	// Convert to ClusterInfo
	cluster := ClusterInfo{
		ID:      aws.ToString(output.Cluster.Arn),
		Name:    aws.ToString(output.Cluster.Name),
		Version: aws.ToString(output.Cluster.Version),
		Status:  string(output.Cluster.Status),
		Region:  creds.Region,
		Tags:    output.Cluster.Tags,
	}

	if output.Cluster.Endpoint != nil {
		cluster.Endpoint = *output.Cluster.Endpoint
	}

	if output.Cluster.CreatedAt != nil {
		cluster.CreatedAt = output.Cluster.CreatedAt.String()
	}

	return &cluster, nil
}

// DeleteEKSCluster: Kubernetes 클러스터를 삭제합니다
func (s *Service) DeleteEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.deleteAWSEKSCluster(ctx, credential, clusterName, region)
	case "gcp":
		return s.deleteGCPGKECluster(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS cluster deletion
		return domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure AKS cluster deletion not implemented yet", 501)
	case "ncp":
		// TODO: Implement NCP NKS cluster deletion
		return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS cluster deletion not implemented yet", 501)
	default:
		return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}
}

// deleteAWSEKSCluster: AWS EKS 클러스터를 삭제합니다
func (s *Service) deleteAWSEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Delete cluster
	_, err = eksClient.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete EKS cluster: %v", err), 502)
	}

	s.logger.Info("EKS cluster deletion initiated",
		zap.String("cluster_name", clusterName),
		zap.String("region", region))

	// 캐시 무효화: 클러스터 목록 및 개별 클러스터 캐시 삭제
	credentialID := credential.ID.String()
	if err := s.invalidator.InvalidateKubernetesClusterList(ctx, credential.Provider, credentialID, region); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", region),
			zap.Error(err))
	}
	if err := s.invalidator.InvalidateKubernetesClusterItem(ctx, credential.Provider, credentialID, clusterName); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster item cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("cluster_name", clusterName),
			zap.Error(err))
	}

	// 이벤트 발행: 클러스터 삭제 이벤트
	clusterData := map[string]interface{}{
		"cluster_name": clusterName,
		"region":       region,
	}
	if err := s.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "deleted", clusterData); err != nil {
		s.logger.Warn("Failed to publish Kubernetes cluster deleted event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("cluster_name", clusterName),
			zap.Error(err))
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionKubernetesClusterDelete,
		fmt.Sprintf("DELETE /api/v1/%s/kubernetes/clusters/%s", credential.Provider, clusterName),
		map[string]interface{}{
			"cluster_name":  clusterName,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	return nil
}

// GetEKSKubeconfig: Kubernetes 클러스터의 kubeconfig를 생성합니다
func (s *Service) GetEKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.getAWSEKSKubeconfig(ctx, credential, clusterName, region)
	case "gcp":
		return s.getGCPGKEKubeconfig(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS kubeconfig generation
		return "", domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure AKS kubeconfig generation not implemented yet", 501)
	case "ncp":
		// TODO: Implement NCP NKS kubeconfig generation
		return "", domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS kubeconfig generation not implemented yet", 501)
	default:
		return "", domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}
}

// getAWSEKSKubeconfig: AWS EKS 클러스터의 kubeconfig를 생성합니다
func (s *Service) getAWSEKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, region)
	if err != nil {
		return "", err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return "", err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe cluster to get endpoint and CA
	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe EKS cluster: %v", err), 502)
	}

	// Generate kubeconfig
	kubeconfig := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
        - eks
        - get-token
        - --cluster-name
        - %s
        - --region
        - %s
      env:
        - name: AWS_ACCESS_KEY_ID
          value: %s
        - name: AWS_SECRET_ACCESS_KEY
          value: %s
`,
		aws.ToString(output.Cluster.CertificateAuthority.Data),
		aws.ToString(output.Cluster.Endpoint),
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		clusterName,
		creds.Region,
		creds.AccessKey,
		creds.SecretKey,
	)

	return kubeconfig, nil
}

// CreateEKSNodePool: EKS 클러스터의 노드 풀을 생성합니다
func (s *Service) CreateEKSNodePool(ctx context.Context, credential *domain.Credential, req CreateNodePoolRequest) (map[string]interface{}, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(req.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Prepare node group input
	input := &eks.CreateNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodePoolName),
		Subnets:       req.SubnetIDs,
		ScalingConfig: &types.NodegroupScalingConfig{
			MinSize:     aws.Int32(req.MinSize),
			MaxSize:     aws.Int32(req.MaxSize),
			DesiredSize: aws.Int32(req.DesiredSize),
		},
		InstanceTypes: []string{req.InstanceType},
	}

	// Add optional fields
	if req.DiskSize > 0 {
		input.DiskSize = aws.Int32(req.DiskSize)
	}

	if len(req.Tags) > 0 {
		input.Tags = req.Tags
	}

	// Create node group
	output, err := eksClient.CreateNodegroup(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create node group: %v", err), 502)
	}

	// Convert to map
	result := map[string]interface{}{
		"nodegroup_name": aws.ToString(output.Nodegroup.NodegroupName),
		"cluster_name":   aws.ToString(output.Nodegroup.ClusterName),
		"status":         string(output.Nodegroup.Status),
		"instance_types": output.Nodegroup.InstanceTypes,
		"subnets":        output.Nodegroup.Subnets,
	}

	if output.Nodegroup.CreatedAt != nil {
		result["created_at"] = output.Nodegroup.CreatedAt.String()
	}

	s.logger.Info("EKS node group creation initiated",
		zap.String("cluster_name", req.ClusterName),
		zap.String("nodegroup_name", req.NodePoolName),
		zap.String("region", req.Region))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionKubernetesNodePoolCreate,
		fmt.Sprintf("POST /api/v1/%s/kubernetes/clusters/%s/node-pools", credential.Provider, req.ClusterName),
		map[string]interface{}{
			"nodegroup_name": req.NodePoolName,
			"cluster_name":   req.ClusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		nodePoolData := map[string]interface{}{
			"nodegroup_name": req.NodePoolName,
			"cluster_name":   req.ClusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
		}
		_ = s.eventPublisher.PublishKubernetesNodePoolEvent(ctx, credential.Provider, credentialID, req.ClusterName, "created", nodePoolData)
	}

	return result, nil
}

// CreateEKSNodeGroup: EKS 노드 그룹을 생성합니다
func (s *Service) CreateEKSNodeGroup(ctx context.Context, credential *domain.Credential, req CreateNodeGroupRequest) (*CreateNodeGroupResponse, error) {
	creds, err := s.extractAWSCredentials(ctx, credential, req.Region)
	if err != nil {
		return nil, err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Prepare node group input
	input := &eks.CreateNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
		NodeRole:      aws.String(req.NodeRoleARN),
		Subnets:       req.SubnetIDs,
		InstanceTypes: req.InstanceTypes,
		ScalingConfig: &types.NodegroupScalingConfig{
			MinSize:     aws.Int32(req.ScalingConfig.MinSize),
			MaxSize:     aws.Int32(req.ScalingConfig.MaxSize),
			DesiredSize: aws.Int32(req.ScalingConfig.DesiredSize),
		},
	}

	// Add optional fields
	if req.DiskSize > 0 {
		input.DiskSize = aws.Int32(req.DiskSize)
	}

	if req.AMI != "" {
		input.AmiType = types.AMITypes(req.AMI)
	}

	if req.CapacityType != "" {
		input.CapacityType = types.CapacityTypes(req.CapacityType)
	}

	if len(req.Tags) > 0 {
		input.Tags = req.Tags
	}

	// Create node group
	output, err := eksClient.CreateNodegroup(ctx, input)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create node group: %v", err), 502)
	}

	// Convert to response
	response := &CreateNodeGroupResponse{
		NodeGroupName: aws.ToString(output.Nodegroup.NodegroupName),
		ClusterName:   aws.ToString(output.Nodegroup.ClusterName),
		Status:        string(output.Nodegroup.Status),
		InstanceTypes: output.Nodegroup.InstanceTypes,
		ScalingConfig: NodeGroupScalingConfig{
			MinSize:     aws.ToInt32(output.Nodegroup.ScalingConfig.MinSize),
			MaxSize:     aws.ToInt32(output.Nodegroup.ScalingConfig.MaxSize),
			DesiredSize: aws.ToInt32(output.Nodegroup.ScalingConfig.DesiredSize),
		},
		Tags: output.Nodegroup.Tags,
	}

	if output.Nodegroup.CreatedAt != nil {
		response.CreatedAt = output.Nodegroup.CreatedAt.String()
	}

	s.logger.Info("EKS node group creation initiated",
		zap.String("cluster_name", req.ClusterName),
		zap.String("nodegroup_name", req.NodeGroupName),
		zap.String("region", req.Region))

	return response, nil
}

// ListNodeGroups: 클러스터의 노드 그룹 목록을 조회합니다
func (s *Service) ListNodeGroups(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	s.logger.Info("ListNodeGroups called",
		zap.String("provider", credential.Provider),
		zap.String("cluster_name", req.ClusterName),
		zap.String("region", req.Region))

	switch credential.Provider {
	case "aws":
		return s.listAWSEKSNodeGroups(ctx, credential, req)
	case "gcp":
		return s.listGCPGKENodePools(ctx, credential, req)
	case "azure":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure node groups not implemented yet", 501)
	case "ncp":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP node groups not implemented yet", 501)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}
}

// listAWSEKSNodeGroups: 클러스터의 AWS EKS 노드 그룹 목록을 조회합니다
func (s *Service) listAWSEKSNodeGroups(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(req.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List node groups
	output, err := eksClient.ListNodegroups(ctx, &eks.ListNodegroupsInput{
		ClusterName: aws.String(req.ClusterName),
	})
	if err != nil {
		return nil, s.handleAWSError(err, "list node groups")
	}

	// Get detailed information for each node group
	var nodeGroups []NodeGroupInfo
	for _, nodeGroupName := range output.Nodegroups {
		describeOutput, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(req.ClusterName),
			NodegroupName: aws.String(nodeGroupName),
		})
		if err != nil {
			s.logger.Warn("Failed to describe node group",
				zap.String("cluster_name", req.ClusterName),
				zap.String("nodegroup_name", nodeGroupName),
				zap.Error(err))
			continue
		}

		nodeGroup := NodeGroupInfo{
			ID:          aws.ToString(describeOutput.Nodegroup.NodegroupArn),
			Name:        aws.ToString(describeOutput.Nodegroup.NodegroupName),
			Status:      string(describeOutput.Nodegroup.Status),
			ClusterName: aws.ToString(describeOutput.Nodegroup.ClusterName),
			Region:      req.Region,
			Tags:        describeOutput.Nodegroup.Tags,
		}

		// Add version if available
		if describeOutput.Nodegroup.Version != nil {
			nodeGroup.Version = aws.ToString(describeOutput.Nodegroup.Version)
		}

		// Add instance types
		if describeOutput.Nodegroup.InstanceTypes != nil {
			nodeGroup.InstanceTypes = describeOutput.Nodegroup.InstanceTypes
		}

		// Add scaling config
		if describeOutput.Nodegroup.ScalingConfig != nil {
			nodeGroup.ScalingConfig = NodeGroupScalingConfig{
				MinSize:     aws.ToInt32(describeOutput.Nodegroup.ScalingConfig.MinSize),
				MaxSize:     aws.ToInt32(describeOutput.Nodegroup.ScalingConfig.MaxSize),
				DesiredSize: aws.ToInt32(describeOutput.Nodegroup.ScalingConfig.DesiredSize),
			}
		}

		// Add capacity type
		if describeOutput.Nodegroup.CapacityType != "" {
			nodeGroup.CapacityType = string(describeOutput.Nodegroup.CapacityType)
		}

		// Add disk size
		if describeOutput.Nodegroup.DiskSize != nil {
			nodeGroup.DiskSize = aws.ToInt32(describeOutput.Nodegroup.DiskSize)
		}

		// Add timestamps
		if describeOutput.Nodegroup.CreatedAt != nil {
			nodeGroup.CreatedAt = describeOutput.Nodegroup.CreatedAt.String()
		}

		if describeOutput.Nodegroup.ModifiedAt != nil {
			nodeGroup.UpdatedAt = describeOutput.Nodegroup.ModifiedAt.String()
		}

		nodeGroups = append(nodeGroups, nodeGroup)
	}

	return &ListNodeGroupsResponse{
		NodeGroups: nodeGroups,
		Total:      len(nodeGroups),
	}, nil
}

// listGCPGKENodePools: 클러스터의 GCP GKE 노드 풀 목록을 조회합니다
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
			s.logger.Debug("Failed to find cluster in location",
				zap.String("location", location),
				zap.String("cluster_name", req.ClusterName),
				zap.Error(err))
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

	s.logger.Info("GKE node pools listed successfully",
		zap.String("cluster_name", req.ClusterName),
		zap.String("region", req.Region),
		zap.String("cluster_location", clusterLocation),
		zap.Int("node_pool_count", len(nodeGroups)))

	return &ListNodeGroupsResponse{
		NodeGroups: nodeGroups,
		Total:      len(nodeGroups),
	}, nil
}

// GetNodeGroup: 노드 그룹 상세 정보를 조회합니다
func (s *Service) GetNodeGroup(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
	s.logger.Info("GetNodeGroup called",
		zap.String("provider", credential.Provider),
		zap.String("cluster_name", req.ClusterName),
		zap.String("node_group_name", req.NodeGroupName),
		zap.String("region", req.Region))

	switch credential.Provider {
	case "aws":
		return s.getAWSEKSNodeGroup(ctx, credential, req)
	case "gcp":
		return s.getGCPGKENodePool(ctx, credential, req)
	case "azure":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "Azure node groups not implemented yet", 501)
	case "ncp":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP node groups not implemented yet", 501)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), 400)
	}
}

// getAWSEKSNodeGroup: AWS EKS 노드 그룹 상세 정보를 조회합니다
func (s *Service) getAWSEKSNodeGroup(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(req.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe node group
	output, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe node group: %v", err), 502)
	}

	// Convert to NodeGroupInfo
	nodeGroup := NodeGroupInfo{
		ID:          aws.ToString(output.Nodegroup.NodegroupArn),
		Name:        aws.ToString(output.Nodegroup.NodegroupName),
		Status:      string(output.Nodegroup.Status),
		ClusterName: aws.ToString(output.Nodegroup.ClusterName),
		Region:      req.Region,
		Tags:        output.Nodegroup.Tags,
	}

	// Add version if available
	if output.Nodegroup.Version != nil {
		nodeGroup.Version = aws.ToString(output.Nodegroup.Version)
	}

	// Add instance types
	if output.Nodegroup.InstanceTypes != nil {
		nodeGroup.InstanceTypes = output.Nodegroup.InstanceTypes
	}

	// Add scaling config
	if output.Nodegroup.ScalingConfig != nil {
		nodeGroup.ScalingConfig = NodeGroupScalingConfig{
			MinSize:     aws.ToInt32(output.Nodegroup.ScalingConfig.MinSize),
			MaxSize:     aws.ToInt32(output.Nodegroup.ScalingConfig.MaxSize),
			DesiredSize: aws.ToInt32(output.Nodegroup.ScalingConfig.DesiredSize),
		}
	}

	// Add capacity type
	if output.Nodegroup.CapacityType != "" {
		nodeGroup.CapacityType = string(output.Nodegroup.CapacityType)
	}

	// Add disk size
	if output.Nodegroup.DiskSize != nil {
		nodeGroup.DiskSize = aws.ToInt32(output.Nodegroup.DiskSize)
	}

	// Add timestamps
	if output.Nodegroup.CreatedAt != nil {
		nodeGroup.CreatedAt = output.Nodegroup.CreatedAt.String()
	}

	if output.Nodegroup.ModifiedAt != nil {
		nodeGroup.UpdatedAt = output.Nodegroup.ModifiedAt.String()
	}

	s.logger.Info("EKS node group retrieved successfully",
		zap.String("cluster_name", req.ClusterName),
		zap.String("node_group_name", req.NodeGroupName),
		zap.String("region", req.Region))

	return &nodeGroup, nil
}

// getGCPGKENodePool: GCP GKE 노드 풀 상세 정보를 조회합니다
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
			s.logger.Debug("Failed to find node pool in location",
				zap.String("location", location),
				zap.String("cluster_name", req.ClusterName),
				zap.String("node_pool_name", req.NodeGroupName),
				zap.Error(err))
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

	s.logger.Info("GKE node pool retrieved successfully",
		zap.String("cluster_name", req.ClusterName),
		zap.String("node_pool_name", req.NodeGroupName),
		zap.String("region", req.Region),
		zap.String("cluster_location", clusterLocation))

	return &nodeGroup, nil
}

// DeleteNodeGroup: 노드 그룹을 삭제합니다
func (s *Service) DeleteNodeGroup(ctx context.Context, credential *domain.Credential, req DeleteNodeGroupRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", 400)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", 400)
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(req.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), 502)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Delete node group
	_, err = eksClient.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete node group: %v", err), 502)
	}

	s.logger.Info("EKS node group deletion initiated",
		zap.String("cluster_name", req.ClusterName),
		zap.String("nodegroup_name", req.NodeGroupName),
		zap.String("region", req.Region))

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionKubernetesNodeGroupDelete,
		fmt.Sprintf("DELETE /api/v1/%s/kubernetes/clusters/%s/node-groups/%s", credential.Provider, req.ClusterName, req.NodeGroupName),
		map[string]interface{}{
			"nodegroup_name": req.NodeGroupName,
			"cluster_name":   req.ClusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		nodePoolData := map[string]interface{}{
			"nodegroup_name": req.NodeGroupName,
			"cluster_name":   req.ClusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
		}
		_ = s.eventPublisher.PublishKubernetesNodePoolEvent(ctx, credential.Provider, credentialID, req.ClusterName, "deleted", nodePoolData)
	}

	return nil
}

// listGCPGKEClusters: GCP GKE 클러스터 목록을 조회합니다
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
			s.logger.Warn("Failed to list clusters in location",
				zap.String("location", location),
				zap.Error(err))
			continue
		}

		// Convert to ClusterInfo
		// clustersResp.Clusters가 nil인 경우 처리
		if clustersResp.Clusters == nil {
			s.logger.Debug("No clusters found in location",
				zap.String("location", location))
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
			s.logger.Debug("Processing GKE cluster",
				zap.String("cluster_name", cluster.Name),
				zap.String("cluster_location", cluster.Location),
				zap.String("query_location", location),
				zap.String("zone", clusterZone))

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

	s.logger.Info("GCP GKE clusters listed successfully",
		zap.String("project_id", projectID),
		zap.String("region", region),
		zap.Int("count", len(allClusters)))

	// 빈 배열인 경우에도 nil이 아닌 빈 슬라이스 반환 보장
	if allClusters == nil {
		allClusters = []ClusterInfo{}
	}

	return &ListClustersResponse{Clusters: allClusters}, nil
}

// getGCPGKECluster: GCP GKE 클러스터 상세 정보를 조회합니다
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

	s.logger.Info("GCP GKE cluster retrieved successfully",
		zap.String("project_id", projectID),
		zap.String("cluster_name", clusterName),
		zap.String("region", region))

	return &clusterInfo, nil
}

// extractZoneFromLocation: 위치 문자열에서 존 이름을 추출합니다
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

// getGCPLocations: GCP 리전의 모든 위치(리전 및 존) 목록을 반환합니다
func (s *Service) getGCPLocations(region string) []string {
	return []string{
		region,                      // Region level (e.g., asia-northeast3)
		fmt.Sprintf("%s-a", region), // Zone level (e.g., asia-northeast3-a)
		fmt.Sprintf("%s-b", region), // Zone level (e.g., asia-northeast3-b)
		fmt.Sprintf("%s-c", region), // Zone level (e.g., asia-northeast3-c)
	}
}

// getGCPContainerServiceAndProjectID: 자격 증명으로부터 GCP Container 서비스 클라이언트와 프로젝트 ID를 조회합니다
func (s *Service) getGCPContainerServiceAndProjectID(ctx context.Context, credential *domain.Credential) (*container.Service, string, error) {
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), 500)
	}

	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to marshal credential data: %v", err), 500)
	}

	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create GCP container service: %v", err), 502)
	}

	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, "", domain.NewDomainError(domain.ErrCodeValidationFailed, "project_id not found in credential", 400)
	}

	return containerService, projectID, nil
}

// convertGCPNodePoolToNodeGroupInfo: GCP NodePool을 NodeGroupInfo로 변환합니다
func (s *Service) convertGCPNodePoolToNodeGroupInfo(nodePool *container.NodePool, clusterName, region string) NodeGroupInfo {
	s.logger.Debug("Processing GCP NodePool",
		zap.String("name", nodePool.Name),
		zap.String("status", nodePool.Status),
		zap.String("version", nodePool.Version),
		zap.Bool("has_config", nodePool.Config != nil),
		zap.Bool("has_autoscaling", nodePool.Autoscaling != nil),
		zap.Bool("has_management", nodePool.Management != nil),
		zap.Bool("has_upgrade_settings", nodePool.UpgradeSettings != nil))

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
		s.logger.Info("NodePool Config details",
			zap.String("name", nodePool.Name),
			zap.String("machine_type", nodePool.Config.MachineType),
			zap.Int64("disk_size_gb", nodePool.Config.DiskSizeGb),
			zap.String("disk_type", nodePool.Config.DiskType),
			zap.String("image_type", nodePool.Config.ImageType),
			zap.Bool("preemptible", nodePool.Config.Preemptible),
			zap.Bool("spot", nodePool.Config.Spot),
			zap.String("service_account", nodePool.Config.ServiceAccount),
			zap.Int("oauth_scopes_count", len(nodePool.Config.OauthScopes)),
			zap.Int("tags_count", len(nodePool.Config.Tags)),
			zap.Int("labels_count", len(nodePool.Config.Labels)),
			zap.Int("taints_count", len(nodePool.Config.Taints)))

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

// populateNodeGroupFromConfig: GCP NodePool Config로부터 NodeGroupInfo를 채웁니다
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

// getGCPGKEKubeconfig: GCP GKE 클러스터의 kubeconfig를 생성합니다
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
			s.logger.Debug("Failed to get cluster in location",
				zap.String("location", location),
				zap.Error(err))
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

	s.logger.Info("GCP GKE kubeconfig generated successfully",
		zap.String("project_id", projectID),
		zap.String("cluster_name", clusterName),
		zap.String("region", region))

	return kubeconfig, nil
}

// deleteGCPGKECluster: GCP GKE 클러스터를 삭제합니다
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
			s.logger.Debug("Failed to find cluster in location",
				zap.String("location", location),
				zap.Error(err))
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

	s.logger.Info("GCP GKE cluster deletion initiated",
		zap.String("project_id", projectID),
		zap.String("cluster_name", clusterName),
		zap.String("location", clusterLocation),
		zap.String("operation_name", operation.Name))

	// 캐시 무효화: 클러스터 목록 및 개별 클러스터 캐시 삭제
	credentialID := credential.ID.String()
	if err := s.invalidator.InvalidateKubernetesClusterList(ctx, credential.Provider, credentialID, region); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster list cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("region", region),
			zap.Error(err))
	}
	if err := s.invalidator.InvalidateKubernetesClusterItem(ctx, credential.Provider, credentialID, clusterName); err != nil {
		s.logger.Warn("Failed to invalidate Kubernetes cluster item cache",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("cluster_name", clusterName),
			zap.Error(err))
	}

	// 이벤트 발행: 클러스터 삭제 이벤트
	clusterData := map[string]interface{}{
		"cluster_name": clusterName,
		"region":       region,
	}
	if err := s.eventPublisher.PublishKubernetesClusterEvent(ctx, credential.Provider, credentialID, region, "deleted", clusterData); err != nil {
		s.logger.Warn("Failed to publish Kubernetes cluster deleted event",
			zap.String("provider", credential.Provider),
			zap.String("credential_id", credentialID),
			zap.String("cluster_name", clusterName),
			zap.Error(err))
	}

	return nil
}
