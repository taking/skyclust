package network

import (
	networkservice "skyclust/internal/application/services/network"
)

// HTTP Handler Layer Types
// These types are used for HTTP request/response transformation
// Separate from service layer DTOs to maintain clean architecture boundaries

// Network Resource Response DTOs (HTTP layer)

// VPCInfoResponse represents VPC information in HTTP responses
type VPCInfoResponse struct {
	ID                string                `json:"id"`
	Name              string                `json:"name"`
	State             string                `json:"state"`
	IsDefault         bool                  `json:"is_default"`
	NetworkMode       string                `json:"network_mode,omitempty"`
	RoutingMode       string                `json:"routing_mode,omitempty"`
	MTU               int64                 `json:"mtu,omitempty"`
	AutoSubnets       bool                  `json:"auto_subnets,omitempty"`
	Description       string                `json:"description,omitempty"`
	FirewallRuleCount int                   `json:"firewall_rule_count,omitempty"`
	Gateway           *GatewayInfoResponse  `json:"gateway,omitempty"`
	CreationTimestamp string                `json:"creation_timestamp,omitempty"`
	Tags              map[string]string     `json:"tags,omitempty"`
	Subnets           []SubnetInfoResponse  `json:"subnets,omitempty"` // Direct array: subnets[]
}

// GatewayInfoResponse represents gateway information in HTTP responses
type GatewayInfoResponse struct {
	Type      string `json:"type,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Name      string `json:"name,omitempty"`
}

// SubnetInfoResponse represents subnet information in HTTP responses
type SubnetInfoResponse struct {
	ID                    string            `json:"id"`
	Name                  string            `json:"name"`
	VPCID                 string            `json:"vpc_id"`
	CIDRBlock             string            `json:"cidr_block"`
	AvailabilityZone      string            `json:"availability_zone"`
	State                 string            `json:"state"`
	IsPublic              bool              `json:"is_public"`
	Region                string            `json:"region"`
	Description           string            `json:"description,omitempty"`
	GatewayAddress        string            `json:"gateway_address,omitempty"`
	PrivateIPGoogleAccess bool              `json:"private_ip_google_access,omitempty"`
	FlowLogs              bool              `json:"flow_logs,omitempty"`
	CreationTimestamp     string            `json:"creation_timestamp,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// SecurityGroupInfoResponse represents security group information in HTTP responses
type SecurityGroupInfoResponse struct {
	ID          string                          `json:"id"`
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	VPCID       string                          `json:"vpc_id"`
	Region      string                          `json:"region"`
	Rules       []SecurityGroupRuleInfoResponse `json:"rules,omitempty"`
	Tags        map[string]string               `json:"tags,omitempty"`
}

// SecurityGroupRuleInfoResponse represents security group rule information in HTTP responses
type SecurityGroupRuleInfoResponse struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Protocol     string   `json:"protocol"`
	FromPort     int32    `json:"from_port,omitempty"`
	ToPort       int32    `json:"to_port,omitempty"`
	CIDRBlocks   []string `json:"cidr_blocks,omitempty"`
	SourceGroups []string `json:"source_groups,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// HTTP Request DTOs

// ListVPCsRequest represents a request to list VPCs (HTTP layer)
type ListVPCsRequest struct {
	Region        string `form:"region" json:"region,omitempty"`
	VPCID         string `form:"vpc_id" json:"vpc_id,omitempty"`
	ResourceGroup string `form:"resource_group" json:"resource_group,omitempty"` // Azure-specific: Resource Group filter
	Page          int    `form:"page" json:"page,omitempty"`
	Limit         int    `form:"limit" json:"limit,omitempty"`
	SortBy        string `form:"sort_by" json:"sort_by,omitempty"`
	SortOrder     string `form:"sort_order" json:"sort_order,omitempty"`
	Search        string `form:"search" json:"search,omitempty"`
}

// ListVPCsResponse represents the response after listing VPCs (HTTP layer)
type ListVPCsResponse struct {
	VPCs []VPCInfoResponse `json:"vpcs"`
}

// ListSubnetsRequest represents a request to list subnets (HTTP layer)
type ListSubnetsRequest struct {
	VPCID         string `form:"vpc_id" json:"vpc_id,omitempty"`
	Region        string `form:"region" json:"region,omitempty"`
	SubnetID      string `form:"subnet_id" json:"subnet_id,omitempty"`
	ResourceGroup string `form:"resource_group" json:"resource_group,omitempty"` // Azure-specific: Resource Group filter
	Page          int    `form:"page" json:"page,omitempty"`
	Limit         int    `form:"limit" json:"limit,omitempty"`
	SortBy        string `form:"sort_by" json:"sort_by,omitempty"`
	SortOrder     string `form:"sort_order" json:"sort_order,omitempty"`
	Search        string `form:"search" json:"search,omitempty"`
}

// ListSubnetsResponse represents the response after listing subnets (HTTP layer)
type ListSubnetsResponse struct {
	Subnets []SubnetInfoResponse `json:"subnets"`
}

// ListSecurityGroupsRequest represents a request to list security groups (HTTP layer)
type ListSecurityGroupsRequest struct {
	VPCID           string `form:"vpc_id" json:"vpc_id,omitempty"`
	Region          string `form:"region" json:"region,omitempty"`
	SecurityGroupID string `form:"security_group_id" json:"security_group_id,omitempty"`
	ResourceGroup   string `form:"resource_group" json:"resource_group,omitempty"` // Azure-specific: Resource Group filter
	Page            int    `form:"page" json:"page,omitempty"`
	Limit           int    `form:"limit" json:"limit,omitempty"`
	SortBy          string `form:"sort_by" json:"sort_by,omitempty"`
	SortOrder       string `form:"sort_order" json:"sort_order,omitempty"`
	Search          string `form:"search" json:"search,omitempty"`
}

// ListSecurityGroupsResponse represents the response after listing security groups (HTTP layer)
type ListSecurityGroupsResponse struct {
	SecurityGroups []SecurityGroupInfoResponse `json:"security_groups"`
}

// CreateVPCRequest represents a request to create a VPC (HTTP layer)
type CreateVPCRequest struct {
	Name              string            `json:"name" validate:"required,min=1,max=255"`
	Description       string            `json:"description,omitempty"`
	CIDRBlock         string            `json:"cidr_block,omitempty"`
	Region            string            `json:"region,omitempty"`
	ProjectID         string            `json:"project_id,omitempty"`
	AutoCreateSubnets *bool             `json:"auto_create_subnets,omitempty"`
	RoutingMode       string            `json:"routing_mode,omitempty"`
	MTU               int64             `json:"mtu,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// UpdateVPCRequest represents a request to update a VPC (HTTP layer)
type UpdateVPCRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// GetVPCRequest represents a request to get a specific VPC (HTTP layer)
type GetVPCRequest struct {
	VPCID  string `form:"vpc_id" uri:"id" json:"vpc_id"`
	Region string `form:"region" json:"region,omitempty"`
}

// DeleteVPCRequest represents a request to delete a VPC (HTTP layer)
type DeleteVPCRequest struct {
	VPCID  string `form:"vpc_id" uri:"id" json:"vpc_id"`
	Region string `form:"region" json:"region,omitempty"`
}

// CreateSubnetRequest represents a request to create a subnet (HTTP layer)
type CreateSubnetRequest struct {
	Name                  string            `json:"name" validate:"required,min=1,max=255"`
	VPCID                 string            `json:"vpc_id" validate:"required"`
	CIDRBlock             string            `json:"cidr_block" validate:"required,cidr"`
	AvailabilityZone      string            `json:"availability_zone" validate:"required"`
	Region                string            `json:"region" validate:"required"`
	Description           string            `json:"description,omitempty"`
	PrivateIPGoogleAccess bool              `json:"private_ip_google_access,omitempty"`
	FlowLogs              bool              `json:"flow_logs,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// UpdateSubnetRequest represents a request to update a subnet (HTTP layer)
type UpdateSubnetRequest struct {
	Name                  string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description           string            `json:"description,omitempty"`
	PrivateIPGoogleAccess *bool             `json:"private_ip_google_access,omitempty"`
	FlowLogs              *bool             `json:"flow_logs,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// GetSubnetRequest represents a request to get a specific subnet (HTTP layer)
type GetSubnetRequest struct {
	SubnetID string `form:"subnet_id" uri:"id" json:"subnet_id"`
	Region   string `form:"region" json:"region"`
}

// DeleteSubnetRequest represents a request to delete a subnet (HTTP layer)
type DeleteSubnetRequest struct {
	SubnetID string `form:"subnet_id" uri:"id" json:"subnet_id"`
	Region   string `form:"region" json:"region"`
}

// CreateSecurityGroupRequest represents a request to create a security group (HTTP layer)
type CreateSecurityGroupRequest struct {
	Name         string                   `json:"name" validate:"required,min=1,max=255"`
	Description  string                   `json:"description" validate:"required,min=1,max=255"`
	VPCID        string                   `json:"vpc_id" validate:"required"`
	Region       string                   `json:"region" validate:"required"`
	ProjectID    string                   `json:"project_id,omitempty"`
	Direction    string                   `json:"direction,omitempty"`
	Priority     int64                    `json:"priority,omitempty"`
	Action       string                   `json:"action,omitempty"`
	Protocol     string                   `json:"protocol,omitempty"`
	Ports        []string                 `json:"ports,omitempty"`
	SourceRanges []string                 `json:"source_ranges,omitempty"`
	TargetTags   []string                 `json:"target_tags,omitempty"`
	Allowed      []FirewallAllowedRequest `json:"allowed,omitempty"`
	Denied       []FirewallDeniedRequest  `json:"denied,omitempty"`
	Tags         map[string]string        `json:"tags,omitempty"`
}

// FirewallAllowedRequest represents GCP firewall allowed rule (HTTP layer)
type FirewallAllowedRequest struct {
	Protocol string   `json:"protocol"`
	Ports    []string `json:"ports,omitempty"`
}

// FirewallDeniedRequest represents GCP firewall denied rule (HTTP layer)
type FirewallDeniedRequest struct {
	Protocol string   `json:"protocol"`
	Ports    []string `json:"ports,omitempty"`
}

// UpdateSecurityGroupRequest represents a request to update a security group (HTTP layer)
type UpdateSecurityGroupRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description string            `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// GetSecurityGroupRequest represents a request to get a specific security group (HTTP layer)
type GetSecurityGroupRequest struct {
	SecurityGroupID string `form:"security_group_id" uri:"id" json:"security_group_id"`
	Region          string `form:"region" json:"region"`
}

// DeleteSecurityGroupRequest represents a request to delete a security group (HTTP layer)
type DeleteSecurityGroupRequest struct {
	SecurityGroupID string `form:"security_group_id" uri:"id" json:"security_group_id"`
	Region          string `form:"region" json:"region"`
}

// AddSecurityGroupRuleRequest represents a request to add a security group rule (HTTP layer)
type AddSecurityGroupRuleRequest struct {
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=ingress egress"`
	Protocol        string   `json:"protocol" validate:"required"`
	FromPort        int32    `json:"from_port"`
	ToPort          int32    `json:"to_port"`
	CIDRBlocks      []string `json:"cidr_blocks,omitempty"`
	SourceGroups    []string `json:"source_groups,omitempty"`
	Description     string   `json:"description,omitempty"`
}

// RemoveSecurityGroupRuleRequest represents a request to remove a security group rule (HTTP layer)
type RemoveSecurityGroupRuleRequest struct {
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=ingress egress"`
	Protocol        string   `json:"protocol" validate:"required"`
	FromPort        int32    `json:"from_port"`
	ToPort          int32    `json:"to_port"`
	CIDRBlocks      []string `json:"cidr_blocks,omitempty"`
	SourceGroups    []string `json:"source_groups,omitempty"`
}

// UpdateSecurityGroupRulesRequest represents a request to update security group rules (HTTP layer)
type UpdateSecurityGroupRulesRequest struct {
	SecurityGroupID string                          `json:"security_group_id" validate:"required"`
	Region          string                          `json:"region" validate:"required"`
	IngressRules    []SecurityGroupRuleInfoResponse `json:"ingress_rules,omitempty"`
	EgressRules     []SecurityGroupRuleInfoResponse `json:"egress_rules,omitempty"`
}

// AddFirewallRuleRequest represents a request to add a specific firewall rule (GCP specific, HTTP layer)
type AddFirewallRuleRequest struct {
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Protocol        string   `json:"protocol" validate:"required"`
	Ports           []string `json:"ports,omitempty"`
	SourceRanges    []string `json:"source_ranges,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
	Action          string   `json:"action,omitempty"` // ALLOW/DENY
}

// RemoveFirewallRuleRequest represents a request to remove a specific firewall rule (GCP specific, HTTP layer)
type RemoveFirewallRuleRequest struct {
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Protocol        string   `json:"protocol" validate:"required"`
	Ports           []string `json:"ports,omitempty"`
	SourceRanges    []string `json:"source_ranges,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
}

// SecurityGroupRuleResponse represents a response for security group rule operations (HTTP layer)
type SecurityGroupRuleResponse struct {
	Success bool                       `json:"success"`
	Message string                     `json:"message"`
	Data    *SecurityGroupInfoResponse `json:"data,omitempty"`
}

// Transformation Functions: Handler Types -> Service DTOs

// ToServiceListVPCsRequest converts handler request to service request
func ToServiceListVPCsRequest(req ListVPCsRequest, credentialID string) networkservice.ListVPCsRequest {
	return networkservice.ListVPCsRequest{
		CredentialID:  credentialID,
		Region:        req.Region,
		VPCID:         req.VPCID,
		ResourceGroup: req.ResourceGroup,
		Page:          req.Page,
		Limit:         req.Limit,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		Search:        req.Search,
	}
}

// ToServiceListSubnetsRequest converts handler request to service request
func ToServiceListSubnetsRequest(req ListSubnetsRequest, credentialID string) networkservice.ListSubnetsRequest {
	return networkservice.ListSubnetsRequest{
		CredentialID:  credentialID,
		VPCID:         req.VPCID,
		Region:        req.Region,
		SubnetID:      req.SubnetID,
		ResourceGroup: req.ResourceGroup,
		Page:          req.Page,
		Limit:         req.Limit,
		SortBy:        req.SortBy,
		SortOrder:     req.SortOrder,
		Search:        req.Search,
	}
}

// ToServiceListSecurityGroupsRequest converts handler request to service request
func ToServiceListSecurityGroupsRequest(req ListSecurityGroupsRequest, credentialID string) networkservice.ListSecurityGroupsRequest {
	return networkservice.ListSecurityGroupsRequest{
		CredentialID:    credentialID,
		VPCID:           req.VPCID,
		Region:          req.Region,
		SecurityGroupID: req.SecurityGroupID,
		ResourceGroup:   req.ResourceGroup,
		Page:            req.Page,
		Limit:           req.Limit,
		SortBy:          req.SortBy,
		SortOrder:       req.SortOrder,
		Search:          req.Search,
	}
}

// ToServiceCreateVPCRequest converts handler request to service request
func ToServiceCreateVPCRequest(req CreateVPCRequest, credentialID string) networkservice.CreateVPCRequest {
	return networkservice.CreateVPCRequest{
		CredentialID:      credentialID,
		Name:              req.Name,
		Description:       req.Description,
		CIDRBlock:         req.CIDRBlock,
		Region:            req.Region,
		ProjectID:         req.ProjectID,
		AutoCreateSubnets: req.AutoCreateSubnets,
		RoutingMode:       req.RoutingMode,
		MTU:               req.MTU,
		Tags:              req.Tags,
	}
}

// ToServiceUpdateVPCRequest converts handler request to service request
func ToServiceUpdateVPCRequest(req UpdateVPCRequest) networkservice.UpdateVPCRequest {
	return networkservice.UpdateVPCRequest{
		Name: req.Name,
		Tags: req.Tags,
	}
}

// ToServiceGetVPCRequest converts handler request to service request
func ToServiceGetVPCRequest(req GetVPCRequest, credentialID string) networkservice.GetVPCRequest {
	return networkservice.GetVPCRequest{
		CredentialID: credentialID,
		VPCID:        req.VPCID,
		Region:       req.Region,
	}
}

// ToServiceDeleteVPCRequest converts handler request to service request
func ToServiceDeleteVPCRequest(req DeleteVPCRequest, credentialID string) networkservice.DeleteVPCRequest {
	return networkservice.DeleteVPCRequest{
		CredentialID: credentialID,
		VPCID:        req.VPCID,
		Region:       req.Region,
	}
}

// ToServiceCreateSubnetRequest converts handler request to service request
func ToServiceCreateSubnetRequest(req CreateSubnetRequest, credentialID string) networkservice.CreateSubnetRequest {
	return networkservice.CreateSubnetRequest{
		CredentialID:          credentialID,
		Name:                  req.Name,
		VPCID:                 req.VPCID,
		CIDRBlock:             req.CIDRBlock,
		AvailabilityZone:      req.AvailabilityZone,
		Region:                req.Region,
		Description:           req.Description,
		PrivateIPGoogleAccess: req.PrivateIPGoogleAccess,
		FlowLogs:              req.FlowLogs,
		Tags:                  req.Tags,
	}
}

// ToServiceUpdateSubnetRequest converts handler request to service request
func ToServiceUpdateSubnetRequest(req UpdateSubnetRequest) networkservice.UpdateSubnetRequest {
	return networkservice.UpdateSubnetRequest{
		Name:                  req.Name,
		Description:           req.Description,
		PrivateIPGoogleAccess: req.PrivateIPGoogleAccess,
		FlowLogs:              req.FlowLogs,
		Tags:                  req.Tags,
	}
}

// ToServiceGetSubnetRequest converts handler request to service request
func ToServiceGetSubnetRequest(req GetSubnetRequest, credentialID string) networkservice.GetSubnetRequest {
	return networkservice.GetSubnetRequest{
		CredentialID: credentialID,
		SubnetID:     req.SubnetID,
		Region:       req.Region,
	}
}

// ToServiceDeleteSubnetRequest converts handler request to service request
func ToServiceDeleteSubnetRequest(req DeleteSubnetRequest, credentialID string) networkservice.DeleteSubnetRequest {
	return networkservice.DeleteSubnetRequest{
		CredentialID: credentialID,
		SubnetID:     req.SubnetID,
		Region:       req.Region,
	}
}

// ToServiceCreateSecurityGroupRequest converts handler request to service request
func ToServiceCreateSecurityGroupRequest(req CreateSecurityGroupRequest, credentialID string) networkservice.CreateSecurityGroupRequest {
	allowed := make([]networkservice.FirewallAllowed, 0, len(req.Allowed))
	for _, a := range req.Allowed {
		allowed = append(allowed, networkservice.FirewallAllowed{
			Protocol: a.Protocol,
			Ports:    a.Ports,
		})
	}

	denied := make([]networkservice.FirewallDenied, 0, len(req.Denied))
	for _, d := range req.Denied {
		denied = append(denied, networkservice.FirewallDenied{
			Protocol: d.Protocol,
			Ports:    d.Ports,
		})
	}

	return networkservice.CreateSecurityGroupRequest{
		CredentialID: credentialID,
		Name:         req.Name,
		Description:  req.Description,
		VPCID:        req.VPCID,
		Region:       req.Region,
		ProjectID:    req.ProjectID,
		Direction:    req.Direction,
		Priority:     req.Priority,
		Action:       req.Action,
		Protocol:     req.Protocol,
		Ports:        req.Ports,
		SourceRanges: req.SourceRanges,
		TargetTags:   req.TargetTags,
		Allowed:      allowed,
		Denied:       denied,
		Tags:         req.Tags,
	}
}

// ToServiceUpdateSecurityGroupRequest converts handler request to service request
func ToServiceUpdateSecurityGroupRequest(req UpdateSecurityGroupRequest) networkservice.UpdateSecurityGroupRequest {
	return networkservice.UpdateSecurityGroupRequest{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
	}
}

// ToServiceGetSecurityGroupRequest converts handler request to service request
func ToServiceGetSecurityGroupRequest(req GetSecurityGroupRequest, credentialID string) networkservice.GetSecurityGroupRequest {
	return networkservice.GetSecurityGroupRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
}

// ToServiceDeleteSecurityGroupRequest converts handler request to service request
func ToServiceDeleteSecurityGroupRequest(req DeleteSecurityGroupRequest, credentialID string) networkservice.DeleteSecurityGroupRequest {
	return networkservice.DeleteSecurityGroupRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
	}
}

// ToServiceAddSecurityGroupRuleRequest converts handler request to service request
func ToServiceAddSecurityGroupRuleRequest(req AddSecurityGroupRuleRequest, credentialID string) networkservice.AddSecurityGroupRuleRequest {
	return networkservice.AddSecurityGroupRuleRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
		Type:            req.Type,
		Protocol:        req.Protocol,
		FromPort:        req.FromPort,
		ToPort:          req.ToPort,
		CIDRBlocks:      req.CIDRBlocks,
		SourceGroups:    req.SourceGroups,
		Description:     req.Description,
	}
}

// ToServiceRemoveSecurityGroupRuleRequest converts handler request to service request
func ToServiceRemoveSecurityGroupRuleRequest(req RemoveSecurityGroupRuleRequest, credentialID string) networkservice.RemoveSecurityGroupRuleRequest {
	return networkservice.RemoveSecurityGroupRuleRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
		Type:            req.Type,
		Protocol:        req.Protocol,
		FromPort:        req.FromPort,
		ToPort:          req.ToPort,
		CIDRBlocks:      req.CIDRBlocks,
		SourceGroups:    req.SourceGroups,
	}
}

// ToServiceUpdateSecurityGroupRulesRequest converts handler request to service request
func ToServiceUpdateSecurityGroupRulesRequest(req UpdateSecurityGroupRulesRequest, credentialID string) networkservice.UpdateSecurityGroupRulesRequest {
	ingressRules := make([]networkservice.SecurityGroupRuleInfo, 0, len(req.IngressRules))
	for _, r := range req.IngressRules {
		ingressRules = append(ingressRules, networkservice.SecurityGroupRuleInfo{
			ID:           r.ID,
			Type:         r.Type,
			Protocol:     r.Protocol,
			FromPort:     r.FromPort,
			ToPort:       r.ToPort,
			CIDRBlocks:   r.CIDRBlocks,
			SourceGroups: r.SourceGroups,
			Description:  r.Description,
		})
	}

	egressRules := make([]networkservice.SecurityGroupRuleInfo, 0, len(req.EgressRules))
	for _, r := range req.EgressRules {
		egressRules = append(egressRules, networkservice.SecurityGroupRuleInfo{
			ID:           r.ID,
			Type:         r.Type,
			Protocol:     r.Protocol,
			FromPort:     r.FromPort,
			ToPort:       r.ToPort,
			CIDRBlocks:   r.CIDRBlocks,
			SourceGroups: r.SourceGroups,
			Description:  r.Description,
		})
	}

	return networkservice.UpdateSecurityGroupRulesRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
		IngressRules:    ingressRules,
		EgressRules:     egressRules,
	}
}

// ToServiceAddFirewallRuleRequest converts handler request to service request
func ToServiceAddFirewallRuleRequest(req AddFirewallRuleRequest, credentialID string) networkservice.AddFirewallRuleRequest {
	return networkservice.AddFirewallRuleRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
		Protocol:        req.Protocol,
		Ports:           req.Ports,
		SourceRanges:    req.SourceRanges,
		TargetTags:      req.TargetTags,
		Action:          req.Action,
	}
}

// ToServiceRemoveFirewallRuleRequest converts handler request to service request
func ToServiceRemoveFirewallRuleRequest(req RemoveFirewallRuleRequest, credentialID string) networkservice.RemoveFirewallRuleRequest {
	return networkservice.RemoveFirewallRuleRequest{
		CredentialID:    credentialID,
		SecurityGroupID: req.SecurityGroupID,
		Region:          req.Region,
		Protocol:        req.Protocol,
		Ports:           req.Ports,
		SourceRanges:    req.SourceRanges,
		TargetTags:      req.TargetTags,
	}
}

// Transformation Functions: Service DTOs -> Handler Types

// FromServiceVPCInfo converts service VPCInfo to handler response
func FromServiceVPCInfo(vpc *networkservice.VPCInfo) *VPCInfoResponse {
	if vpc == nil {
		return nil
	}

	var gateway *GatewayInfoResponse
	if vpc.Gateway != nil {
		gateway = &GatewayInfoResponse{
			Type:      vpc.Gateway.Type,
			IPAddress: vpc.Gateway.IPAddress,
			Name:      vpc.Gateway.Name,
		}
	}

	return &VPCInfoResponse{
		ID:                vpc.ID,
		Name:              vpc.Name,
		State:             vpc.State,
		IsDefault:         vpc.IsDefault,
		NetworkMode:       vpc.NetworkMode,
		RoutingMode:       vpc.RoutingMode,
		MTU:               vpc.MTU,
		AutoSubnets:       vpc.AutoSubnets,
		Description:       vpc.Description,
		FirewallRuleCount: vpc.FirewallRuleCount,
		Gateway:           gateway,
		CreationTimestamp: vpc.CreationTimestamp,
		Tags:              vpc.Tags,
	}
}

// FromServiceSubnetInfo converts service SubnetInfo to handler response
func FromServiceSubnetInfo(subnet *networkservice.SubnetInfo) *SubnetInfoResponse {
	if subnet == nil {
		return nil
	}

	return &SubnetInfoResponse{
		ID:                    subnet.ID,
		Name:                  subnet.Name,
		VPCID:                 subnet.VPCID,
		CIDRBlock:             subnet.CIDRBlock,
		AvailabilityZone:      subnet.AvailabilityZone,
		State:                 subnet.State,
		IsPublic:              subnet.IsPublic,
		Region:                subnet.Region,
		Description:           subnet.Description,
		GatewayAddress:        subnet.GatewayAddress,
		PrivateIPGoogleAccess: subnet.PrivateIPGoogleAccess,
		FlowLogs:              subnet.FlowLogs,
		CreationTimestamp:     subnet.CreationTimestamp,
		Tags:                  subnet.Tags,
	}
}

// FromServiceSecurityGroupInfo converts service SecurityGroupInfo to handler response
func FromServiceSecurityGroupInfo(sg *networkservice.SecurityGroupInfo) *SecurityGroupInfoResponse {
	if sg == nil {
		return nil
	}

	rules := make([]SecurityGroupRuleInfoResponse, 0, len(sg.Rules))
	for _, r := range sg.Rules {
		rules = append(rules, SecurityGroupRuleInfoResponse{
			ID:           r.ID,
			Type:         r.Type,
			Protocol:     r.Protocol,
			FromPort:     r.FromPort,
			ToPort:       r.ToPort,
			CIDRBlocks:   r.CIDRBlocks,
			SourceGroups: r.SourceGroups,
			Description:  r.Description,
		})
	}

	return &SecurityGroupInfoResponse{
		ID:          sg.ID,
		Name:        sg.Name,
		Description: sg.Description,
		VPCID:       sg.VPCID,
		Region:      sg.Region,
		Rules:       rules,
		Tags:        sg.Tags,
	}
}

// FromServiceListVPCsResponse converts service response to handler response
func FromServiceListVPCsResponse(resp *networkservice.ListVPCsResponse) *ListVPCsResponse {
	if resp == nil {
		return &ListVPCsResponse{VPCs: []VPCInfoResponse{}}
	}

	vpcs := make([]VPCInfoResponse, 0, len(resp.VPCs))
	for _, vpc := range resp.VPCs {
		vpcs = append(vpcs, *FromServiceVPCInfo(&vpc))
	}

	return &ListVPCsResponse{VPCs: vpcs}
}

// FromServiceListSubnetsResponse converts service response to handler response
func FromServiceListSubnetsResponse(resp *networkservice.ListSubnetsResponse) *ListSubnetsResponse {
	if resp == nil {
		return &ListSubnetsResponse{Subnets: []SubnetInfoResponse{}}
	}

	subnets := make([]SubnetInfoResponse, 0, len(resp.Subnets))
	for _, subnet := range resp.Subnets {
		subnets = append(subnets, *FromServiceSubnetInfo(&subnet))
	}

	return &ListSubnetsResponse{Subnets: subnets}
}

// FromServiceListSecurityGroupsResponse converts service response to handler response
func FromServiceListSecurityGroupsResponse(resp *networkservice.ListSecurityGroupsResponse) *ListSecurityGroupsResponse {
	if resp == nil {
		return &ListSecurityGroupsResponse{SecurityGroups: []SecurityGroupInfoResponse{}}
	}

	sgs := make([]SecurityGroupInfoResponse, 0, len(resp.SecurityGroups))
	for _, sg := range resp.SecurityGroups {
		sgs = append(sgs, *FromServiceSecurityGroupInfo(&sg))
	}

	return &ListSecurityGroupsResponse{SecurityGroups: sgs}
}

// FromServiceSecurityGroupRuleResponse converts service response to handler response
func FromServiceSecurityGroupRuleResponse(resp *networkservice.SecurityGroupRuleResponse) *SecurityGroupRuleResponse {
	if resp == nil {
		return &SecurityGroupRuleResponse{
			Success: false,
			Message: "",
		}
	}

	var data *SecurityGroupInfoResponse
	if resp.Data != nil {
		data = FromServiceSecurityGroupInfo(resp.Data)
	}

	return &SecurityGroupRuleResponse{
		Success: resp.Success,
		Message: resp.Message,
		Data:    data,
	}
}

// FromServiceSecurityGroupInfoToRuleResponse converts SecurityGroupInfo to SecurityGroupRuleResponse
func FromServiceSecurityGroupInfoToRuleResponse(sg *networkservice.SecurityGroupInfo, success bool, message string) *SecurityGroupRuleResponse {
	if sg == nil {
		return &SecurityGroupRuleResponse{
			Success: success,
			Message: message,
			Data:    nil,
		}
	}

	data := FromServiceSecurityGroupInfo(sg)
	return &SecurityGroupRuleResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
}
