package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cmp/pkg/plugin/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// AWSProvider implements the CloudProvider interface for AWS
type AWSProvider struct {
	config    map[string]interface{}
	ec2Client *ec2.Client
	iamClient *iam.Client
	region    string
}

// New creates a new AWS provider instance
// This function is required for plugin loading
func New() interfaces.CloudProvider {
	return &AWSProvider{}
}

// GetName returns the provider name
func (p *AWSProvider) GetName() string {
	return "AWS"
}

// GetVersion returns the provider version
func (p *AWSProvider) GetVersion() string {
	return "1.0.0"
}

// Initialize initializes the AWS provider with configuration
func (p *AWSProvider) Initialize(config map[string]interface{}) error {
	p.config = config

	// Validate required configuration
	if _, ok := config["region"]; !ok {
		config["region"] = "us-east-1" // Default region
	}

	p.region = config["region"].(string)

	// Check for explicit credentials
	var awsCfg aws.Config
	var err error

	if accessKey, hasAccessKey := config["access_key"]; hasAccessKey && accessKey != "" {
		if secretKey, hasSecretKey := config["secret_key"]; hasSecretKey && secretKey != "" {
			// Use explicit credentials
			awsCfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
				awsconfig.WithRegion(p.region),
				awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					accessKey.(string),
					secretKey.(string),
					"", // session token
				)),
			)
			if err != nil {
				return fmt.Errorf("failed to load AWS config with explicit credentials: %w", err)
			}
		} else {
			return fmt.Errorf("AWS access key provided but secret key is missing")
		}
	} else {
		// Try to load from default sources (environment, profile, etc.)
		awsCfg, err = awsconfig.LoadDefaultConfig(context.TODO(),
			awsconfig.WithRegion(p.region),
		)
		if err != nil {
			return fmt.Errorf("failed to load AWS config from default sources: %w", err)
		}
	}

	// Test the configuration by creating clients and making a test call
	p.ec2Client = ec2.NewFromConfig(awsCfg)
	p.iamClient = iam.NewFromConfig(awsCfg)

	// Test AWS connectivity with a simple call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to describe regions to test connectivity
	_, err = p.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		// Provide detailed error information
		if strings.Contains(err.Error(), "InvalidUserID.NotFound") {
			return fmt.Errorf("AWS credentials are invalid or user not found: %w", err)
		} else if strings.Contains(err.Error(), "UnauthorizedOperation") {
			return fmt.Errorf("AWS credentials lack required permissions: %w", err)
		} else if strings.Contains(err.Error(), "NoCredentialProviders") {
			return fmt.Errorf("no AWS credentials found in environment or config: %w", err)
		} else if strings.Contains(err.Error(), "InvalidAccessKeyId") {
			return fmt.Errorf("AWS access key ID is invalid: %w", err)
		} else if strings.Contains(err.Error(), "SignatureDoesNotMatch") {
			return fmt.Errorf("AWS secret key is invalid: %w", err)
		} else if strings.Contains(err.Error(), "RequestLimitExceeded") {
			return fmt.Errorf("AWS API rate limit exceeded: %w", err)
		} else if strings.Contains(err.Error(), "timeout") {
			return fmt.Errorf("AWS connection timeout - check network connectivity: %w", err)
		} else {
			return fmt.Errorf("AWS connectivity test failed: %w", err)
		}
	}

	fmt.Printf("AWS provider initialized for region: %s\n", p.region)
	return nil
}

// ListInstances returns a list of AWS EC2 instances
func (p *AWSProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
	// Check if client is initialized
	if p.ec2Client == nil {
		return nil, fmt.Errorf("AWS provider not initialized. Please configure AWS credentials")
	}

	// Use AWS SDK to list instances
	result, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe instances: %w", err)
	}

	var instances []interfaces.Instance
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			// Get instance name from tags
			name := ""
			for _, tag := range instance.Tags {
				if aws.ToString(tag.Key) == "Name" {
					name = aws.ToString(tag.Value)
					break
				}
			}

			// Get public IP
			publicIP := ""
			if instance.PublicIpAddress != nil {
				publicIP = aws.ToString(instance.PublicIpAddress)
			}

			// Get private IP
			privateIP := ""
			if instance.PrivateIpAddress != nil {
				privateIP = aws.ToString(instance.PrivateIpAddress)
			}

			// Convert tags
			tags := make(map[string]string)
			for _, tag := range instance.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}

			instances = append(instances, interfaces.Instance{
				ID:        aws.ToString(instance.InstanceId),
				Name:      name,
				Status:    string(instance.State.Name),
				Type:      string(instance.InstanceType),
				Region:    p.region,
				CreatedAt: instance.LaunchTime.Format(time.RFC3339),
				Tags:      tags,
				PublicIP:  publicIP,
				PrivateIP: privateIP,
			})
		}
	}

	return instances, nil
}

// CreateInstance creates a new AWS EC2 instance
func (p *AWSProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error) {
	// Prepare tags
	var tags []types.Tag
	for key, value := range req.Tags {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}
	// Add Name tag
	tags = append(tags, types.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(req.Name),
	})

	// Prepare security groups
	var securityGroupIds []string
	if len(req.SecurityGroups) > 0 {
		securityGroupIds = req.SecurityGroups
	}

	// Prepare user data
	var userData *string
	if req.UserData != "" {
		userData = aws.String(req.UserData)
	}

	// Create instance
	runResult, err := p.ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:          aws.String(req.ImageID),
		InstanceType:     types.InstanceType(req.Type),
		MinCount:         aws.Int32(1),
		MaxCount:         aws.Int32(1),
		SecurityGroupIds: securityGroupIds,
		UserData:         userData,
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags:         tags,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to run instance: %w", err)
	}

	if len(runResult.Instances) == 0 {
		return nil, fmt.Errorf("no instances created")
	}

	instance := runResult.Instances[0]

	// Get public IP if available
	publicIP := ""
	if instance.PublicIpAddress != nil {
		publicIP = aws.ToString(instance.PublicIpAddress)
	}

	// Get private IP if available
	privateIP := ""
	if instance.PrivateIpAddress != nil {
		privateIP = aws.ToString(instance.PrivateIpAddress)
	}

	fmt.Printf("Creating AWS instance: %s (%s) in %s\n", req.Name, req.Type, req.Region)

	return &interfaces.Instance{
		ID:        aws.ToString(instance.InstanceId),
		Name:      req.Name,
		Status:    string(instance.State.Name),
		Type:      string(instance.InstanceType),
		Region:    req.Region,
		CreatedAt: instance.LaunchTime.Format(time.RFC3339),
		Tags:      req.Tags,
		PublicIP:  publicIP,
		PrivateIP: privateIP,
	}, nil
}

// DeleteInstance deletes an AWS EC2 instance
func (p *AWSProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	_, err := p.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	fmt.Printf("Deleting AWS instance: %s\n", instanceID)
	return nil
}

// GetInstanceStatus returns the status of an AWS instance
func (p *AWSProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
	result, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", fmt.Errorf("instance not found")
	}

	return string(result.Reservations[0].Instances[0].State.Name), nil
}

// ListRegions returns available AWS regions
func (p *AWSProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error) {
	result, err := p.ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	var regions []interfaces.Region
	for _, region := range result.Regions {
		regions = append(regions, interfaces.Region{
			ID:          aws.ToString(region.RegionName),
			Name:        aws.ToString(region.RegionName),
			DisplayName: aws.ToString(region.RegionName),
			Status:      "available",
		})
	}

	return regions, nil
}

// GetCostEstimate returns cost estimate for AWS resources
func (p *AWSProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
	// Mock cost calculation
	var costPerHour float64
	switch req.InstanceType {
	case "t3.micro":
		costPerHour = 0.0104
	case "t3.small":
		costPerHour = 0.0208
	case "t3.medium":
		costPerHour = 0.0416
	default:
		costPerHour = 0.05
	}

	// Simple duration calculation (in real implementation, parse duration properly)
	var multiplier float64
	switch req.Duration {
	case "1h":
		multiplier = 1
	case "1d":
		multiplier = 24
	case "1m":
		multiplier = 24 * 30
	default:
		multiplier = 1
	}

	return &interfaces.CostEstimate{
		InstanceType: req.InstanceType,
		Region:       req.Region,
		Duration:     req.Duration,
		Cost:         costPerHour * multiplier,
		Currency:     "USD",
	}, nil
}

// GetNetworkProvider returns the network provider (not implemented for AWS in this example)
func (p *AWSProvider) GetNetworkProvider() interfaces.NetworkProvider {
	return nil
}

// GetIAMProvider returns the IAM provider (not implemented for AWS in this example)
func (p *AWSProvider) GetIAMProvider() interfaces.IAMProvider {
	return nil
}
