package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"skyclust/internal/application/dto"
	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"go.uber.org/zap"
	"google.golang.org/api/container/v1"
	"google.golang.org/api/option"
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

// CreateEKSCluster creates an AWS EKS cluster (for backward compatibility)
func (s *KubernetesService) CreateEKSCluster(ctx context.Context, credential *domain.Credential, req dto.CreateClusterRequest) (*dto.CreateClusterResponse, error) {
	return s.createAWSEKSCluster(ctx, credential, req)
}

// CreateGCPGKECluster creates a GCP GKE cluster with new sectioned structure
func (s *KubernetesService) CreateGCPGKECluster(ctx context.Context, credential *domain.Credential, req dto.CreateGKEClusterRequest) (*dto.CreateClusterResponse, error) {
	return s.createGCPGKEClusterWithAdvanced(ctx, credential, req)
}

// createGCPGKEClusterWithAdvanced creates a GCP GKE cluster with advanced configuration
func (s *KubernetesService) createGCPGKEClusterWithAdvanced(ctx context.Context, credential *domain.Credential, req dto.CreateGKEClusterRequest) (*dto.CreateClusterResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Create GCP Container service
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP container service: %w", err)
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
		return nil, fmt.Errorf("network configuration is required")
	}

	// Node pool configuration for standard clusters
	if clusterType == "standard" {
		if req.NodePool == nil {
			return nil, fmt.Errorf("node pool configuration is required for standard clusters")
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
		return nil, fmt.Errorf("failed to create GCP GKE cluster: %w", err)
	}

	s.logger.Info("GCP GKE cluster creation initiated",
		zap.String("cluster_name", req.Name),
		zap.String("project_id", req.ProjectID),
		zap.String("location", location),
		zap.String("cluster_type", clusterType))

	// Build response
	response := &dto.CreateClusterResponse{
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

	return response, nil
}

// buildNodePoolConfig builds a GCP node pool configuration
func (s *KubernetesService) buildNodePoolConfig(nodePool *dto.GKENodePoolConfig) *container.NodePool {
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

// convertToGCPTagKey converts tag keys to GCP-compatible format
// GCP requires tag keys to start with lowercase letters and contain only
// lowercase letters, numbers, underscores, and dashes
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

// isLetter checks if a character is a letter
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
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
		return s.listGCPGKEClusters(ctx, credential, region)
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
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.getAWSEKSCluster(ctx, credential, clusterName, region)
	case "gcp":
		return s.getGCPGKECluster(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS cluster retrieval
		return nil, fmt.Errorf("Azure AKS cluster retrieval not implemented yet")
	case "ncp":
		// TODO: Implement NCP NKS cluster retrieval
		return nil, fmt.Errorf("NCP NKS cluster retrieval not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported provider: %s", credential.Provider)
	}
}

// getAWSEKSCluster gets AWS EKS cluster details
func (s *KubernetesService) getAWSEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*dto.ClusterInfo, error) {
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

// DeleteEKSCluster deletes a Kubernetes cluster
func (s *KubernetesService) DeleteEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.deleteAWSEKSCluster(ctx, credential, clusterName, region)
	case "gcp":
		return s.deleteGCPGKECluster(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS cluster deletion
		return fmt.Errorf("Azure AKS cluster deletion not implemented yet")
	case "ncp":
		// TODO: Implement NCP NKS cluster deletion
		return fmt.Errorf("NCP NKS cluster deletion not implemented yet")
	default:
		return fmt.Errorf("unsupported provider: %s", credential.Provider)
	}
}

// deleteAWSEKSCluster deletes an AWS EKS cluster
func (s *KubernetesService) deleteAWSEKSCluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
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

// GetEKSKubeconfig generates kubeconfig for a Kubernetes cluster
func (s *KubernetesService) GetEKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	// Route to provider-specific implementation
	switch credential.Provider {
	case "aws":
		return s.getAWSEKSKubeconfig(ctx, credential, clusterName, region)
	case "gcp":
		return s.getGCPGKEKubeconfig(ctx, credential, clusterName, region)
	case "azure":
		// TODO: Implement Azure AKS kubeconfig generation
		return "", fmt.Errorf("Azure AKS kubeconfig generation not implemented yet")
	case "ncp":
		// TODO: Implement NCP NKS kubeconfig generation
		return "", fmt.Errorf("NCP NKS kubeconfig generation not implemented yet")
	default:
		return "", fmt.Errorf("unsupported provider: %s", credential.Provider)
	}
}

// getAWSEKSKubeconfig generates kubeconfig for an AWS EKS cluster
func (s *KubernetesService) getAWSEKSKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
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

// listGCPGKEClusters lists all GCP GKE clusters
func (s *KubernetesService) listGCPGKEClusters(ctx context.Context, credential *domain.Credential, region string) (*dto.ListClustersResponse, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Extract project ID
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id not found in credential")
	}

	// Create Container service client
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP container service: %w", err)
	}

	// List clusters in the region and all zones of the specified region
	// GCP clusters can be at region level or zone level
	locations := []string{
		region,                      // Region level (e.g., asia-northeast3)
		fmt.Sprintf("%s-a", region), // Zone level (e.g., asia-northeast3-a)
		fmt.Sprintf("%s-b", region), // Zone level (e.g., asia-northeast3-b)
		fmt.Sprintf("%s-c", region), // Zone level (e.g., asia-northeast3-c)
	}

	var allClusters []dto.ClusterInfo

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
			clusterInfo := dto.ClusterInfo{
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
				clusterInfo.NetworkConfig = &dto.NetworkConfigInfo{
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

				clusterInfo.NodePoolInfo = &dto.NodePoolSummaryInfo{
					TotalNodePools: int32(len(cluster.NodePools)),
					TotalNodes:     totalNodes,
					MinNodes:       minNodes,
					MaxNodes:       maxNodes,
				}
			}

			// Add security configuration
			if cluster.WorkloadIdentityConfig != nil || cluster.BinaryAuthorization != nil || cluster.NetworkPolicy != nil {
				clusterInfo.SecurityConfig = &dto.SecurityConfigInfo{
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

	return &dto.ListClustersResponse{Clusters: allClusters}, nil
}

// getGCPGKECluster gets GCP GKE cluster details
func (s *KubernetesService) getGCPGKECluster(ctx context.Context, credential *domain.Credential, clusterName, region string) (*dto.ClusterInfo, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Extract project ID
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id not found in credential")
	}

	// Create Container service client
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP container service: %w", err)
	}

	// Get cluster details
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, region, clusterName)
	cluster, err := containerService.Projects.Locations.Clusters.Get(clusterPath).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get GKE cluster: %w", err)
	}

	// Convert to ClusterInfo
	clusterZone := extractZoneFromLocation(cluster.Location)

	// Build detailed cluster information
	clusterInfo := dto.ClusterInfo{
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
		clusterInfo.NetworkConfig = &dto.NetworkConfigInfo{
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

		clusterInfo.NodePoolInfo = &dto.NodePoolSummaryInfo{
			TotalNodePools: int32(len(cluster.NodePools)),
			TotalNodes:     totalNodes,
			MinNodes:       minNodes,
			MaxNodes:       maxNodes,
		}
	}

	// Add security configuration
	if cluster.WorkloadIdentityConfig != nil || cluster.BinaryAuthorization != nil || cluster.NetworkPolicy != nil {
		clusterInfo.SecurityConfig = &dto.SecurityConfigInfo{
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

// extractZoneFromLocation extracts zone from GCP cluster location
// Location format: "projects/PROJECT_ID/locations/ZONE" or "projects/PROJECT_ID/locations/REGION"
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

// getGCPGKEKubeconfig generates kubeconfig for a GCP GKE cluster
func (s *KubernetesService) getGCPGKEKubeconfig(ctx context.Context, credential *domain.Credential, clusterName, region string) (string, error) {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Extract project ID
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return "", fmt.Errorf("project_id not found in credential")
	}

	// Create Container service client
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create GCP container service: %w", err)
	}

	// Get cluster details - search in region and all zones
	// GCP clusters can be at region level or zone level
	locations := []string{
		region,                      // Region level (e.g., asia-northeast3)
		fmt.Sprintf("%s-a", region), // Zone level (e.g., asia-northeast3-a)
		fmt.Sprintf("%s-b", region), // Zone level (e.g., asia-northeast3-b)
		fmt.Sprintf("%s-c", region), // Zone level (e.g., asia-northeast3-c)
	}

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
		return "", fmt.Errorf("failed to find GKE cluster %s in region %s or any of its zones", clusterName, region)
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

// deleteGCPGKECluster deletes a GCP GKE cluster
func (s *KubernetesService) deleteGCPGKECluster(ctx context.Context, credential *domain.Credential, clusterName, region string) error {
	// Decrypt credential data
	credData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Convert credential data to JSON
	jsonData, err := json.Marshal(credData)
	if err != nil {
		return fmt.Errorf("failed to marshal credential data: %w", err)
	}

	// Extract project ID
	projectID, ok := credData["project_id"].(string)
	if !ok {
		return fmt.Errorf("project_id not found in credential")
	}

	// Create Container service client
	containerService, err := container.NewService(ctx, option.WithCredentialsJSON(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create GCP container service: %w", err)
	}

	// Find cluster location - search in region and all zones
	locations := []string{
		region,                      // Region level (e.g., asia-northeast3)
		fmt.Sprintf("%s-a", region), // Zone level (e.g., asia-northeast3-a)
		fmt.Sprintf("%s-b", region), // Zone level (e.g., asia-northeast3-b)
		fmt.Sprintf("%s-c", region), // Zone level (e.g., asia-northeast3-c)
	}

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
		return fmt.Errorf("failed to find GKE cluster %s in region %s or any of its zones", clusterName, region)
	}

	// Delete cluster
	clusterPath := fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, clusterLocation, clusterName)
	operation, err := containerService.Projects.Locations.Clusters.Delete(clusterPath).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete GKE cluster: %w", err)
	}

	s.logger.Info("GCP GKE cluster deletion initiated",
		zap.String("project_id", projectID),
		zap.String("cluster_name", clusterName),
		zap.String("location", clusterLocation),
		zap.String("operation_name", operation.Name))

	return nil
}
