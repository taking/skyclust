package kubernetes

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	providererrors "skyclust/internal/shared/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
)

// Service: Kubernetes 서비스 구현체
type Service struct {
	credentialService      domain.CredentialService
	cacheService           domain.CacheService
	eventService           domain.EventService
	auditLogRepo           domain.AuditLogRepository
	logger                 domain.LoggerService
	providerErrorConverter *providererrors.ProviderErrorConverter
}

// NewService: 새로운 Kubernetes 서비스를 생성합니다
func NewService(
	credentialService domain.CredentialService,
	cacheService domain.CacheService,
	eventService domain.EventService,
	auditLogRepo domain.AuditLogRepository,
	logger domain.LoggerService,
) *Service {
	return &Service{
		credentialService:      credentialService,
		cacheService:           cacheService,
		eventService:           eventService,
		auditLogRepo:           auditLogRepo,
		logger:                 logger,
		providerErrorConverter: providererrors.NewProviderErrorConverter(),
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

// CreateAKSCluster: Azure AKS 클러스터를 생성합니다
func (s *Service) CreateAKSCluster(ctx context.Context, credential *domain.Credential, req CreateAKSClusterRequest) (*CreateClusterResponse, error) {
	return s.createAzureAKSCluster(ctx, credential, req)
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

	// AWS EKS: 서브넷이 최소 2개의 다른 AZ에 있는지 사전 검증
	if len(req.SubnetIDs) > 0 {
		// EC2 클라이언트 생성 (서브넷 정보 조회용)
		ec2Client := ec2.NewFromConfig(cfg)

		// 서브넷 정보 조회
		describeSubnetsOutput, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
			SubnetIds: req.SubnetIDs,
		})
		if err != nil {
			return nil, domain.NewDomainError(
				domain.ErrCodeProviderError,
				fmt.Sprintf("failed to describe subnets for validation: %v", err),
				502,
			)
		}

		// 고유한 Availability Zone 추출
		uniqueAZs := make(map[string]bool)
		for _, subnet := range describeSubnetsOutput.Subnets {
			if subnet.AvailabilityZone != nil {
				uniqueAZs[*subnet.AvailabilityZone] = true
			}
		}

		// 최소 2개의 다른 AZ가 필요
		if len(uniqueAZs) < 2 {
			// AZ 목록을 문자열로 변환
			azList := make([]string, 0, len(uniqueAZs))
			for az := range uniqueAZs {
				azList = append(azList, az)
			}

			return nil, domain.NewDomainError(
				domain.ErrCodeValidationFailed,
				fmt.Sprintf("AWS EKS requires subnets from at least two different availability zones. Currently selected subnets are in %d zone(s): %v. Please select subnets from at least two different availability zones.", len(uniqueAZs), azList),
				400,
			)
		}
	}

	// Prepare cluster input
	input := &eks.CreateClusterInput{
		Name:    aws.String(req.Name),
		Version: aws.String(req.Version),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: req.SubnetIDs,
		},
	}

	// Role ARN 자동 생성 (role_arn이 없을 경우)
	roleARN := req.RoleARN
	if roleARN == "" {
		generatedARN, err := s.generateDefaultRoleARN(ctx, cfg, s.getDefaultRoleName())
		if err != nil {
			return nil, fmt.Errorf("failed to generate default role ARN: %w", err)
		}
		roleARN = generatedARN
	}

	// Role ARN 설정
	input.RoleArn = aws.String(roleARN)

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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create EKS cluster: %v", err), HTTPStatusBadGateway)
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

	s.logger.Info(ctx, "EKS cluster creation initiated",
		domain.NewLogField("cluster_name", req.Name),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("version", req.Version))

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

	return response, nil
}

// ListEKSClusters: 모든 Kubernetes 클러스터 목록을 조회합니다 (다중 프로바이더 지원)
func (s *Service) ListEKSClusters(ctx context.Context, credential *domain.Credential, region string) (*ListClustersResponse, error) {
	// 캐시 키 생성
	credentialID := credential.ID.String()
	cacheKey := buildKubernetesClusterListKey(credential.Provider, credentialID, region)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedResponse, ok := cachedValue.(*ListClustersResponse); ok {
				s.logger.Debug(ctx, "Kubernetes clusters retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return cachedResponse, nil
			} else if cachedResponse, ok := cachedValue.(ListClustersResponse); ok {
				s.logger.Debug(ctx, "Kubernetes clusters retrieved from cache",
					domain.NewLogField("provider", credential.Provider),
					domain.NewLogField("credential_id", credentialID),
					domain.NewLogField("region", region))
				return &cachedResponse, nil
			}
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
		response, err = s.listAzureAKSClusters(ctx, credential, region)
	case "ncp":
		// TODO: Implement NCP NKS cluster listing
		response = &ListClustersResponse{Clusters: []ClusterInfo{}}
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
	}

	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cacheService != nil && response != nil {
		if err := s.cacheService.Set(ctx, cacheKey, response, defaultK8sTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache Kubernetes clusters, continuing without cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
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

	s.logger.Debug(ctx, "AWS credentials extracted",
		domain.NewLogField("access_key_prefix", creds.AccessKey[:min(10, len(creds.AccessKey))]),
		domain.NewLogField("access_key_length", len(creds.AccessKey)),
		domain.NewLogField("secret_key_length", len(creds.SecretKey)),
		domain.NewLogField("region", creds.Region))

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List clusters
	output, err := eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to list EKS clusters: %v", err), HTTPStatusBadGateway)
	}

	// Get detailed information for each cluster
	var clusters []ClusterInfo
	for _, clusterName := range output.Clusters {
		// Describe each cluster to get detailed information
		describeOutput, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		})
		if err != nil {
			s.logger.Warn(ctx, "Failed to describe cluster",
				domain.NewLogField("cluster_name", clusterName),
				domain.NewLogField("error", err))
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
	credentialID := credential.ID.String()
	cacheKey := buildKubernetesClusterItemKey(credential.Provider, credentialID, clusterName)

	// 캐시에서 조회 시도 (Early return)
	if cachedCluster := s.getClusterFromCache(ctx, cacheKey, credential.Provider, credentialID, clusterName); cachedCluster != nil {
		return cachedCluster, nil
	}

	// 캐시 미스 시 실제 API 호출
	cluster, err := s.getClusterFromProvider(ctx, credential, clusterName, region)
	if err != nil {
		return nil, err
	}

	// 응답을 캐시에 저장
	if s.cacheService != nil && cluster != nil {
		if err := s.cacheService.Set(ctx, cacheKey, cluster, defaultK8sTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache Kubernetes cluster",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("cluster_name", clusterName),
				domain.NewLogField("error", err))
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe EKS cluster: %v", err), HTTPStatusBadGateway)
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
		return s.deleteAzureAKSCluster(ctx, credential, clusterName, region)
	case "ncp":
		// TODO: Implement NCP NKS cluster deletion
		return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS cluster deletion not implemented yet", HTTPStatusNotImplemented)
	default:
		return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
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
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to delete EKS cluster: %v", err), HTTPStatusBadGateway)
	}

	s.logger.Info(ctx, "EKS cluster deletion initiated",
		domain.NewLogField("cluster_name", clusterName),
		domain.NewLogField("region", region))

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
		return s.getAzureAKSKubeconfig(ctx, credential, clusterName, region)
	case "ncp":
		// TODO: Implement NCP NKS kubeconfig generation
		return "", domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS kubeconfig generation not implemented yet", HTTPStatusNotImplemented)
	default:
		return "", domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
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
		return "", domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to describe EKS cluster: %v", err), HTTPStatusBadGateway)
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
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), HTTPStatusInternalServerError)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", HTTPStatusBadRequest)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", HTTPStatusBadRequest)
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), HTTPStatusBadGateway)
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create node group: %v", err), HTTPStatusBadGateway)
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

	s.logger.Info(ctx, "EKS node group creation initiated",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("nodegroup_name", req.NodePoolName),
		domain.NewLogField("region", req.Region))

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

	// 이벤트 발행
	if s.eventService != nil {
		nodePoolData := map[string]interface{}{
			"nodegroup_name": req.NodePoolName,
			"cluster_name":   req.ClusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"region":         req.Region,
		}
		eventType := fmt.Sprintf("kubernetes.nodepool.%s.created", credential.Provider)
		_ = s.eventService.Publish(ctx, eventType, nodePoolData)
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

	// Node Role ARN 자동 생성 (node_role_arn이 없을 경우)
	nodeRoleARN := req.NodeRoleARN
	if nodeRoleARN == "" {
		// 환경 변수로 기본 Node Role 이름 커스터마이징 가능 (기본값: EKSNodeRole)
		defaultNodeRoleName := os.Getenv("AWS_EKS_DEFAULT_NODE_ROLE_NAME")
		if defaultNodeRoleName == "" {
			defaultNodeRoleName = "EKSNodeRole"
		}

		generatedARN, err := s.generateDefaultRoleARN(ctx, cfg, defaultNodeRoleName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate default node role ARN: %w", err)
		}
		nodeRoleARN = generatedARN
	}

	// Prepare node group input
	input := &eks.CreateNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
		NodeRole:      aws.String(nodeRoleARN),
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create node group: %v", err), HTTPStatusBadGateway)
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

	s.logger.Info(ctx, "EKS node group creation initiated",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("nodegroup_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region))

	return response, nil
}

// ListNodeGroups: 클러스터의 노드 그룹 목록을 조회합니다
func (s *Service) ListNodeGroups(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	s.logger.Info(ctx, "ListNodeGroups called",
		domain.NewLogField("provider", credential.Provider),
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("region", req.Region))

	switch credential.Provider {
	case "aws":
		return s.listAWSEKSNodeGroups(ctx, credential, req)
	case "gcp":
		return s.listGCPGKENodePools(ctx, credential, req)
	case "azure":
		return s.listAzureNodePools(ctx, credential, req)
	case "ncp":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP node groups not implemented yet", HTTPStatusNotImplemented)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
	}
}

// listAWSEKSNodeGroups: 클러스터의 AWS EKS 노드 그룹 목록을 조회합니다
func (s *Service) listAWSEKSNodeGroups(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to decrypt credential: %v", err), HTTPStatusInternalServerError)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "access_key not found in credential", HTTPStatusBadRequest)
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, "secret_key not found in credential", HTTPStatusBadRequest)
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
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to load AWS config: %v", err), HTTPStatusBadGateway)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List node groups
	output, err := eksClient.ListNodegroups(ctx, &eks.ListNodegroupsInput{
		ClusterName: aws.String(req.ClusterName),
	})
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "list node groups")
	}

	// Get detailed information for each node group
	var nodeGroups []NodeGroupInfo
	for _, nodeGroupName := range output.Nodegroups {
		describeOutput, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
			ClusterName:   aws.String(req.ClusterName),
			NodegroupName: aws.String(nodeGroupName),
		})
		if err != nil {
			s.logger.Warn(ctx, "Failed to describe node group",
				domain.NewLogField("cluster_name", req.ClusterName),
				domain.NewLogField("nodegroup_name", nodeGroupName),
				domain.NewLogField("error", err))
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

// GetNodeGroup: 클러스터의 노드 그룹 상세 정보를 조회합니다
func (s *Service) GetNodeGroup(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
	s.logger.Info(ctx, "GetNodeGroup called",
		domain.NewLogField("provider", credential.Provider),
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("node_group_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region))

	switch credential.Provider {
	case "aws":
		return s.getAWSEKSNodeGroup(ctx, credential, req)
	case "gcp":
		return s.getGCPGKENodePool(ctx, credential, req)
	case "azure":
		return s.getAzureNodePool(ctx, credential, req)
	case "ncp":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP node groups not implemented yet", HTTPStatusNotImplemented)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
	}
}

// getAWSEKSNodeGroup: AWS EKS 노드 그룹 상세 정보를 조회합니다
func (s *Service) getAWSEKSNodeGroup(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
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

	// Describe node group
	describeOutput, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "get EKS node group")
	}

	if describeOutput.Nodegroup == nil {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("node group %s not found in cluster %s", req.NodeGroupName, req.ClusterName), HTTPStatusNotFound)
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

	return &nodeGroup, nil
}

// validateEKSSubnetAZs EKS 서브넷이 최소 2개의 다른 AZ에 있는지 검증합니다
func (s *Service) validateEKSSubnetAZs(ctx context.Context, cfg aws.Config, subnetIDs []string) error {
	ec2Client := ec2.NewFromConfig(cfg)

	describeSubnetsOutput, err := ec2Client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
		SubnetIds: subnetIDs,
	})
	if err != nil {
		return domain.NewDomainError(
			domain.ErrCodeProviderError,
			fmt.Sprintf("failed to describe subnets for validation: %v", err),
			HTTPStatusBadGateway,
		)
	}

	uniqueAZs := s.extractUniqueAZs(describeSubnetsOutput.Subnets)
	if len(uniqueAZs) < MinEKSSubnetAZs {
		return domain.NewDomainError(
			domain.ErrCodeValidationFailed,
			fmt.Sprintf("AWS EKS requires subnets from at least %d different availability zones. Currently selected subnets are in %d zone(s): %v. Please select subnets from at least %d different availability zones.", MinEKSSubnetAZs, len(uniqueAZs), s.azMapToSlice(uniqueAZs), MinEKSSubnetAZs),
			HTTPStatusBadRequest,
		)
	}

	return nil
}

// extractUniqueAZs 서브넷 목록에서 고유한 Availability Zone을 추출합니다
func (s *Service) extractUniqueAZs(subnets []ec2types.Subnet) map[string]bool {
	uniqueAZs := make(map[string]bool)
	for _, subnet := range subnets {
		if subnet.AvailabilityZone != nil {
			uniqueAZs[*subnet.AvailabilityZone] = true
		}
	}
	return uniqueAZs
}

// azMapToSlice AZ 맵을 슬라이스로 변환합니다
func (s *Service) azMapToSlice(azMap map[string]bool) []string {
	azList := make([]string, 0, len(azMap))
	for az := range azMap {
		azList = append(azList, az)
	}
	return azList
}

// getDefaultRoleName 기본 Role 이름을 반환합니다
func (s *Service) getDefaultRoleName() string {
	defaultRoleName := os.Getenv("AWS_EKS_DEFAULT_CLUSTER_ROLE_NAME")
	if defaultRoleName == "" {
		return "EKSClusterRole"
	}
	return defaultRoleName
}

// getClusterFromCache 캐시에서 클러스터 정보를 조회합니다
func (s *Service) getClusterFromCache(ctx context.Context, cacheKey, provider, credentialID, clusterName string) *ClusterInfo {
	if s.cacheService == nil {
		return nil
	}

	cachedValue, err := s.cacheService.Get(ctx, cacheKey)
	if err != nil || cachedValue == nil {
		return nil
	}

	// 타입 단언 시도 (*ClusterInfo)
	if cachedCluster, ok := cachedValue.(*ClusterInfo); ok {
		s.logger.Debug(ctx, "Kubernetes cluster retrieved from cache",
			domain.NewLogField("provider", provider),
			domain.NewLogField("credential_id", credentialID),
			domain.NewLogField("cluster_name", clusterName))
		return cachedCluster
	}

	// 타입 단언 시도 (ClusterInfo)
	if cachedCluster, ok := cachedValue.(ClusterInfo); ok {
		s.logger.Debug(ctx, "Kubernetes cluster retrieved from cache",
			domain.NewLogField("provider", provider),
			domain.NewLogField("credential_id", credentialID),
			domain.NewLogField("cluster_name", clusterName))
		return &cachedCluster
	}

	return nil
}

// getClusterFromProvider 프로바이더별로 클러스터 정보를 조회합니다
func (s *Service) getClusterFromProvider(ctx context.Context, credential *domain.Credential, clusterName, region string) (*ClusterInfo, error) {
	switch credential.Provider {
	case "aws":
		return s.getAWSEKSCluster(ctx, credential, clusterName, region)
	case "gcp":
		return s.getGCPGKECluster(ctx, credential, clusterName, region)
	case "azure":
		return s.getAzureAKSCluster(ctx, credential, clusterName, region)
	case "ncp":
		return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP NKS cluster retrieval not implemented yet", HTTPStatusNotImplemented)
	default:
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
	}
}

// DeleteNodeGroup: 클러스터의 노드 그룹을 삭제합니다
func (s *Service) DeleteNodeGroup(ctx context.Context, credential *domain.Credential, req DeleteNodeGroupRequest) error {
	s.logger.Info(ctx, "DeleteNodeGroup called",
		domain.NewLogField("provider", credential.Provider),
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("node_group_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region))

	switch credential.Provider {
	case "aws":
		return s.deleteAWSEKSNodeGroup(ctx, credential, req)
	case "gcp":
		return s.deleteGCPGKENodePool(ctx, credential, req)
	case "azure":
		return s.deleteAzureNodePool(ctx, credential, req)
	case "ncp":
		return domain.NewDomainError(domain.ErrCodeNotImplemented, "NCP node groups not implemented yet", HTTPStatusNotImplemented)
	default:
		return domain.NewDomainError(domain.ErrCodeValidationFailed, fmt.Sprintf("unsupported provider: %s", credential.Provider), HTTPStatusBadRequest)
	}
}

// deleteAWSEKSNodeGroup: AWS EKS 노드 그룹을 삭제합니다
func (s *Service) deleteAWSEKSNodeGroup(ctx context.Context, credential *domain.Credential, req DeleteNodeGroupRequest) error {
	creds, err := s.extractAWSCredentials(ctx, credential, req.Region)
	if err != nil {
		return err
	}

	cfg, err := s.createAWSConfig(ctx, creds)
	if err != nil {
		return err
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Delete node group
	_, err = eksClient.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "delete EKS node group")
	}

	s.logger.Info(ctx, "EKS node group deletion initiated",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("nodegroup_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region))

	return nil
}

// deleteGCPGKENodePool: GCP GKE 노드 풀을 삭제합니다
func (s *Service) deleteGCPGKENodePool(ctx context.Context, credential *domain.Credential, req DeleteNodeGroupRequest) error {
	containerService, projectID, err := s.setupGCPContainerService(ctx, credential)
	if err != nil {
		return err
	}

	// Find the cluster first to determine its actual location
	locations := s.getGCPLocations(req.Region)

	var clusterLocation string
	var deleted bool

	// Search for the cluster and node pool in all possible locations
	for _, location := range locations {
		nodePoolPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", projectID, location, req.ClusterName, req.NodeGroupName)

		// Try to delete the node pool
		_, err := containerService.Projects.Locations.Clusters.NodePools.Delete(nodePoolPath).Context(ctx).Do()
		if err != nil {
			// Log debug but continue with other locations
			s.logger.Debug(ctx, "Failed to delete node pool in location",
				domain.NewLogField("location", location),
				domain.NewLogField("cluster_name", req.ClusterName),
				domain.NewLogField("node_pool_name", req.NodeGroupName),
				domain.NewLogField("error", err))
			continue
		}

		clusterLocation = location
		deleted = true
		break
	}

	if !deleted {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("failed to find and delete GKE node pool %s in cluster %s in region %s or any of its zones", req.NodeGroupName, req.ClusterName, req.Region), HTTPStatusNotFound)
	}

	s.logger.Info(ctx, "GKE node pool deletion initiated",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("node_pool_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("cluster_location", clusterLocation))

	return nil
}

// deleteAzureNodePool: Azure AKS 노드 풀을 삭제합니다
func (s *Service) deleteAzureNodePool(ctx context.Context, credential *domain.Credential, req DeleteNodeGroupRequest) error {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return err
	}

	// Find resource group from cluster
	if creds.ResourceGroup == "" {
		clusters, err := s.listAzureAKSClusters(ctx, credential, req.Region)
		if err != nil {
			return err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == req.ClusterName {
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", req.ClusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return err
	}

	// Get agent pools client
	agentPoolsClient := clientFactory.NewAgentPoolsClient()

	// Delete agent pool (BeginDelete returns a poller for async operations)
	poller, err := agentPoolsClient.BeginDelete(ctx, creds.ResourceGroup, req.ClusterName, req.NodeGroupName, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete AKS node pool")
	}

	// Wait for deletion to complete (optional, can be async)
	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "wait for AKS node pool deletion")
	}

	s.logger.Info(ctx, "AKS node pool deletion initiated",
		domain.NewLogField("cluster_name", req.ClusterName),
		domain.NewLogField("node_pool_name", req.NodeGroupName),
		domain.NewLogField("region", req.Region),
		domain.NewLogField("resource_group", creds.ResourceGroup))

	return nil
}

// createAzureAKSCluster: Azure AKS 클러스터를 생성합니다
func (s *Service) createAzureAKSCluster(ctx context.Context, credential *domain.Credential, req CreateAKSClusterRequest) (*CreateClusterResponse, error) {
	// Extract Azure credentials
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Override resource group from request if provided
	if req.ResourceGroup != "" {
		creds.ResourceGroup = req.ResourceGroup
	}
	if creds.ResourceGroup == "" {
		return nil, domain.NewDomainError(domain.ErrCodeValidationFailed, ErrMsgResourceGroupRequired, HTTPStatusBadRequest)
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get managed clusters client
	managedClustersClient := clientFactory.NewManagedClustersClient()

	// Build network profile
	networkPlugin := armcontainerservice.NetworkPluginAzure
	if req.Network != nil && req.Network.NetworkPlugin != "" {
		switch req.Network.NetworkPlugin {
		case "azure":
			networkPlugin = armcontainerservice.NetworkPluginAzure
		case "kubenet":
			networkPlugin = armcontainerservice.NetworkPluginKubenet
		default:
			networkPlugin = armcontainerservice.NetworkPluginAzure
		}
	}

	var networkPolicy *armcontainerservice.NetworkPolicy
	if req.Network != nil && req.Network.NetworkPolicy != "" {
		switch req.Network.NetworkPolicy {
		case "azure":
			networkPolicy = to.Ptr(armcontainerservice.NetworkPolicyAzure)
		case "calico":
			networkPolicy = to.Ptr(armcontainerservice.NetworkPolicyCalico)
		}
	}

	networkProfile := &armcontainerservice.NetworkProfile{
		NetworkPlugin: &networkPlugin,
		NetworkPolicy: networkPolicy,
	}

	if req.Network != nil {
		if req.Network.PodCIDR != "" {
			networkProfile.PodCidr = to.Ptr(req.Network.PodCIDR)
		}
		if req.Network.ServiceCIDR != "" {
			networkProfile.ServiceCidr = to.Ptr(req.Network.ServiceCIDR)
		}
		if req.Network.DNSServiceIP != "" {
			networkProfile.DNSServiceIP = to.Ptr(req.Network.DNSServiceIP)
		}
	}

	// Build agent pool profile (node pool)
	agentPoolProfiles := []*armcontainerservice.ManagedClusterAgentPoolProfile{}
	if req.NodePool != nil {
		agentPoolProfile := &armcontainerservice.ManagedClusterAgentPoolProfile{
			Name:   to.Ptr(req.NodePool.Name),
			VMSize: to.Ptr(req.NodePool.VMSize),
			Count:  to.Ptr(int32(req.NodePool.NodeCount)),
		}

		if req.NodePool.OSDiskSizeGB > 0 {
			agentPoolProfile.OSDiskSizeGB = to.Ptr(int32(req.NodePool.OSDiskSizeGB))
		}

		if req.NodePool.OSDiskType != "" {
			switch req.NodePool.OSDiskType {
			case "Managed":
				agentPoolProfile.OSDiskType = to.Ptr(armcontainerservice.OSDiskTypeManaged)
			case "Ephemeral":
				agentPoolProfile.OSDiskType = to.Ptr(armcontainerservice.OSDiskTypeEphemeral)
			}
		}

		if req.NodePool.OSType != "" {
			switch req.NodePool.OSType {
			case "Linux":
				agentPoolProfile.OSType = to.Ptr(armcontainerservice.OSTypeLinux)
			case "Windows":
				agentPoolProfile.OSType = to.Ptr(armcontainerservice.OSTypeWindows)
			}
		} else {
			agentPoolProfile.OSType = to.Ptr(armcontainerservice.OSTypeLinux)
		}

		if req.NodePool.OSSKU != "" {
			switch req.NodePool.OSSKU {
			case "Ubuntu":
				agentPoolProfile.OSSKU = to.Ptr(armcontainerservice.OSSKUUbuntu)
			case "CBLMariner":
				agentPoolProfile.OSSKU = to.Ptr(armcontainerservice.OSSKUCBLMariner)
			}
		}

		if req.NodePool.EnableAutoScaling {
			agentPoolProfile.EnableAutoScaling = to.Ptr(true)
			if req.NodePool.MinCount > 0 {
				agentPoolProfile.MinCount = to.Ptr(int32(req.NodePool.MinCount))
			}
			if req.NodePool.MaxCount > 0 {
				agentPoolProfile.MaxCount = to.Ptr(int32(req.NodePool.MaxCount))
			}
		}

		if req.NodePool.MaxPods > 0 {
			agentPoolProfile.MaxPods = to.Ptr(int32(req.NodePool.MaxPods))
		}

		if req.NodePool.VnetSubnetID != "" {
			agentPoolProfile.VnetSubnetID = to.Ptr(req.NodePool.VnetSubnetID)
		} else if req.Network != nil && req.Network.SubnetID != "" {
			agentPoolProfile.VnetSubnetID = to.Ptr(req.Network.SubnetID)
		}

		if len(req.NodePool.AvailabilityZones) > 0 {
			availabilityZones := make([]*string, len(req.NodePool.AvailabilityZones))
			for i, zone := range req.NodePool.AvailabilityZones {
				availabilityZones[i] = to.Ptr(zone)
			}
			agentPoolProfile.AvailabilityZones = availabilityZones
		}

		if len(req.NodePool.Labels) > 0 {
			labels := make(map[string]*string)
			for k, v := range req.NodePool.Labels {
				labels[k] = to.Ptr(v)
			}
			agentPoolProfile.NodeLabels = labels
		}

		if req.NodePool.Mode != "" {
			switch req.NodePool.Mode {
			case "System":
				agentPoolProfile.Mode = to.Ptr(armcontainerservice.AgentPoolModeSystem)
			case "User":
				agentPoolProfile.Mode = to.Ptr(armcontainerservice.AgentPoolModeUser)
			}
		} else {
			agentPoolProfile.Mode = to.Ptr(armcontainerservice.AgentPoolModeSystem)
		}

		agentPoolProfiles = append(agentPoolProfiles, agentPoolProfile)
	}

	// Build security profile
	enableRBAC := true
	if req.Security != nil {
		enableRBAC = req.Security.EnableRBAC
	}

	// Build API server access profile
	var apiServerAccessProfile *armcontainerservice.ManagedClusterAPIServerAccessProfile
	if req.Security != nil {
		apiServerAccessProfile = &armcontainerservice.ManagedClusterAPIServerAccessProfile{}
		if req.Security.EnablePrivateCluster {
			apiServerAccessProfile.EnablePrivateCluster = to.Ptr(true)
		}
		if len(req.Security.APIServerAuthorizedIPRanges) > 0 {
			authorizedIPRanges := make([]*string, len(req.Security.APIServerAuthorizedIPRanges))
			for i, ipRange := range req.Security.APIServerAuthorizedIPRanges {
				authorizedIPRanges[i] = to.Ptr(ipRange)
			}
			apiServerAccessProfile.AuthorizedIPRanges = authorizedIPRanges
		}
	}

	// Build addon profiles
	addonProfiles := make(map[string]*armcontainerservice.ManagedClusterAddonProfile)
	if req.Security != nil {
		if req.Security.EnableAzurePolicy {
			addonProfiles["azurepolicy"] = &armcontainerservice.ManagedClusterAddonProfile{
				Enabled: to.Ptr(true),
			}
		}
	}

	// Build workload identity profile
	var workloadIdentityProfile *armcontainerservice.ManagedClusterSecurityProfileWorkloadIdentity
	if req.Security != nil && req.Security.EnableWorkloadIdentity {
		workloadIdentityProfile = &armcontainerservice.ManagedClusterSecurityProfileWorkloadIdentity{
			Enabled: to.Ptr(true),
		}
	}

	// Build managed cluster
	managedCluster := armcontainerservice.ManagedCluster{
		Location: to.Ptr(req.Location),
		Properties: &armcontainerservice.ManagedClusterProperties{
			KubernetesVersion: to.Ptr(req.Version),
			DNSPrefix:         to.Ptr(req.Name),
			AgentPoolProfiles: agentPoolProfiles,
			NetworkProfile:    networkProfile,
			EnableRBAC:        to.Ptr(enableRBAC),
		},
		Tags: req.Tags,
	}

	if apiServerAccessProfile != nil {
		managedCluster.Properties.APIServerAccessProfile = apiServerAccessProfile
	}

	if len(addonProfiles) > 0 {
		managedCluster.Properties.AddonProfiles = addonProfiles
	}

	if workloadIdentityProfile != nil {
		if managedCluster.Properties.SecurityProfile == nil {
			managedCluster.Properties.SecurityProfile = &armcontainerservice.ManagedClusterSecurityProfile{}
		}
		managedCluster.Properties.SecurityProfile.WorkloadIdentity = workloadIdentityProfile
	}

	// Create cluster
	poller, err := managedClustersClient.BeginCreateOrUpdate(ctx, creds.ResourceGroup, req.Name, managedCluster, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "create AKS cluster")
	}

	// Wait for the operation to complete (optional, can be async)
	// For now, we'll return immediately and let the user poll for status
	s.logger.Info(ctx, "Azure AKS cluster creation initiated",
		domain.NewLogField("cluster_name", req.Name),
		domain.NewLogField("location", req.Location),
		domain.NewLogField("resource_group", creds.ResourceGroup),
		domain.NewLogField("version", req.Version))

	// Build response
	response := &CreateClusterResponse{
		ClusterID: fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ContainerService/managedClusters/%s",
			creds.SubscriptionID, creds.ResourceGroup, req.Name),
		Name:      req.Name,
		Version:   req.Version,
		Region:    req.Location,
		Status:    "creating",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	// 캐시 무효화: 클러스터 목록 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		cacheKey := buildKubernetesClusterListKey(credential.Provider, credentialID, req.Location)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate Kubernetes cluster list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("location", req.Location),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: 클러스터 생성 이벤트
	if s.eventService != nil {
		clusterData := map[string]interface{}{
			"cluster_name":   req.Name,
			"location":       req.Location,
			"resource_group": creds.ResourceGroup,
			"version":        req.Version,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
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
			"cluster_name":   req.Name,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"location":       req.Location,
			"resource_group": creds.ResourceGroup,
			"version":        req.Version,
		},
	)

	// Note: Azure operations are async, so we return immediately
	// The actual cluster will be created in the background
	// Users should poll GetCluster to check status
	_ = poller // Suppress unused variable warning

	return response, nil
}

// listAzureAKSClusters: Azure AKS 클러스터 목록을 조회합니다
func (s *Service) listAzureAKSClusters(ctx context.Context, credential *domain.Credential, location string) (*ListClustersResponse, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get managed clusters client
	managedClustersClient := clientFactory.NewManagedClustersClient()

	// List all clusters in the subscription
	var clusters []ClusterInfo
	pager := managedClustersClient.NewListPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "list AKS clusters")
		}

		for _, cluster := range page.Value {
			// Filter by location if provided
			if location != "" && cluster.Location != nil && *cluster.Location != location {
				continue
			}

			clusterInfo := ClusterInfo{
				ID:      *cluster.ID,
				Name:    *cluster.Name,
				Version: "",
				Status:  "",
				Region:  "",
			}

			if cluster.Properties != nil {
				if cluster.Properties.KubernetesVersion != nil {
					clusterInfo.Version = *cluster.Properties.KubernetesVersion
				}
				if cluster.Properties.PowerState != nil && cluster.Properties.PowerState.Code != nil {
					clusterInfo.Status = string(*cluster.Properties.PowerState.Code)
				}
				if cluster.Properties.Fqdn != nil {
					clusterInfo.Endpoint = *cluster.Properties.Fqdn
				}
			}

			if cluster.Location != nil {
				clusterInfo.Region = *cluster.Location
			}

			if cluster.Tags != nil {
				tags := make(map[string]string)
				for k, v := range cluster.Tags {
					if v != nil {
						tags[k] = *v
					}
				}
				clusterInfo.Tags = tags
			}

			clusters = append(clusters, clusterInfo)
		}
	}

	return &ListClustersResponse{Clusters: clusters}, nil
}

// getAzureAKSCluster: Azure AKS 클러스터 상세 정보를 조회합니다
func (s *Service) getAzureAKSCluster(ctx context.Context, credential *domain.Credential, clusterName, location string) (*ClusterInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Resource group is required for getting a specific cluster
	// We need to find the resource group first by listing all clusters
	// or it should be provided in the request
	if creds.ResourceGroup == "" {
		// Try to find the cluster by listing all clusters
		clusters, err := s.listAzureAKSClusters(ctx, credential, location)
		if err != nil {
			return nil, err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == clusterName {
				// Extract resource group from cluster ID
				// Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/...
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", clusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get managed clusters client
	managedClustersClient := clientFactory.NewManagedClustersClient()

	// Get cluster details
	cluster, err := managedClustersClient.Get(ctx, creds.ResourceGroup, clusterName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get AKS cluster")
	}

	clusterInfo := ClusterInfo{
		ID:      *cluster.ID,
		Name:    *cluster.Name,
		Version: "",
		Status:  "",
		Region:  "",
	}

	if cluster.Properties != nil {
		if cluster.Properties.KubernetesVersion != nil {
			clusterInfo.Version = *cluster.Properties.KubernetesVersion
		}
		if cluster.Properties.PowerState != nil && cluster.Properties.PowerState.Code != nil {
			clusterInfo.Status = string(*cluster.Properties.PowerState.Code)
		}
		if cluster.Properties.Fqdn != nil {
			clusterInfo.Endpoint = *cluster.Properties.Fqdn
		}
		// Note: Azure SDK doesn't expose CreatedAt directly in ManagedClusterProperties
	}

	if cluster.Location != nil {
		clusterInfo.Region = *cluster.Location
	}

	if cluster.Tags != nil {
		tags := make(map[string]string)
		for k, v := range cluster.Tags {
			if v != nil {
				tags[k] = *v
			}
		}
		clusterInfo.Tags = tags
	}

	return &clusterInfo, nil
}

// deleteAzureAKSCluster: Azure AKS 클러스터를 삭제합니다
func (s *Service) deleteAzureAKSCluster(ctx context.Context, credential *domain.Credential, clusterName, location string) error {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return err
	}

	// Resource group is required for deleting a cluster
	// We need to find the resource group first
	if creds.ResourceGroup == "" {
		// Try to find the cluster by listing all clusters
		clusters, err := s.listAzureAKSClusters(ctx, credential, location)
		if err != nil {
			return err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == clusterName {
				// Extract resource group from cluster ID
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", clusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return err
	}

	// Get managed clusters client
	managedClustersClient := clientFactory.NewManagedClustersClient()

	// Delete cluster
	poller, err := managedClustersClient.BeginDelete(ctx, creds.ResourceGroup, clusterName, nil)
	if err != nil {
		return s.providerErrorConverter.ConvertAzureError(err, "delete AKS cluster")
	}

	s.logger.Info(ctx, "Azure AKS cluster deletion initiated",
		domain.NewLogField("cluster_name", clusterName),
		domain.NewLogField("location", location),
		domain.NewLogField("resource_group", creds.ResourceGroup))

	// 캐시 무효화: 클러스터 목록 및 개별 클러스터 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildKubernetesClusterListKey(credential.Provider, credentialID, location)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate Kubernetes cluster list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("location", location),
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
			"cluster_name":   clusterName,
			"location":       location,
			"resource_group": creds.ResourceGroup,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
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

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionKubernetesClusterDelete,
		fmt.Sprintf("DELETE /api/v1/%s/kubernetes/clusters/%s", credential.Provider, clusterName),
		map[string]interface{}{
			"cluster_name":   clusterName,
			"provider":       credential.Provider,
			"credential_id":  credentialID,
			"location":       location,
			"resource_group": creds.ResourceGroup,
		},
	)

	// Note: Azure operations are async, so we return immediately
	_ = poller // Suppress unused variable warning

	return nil
}

// getAzureAKSKubeconfig: Azure AKS 클러스터의 kubeconfig를 생성합니다
func (s *Service) getAzureAKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, location string) (string, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return "", err
	}

	// Resource group is required for getting kubeconfig
	// We need to find the resource group first
	if creds.ResourceGroup == "" {
		// Try to find the cluster by listing all clusters
		clusters, err := s.listAzureAKSClusters(ctx, credential, location)
		if err != nil {
			return "", err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == clusterName {
				// Extract resource group from cluster ID
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return "", domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", clusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return "", err
	}

	// Get managed clusters client
	managedClustersClient := clientFactory.NewManagedClustersClient()

	// Get cluster credentials (kubeconfig)
	credentialResult, err := managedClustersClient.ListClusterUserCredentials(ctx, creds.ResourceGroup, clusterName, nil)
	if err != nil {
		return "", s.providerErrorConverter.ConvertAzureError(err, "get AKS kubeconfig")
	}

	// Extract kubeconfig from the credential result
	if credentialResult.Kubeconfigs == nil || len(credentialResult.Kubeconfigs) == 0 {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, "no kubeconfig found in Azure response", HTTPStatusInternalServerError)
	}

	// Return the first kubeconfig (usually there's only one)
	kubeconfigBytes := credentialResult.Kubeconfigs[0].Value
	return string(kubeconfigBytes), nil
}

// listAzureNodePools: Azure AKS 클러스터의 노드 풀 목록을 조회합니다
func (s *Service) listAzureNodePools(ctx context.Context, credential *domain.Credential, req ListNodeGroupsRequest) (*ListNodeGroupsResponse, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Find resource group from cluster
	if creds.ResourceGroup == "" {
		clusters, err := s.listAzureAKSClusters(ctx, credential, req.Region)
		if err != nil {
			return nil, err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == req.ClusterName {
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", req.ClusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get agent pools client
	agentPoolsClient := clientFactory.NewAgentPoolsClient()

	// List agent pools
	var nodeGroups []NodeGroupInfo
	pager := agentPoolsClient.NewListPager(creds.ResourceGroup, req.ClusterName, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, s.providerErrorConverter.ConvertAzureError(err, "list AKS node pools")
		}

		for _, agentPool := range page.Value {
			nodeGroup := NodeGroupInfo{
				ID:          *agentPool.ID,
				Name:        *agentPool.Name,
				ClusterName: req.ClusterName,
				Region:      req.Region,
				Status:      "",
			}

			if agentPool.Properties != nil {
				if agentPool.Properties.Count != nil {
					nodeGroup.ScalingConfig = NodeGroupScalingConfig{
						DesiredSize: *agentPool.Properties.Count,
					}
					if agentPool.Properties.MinCount != nil {
						nodeGroup.ScalingConfig.MinSize = *agentPool.Properties.MinCount
					}
					if agentPool.Properties.MaxCount != nil {
						nodeGroup.ScalingConfig.MaxSize = *agentPool.Properties.MaxCount
					}
				}

				if agentPool.Properties.VMSize != nil {
					nodeGroup.InstanceTypes = []string{*agentPool.Properties.VMSize}
				}

				if agentPool.Properties.OSDiskSizeGB != nil {
					nodeGroup.DiskSize = *agentPool.Properties.OSDiskSizeGB
				}

				if agentPool.Properties.ProvisioningState != nil {
					nodeGroup.Status = *agentPool.Properties.ProvisioningState
				}

				if agentPool.Properties.NodeLabels != nil {
					labels := make(map[string]string)
					for k, v := range agentPool.Properties.NodeLabels {
						if v != nil {
							labels[k] = *v
						}
					}
					nodeGroup.Labels = labels
				}
			}

			nodeGroups = append(nodeGroups, nodeGroup)
		}
	}

	return &ListNodeGroupsResponse{
		NodeGroups: nodeGroups,
		Total:      len(nodeGroups),
	}, nil
}

// getAzureNodePool: Azure AKS 클러스터의 노드 풀 상세 정보를 조회합니다
func (s *Service) getAzureNodePool(ctx context.Context, credential *domain.Credential, req GetNodeGroupRequest) (*NodeGroupInfo, error) {
	creds, err := s.extractAzureCredentials(ctx, credential)
	if err != nil {
		return nil, err
	}

	// Find resource group from cluster
	if creds.ResourceGroup == "" {
		clusters, err := s.listAzureAKSClusters(ctx, credential, req.Region)
		if err != nil {
			return nil, err
		}

		for _, cluster := range clusters.Clusters {
			if cluster.Name == req.ClusterName {
				parts := strings.Split(cluster.ID, "/")
				for i, part := range parts {
					if part == "resourceGroups" && i+1 < len(parts) {
						creds.ResourceGroup = parts[i+1]
						break
					}
				}
				break
			}
		}

		if creds.ResourceGroup == "" {
			return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("AKS cluster %s not found", req.ClusterName), HTTPStatusNotFound)
		}
	}

	// Create Azure Container Service client
	clientFactory, err := s.createAzureContainerServiceClient(ctx, creds)
	if err != nil {
		return nil, err
	}

	// Get agent pools client
	agentPoolsClient := clientFactory.NewAgentPoolsClient()

	// Get agent pool details
	agentPool, err := agentPoolsClient.Get(ctx, creds.ResourceGroup, req.ClusterName, req.NodeGroupName, nil)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAzureError(err, "get AKS node pool")
	}

	nodeGroup := NodeGroupInfo{
		ID:          *agentPool.ID,
		Name:        *agentPool.Name,
		ClusterName: req.ClusterName,
		Region:      req.Region,
		Status:      "",
	}

	if agentPool.Properties != nil {
		if agentPool.Properties.Count != nil {
			nodeGroup.ScalingConfig = NodeGroupScalingConfig{
				DesiredSize: *agentPool.Properties.Count,
			}
			if agentPool.Properties.MinCount != nil {
				nodeGroup.ScalingConfig.MinSize = *agentPool.Properties.MinCount
			}
			if agentPool.Properties.MaxCount != nil {
				nodeGroup.ScalingConfig.MaxSize = *agentPool.Properties.MaxCount
			}
		}

		if agentPool.Properties.VMSize != nil {
			nodeGroup.InstanceTypes = []string{*agentPool.Properties.VMSize}
		}

		if agentPool.Properties.OSDiskSizeGB != nil {
			nodeGroup.DiskSize = *agentPool.Properties.OSDiskSizeGB
		}

		if agentPool.Properties.ProvisioningState != nil {
			nodeGroup.Status = *agentPool.Properties.ProvisioningState
		}

		if agentPool.Properties.NodeLabels != nil {
			labels := make(map[string]string)
			for k, v := range agentPool.Properties.NodeLabels {
				if v != nil {
					labels[k] = *v
				}
			}
			nodeGroup.Labels = labels
		}
	}

	return &nodeGroup, nil
}
