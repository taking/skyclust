package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	// TODO: Import generated proto when available
	// providerv1 "skyclust/api/gen/v1"
)

// AWSProviderServer implements the CloudProviderService gRPC interface
type AWSProviderServer struct {
	// TODO: Uncomment when proto is generated
	// providerv1.UnimplementedCloudProviderServiceServer

	config    aws.Config
	ec2Client *ec2.Client
	region    string

	// Credentials
	accessKey string
	secretKey string
}

// NewAWSProviderServer creates a new AWS provider gRPC server
func NewAWSProviderServerWithConfig(accessKey, secretKey, region string) (*AWSProviderServer, error) {
	if region == "" {
		region = "us-east-1"
	}

	server := &AWSProviderServer{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
	}

	if err := server.initializeAWS(); err != nil {
		return nil, fmt.Errorf("failed to initialize AWS: %w", err)
	}

	return server, nil
}

// initializeAWS initializes AWS SDK clients
func (s *AWSProviderServer) initializeAWS() error {
	ctx := context.Background()

	// Create AWS config
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(s.region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			s.accessKey,
			s.secretKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	s.config = cfg
	s.ec2Client = ec2.NewFromConfig(cfg)

	return nil
}

// GetProviderInfo returns information about the AWS provider
func (s *AWSProviderServer) GetProviderInfo(ctx context.Context, req *emptypb.Empty) (*ProviderInfoResponse, error) {
	return &ProviderInfoResponse{
		Name:    "AWS",
		Version: "1.0.0",
		SupportedRegions: []string{
			"us-east-1", "us-east-2", "us-west-1", "us-west-2",
			"eu-west-1", "eu-west-2", "eu-central-1",
			"ap-northeast-1", "ap-northeast-2", "ap-southeast-1",
		},
		Capabilities: map[string]string{
			"compute":    "supported",
			"storage":    "supported",
			"networking": "supported",
			"iam":        "supported",
			"kubernetes": "supported",
		},
	}, nil
}

// ListInstances lists EC2 instances
func (s *AWSProviderServer) ListInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	input := &ec2.DescribeInstancesInput{}

	// Apply filters if provided
	if len(req.Filters) > 0 {
		var filters []types.Filter
		for key, value := range req.Filters {
			filters = append(filters, types.Filter{
				Name:   aws.String(key),
				Values: []string{value},
			})
		}
		input.Filters = filters
	}

	result, err := s.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list instances: %v", err)
	}

	var instances []*Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			inst := &Instance{
				Id:       aws.ToString(instance.InstanceId),
				Name:     getInstanceName(instance.Tags),
				Status:   string(instance.State.Name),
				Type:     string(instance.InstanceType),
				Region:   s.region,
				Provider: "aws",
			}

			if instance.LaunchTime != nil {
				inst.CreatedAt = timestamppb.New(*instance.LaunchTime)
			}

			if instance.PublicIpAddress != nil {
				inst.PublicIp = aws.ToString(instance.PublicIpAddress)
			}

			if instance.PrivateIpAddress != nil {
				inst.PrivateIp = aws.ToString(instance.PrivateIpAddress)
			}

			// Convert tags
			inst.Tags = make(map[string]string)
			for _, tag := range instance.Tags {
				if tag.Key != nil && tag.Value != nil {
					inst.Tags[*tag.Key] = *tag.Value
				}
			}

			instances = append(instances, inst)
		}
	}

	return &ListInstancesResponse{
		Instances:  instances,
		TotalCount: int32(len(instances)),
	}, nil
}

// CreateInstance creates a new EC2 instance
func (s *AWSProviderServer) CreateInstance(ctx context.Context, req *CreateInstanceRequest) (*CreateInstanceResponse, error) {
	if req.Name == "" || req.Type == "" || req.ImageId == "" {
		return nil, status.Error(codes.InvalidArgument, "name, type, and image_id are required")
	}

	// Prepare run instances input
	runInput := &ec2.RunInstancesInput{
		ImageId:      aws.String(req.ImageId),
		InstanceType: types.InstanceType(req.Type),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
	}

	// Add tags
	tagSpecs := []types.TagSpecification{
		{
			ResourceType: types.ResourceTypeInstance,
			Tags: []types.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(req.Name),
				},
			},
		},
	}

	for key, value := range req.Tags {
		tagSpecs[0].Tags = append(tagSpecs[0].Tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	runInput.TagSpecifications = tagSpecs

	// Network configuration
	if req.SubnetId != "" {
		runInput.SubnetId = aws.String(req.SubnetId)
	}

	if len(req.SecurityGroups) > 0 {
		runInput.SecurityGroupIds = req.SecurityGroups
	}

	// Key pair
	if req.KeyPairName != "" {
		runInput.KeyName = aws.String(req.KeyPairName)
	}

	// User data
	if req.UserData != "" {
		runInput.UserData = aws.String(req.UserData)
	}

	// Launch instance
	result, err := s.ec2Client.RunInstances(ctx, runInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create instance: %v", err)
	}

	if len(result.Instances) == 0 {
		return nil, status.Error(codes.Internal, "no instance created")
	}

	instance := result.Instances[0]
	inst := &Instance{
		Id:       aws.ToString(instance.InstanceId),
		Name:     req.Name,
		Status:   string(instance.State.Name),
		Type:     string(instance.InstanceType),
		Region:   s.region,
		Provider: "aws",
		Tags:     req.Tags,
	}

	if instance.LaunchTime != nil {
		inst.CreatedAt = timestamppb.New(*instance.LaunchTime)
	}

	return &CreateInstanceResponse{
		Instance: inst,
	}, nil
}

// DeleteInstance terminates an EC2 instance
func (s *AWSProviderServer) DeleteInstance(ctx context.Context, req *DeleteInstanceRequest) (*DeleteInstanceResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	input := &ec2.TerminateInstancesInput{
		InstanceIds: []string{req.InstanceId},
	}

	_, err := s.ec2Client.TerminateInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete instance: %v", err)
	}

	return &DeleteInstanceResponse{
		Success: true,
		Message: fmt.Sprintf("Instance %s terminated successfully", req.InstanceId),
	}, nil
}

// GetInstanceStatus returns the status of an EC2 instance
func (s *AWSProviderServer) GetInstanceStatus(ctx context.Context, req *GetInstanceStatusRequest) (*GetInstanceStatusResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{req.InstanceId},
	}

	result, err := s.ec2Client.DescribeInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get instance status: %v", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	instance := result.Reservations[0].Instances[0]
	return &GetInstanceStatusResponse{
		InstanceId:  req.InstanceId,
		Status:      string(instance.State.Name),
		LastUpdated: timestamppb.Now(),
	}, nil
}

// StartInstance starts a stopped EC2 instance
func (s *AWSProviderServer) StartInstance(ctx context.Context, req *StartInstanceRequest) (*StartInstanceResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	input := &ec2.StartInstancesInput{
		InstanceIds: []string{req.InstanceId},
	}

	_, err := s.ec2Client.StartInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start instance: %v", err)
	}

	return &StartInstanceResponse{
		Success: true,
		Message: fmt.Sprintf("Instance %s started successfully", req.InstanceId),
	}, nil
}

// StopInstance stops a running EC2 instance
func (s *AWSProviderServer) StopInstance(ctx context.Context, req *StopInstanceRequest) (*StopInstanceResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	input := &ec2.StopInstancesInput{
		InstanceIds: []string{req.InstanceId},
		Force:       aws.Bool(req.Force),
	}

	_, err := s.ec2Client.StopInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to stop instance: %v", err)
	}

	return &StopInstanceResponse{
		Success: true,
		Message: fmt.Sprintf("Instance %s stopped successfully", req.InstanceId),
	}, nil
}

// RestartInstance restarts an EC2 instance
func (s *AWSProviderServer) RestartInstance(ctx context.Context, req *RestartInstanceRequest) (*RestartInstanceResponse, error) {
	if req.InstanceId == "" {
		return nil, status.Error(codes.InvalidArgument, "instance_id is required")
	}

	input := &ec2.RebootInstancesInput{
		InstanceIds: []string{req.InstanceId},
	}

	_, err := s.ec2Client.RebootInstances(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to restart instance: %v", err)
	}

	return &RestartInstanceResponse{
		Success: true,
		Message: fmt.Sprintf("Instance %s restarted successfully", req.InstanceId),
	}, nil
}

// ListRegions lists available AWS regions
func (s *AWSProviderServer) ListRegions(ctx context.Context, req *ListRegionsRequest) (*ListRegionsResponse, error) {
	input := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false),
	}

	result, err := s.ec2Client.DescribeRegions(ctx, input)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list regions: %v", err)
	}

	var regions []*Region
	for _, region := range result.Regions {
		regions = append(regions, &Region{
			Id:          aws.ToString(region.RegionName),
			Name:        aws.ToString(region.RegionName),
			DisplayName: aws.ToString(region.RegionName),
			Status:      "available",
		})
	}

	return &ListRegionsResponse{
		Regions: regions,
	}, nil
}

// GetCostEstimate provides cost estimation for an instance type
func (s *AWSProviderServer) GetCostEstimate(ctx context.Context, req *GetCostEstimateRequest) (*GetCostEstimateResponse, error) {
	// This is a simplified cost estimation
	// In production, you would use AWS Pricing API or Cost Explorer API

	baseCost := getBaseCost(req.InstanceType)
	duration := parseDuration(req.Duration)

	return &GetCostEstimateResponse{
		InstanceType: req.InstanceType,
		Region:       req.Region,
		Duration:     req.Duration,
		Cost:         baseCost * duration,
		Currency:     "USD",
	}, nil
}

// HealthCheck performs a health check
func (s *AWSProviderServer) HealthCheck(ctx context.Context, req *emptypb.Empty) (*HealthCheckResponse, error) {
	// Test EC2 connection
	_, err := s.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})

	if err != nil {
		return &HealthCheckResponse{
			Healthy:   false,
			Message:   fmt.Sprintf("AWS connection failed: %v", err),
			Timestamp: timestamppb.Now(),
		}, nil
	}

	return &HealthCheckResponse{
		Healthy:   true,
		Message:   "AWS Provider is healthy",
		Timestamp: timestamppb.Now(),
	}, nil
}

// Helper functions

func getInstanceName(tags []types.Tag) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
			return *tag.Value
		}
	}
	return ""
}

func getBaseCost(instanceType string) float64 {
	// Simplified pricing - in production use AWS Pricing API
	costs := map[string]float64{
		"t2.micro":  0.0116,
		"t2.small":  0.023,
		"t2.medium": 0.0464,
		"t3.micro":  0.0104,
		"t3.small":  0.0208,
		"t3.medium": 0.0416,
		"m5.large":  0.096,
		"m5.xlarge": 0.192,
	}

	if cost, ok := costs[instanceType]; ok {
		return cost
	}
	return 0.10 // Default cost per hour
}

func parseDuration(duration string) float64 {
	// Parse duration string like "1h", "1d", "1m"
	// Simplified implementation
	switch {
	case len(duration) > 0 && duration[len(duration)-1] == 'h':
		return 1.0
	case len(duration) > 0 && duration[len(duration)-1] == 'd':
		return 24.0
	case len(duration) > 0 && duration[len(duration)-1] == 'm':
		return 720.0 // 30 days
	default:
		return 1.0
	}
}

// Placeholder types - will be replaced by generated proto
type (
	ProviderInfoResponse struct {
		Name             string
		Version          string
		SupportedRegions []string
		Capabilities     map[string]string
	}
	ListInstancesRequest struct {
		Region    string
		Filters   map[string]string
		PageSize  int32
		PageToken string
	}
	ListInstancesResponse struct {
		Instances     []*Instance
		NextPageToken string
		TotalCount    int32
	}
	CreateInstanceRequest struct {
		Name           string
		Type           string
		Region         string
		ImageId        string
		Tags           map[string]string
		UserData       string
		VpcId          string
		SubnetId       string
		SecurityGroups []string
		PublicIp       bool
		KeyPairId      string
		KeyPairName    string
		RootVolumeSize int32
		RootVolumeType string
	}
	CreateInstanceResponse struct {
		Instance *Instance
	}
	DeleteInstanceRequest struct {
		InstanceId string
		Region     string
	}
	DeleteInstanceResponse struct {
		Success bool
		Message string
	}
	GetInstanceStatusRequest struct {
		InstanceId string
		Region     string
	}
	GetInstanceStatusResponse struct {
		InstanceId  string
		Status      string
		LastUpdated *timestamppb.Timestamp
	}
	StartInstanceRequest struct {
		InstanceId string
		Region     string
	}
	StartInstanceResponse struct {
		Success bool
		Message string
	}
	StopInstanceRequest struct {
		InstanceId string
		Region     string
		Force      bool
	}
	StopInstanceResponse struct {
		Success bool
		Message string
	}
	RestartInstanceRequest struct {
		InstanceId string
		Region     string
	}
	RestartInstanceResponse struct {
		Success bool
		Message string
	}
	ListRegionsRequest  struct{}
	ListRegionsResponse struct {
		Regions []*Region
	}
	GetCostEstimateRequest struct {
		InstanceType string
		Region       string
		Duration     string
	}
	GetCostEstimateResponse struct {
		InstanceType string
		Region       string
		Duration     string
		Cost         float64
		Currency     string
	}
	HealthCheckResponse struct {
		Healthy   bool
		Message   string
		Timestamp *timestamppb.Timestamp
	}
	Instance struct {
		Id        string
		Name      string
		Status    string
		Type      string
		Region    string
		CreatedAt *timestamppb.Timestamp
		Tags      map[string]string
		PublicIp  string
		PrivateIp string
		Provider  string
	}
	Region struct {
		Id          string
		Name        string
		DisplayName string
		Status      string
	}
)
