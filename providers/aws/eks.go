package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eksTypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
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
	ec2Client *ec2.Client
	iamClient *iam.Client
	config    aws.Config
	region    string
}

// NewAWSKubernetesServer creates a new AWS EKS Kubernetes service
func NewAWSKubernetesServer(cfg aws.Config, region string) *AWSKubernetesServer {
	return &AWSKubernetesServer{
		eksClient: eks.NewFromConfig(cfg),
		ec2Client: ec2.NewFromConfig(cfg),
		iamClient: iam.NewFromConfig(cfg),
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
		ResourcesVpcConfig: &eksTypes.VpcConfigRequest{
			SubnetIds: req.SubnetIds,
		},
		Tags: req.Tags,
	}

	// Add IAM role if provided
	if req.RoleArn != "" {
		input.RoleArn = aws.String(req.RoleArn)
	}

	// Add networking configuration
	if req.Networking != nil {
		input.KubernetesNetworkConfig = &eksTypes.KubernetesNetworkConfigRequest{
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
		ScalingConfig: &eksTypes.NodegroupScalingConfig{
			DesiredSize: aws.Int32(req.DesiredSize),
			MinSize:     aws.Int32(req.MinSize),
			MaxSize:     aws.Int32(req.MaxSize),
		},
		Subnets:       req.SubnetIds,
		InstanceTypes: []string{req.InstanceType},
		Labels:        req.Labels,
		Tags:          req.Tags,
	}

	// Add IAM role if provided
	if req.NodeRoleArn != "" {
		input.NodeRole = aws.String(req.NodeRoleArn)
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
		ScalingConfig: &eksTypes.NodegroupScalingConfig{
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

// ListVPCs lists all VPCs in the region
func (s *AWSKubernetesServer) ListVPCs(ctx context.Context, req *ListVPCsRequest) (*ListVPCsResponse, error) {
	input := &ec2.DescribeVpcsInput{}

	result, err := s.ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list VPCs: %v", err)
	}

	var vpcs []*VPC
	for _, vpc := range result.Vpcs {
		vpcInfo := &VPC{
			Id:        aws.ToString(vpc.VpcId),
			CidrBlock: aws.ToString(vpc.CidrBlock),
			State:     string(vpc.State),
			IsDefault: aws.ToBool(vpc.IsDefault),
			Region:    s.region,
		}

		// Get VPC name from tags
		if vpc.Tags != nil {
			for _, tag := range vpc.Tags {
				if aws.ToString(tag.Key) == "Name" {
					vpcInfo.Name = aws.ToString(tag.Value)
					break
				}
			}
		}

		// Get availability zones for this VPC
		subnets, err := s.getSubnetsByVPC(ctx, aws.ToString(vpc.VpcId))
		if err == nil {
			azs := make(map[string]bool)
			for _, subnet := range subnets {
				azs[subnet.AvailabilityZone] = true
			}
			for az := range azs {
				vpcInfo.AvailabilityZones = append(vpcInfo.AvailabilityZones, az)
			}
		}

		vpcs = append(vpcs, vpcInfo)
	}

	return &ListVPCsResponse{
		VPCs: vpcs,
	}, nil
}

// ListSubnets lists subnets for a specific VPC
func (s *AWSKubernetesServer) ListSubnets(ctx context.Context, req *ListSubnetsRequest) (*ListSubnetsResponse, error) {
	if req.VpcId == "" {
		return nil, status.Error(codes.InvalidArgument, "vpc_id is required")
	}

	subnets, err := s.getSubnetsByVPC(ctx, req.VpcId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list subnets: %v", err)
	}

	return &ListSubnetsResponse{
		Subnets: subnets,
	}, nil
}

// getSubnetsByVPC is a helper function to get subnets for a VPC
func (s *AWSKubernetesServer) getSubnetsByVPC(ctx context.Context, vpcId string) ([]*Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{vpcId},
			},
		},
	}

	result, err := s.ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, err
	}

	var subnets []*Subnet
	for _, subnet := range result.Subnets {
		subnetInfo := &Subnet{
			Id:               aws.ToString(subnet.SubnetId),
			VpcId:            aws.ToString(subnet.VpcId),
			CidrBlock:        aws.ToString(subnet.CidrBlock),
			AvailabilityZone: aws.ToString(subnet.AvailabilityZone),
			State:            string(subnet.State),
			Region:           s.region,
		}

		// Get subnet name from tags
		if subnet.Tags != nil {
			for _, tag := range subnet.Tags {
				if aws.ToString(tag.Key) == "Name" {
					subnetInfo.Name = aws.ToString(tag.Value)
					break
				}
			}
		}

		// Check if subnet is public (has route to internet gateway)
		isPublic, err := s.isSubnetPublic(ctx, aws.ToString(subnet.SubnetId))
		if err == nil {
			subnetInfo.IsPublic = isPublic
		}

		subnets = append(subnets, subnetInfo)
	}

	return subnets, nil
}

// isSubnetPublic checks if a subnet is public by looking at its route table
func (s *AWSKubernetesServer) isSubnetPublic(ctx context.Context, subnetId string) (bool, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("association.subnet-id"),
				Values: []string{subnetId},
			},
		},
	}

	result, err := s.ec2Client.DescribeRouteTables(ctx, input)
	if err != nil {
		return false, err
	}

	for _, routeTable := range result.RouteTables {
		for _, route := range routeTable.Routes {
			if route.GatewayId != nil && aws.ToString(route.GatewayId) != "local" {
				// Check if this is an internet gateway
				igwInput := &ec2.DescribeInternetGatewaysInput{
					InternetGatewayIds: []string{aws.ToString(route.GatewayId)},
				}
				igwResult, err := s.ec2Client.DescribeInternetGateways(ctx, igwInput)
				if err == nil && len(igwResult.InternetGateways) > 0 {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// ListSecurityGroups lists security groups for a specific VPC
func (s *AWSKubernetesServer) ListSecurityGroups(ctx context.Context, req *ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	if req.VpcId == "" {
		return nil, status.Error(codes.InvalidArgument, "vpc_id is required")
	}

	input := &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{req.VpcId},
			},
		},
	}

	result, err := s.ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list security groups: %v", err)
	}

	var securityGroups []*SecurityGroup
	for _, sg := range result.SecurityGroups {
		sgInfo := &SecurityGroup{
			Id:          aws.ToString(sg.GroupId),
			Name:        aws.ToString(sg.GroupName),
			Description: aws.ToString(sg.Description),
			VpcId:       aws.ToString(sg.VpcId),
			Region:      s.region,
		}

		// Convert inbound rules
		for _, rule := range sg.IpPermissions {
			ingressRule := &SecurityGroupRule{
				Protocol: aws.ToString(rule.IpProtocol),
				FromPort: aws.ToInt32(rule.FromPort),
				ToPort:   aws.ToInt32(rule.ToPort),
			}
			if len(rule.UserIdGroupPairs) > 0 {
				ingressRule.Source = aws.ToString(rule.UserIdGroupPairs[0].GroupId)
			}
			sgInfo.IngressRules = append(sgInfo.IngressRules, ingressRule)
		}

		// Convert outbound rules
		for _, rule := range sg.IpPermissionsEgress {
			egressRule := &SecurityGroupRule{
				Protocol: aws.ToString(rule.IpProtocol),
				FromPort: aws.ToInt32(rule.FromPort),
				ToPort:   aws.ToInt32(rule.ToPort),
			}
			if len(rule.UserIdGroupPairs) > 0 {
				egressRule.Source = aws.ToString(rule.UserIdGroupPairs[0].GroupId)
			}
			sgInfo.EgressRules = append(sgInfo.EgressRules, egressRule)
		}

		securityGroups = append(securityGroups, sgInfo)
	}

	return &ListSecurityGroupsResponse{
		SecurityGroups: securityGroups,
	}, nil
}

// ListIAMRoles lists IAM roles for EKS
func (s *AWSKubernetesServer) ListIAMRoles(ctx context.Context, req *ListIAMRolesRequest) (*ListIAMRolesResponse, error) {
	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/"),
	}

	result, err := s.iamClient.ListRoles(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list IAM roles: %v", err)
	}

	var roles []*IAMRole
	for _, role := range result.Roles {
		roleInfo := &IAMRole{
			Id:     aws.ToString(role.RoleId),
			Name:   aws.ToString(role.RoleName),
			Arn:    aws.ToString(role.Arn),
			Path:   aws.ToString(role.Path),
			Region: s.region,
		}

		// Get role description
		if role.Description != nil {
			roleInfo.Description = aws.ToString(role.Description)
		}

		// Check if this is an EKS-related role
		roleName := aws.ToString(role.RoleName)
		if contains(roleName, "eks") || contains(roleName, "EKS") {
			roleInfo.IsEKSRelated = true
		}

		// Get role creation date
		if role.CreateDate != nil {
			roleInfo.CreatedAt = timestamppb.New(*role.CreateDate)
		}

		roles = append(roles, roleInfo)
	}

	return &ListIAMRolesResponse{
		Roles: roles,
	}, nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || contains(s[1:], substr))))
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
		RoleArn    string
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
		NodeRoleArn  string
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

	// VPC related types
	ListVPCsRequest struct {
		Region string
	}
	ListVPCsResponse struct {
		VPCs []*VPC
	}
	VPC struct {
		Id                string
		Name              string
		CidrBlock         string
		State             string
		IsDefault         bool
		AvailabilityZones []string
		Region            string
	}

	// Subnet related types
	ListSubnetsRequest struct {
		VpcId  string
		Region string
	}
	ListSubnetsResponse struct {
		Subnets []*Subnet
	}
	Subnet struct {
		Id               string
		Name             string
		VpcId            string
		CidrBlock        string
		AvailabilityZone string
		State            string
		IsPublic         bool
		Region           string
	}

	// Security Group related types
	ListSecurityGroupsRequest struct {
		VpcId  string
		Region string
	}
	ListSecurityGroupsResponse struct {
		SecurityGroups []*SecurityGroup
	}
	SecurityGroup struct {
		Id           string
		Name         string
		Description  string
		VpcId        string
		IngressRules []*SecurityGroupRule
		EgressRules  []*SecurityGroupRule
		Region       string
	}
	SecurityGroupRule struct {
		Protocol    string
		FromPort    int32
		ToPort      int32
		Source      string
		Description string
	}

	// IAM Role related types
	ListIAMRolesRequest struct {
		Region string
	}
	ListIAMRolesResponse struct {
		Roles []*IAMRole
	}
	IAMRole struct {
		Id           string
		Name         string
		Arn          string
		Path         string
		Description  string
		IsEKSRelated bool
		CreatedAt    *timestamppb.Timestamp
		Region       string
	}
)
