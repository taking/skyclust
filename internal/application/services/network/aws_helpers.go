package network

import (
	"context"
	"fmt"

	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AWS VPC Functions

// listAWSVPCs: AWS VPC 목록을 조회합니다
func (s *Service) listAWSVPCs(ctx context.Context, credential *domain.Credential, req ListVPCsRequest) (*ListVPCsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
	}

	// Describe VPCs
	input := &ec2.DescribeVpcsInput{}
	if req.VPCID != "" {
		input.VpcIds = []string{req.VPCID}
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "list VPCs")
	}

	// Convert to DTOs
	vpcs := make([]VPCInfo, 0, len(result.Vpcs))
	for _, vpc := range result.Vpcs {
		vpcInfo := VPCInfo{
			ID:        aws.ToString(vpc.VpcId),
			Name:      s.getTagValue(vpc.Tags, "Name"),
			State:     string(vpc.State),
			IsDefault: aws.ToBool(vpc.IsDefault),
			Region:    req.Region,
		}
		vpcs = append(vpcs, vpcInfo)
	}

	return &ListVPCsResponse{VPCs: vpcs}, nil
}

// getAWSVPC: 특정 AWS VPC를 조회합니다
func (s *Service) getAWSVPC(ctx context.Context, credential *domain.Credential, req GetVPCRequest) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
	}

	// Describe specific VPC
	input := &ec2.DescribeVpcsInput{
		VpcIds: []string{req.VPCID},
	}

	result, err := ec2Client.DescribeVpcs(ctx, input)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "get VPC")
	}

	if len(result.Vpcs) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf(ErrMsgVPCNotFound, req.VPCID), 404)
	}

	vpc := result.Vpcs[0]
	vpcInfo := &VPCInfo{
		ID:        aws.ToString(vpc.VpcId),
		Name:      s.getTagValue(vpc.Tags, "Name"),
		State:     string(vpc.State),
		IsDefault: aws.ToBool(vpc.IsDefault),
		Region:    req.Region,
		Tags:      s.convertTags(vpc.Tags),
	}

	return vpcInfo, nil
}

// createAWSVPC: AWS VPC를 생성합니다
func (s *Service) createAWSVPC(ctx context.Context, credential *domain.Credential, req CreateVPCRequest) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "create VPC")
	}

	vpcInfo := &VPCInfo{
		ID:        aws.ToString(result.Vpc.VpcId),
		Name:      req.Name,
		State:     string(result.Vpc.State),
		IsDefault: aws.ToBool(result.Vpc.IsDefault),
		Region:    req.Region,
		Tags:      req.Tags,
	}

	return vpcInfo, nil
}

// updateAWSVPC: AWS VPC를 업데이트합니다
func (s *Service) updateAWSVPC(ctx context.Context, credential *domain.Credential, req UpdateVPCRequest, vpcID, region string) (*VPCInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
			return nil, s.providerErrorConverter.ConvertAWSError(err, "update VPC tags")
		}
	}

	// Get updated VPC info
	getReq := GetVPCRequest{
		CredentialID: credential.ID.String(),
		VPCID:        vpcID,
		Region:       region,
	}
	vpc, err := s.GetVPC(ctx, credential, getReq)
	if err != nil {
		return nil, err
	}

	// 캐시 무효화: VPC 목록 및 개별 VPC 캐시 삭제
	credentialID := credential.ID.String()
	if s.cacheService != nil {
		listKey := buildNetworkVPCListKey(credential.Provider, credentialID, region)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC list cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("region", region),
				domain.NewLogField("error", err))
		}
		itemKey := buildNetworkVPCItemKey(credential.Provider, credentialID, vpcID)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VPC item cache",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", vpcID),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행: VPC 업데이트 이벤트
	vpcData := map[string]interface{}{
		"vpc_id": vpc.ID,
		"name":   vpc.Name,
		"state":  vpc.State,
		"region": vpc.Region,
	}
	if s.eventService != nil {
		vpcData["provider"] = credential.Provider
		vpcData["credential_id"] = credentialID
		eventType := fmt.Sprintf("network.vpc.%s.updated", credential.Provider)
		if err := s.eventService.Publish(ctx, eventType, vpcData); err != nil {
			s.logger.Warn(ctx, "Failed to publish VPC updated event",
				domain.NewLogField("provider", credential.Provider),
				domain.NewLogField("credential_id", credentialID),
				domain.NewLogField("vpc_id", vpcID),
				domain.NewLogField("error", err))
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVPCUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/vpcs/%s", credential.Provider, vpcID),
		map[string]interface{}{
			"vpc_id":        vpc.ID,
			"name":          vpc.Name,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	return vpc, nil
}

// deleteAWSVPC: AWS VPC를 삭제합니다
func (s *Service) deleteAWSVPC(ctx context.Context, credential *domain.Credential, req DeleteVPCRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "create EC2 client")
	}

	// Delete VPC
	_, err = ec2Client.DeleteVpc(ctx, &ec2.DeleteVpcInput{
		VpcId: aws.String(req.VPCID),
	})
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, fmt.Sprintf("delete VPC %s", req.VPCID))
	}

	return nil
}

// AWS Security Group Functions

// listAWSSecurityGroups: AWS 보안 그룹 목록을 조회합니다
func (s *Service) listAWSSecurityGroups(ctx context.Context, credential *domain.Credential, req ListSecurityGroupsRequest) (*ListSecurityGroupsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "describe security groups")
	}

	// Convert to DTOs
	securityGroups := make([]SecurityGroupInfo, 0, len(result.SecurityGroups))
	for _, sg := range result.SecurityGroups {
		sgInfo := SecurityGroupInfo{
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

	return &ListSecurityGroupsResponse{SecurityGroups: securityGroups}, nil
}

// getAWSSecurityGroup: 특정 AWS 보안 그룹을 조회합니다
func (s *Service) getAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req GetSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
	}

	// Describe specific security group
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{req.SecurityGroupID},
	}

	result, err := ec2Client.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "describe security group")
	}

	if len(result.SecurityGroups) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf(ErrMsgSecurityGroupNotFound, req.SecurityGroupID), 404)
	}

	sg := result.SecurityGroups[0]
	sgInfo := &SecurityGroupInfo{
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

// createAWSSecurityGroup: AWS 보안 그룹을 생성합니다
func (s *Service) createAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req CreateSecurityGroupRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "create security group")
	}

	sgInfo := &SecurityGroupInfo{
		ID:          aws.ToString(result.GroupId),
		Name:        req.Name,
		Description: req.Description,
		VPCID:       req.VPCID,
		Region:      req.Region,
		Rules:       []SecurityGroupRuleInfo{}, // Empty initially
		Tags:        req.Tags,
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs/%s/security-groups", credential.Provider, req.VPCID),
		map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": sgInfo.ID,
			"name":              req.Name,
			"vpc_id":            req.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.created", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return sgInfo, nil
}

// updateAWSSecurityGroup: AWS 보안 그룹을 업데이트합니다
func (s *Service) updateAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req UpdateSecurityGroupRequest, securityGroupID, region string) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
			return nil, s.providerErrorConverter.ConvertAWSError(err, "update security group tags")
		}
	}

	// Get updated security group info
	sgInfo, err := s.GetSecurityGroup(ctx, credential, GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: securityGroupID,
		Region:          region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/security-groups/%s", credential.Provider, securityGroupID),
		map[string]interface{}{
			"security_group_id": securityGroupID,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": securityGroupID,
			"name":              sgInfo.Name,
			"vpc_id":            sgInfo.VPCID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.updated", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return sgInfo, nil
}

// deleteAWSSecurityGroup: AWS 보안 그룹을 삭제합니다
func (s *Service) deleteAWSSecurityGroup(ctx context.Context, credential *domain.Credential, req DeleteSecurityGroupRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "create EC2 client")
	}

	// Delete security group
	_, err = ec2Client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(req.SecurityGroupID),
	})
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "delete security group")
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSecurityGroupDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/security-groups/%s", credential.Provider, req.SecurityGroupID),
		map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		sgData := map[string]interface{}{
			"security_group_id": req.SecurityGroupID,
			"provider":          credential.Provider,
			"credential_id":     credentialID,
			"region":            req.Region,
		}
		if s.eventService != nil {
			sgData := sgData
			sgData["provider"] = credential.Provider
			sgData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.security-group.%s.deleted", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, sgData)
		}
	}

	return nil
}

// AWS Subnet Functions

// listAWSSubnets: AWS 서브넷 목록을 조회합니다
func (s *Service) listAWSSubnets(ctx context.Context, credential *domain.Credential, req ListSubnetsRequest) (*ListSubnetsResponse, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "describe subnets")
	}

	// Convert to DTOs
	subnets := make([]SubnetInfo, 0, len(result.Subnets))
	for _, subnet := range result.Subnets {
		subnetInfo := SubnetInfo{
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

	return &ListSubnetsResponse{Subnets: subnets}, nil
}

// getAWSSubnet gets a specific AWS subnet
func (s *Service) getAWSSubnet(ctx context.Context, credential *domain.Credential, req GetSubnetRequest) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
	}

	// Describe specific subnet
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: []string{req.SubnetID},
	}

	result, err := ec2Client.DescribeSubnets(ctx, input)
	if err != nil {
		return nil, s.providerErrorConverter.ConvertAWSError(err, "describe subnet")
	}

	if len(result.Subnets) == 0 {
		return nil, domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf(ErrMsgSubnetNotFound, req.SubnetID), 404)
	}

	subnet := result.Subnets[0]
	subnetInfo := &SubnetInfo{
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

// createAWSSubnet creates a new AWS subnet
func (s *Service) createAWSSubnet(ctx context.Context, credential *domain.Credential, req CreateSubnetRequest) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "create subnet")
	}

	subnetInfo := &SubnetInfo{
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

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetCreate,
		fmt.Sprintf("POST /api/v1/%s/networks/vpcs/%s/subnets", credential.Provider, req.VPCID),
		map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetInfo.ID,
			"name":          req.Name,
			"vpc_id":        req.VPCID,
			"cidr_block":    subnetInfo.CIDRBlock,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.created", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return subnetInfo, nil
}

// updateAWSSubnet updates an AWS subnet
func (s *Service) updateAWSSubnet(ctx context.Context, credential *domain.Credential, req UpdateSubnetRequest, subnetID, region string) (*SubnetInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
			return nil, s.providerErrorConverter.ConvertAWSError(err, "update subnet tags")
		}
	}

	// Get updated subnet info
	subnetInfo, err := s.getAWSSubnet(ctx, credential, GetSubnetRequest{
		CredentialID: credential.ID.String(),
		SubnetID:     subnetID,
		Region:       region,
	})
	if err != nil {
		return nil, err
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetUpdate,
		fmt.Sprintf("PUT /api/v1/%s/networks/vpcs/%s/subnets/%s", credential.Provider, subnetInfo.VPCID, subnetID),
		map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		},
	)

	// NATS 이벤트 발행
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     subnetID,
			"name":          subnetInfo.Name,
			"vpc_id":        subnetInfo.VPCID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.updated", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return subnetInfo, nil
}

// deleteAWSSubnet deletes an AWS subnet
func (s *Service) deleteAWSSubnet(ctx context.Context, credential *domain.Credential, req DeleteSubnetRequest) error {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "create EC2 client")
	}

	// Delete subnet
	_, err = ec2Client.DeleteSubnet(ctx, &ec2.DeleteSubnetInput{
		SubnetId: aws.String(req.SubnetID),
	})
	if err != nil {
		return s.providerErrorConverter.ConvertAWSError(err, "delete subnet")
	}

	// 감사로그 기록
	credentialID := credential.ID.String()
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionSubnetDelete,
		fmt.Sprintf("DELETE /api/v1/%s/networks/subnets/%s", credential.Provider, req.SubnetID),
		map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		},
	)

	// NATS 이벤트 발행 (VPCID는 req에 없으므로 빈 문자열로 처리)
	if s.eventService != nil {
		subnetData := map[string]interface{}{
			"subnet_id":     req.SubnetID,
			"provider":      credential.Provider,
			"credential_id": credentialID,
			"region":        req.Region,
		}
		if s.eventService != nil {
			subnetData := subnetData
			subnetData["provider"] = credential.Provider
			subnetData["credential_id"] = credentialID
			eventType := fmt.Sprintf("network.subnet.%s.deleted", credential.Provider)
			_ = s.eventService.Publish(ctx, eventType, subnetData)
		}
	}

	return nil
}

// isSubnetPublic: 서브넷이 공개 서브넷인지 확인합니다
func (s *Service) isSubnetPublic(ctx context.Context, ec2Client *ec2.Client, subnetID string) bool {
	// Get route tables for the subnet
	input := &ec2.DescribeRouteTablesInput{
		Filters: []ec2Types.Filter{
			{Name: aws.String("association.subnet-id"), Values: []string{subnetID}},
		},
	}

	result, err := ec2Client.DescribeRouteTables(ctx, input)
	if err != nil {
		s.logger.Warn(ctx, "Failed to describe route tables", domain.NewLogField("subnet_id", subnetID), domain.NewLogField("error", err))
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

// addAWSSecurityGroupRule: AWS 보안 그룹 규칙을 추가합니다
func (s *Service) addAWSSecurityGroupRule(ctx context.Context, credential *domain.Credential, req AddSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "add security group rule")
	}

	// Get updated security group info
	getReq := GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

// removeAWSSecurityGroupRule: AWS 보안 그룹 규칙을 제거합니다
func (s *Service) removeAWSSecurityGroupRule(ctx context.Context, credential *domain.Credential, req RemoveSecurityGroupRuleRequest) (*SecurityGroupInfo, error) {
	// Create AWS EC2 client
	ec2Client, err := s.createEC2Client(ctx, credential, req.Region)
	if err != nil {
		return nil, err // createEC2Client already returns domain error
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
		return nil, s.providerErrorConverter.ConvertAWSError(err, "remove security group rule")
	}

	// Get updated security group info
	getReq := GetSecurityGroupRequest{
		CredentialID:    credential.ID.String(),
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
	return s.GetSecurityGroup(ctx, credential, getReq)
}

