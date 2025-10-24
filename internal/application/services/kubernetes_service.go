package service

import (
	"context"
	"fmt"

	"skyclust/internal/application/dto"
	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"go.uber.org/zap"
)

// KubernetesService handles Kubernetes cluster operations
type KubernetesService struct {
	credentialService domain.CredentialService
	logger            *zap.Logger
}

// NewKubernetesService creates a new Kubernetes service
func NewKubernetesService(credentialService domain.CredentialService, logger *zap.Logger) *KubernetesService {
	return &KubernetesService{
		credentialService: credentialService,
		logger:            logger,
	}
}

// CreateEKSCluster creates a new Kubernetes cluster (supports multiple providers)
// For AWS: EKS, For GCP: GKE, For Azure: AKS, For NCP: NKS
func (s *KubernetesService) CreateEKSCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.createAWSEKSCluster(ctx, credential, req)
	case "gcp":
		return s.createGCPGKECluster(ctx, credential, req)
	case "azure":
		return s.createAzureAKSCluster(ctx, credential, req)
	case "ncp":
		return s.createNCPNKSCluster(ctx, credential, req)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", credential.Provider)
	}
}

// createAWSEKSCluster creates an AWS EKS cluster
func (s *KubernetesService) createAWSEKSCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
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
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
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
		return nil, fmt.Errorf("failed to create EKS cluster: %w", err)
	}

	// Convert to response
	response := &dto.CreateClusterResponse{
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

	return response, nil
}

// ListEKSClusters lists all Kubernetes clusters (supports multiple providers)
func (s *KubernetesService) ListEKSClusters(ctx context.Context, credential *domain.Credential, region string) (*dto.ListClustersResponse, error) {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.listAWSEKSClusters(ctx, credential, region)
	case "gcp":
		// TODO: Implement GCP GKE cluster listing
		return &dto.ListClustersResponse{Clusters: []dto.ClusterInfo{}}, nil
	case "azure":
		// TODO: Implement Azure AKS cluster listing
		return &dto.ListClustersResponse{Clusters: []dto.ClusterInfo{}}, nil
	case "ncp":
		// TODO: Implement NCP NKS cluster listing
		return &dto.ListClustersResponse{Clusters: []dto.ClusterInfo{}}, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", credential.Provider)
	}
}

// listAWSEKSClusters lists all AWS EKS clusters
func (s *KubernetesService) listAWSEKSClusters(ctx context.Context, credential *domain.Credential, region string) (*dto.ListClustersResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Debug log - check decrypted data
	s.logger.Debug("Decrypted credential keys",
		zap.Strings("keys", getMapKeys(credData)),
		zap.Int("key_count", len(credData)))

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		s.logger.Error("access_key not found or wrong type",
			zap.Any("access_key_value", credData["access_key"]),
			zap.String("access_key_type", fmt.Sprintf("%T", credData["access_key"])))
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		s.logger.Error("secret_key not found or wrong type",
			zap.Any("secret_key_value", credData["secret_key"]),
			zap.String("secret_key_type", fmt.Sprintf("%T", credData["secret_key"])))
		return nil, fmt.Errorf("secret_key not found in credential")
	}

	// Debug log - check credential format
	s.logger.Debug("AWS credentials extracted",
		zap.String("access_key_prefix", accessKey[:min(10, len(accessKey))]),
		zap.Int("access_key_length", len(accessKey)),
		zap.Int("secret_key_length", len(secretKey)))

	// Use region from credential if not specified
	if region == "" {
		if r, ok := credData["region"].(string); ok {
			region = r
		} else {
			region = "us-east-1" // Default region
		}
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List clusters
	output, err := eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	// Get detailed information for each cluster
	var clusters []dto.ClusterInfo
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

		cluster := dto.ClusterInfo{
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

	return &dto.ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// Helper function to get map keys
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// GetEKSCluster gets details of an EKS cluster by name
func (s *KubernetesService) GetEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*dto.ClusterInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe cluster
	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EKS cluster: %w", err)
	}

	// Convert to ClusterInfo
	cluster := dto.ClusterInfo{
		ID:      aws.ToString(output.Cluster.Arn),
		Name:    aws.ToString(output.Cluster.Name),
		Version: aws.ToString(output.Cluster.Version),
		Status:  string(output.Cluster.Status),
		Region:  region,
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

// DeleteEKSCluster deletes an EKS cluster
func (s *KubernetesService) DeleteEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return fmt.Errorf("secret_key not found in credential")
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Delete cluster
	_, err = eksClient.DeleteCluster(ctx, &eks.DeleteClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete EKS cluster: %w", err)
	}

	s.logger.Info("EKS cluster deletion initiated",
		zap.String("cluster_name", clusterName),
		zap.String("region", region))

	return nil
}

// GetEKSKubeconfig generates kubeconfig for an EKS cluster
func (s *KubernetesService) GetEKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return "", fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return "", fmt.Errorf("secret_key not found in credential")
	}

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe cluster to get endpoint and CA
	output, err := eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe EKS cluster: %w", err)
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
		region,
		accessKey,
		secretKey,
	)

	return kubeconfig, nil
}

// CreateEKSNodePool creates a node pool (node group) for an EKS cluster
func (s *KubernetesService) CreateEKSNodePool(ctx context.Context, credential *domain.Credential, req dto.CreateNodePoolRequest) (map[string]interface{}, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
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
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
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
		return nil, fmt.Errorf("failed to create node group: %w", err)
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

	return result, nil
}

// CreateEKSNodeGroup creates an EKS node group
func (s *KubernetesService) CreateEKSNodeGroup(ctx context.Context, credential *domain.Credential, req dto.CreateNodeGroupRequest) (*dto.CreateNodeGroupResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
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
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
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
		return nil, fmt.Errorf("failed to create node group: %w", err)
	}

	// Convert to response
	response := &dto.CreateNodeGroupResponse{
		NodeGroupName: aws.ToString(output.Nodegroup.NodegroupName),
		ClusterName:   aws.ToString(output.Nodegroup.ClusterName),
		Status:        string(output.Nodegroup.Status),
		InstanceTypes: output.Nodegroup.InstanceTypes,
		ScalingConfig: dto.NodeGroupScalingConfig{
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

// ListNodeGroups lists all node groups for a cluster
func (s *KubernetesService) ListNodeGroups(ctx context.Context, credential *domain.Credential, req dto.ListNodeGroupsRequest) (*dto.ListNodeGroupsResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
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
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// List node groups
	output, err := eksClient.ListNodegroups(ctx, &eks.ListNodegroupsInput{
		ClusterName: aws.String(req.ClusterName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list node groups: %w", err)
	}

	// Get detailed information for each node group
	var nodeGroups []dto.NodeGroupInfo
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

		nodeGroup := dto.NodeGroupInfo{
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
			nodeGroup.ScalingConfig = dto.NodeGroupScalingConfig{
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

	return &dto.ListNodeGroupsResponse{
		NodeGroups: nodeGroups,
		Total:      len(nodeGroups),
	}, nil
}

// GetNodeGroup gets details of a node group
func (s *KubernetesService) GetNodeGroup(ctx context.Context, credential *domain.Credential, req dto.GetNodeGroupRequest) (*dto.NodeGroupInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
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
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Describe node group
	output, err := eksClient.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe node group: %w", err)
	}

	// Convert to NodeGroupInfo
	nodeGroup := dto.NodeGroupInfo{
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
		nodeGroup.ScalingConfig = dto.NodeGroupScalingConfig{
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

	return &nodeGroup, nil
}

// DeleteNodeGroup deletes a node group
func (s *KubernetesService) DeleteNodeGroup(ctx context.Context, credential *domain.Credential, req dto.DeleteNodeGroupRequest) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials
	accessKey, ok := credData["access_key"].(string)
	if !ok {
		return fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := credData["secret_key"].(string)
	if !ok {
		return fmt.Errorf("secret_key not found in credential")
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
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EKS client
	eksClient := eks.NewFromConfig(cfg)

	// Delete node group
	_, err = eksClient.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(req.ClusterName),
		NodegroupName: aws.String(req.NodeGroupName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete node group: %w", err)
	}

	s.logger.Info("EKS node group deletion initiated",
		zap.String("cluster_name", req.ClusterName),
		zap.String("nodegroup_name", req.NodeGroupName),
		zap.String("region", req.Region))

	return nil
}
