package service

import (
	"context"
	"fmt"

	"skyclust/internal/application/dto"
	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/zap"
)

// NetworkService handles network resource operations
type NetworkService struct {
	credentialService domain.CredentialService
	logger            *zap.Logger
}

// NewNetworkService creates a new network service
func NewNetworkService(credentialService domain.CredentialService, logger *zap.Logger) *NetworkService {
	return &NetworkService{
		credentialService: credentialService,
		logger:            logger,
	}
}

// ListVPCs lists VPCs for a given credential and region
func (s *NetworkService) ListVPCs(ctx context.Context, credential *domain.Credential, req dto.ListVPCsRequest) (*dto.ListVPCsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe VPCs
	input := &ec2.DescribeVpcsInput{}
	if req.VPCID != "" {
		input.VpcIds = []string{req.VPCID}
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPCs: %w", err)
	}

	// Convert to DTOs
	vpcs := make([]dto.VPCInfo, 0, len(result.Vpcs))
	for _, vpc := range result.Vpcs {
		vpcInfo := dto.VPCInfo{
			ID:        aws.ToString(vpc.VpcId),
			Name:      s.getTagValue(vpc.Tags, "Name"),
			CIDRBlock: aws.ToString(vpc.CidrBlock),
			State:     string(vpc.State),
			IsDefault: aws.ToBool(vpc.IsDefault),
			Region:    req.Region,
		}
		vpcs = append(vpcs, vpcInfo)
	}

	return &dto.ListVPCsResponse{VPCs: vpcs}, nil
}

// GetVPC gets a specific VPC
func (s *NetworkService) GetVPC(ctx context.Context, credential *domain.Credential, req dto.GetVPCRequest) (*dto.VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe specific VPC
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{req.VPCID},
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe VPC: %w", err)
	}

	if len(result.Vpcs) == 0 {
		return nil, fmt.Errorf("VPC not found: %s", req.VPCID)
	}

	vpc := result.Vpcs[0]
	vpcInfo := &dto.VPCInfo{
		ID:        aws.ToString(vpc.VpcId),
		Name:      s.getTagValue(vpc.Tags, "Name"),
		CIDRBlock: aws.ToString(vpc.CidrBlock),
		State:     string(vpc.State),
		IsDefault: aws.ToBool(vpc.IsDefault),
		Region:    req.Region,
		Tags:      s.convertTags(vpc.Tags),
	}

	return vpcInfo, nil
}

// CreateVPC creates a new VPC
func (s *NetworkService) CreateVPC(ctx context.Context, credential *domain.Credential, req dto.CreateVPCRequest) (*dto.VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Create VPC
	input := &ec2.CreateVpcInput{
		CidrBlock: aws.String(req.CIDRBlock),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeVpc,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateVpc(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create VPC: %w", err)
	}

	vpcInfo := &dto.VPCInfo{
		ID:        aws.ToString(result.Vpc.VpcId),
		Name:      req.Name,
		CIDRBlock: aws.ToString(result.Vpc.CidrBlock),
		State:     string(result.Vpc.State),
		IsDefault: aws.ToBool(result.Vpc.IsDefault),
		Region:    req.Region,
		Tags:      req.Tags,
	}

	return vpcInfo, nil
}

// UpdateVPC updates a VPC
func (s *NetworkService) UpdateVPC(ctx context.Context, credential *domain.Credential, req dto.UpdateVPCRequest, vpcID, region string) (*dto.VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Update VPC tags if provided
	if req.Name != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{vpcID},
			Tags:      tags,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update VPC tags: %w", err)
		}
	}

	// Get updated VPC info
	getReq := dto.GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}
	return s.GetVPC(ctx, credential, getReq)
}

// DeleteVPC deletes a VPC
func (s *NetworkService) DeleteVPC(ctx context.Context, credential *domain.Credential, req dto.DeleteVPCRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Delete VPC
	_, err = ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(req.VPCID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete VPC: %w", err)
	}

	return nil
}

// ListSubnets lists subnets for a given credential and region
func (s *NetworkService) ListSubnets(ctx context.Context, credential *domain.Credential, req dto.ListSubnetsRequest) (*dto.ListSubnetsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe subnets
	input := &ec2.DescribeSubnetsInput{}
	if req.SubnetID != "" {
		input.SubnetIds = []string{req.SubnetID}
	}
	if req.VPCID != "" {
		input.Filters = []ec2Types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{req.VPCID}},
		}
	}

	result, err := ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnets: %w", err)
	}

	// Convert to DTOs
	subnets := make([]dto.SubnetInfo, 0, len(result.Subnets))
	for _, subnet := range result.Subnets {
		subnetInfo := dto.SubnetInfo{
			ID:               aws.ToString(subnet.SubnetId),
			Name:             s.getTagValue(subnet.Tags, "Name"),
			VPCID:            aws.ToString(subnet.VpcId),
			CIDRBlock:        aws.ToString(subnet.CidrBlock),
			AvailabilityZone: aws.ToString(subnet.AvailabilityZone),
			State:            string(subnet.State),
			IsPublic:         s.isSubnetPublic(ctx, ec2Client, aws.ToString(subnet.SubnetId)),
			Region:           req.Region,
			Tags:             s.convertTags(subnet.Tags),
		}
		subnets = append(subnets, subnetInfo)
	}

	return &dto.ListSubnetsResponse{Subnets: subnets}, nil
}

// GetSubnet gets a specific subnet
func (s *NetworkService) GetSubnet(ctx context.Context, credential *domain.Credential, req dto.GetSubnetRequest) (*dto.SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe specific subnet
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{req.SubnetID},
	}

	result, err := ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe subnet: %w", err)
	}

	if len(result.Subnets) == 0 {
		return nil, fmt.Errorf("subnet not found: %s", req.SubnetID)
	}

	subnet := result.Subnets[0]
	subnetInfo := &dto.SubnetInfo{
		ID:               aws.ToString(subnet.SubnetId),
		Name:             s.getTagValue(subnet.Tags, "Name"),
		VPCID:            aws.ToString(subnet.VpcId),
		CIDRBlock:        aws.ToString(subnet.CidrBlock),
		AvailabilityZone: aws.ToString(subnet.AvailabilityZone),
		State:            string(subnet.State),
		IsPublic:         s.isSubnetPublic(ctx, ec2Client, aws.ToString(subnet.SubnetId)),
		Region:           req.Region,
		Tags:             s.convertTags(subnet.Tags),
	}

	return subnetInfo, nil
}

// CreateSubnet creates a new subnet
func (s *NetworkService) CreateSubnet(ctx context.Context, credential *domain.Credential, req dto.CreateSubnetRequest) (*dto.SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Create subnet
	input := &ec2.CreateSubnetInput{
		VpcId:            aws.String(req.VPCID),
		CidrBlock:        aws.String(req.CIDRBlock),
		AvailabilityZone: aws.String(req.AvailabilityZone),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeSubnet,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateSubnet(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create subnet: %w", err)
	}

	subnetInfo := &dto.SubnetInfo{
		ID:               aws.ToString(result.Subnet.SubnetId),
		Name:             req.Name,
		VPCID:            aws.ToString(result.Subnet.VpcId),
		CIDRBlock:        aws.ToString(result.Subnet.CidrBlock),
		AvailabilityZone: aws.ToString(result.Subnet.AvailabilityZone),
		State:            string(result.Subnet.State),
		IsPublic:         false, // Will be determined later
		Region:           req.Region,
		Tags:             req.Tags,
	}

	return subnetInfo, nil
}

// UpdateSubnet updates a subnet
func (s *NetworkService) UpdateSubnet(ctx context.Context, credential *domain.Credential, req dto.UpdateSubnetRequest, subnetID, region string) (*dto.SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Update subnet tags if provided
	if req.Name != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{subnetID},
			Tags:      tags,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update subnet tags: %w", err)
		}
	}

	// Get updated subnet info
	getReq := dto.GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	}
	return s.GetSubnet(ctx, credential, getReq)
}

// DeleteSubnet deletes a subnet
func (s *NetworkService) DeleteSubnet(ctx context.Context, credential *domain.Credential, req dto.DeleteSubnetRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Delete subnet
	_, err = ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(req.SubnetID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete subnet: %w", err)
	}

	return nil
}

// ListSecurityGroups lists security groups for a given credential and region
func (s *NetworkService) ListSecurityGroups(ctx context.Context, credential *domain.Credential, req dto.ListSecurityGroupsRequest) (*dto.ListSecurityGroupsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe security groups
	input := &ec2.DescribeSecurityGroupsInput{}
	if req.SecurityGroupID != "" {
		input.GroupIds = []string{req.SecurityGroupID}
	}
	if req.VPCID != "" {
		input.Filters = []ec2Types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{req.VPCID}},
		}
	}

	result, err := ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe security groups: %w", err)
	}

	// Convert to DTOs
	securityGroups := make([]dto.SecurityGroupInfo, 0, len(result.SecurityGroups))
	for _, sg := range result.SecurityGroups {
		sgInfo := dto.SecurityGroupInfo{
			ID:          aws.ToString(sg.GroupId),
			Name:        aws.ToString(sg.GroupName),
			Description: aws.ToString(sg.Description),
			VPCID:       aws.ToString(sg.VpcId),
			Region:      req.Region,
			Rules:       s.convertSecurityGroupRules(sg.IpPermissions, sg.IpPermissionsEgress),
			Tags:        s.convertTags(sg.Tags),
		}
		securityGroups = append(securityGroups, sgInfo)
	}

	return &dto.ListSecurityGroupsResponse{SecurityGroups: securityGroups}, nil
}

// GetSecurityGroup gets a specific security group
func (s *NetworkService) GetSecurityGroup(ctx context.Context, credential *domain.Credential, req dto.GetSecurityGroupRequest) (*dto.SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Describe specific security group
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{req.SecurityGroupID},
	}

	result, err := ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe security group: %w", err)
	}

	if len(result.SecurityGroups) == 0 {
		return nil, fmt.Errorf("security group not found: %s", req.SecurityGroupID)
	}

	sg := result.SecurityGroups[0]
	sgInfo := &dto.SecurityGroupInfo{
		ID:          aws.ToString(sg.GroupId),
		Name:        aws.ToString(sg.GroupName),
		Description: aws.ToString(sg.Description),
		VPCID:       aws.ToString(sg.VpcId),
		Region:      req.Region,
		Rules:       s.convertSecurityGroupRules(sg.IpPermissions, sg.IpPermissionsEgress),
		Tags:        s.convertTags(sg.Tags),
	}

	return sgInfo, nil
}

// CreateSecurityGroup creates a new security group
func (s *NetworkService) CreateSecurityGroup(ctx context.Context, credential *domain.Credential, req dto.CreateSecurityGroupRequest) (*dto.SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Create security group
	input := &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(req.Name),
		Description: aws.String(req.Description),
		VpcId:       aws.String(req.VPCID),
		TagSpecifications: []ec2Types.TagSpecification{
			{
				ResourceType: ec2Types.ResourceTypeSecurityGroup,
				Tags: []ec2Types.Tag{
					{Key: aws.String("Name"), Value: aws.String(req.Name)},
				},
			},
		},
	}

	// Add custom tags
	for key, value := range req.Tags {
		input.TagSpecifications[0].Tags = append(input.TagSpecifications[0].Tags, ec2Types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	result, err := ec2Client.CreateSecurityGroup(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create security group: %w", err)
	}

	sgInfo := &dto.SecurityGroupInfo{
		ID:          aws.ToString(result.GroupId),
		Name:        req.Name,
		Description: req.Description,
		VPCID:       req.VPCID,
		Region:      req.Region,
		Rules:       []dto.SecurityGroupRuleInfo{}, // Empty initially
		Tags:        req.Tags,
	}

	return sgInfo, nil
}

// UpdateSecurityGroup updates a security group
func (s *NetworkService) UpdateSecurityGroup(ctx context.Context, credential *domain.Credential, req dto.UpdateSecurityGroupRequest, securityGroupID, region string) (*dto.SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Update security group tags if provided
	if req.Name != "" || req.Description != "" || len(req.Tags) > 0 {
		tags := []ec2Types.Tag{}

		if req.Name != "" {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(req.Name),
			})
		}

		for key, value := range req.Tags {
			tags = append(tags, ec2Types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}

		_, err = ec2Client.CreateTags(ctx, &ec2.CreateTagsInput{
			Resources: []string{securityGroupID},
			Tags:      tags,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update security group tags: %w", err)
		}
	}

	// Get updated security group info
	getReq := dto.GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// DeleteSecurityGroup deletes a security group
func (s *NetworkService) DeleteSecurityGroup(ctx context.Context, credential *domain.Credential, req dto.DeleteSecurityGroupRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Delete security group
	_, err = ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(req.SecurityGroupID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete security group: %w", err)
	}

	return nil
}

// Helper methods

// createEC2Client creates an AWS EC2 client
func (s *NetworkService) createEC2Client(ctx context.Context, credential *domain.Credential, region string) (*ec2.Client, error) {
	// Decrypt credential
	decryptedData, err := s.credentialService.DecryptCredentialData(ctx, credential.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credential: %w", err)
	}

	// Extract AWS credentials (same as kubernetes service)
	accessKey, ok := decryptedData["access_key"].(string)
	if !ok {
		return nil, fmt.Errorf("access_key not found in credential")
	}

	secretKey, ok := decryptedData["secret_key"].(string)
	if !ok {
		return nil, fmt.Errorf("secret_key not found in credential")
	}

	// Debug: Log the extracted credentials
	s.logger.Info("Extracted AWS credentials",
		zap.String("access_key", accessKey),
		zap.String("secret_key", secretKey[:10]+"..."))

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

	return ec2.NewFromConfig(cfg), nil
}

// getTagValue extracts a tag value by key
func (s *NetworkService) getTagValue(tags []ec2Types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}
	return ""
}

// convertTags converts AWS tags to map
func (s *NetworkService) convertTags(tags []ec2Types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		result[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}
	return result
}

// isSubnetPublic determines if a subnet is public by checking its route table
func (s *NetworkService) isSubnetPublic(ctx context.Context, ec2Client *ec2.Client, subnetID string) bool {
	// Get route tables for the subnet
	input := &ec2.DescribeRouteTablesInput{
		Filters: []ec2Types.Filter{
			{Name: aws.String("association.subnet-id"), Values: []string{subnetID}},
		},
	}

	result, err := ec2Client.DescribeRouteTables(ctx, input)
	if err != nil {
		s.logger.Warn("Failed to describe route tables", zap.String("subnet_id", subnetID), zap.Error(err))
		return false
	}

	// Check if any route table has a route to internet gateway
	for _, rt := range result.RouteTables {
		for _, route := range rt.Routes {
			if route.GatewayId != nil && aws.ToString(route.GatewayId) != "local" {
				// Check if this is an internet gateway
				igwInput := &ec2.DescribeInternetGatewaysInput{
					InternetGatewayIds: []string{aws.ToString(route.GatewayId)},
				}
				igwResult, err := ec2Client.DescribeInternetGateways(ctx, igwInput)
				if err == nil && len(igwResult.InternetGateways) > 0 {
					return true
				}
			}
		}
	}

	return false
}

// convertSecurityGroupRules converts AWS security group rules to DTOs
func (s *NetworkService) convertSecurityGroupRules(ingress, egress []ec2Types.IpPermission) []dto.SecurityGroupRuleInfo {
	rules := make([]dto.SecurityGroupRuleInfo, 0)

	// Convert ingress rules
	for _, perm := range ingress {
		rule := dto.SecurityGroupRuleInfo{
			Type:         "ingress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	// Convert egress rules
	for _, perm := range egress {
		rule := dto.SecurityGroupRuleInfo{
			Type:         "egress",
			Protocol:     aws.ToString(perm.IpProtocol),
			FromPort:     aws.ToInt32(perm.FromPort),
			ToPort:       aws.ToInt32(perm.ToPort),
			CIDRBlocks:   make([]string, 0),
			SourceGroups: make([]string, 0),
		}

		// Add CIDR blocks
		for _, ipRange := range perm.IpRanges {
			if ipRange.CidrIp != nil {
				rule.CIDRBlocks = append(rule.CIDRBlocks, aws.ToString(ipRange.CidrIp))
			}
		}

		// Add source groups
		for _, userGroupPair := range perm.UserIdGroupPairs {
			if userGroupPair.GroupId != nil {
				rule.SourceGroups = append(rule.SourceGroups, aws.ToString(userGroupPair.GroupId))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

// AddSecurityGroupRule adds a rule to a security group
func (s *NetworkService) AddSecurityGroupRule(ctx context.Context, credential *domain.Credential, req dto.AddSecurityGroupRuleRequest) (*dto.SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Build IP permissions
	ipPermission := ec2Types.IpPermission{
		IpProtocol: aws.String(req.Protocol),
		FromPort:   aws.Int32(req.FromPort),
		ToPort:     aws.Int32(req.ToPort),
	}

	// Add CIDR blocks
	if len(req.CIDRBlocks) > 0 {
		for _, cidr := range req.CIDRBlocks {
			ipPermission.IpRanges = append(ipPermission.IpRanges, ec2Types.IpRange{
				CidrIp:      aws.String(cidr),
				Description: aws.String(req.Description),
			})
		}
	}

	// Add source groups
	if len(req.SourceGroups) > 0 {
		for _, groupID := range req.SourceGroups {
			ipPermission.UserIdGroupPairs = append(ipPermission.UserIdGroupPairs, ec2Types.UserIdGroupPair{
				GroupId:     aws.String(groupID),
				Description: aws.String(req.Description),
			})
		}
	}

	// Add rule based on type
	if req.Type == "ingress" {
		_, err = ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	} else {
		_, err = ec2Client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to add security group rule: %w", err)
	}

	// Get updated security group info
	getReq := dto.GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// RemoveSecurityGroupRule removes a rule from a security group
func (s *NetworkService) RemoveSecurityGroupRule(ctx context.Context, credential *domain.Credential, req dto.RemoveSecurityGroupRuleRequest) (*dto.SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}

	// Build IP permissions
	ipPermission := ec2Types.IpPermission{
		IpProtocol: aws.String(req.Protocol),
		FromPort:   aws.Int32(req.FromPort),
		ToPort:     aws.Int32(req.ToPort),
	}

	// Add CIDR blocks
	if len(req.CIDRBlocks) > 0 {
		for _, cidr := range req.CIDRBlocks {
			ipPermission.IpRanges = append(ipPermission.IpRanges, ec2Types.IpRange{
				CidrIp: aws.String(cidr),
			})
		}
	}

	// Add source groups
	if len(req.SourceGroups) > 0 {
		for _, groupID := range req.SourceGroups {
			ipPermission.UserIdGroupPairs = append(ipPermission.UserIdGroupPairs, ec2Types.UserIdGroupPair{
				GroupId: aws.String(groupID),
			})
		}
	}

	// Remove rule based on type
	if req.Type == "ingress" {
		_, err = ec2Client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	} else {
		_, err = ec2Client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
			GroupId:       aws.String(req.SecurityGroupID),
			IpPermissions: []ec2Types.IpPermission{ipPermission},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to remove security group rule: %w", err)
	}

	// Get updated security group info
	getReq := dto.GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// UpdateSecurityGroupRules updates all rules for a security group
func (s *NetworkService) UpdateSecurityGroupRules(ctx context.Context, credential *domain.Credential, req dto.UpdateSecurityGroupRulesRequest) (*dto.SecurityGroupInfo, error) {

	// Get current security group
	getReq := dto.GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	currentSG, err := s.GetSecurityGroup(ctx, credential, getReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get current security group: %w", err)
	}

	// Remove all existing rules
	for _, rule := range currentSG.Rules {
		removeReq := dto.RemoveSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            rule.Type,
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.RemoveSecurityGroupRule(ctx, credential, removeReq)
		if err != nil {
			s.logger.Warn("Failed to remove existing rule", zap.Error(err))
		}
	}

	// Add new ingress rules
	for _, rule := range req.IngressRules {
		addReq := dto.AddSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            "ingress",
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.AddSecurityGroupRule(ctx, credential, addReq)
		if err != nil {
			return nil, fmt.Errorf("failed to add ingress rule: %w", err)
		}
	}

	// Add new egress rules
	for _, rule := range req.EgressRules {
		addReq := dto.AddSecurityGroupRuleRequest{
			CredentialID:    credential.ID.String(),
			SecurityGroupID: req.SecurityGroupID,
			Region:          req.Region,
			Type:            "egress",
			Protocol:        rule.Protocol,
			FromPort:        rule.FromPort,
			ToPort:          rule.ToPort,
			CIDRBlocks:      rule.CIDRBlocks,
			SourceGroups:    rule.SourceGroups,
		}
		_, err = s.AddSecurityGroupRule(ctx, credential, addReq)
		if err != nil {
			return nil, fmt.Errorf("failed to add egress rule: %w", err)
		}
	}

	// Get updated security group info
	return s.GetSecurityGroup(ctx, credential, getReq)
}
