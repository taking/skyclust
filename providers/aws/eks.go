package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	// TODO: Import generated proto when available
	// providerv1 "skyclust/api/gen/v1"
)

// AWSKubernetesServer implements the KubernetesService gRPC interface for AWS EKS
type AWSKubernetesServer struct {
	// TODO: Uncomment when proto is generated
	// providerv1.UnimplementedKubernetesServiceServer

	eksClient *eks.Client
	config    aws.Config
	region    string
}

// NewAWSKubernetesServer creates a new AWS EKS Kubernetes service
func NewAWSKubernetesServer(cfg aws.Config, region string) *AWSKubernetesServer {
	return &AWSKubernetesServer{
		eksClient: eks.NewFromConfig(cfg),
		config:    cfg,
		region:    region,
	}
}

// CreateCluster creates a new EKS cluster
func (s *AWSKubernetesServer) CreateCluster(ctx context.Context, req *CreateClusterRequest) (*CreateClusterResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster name is required")
	}

	// Prepare EKS cluster configuration
	input := &eks.CreateClusterInput{
		Name:    aws.String(req.Name),
		Version: aws.String(req.Version),
		ResourcesVpcConfig: &types.VpcConfigRequest{
			SubnetIds: req.SubnetIds,
		},
		Tags: req.Tags,
	}

	// Add networking configuration
	if req.Networking != nil {
		input.KubernetesNetworkConfig = &types.KubernetesNetworkConfigRequest{
			ServiceIpv4Cidr: aws.String(req.Networking.ServiceCidr),
		}

		if len(req.Networking.SecurityGroupIds) > 0 {
			input.ResourcesVpcConfig.SecurityGroupIds = req.Networking.SecurityGroupIds
		}
	}

	// Create cluster
	result, err := s.eksClient.CreateCluster(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create EKS cluster: %v", err)
	}

	cluster := &Cluster{
		Id:        aws.ToString(result.Cluster.Name),
		Name:      aws.ToString(result.Cluster.Name),
		Status:    string(result.Cluster.Status),
		Version:   aws.ToString(result.Cluster.Version),
		Region:    s.region,
		Tags:      req.Tags,
		CreatedAt: timestamppb.New(*result.Cluster.CreatedAt),
	}

	if result.Cluster.Endpoint != nil {
		cluster.Endpoint = aws.ToString(result.Cluster.Endpoint)
	}

	if result.Cluster.ResourcesVpcConfig != nil {
		cluster.VpcId = aws.ToString(result.Cluster.ResourcesVpcConfig.VpcId)
		cluster.SubnetIds = result.Cluster.ResourcesVpcConfig.SubnetIds
	}

	return &CreateClusterResponse{
		Cluster: cluster,
	}, nil
}

// DeleteCluster deletes an EKS cluster
func (s *AWSKubernetesServer) DeleteCluster(ctx context.Context, req *DeleteClusterRequest) (*DeleteClusterResponse, error) {
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	input := &eks.DeleteClusterInput{
		Name: aws.String(req.ClusterId),
	}

	_, err := s.eksClient.DeleteCluster(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete EKS cluster: %v", err)
	}

	return &DeleteClusterResponse{
		Success: true,
		Message: fmt.Sprintf("EKS cluster %s deletion initiated", req.ClusterId),
	}, nil
}

// ListClusters lists all EKS clusters
func (s *AWSKubernetesServer) ListClusters(ctx context.Context, req *ListClustersRequest) (*ListClustersResponse, error) {
	input := &eks.ListClustersInput{}

	result, err := s.eksClient.ListClusters(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list EKS clusters: %v", err)
	}

	var clusters []*Cluster
	for _, clusterName := range result.Clusters {
		// Get detailed cluster information
		describeInput := &eks.DescribeClusterInput{
			Name: aws.String(clusterName),
		}

		describeResult, err := s.eksClient.DescribeCluster(ctx, describeInput)
		if err != nil {
			continue // Skip clusters we can't describe
		}

		cluster := &Cluster{
			Id:      clusterName,
			Name:    clusterName,
			Status:  string(describeResult.Cluster.Status),
			Version: aws.ToString(describeResult.Cluster.Version),
			Region:  s.region,
		}

		if describeResult.Cluster.Endpoint != nil {
			cluster.Endpoint = aws.ToString(describeResult.Cluster.Endpoint)
		}

		if describeResult.Cluster.ResourcesVpcConfig != nil {
			cluster.VpcId = aws.ToString(describeResult.Cluster.ResourcesVpcConfig.VpcId)
			cluster.SubnetIds = describeResult.Cluster.ResourcesVpcConfig.SubnetIds
		}

		if describeResult.Cluster.CreatedAt != nil {
			cluster.CreatedAt = timestamppb.New(*describeResult.Cluster.CreatedAt)
		}

		clusters = append(clusters, cluster)
	}

	return &ListClustersResponse{
		Clusters: clusters,
	}, nil
}

// GetCluster gets details of a specific EKS cluster
func (s *AWSKubernetesServer) GetCluster(ctx context.Context, req *GetClusterRequest) (*GetClusterResponse, error) {
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	input := &eks.DescribeClusterInput{
		Name: aws.String(req.ClusterId),
	}

	result, err := s.eksClient.DescribeCluster(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get EKS cluster: %v", err)
	}

	cluster := &Cluster{
		Id:      aws.ToString(result.Cluster.Name),
		Name:    aws.ToString(result.Cluster.Name),
		Status:  string(result.Cluster.Status),
		Version: aws.ToString(result.Cluster.Version),
		Region:  s.region,
	}

	if result.Cluster.Endpoint != nil {
		cluster.Endpoint = aws.ToString(result.Cluster.Endpoint)
	}

	if result.Cluster.ResourcesVpcConfig != nil {
		cluster.VpcId = aws.ToString(result.Cluster.ResourcesVpcConfig.VpcId)
		cluster.SubnetIds = result.Cluster.ResourcesVpcConfig.SubnetIds
	}

	if result.Cluster.CreatedAt != nil {
		cluster.CreatedAt = timestamppb.New(*result.Cluster.CreatedAt)
	}

	if result.Cluster.Tags != nil {
		cluster.Tags = result.Cluster.Tags
	}

	return &GetClusterResponse{
		Cluster: cluster,
	}, nil
}

// GetClusterKubeconfig gets the kubeconfig for an EKS cluster
func (s *AWSKubernetesServer) GetClusterKubeconfig(ctx context.Context, req *GetClusterKubeconfigRequest) (*GetClusterKubeconfigResponse, error) {
	if req.ClusterId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id is required")
	}

	// Get cluster details
	describeInput := &eks.DescribeClusterInput{
		Name: aws.String(req.ClusterId),
	}

	result, err := s.eksClient.DescribeCluster(ctx, describeInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to describe EKS cluster: %v", err)
	}

	// Generate kubeconfig
	kubeconfig := generateEKSKubeconfig(
		aws.ToString(result.Cluster.Name),
		aws.ToString(result.Cluster.Endpoint),
		aws.ToString(result.Cluster.CertificateAuthority.Data),
		s.region,
	)

	return &GetClusterKubeconfigResponse{
		Kubeconfig: kubeconfig,
	}, nil
}

// CreateNodePool creates a new EKS node group
func (s *AWSKubernetesServer) CreateNodePool(ctx context.Context, req *CreateNodePoolRequest) (*CreateNodePoolResponse, error) {
	if req.ClusterId == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id and name are required")
	}

	input := &eks.CreateNodegroupInput{
		ClusterName:   aws.String(req.ClusterId),
		NodegroupName: aws.String(req.Name),
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(req.DesiredSize),
			MinSize:     aws.Int32(req.MinSize),
			MaxSize:     aws.Int32(req.MaxSize),
		},
		Subnets:       req.SubnetIds,
		InstanceTypes: []string{req.InstanceType},
		Labels:        req.Labels,
		Tags:          req.Tags,
	}

	result, err := s.eksClient.CreateNodegroup(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create EKS node group: %v", err)
	}

	nodePool := &NodePool{
		Id:           aws.ToString(result.Nodegroup.NodegroupName),
		Name:         aws.ToString(result.Nodegroup.NodegroupName),
		ClusterId:    req.ClusterId,
		Status:       string(result.Nodegroup.Status),
		InstanceType: req.InstanceType,
		DesiredSize:  req.DesiredSize,
		MinSize:      req.MinSize,
		MaxSize:      req.MaxSize,
		SubnetIds:    req.SubnetIds,
		Labels:       req.Labels,
		Tags:         req.Tags,
	}

	if result.Nodegroup.CreatedAt != nil {
		nodePool.CreatedAt = timestamppb.New(*result.Nodegroup.CreatedAt)
	}

	return &CreateNodePoolResponse{
		NodePool: nodePool,
	}, nil
}

// DeleteNodePool deletes an EKS node group
func (s *AWSKubernetesServer) DeleteNodePool(ctx context.Context, req *DeleteNodePoolRequest) (*DeleteNodePoolResponse, error) {
	if req.ClusterId == "" || req.NodePoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id and node_pool_id are required")
	}

	input := &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(req.ClusterId),
		NodegroupName: aws.String(req.NodePoolId),
	}

	_, err := s.eksClient.DeleteNodegroup(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete EKS node group: %v", err)
	}

	return &DeleteNodePoolResponse{
		Success: true,
		Message: fmt.Sprintf("Node group %s deletion initiated", req.NodePoolId),
	}, nil
}

// ScaleNodePool scales an EKS node group
func (s *AWSKubernetesServer) ScaleNodePool(ctx context.Context, req *ScaleNodePoolRequest) (*ScaleNodePoolResponse, error) {
	if req.ClusterId == "" || req.NodePoolId == "" {
		return nil, status.Error(codes.InvalidArgument, "cluster_id and node_pool_id are required")
	}

	// Get current node group configuration
	describeInput := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(req.ClusterId),
		NodegroupName: aws.String(req.NodePoolId),
	}

	describeResult, err := s.eksClient.DescribeNodegroup(ctx, describeInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to describe node group: %v", err)
	}

	// Update scaling configuration
	updateInput := &eks.UpdateNodegroupConfigInput{
		ClusterName:   aws.String(req.ClusterId),
		NodegroupName: aws.String(req.NodePoolId),
		ScalingConfig: &types.NodegroupScalingConfig{
			DesiredSize: aws.Int32(req.DesiredSize),
			MinSize:     describeResult.Nodegroup.ScalingConfig.MinSize,
			MaxSize:     describeResult.Nodegroup.ScalingConfig.MaxSize,
		},
	}

	_, err = s.eksClient.UpdateNodegroupConfig(ctx, updateInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to scale node group: %v", err)
	}

	return &ScaleNodePoolResponse{
		Success: true,
		Message: fmt.Sprintf("Node group %s scaled to %d nodes", req.NodePoolId, req.DesiredSize),
	}, nil
}

// Helper function to generate EKS kubeconfig
func generateEKSKubeconfig(clusterName, endpoint, ca string, region string) string {
	caDecoded, _ := base64.StdEncoding.DecodeString(ca)
	caEncoded := base64.StdEncoding.EncodeToString(caDecoded)

	return fmt.Sprintf(`apiVersion: v1
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
`, caEncoded, endpoint, clusterName, clusterName, clusterName, clusterName, clusterName, clusterName, clusterName, region)
}

// Placeholder types - will be replaced by generated proto
type (
	CreateClusterRequest struct {
		Name       string
		Region     string
		Version    string
		VpcId      string
		SubnetIds  []string
		Networking *ClusterNetworking
		Tags       map[string]string
	}
	CreateClusterResponse struct {
		Cluster *Cluster
	}
	DeleteClusterRequest struct {
		ClusterId string
		Region    string
	}
	DeleteClusterResponse struct {
		Success bool
		Message string
	}
	ListClustersRequest struct {
		Region  string
		Filters map[string]string
	}
	ListClustersResponse struct {
		Clusters []*Cluster
	}
	GetClusterRequest struct {
		ClusterId string
		Region    string
	}
	GetClusterResponse struct {
		Cluster *Cluster
	}
	GetClusterKubeconfigRequest struct {
		ClusterId string
		Region    string
	}
	GetClusterKubeconfigResponse struct {
		Kubeconfig string
	}
	CreateNodePoolRequest struct {
		ClusterId    string
		Name         string
		InstanceType string
		DesiredSize  int32
		MinSize      int32
		MaxSize      int32
		SubnetIds    []string
		Labels       map[string]string
		Taints       []string
		Tags         map[string]string
	}
	CreateNodePoolResponse struct {
		NodePool *NodePool
	}
	DeleteNodePoolRequest struct {
		ClusterId  string
		NodePoolId string
		Region     string
	}
	DeleteNodePoolResponse struct {
		Success bool
		Message string
	}
	ScaleNodePoolRequest struct {
		ClusterId   string
		NodePoolId  string
		DesiredSize int32
		Region      string
	}
	ScaleNodePoolResponse struct {
		Success bool
		Message string
	}
	Cluster struct {
		Id         string
		Name       string
		Status     string
		Version    string
		Endpoint   string
		Region     string
		VpcId      string
		SubnetIds  []string
		Networking *ClusterNetworking
		Tags       map[string]string
		CreatedAt  *timestamppb.Timestamp
	}
	ClusterNetworking struct {
		ServiceCidr      string
		PodCidr          string
		SecurityGroupIds []string
	}
	NodePool struct {
		Id           string
		Name         string
		ClusterId    string
		Status       string
		InstanceType string
		DesiredSize  int32
		MinSize      int32
		MaxSize      int32
		CurrentSize  int32
		SubnetIds    []string
		Labels       map[string]string
		Taints       []string
		Tags         map[string]string
		CreatedAt    *timestamppb.Timestamp
	}
)
