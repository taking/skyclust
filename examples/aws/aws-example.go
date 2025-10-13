package main

import (
	"context"
	"fmt"
	"log"

	"skyclust/internal/plugin/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// AWSExampleProvider implements the CloudProvider interface for AWS
// This is a simplified example showing how to create a basic AWS provider
type AWSExampleProvider struct {
	config    map[string]interface{}
	ec2Client *ec2.Client
	iamClient *iam.Client
	region    string
}

// New creates a new AWS provider instance
// This function is required for plugin loading
func New() interfaces.CloudProvider {
	return &AWSExampleProvider{}
}

// GetName returns the provider name
func (p *AWSExampleProvider) GetName() string {
	return "AWS Example"
}

// GetVersion returns the provider version
func (p *AWSExampleProvider) GetVersion() string {
	return "1.0.0"
}

// Initialize initializes the AWS provider with configuration
func (p *AWSExampleProvider) Initialize(config map[string]interface{}) error {
	p.config = config

	// Validate required configuration
	if _, ok := config["region"]; !ok {
		config["region"] = "us-east-1" // Default region
	}

	p.region = config["region"].(string)

	// Create AWS config
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(p.region),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create EC2 client
	p.ec2Client = ec2.NewFromConfig(awsCfg)

	// Create IAM client
	p.iamClient = iam.NewFromConfig(awsCfg)

	fmt.Printf("AWS Example provider initialized for region: %s\n", p.region)
	return nil
}

// ListInstances returns a list of AWS EC2 instances
func (p *AWSExampleProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
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
				CreatedAt: instance.LaunchTime.Format("2006-01-02T15:04:05Z"),
				Tags:      tags,
				PublicIP:  publicIP,
				PrivateIP: privateIP,
			})
		}
	}

	return instances, nil
}

// CreateInstance creates a new AWS EC2 instance
func (p *AWSExampleProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error) {
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
		CreatedAt: instance.LaunchTime.Format("2006-01-02T15:04:05Z"),
		Tags:      req.Tags,
		PublicIP:  publicIP,
		PrivateIP: privateIP,
	}, nil
}

// DeleteInstance deletes an AWS EC2 instance
func (p *AWSExampleProvider) DeleteInstance(ctx context.Context, instanceID string) error {
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
func (p *AWSExampleProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
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
func (p *AWSExampleProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error) {
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
func (p *AWSExampleProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
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

	// Simple duration calculation
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
func (p *AWSExampleProvider) GetNetworkProvider() interfaces.NetworkProvider {
	return nil
}

// GetIAMProvider returns the IAM provider (not implemented for AWS in this example)
func (p *AWSExampleProvider) GetIAMProvider() interfaces.IAMProvider {
	return nil
}

// Example usage function
func main() {
	// This is just for demonstration - in a real plugin, this wouldn't be needed
	log.Println("AWS Example Provider - This is a template for creating AWS plugins")
}
