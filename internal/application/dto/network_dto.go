package dto

// Network Resource DTOs

// VPCInfo represents VPC information
type VPCInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CIDRBlock string            `json:"cidr_block"`
	State     string            `json:"state"`
	IsDefault bool              `json:"is_default"`
	Region    string            `json:"region"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// SubnetInfo represents subnet information
type SubnetInfo struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	VPCID            string            `json:"vpc_id"`
	CIDRBlock        string            `json:"cidr_block"`
	AvailabilityZone string            `json:"availability_zone"`
	State            string            `json:"state"`
	IsPublic         bool              `json:"is_public"`
	Region           string            `json:"region"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// SecurityGroupInfo represents security group information
type SecurityGroupInfo struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	VPCID       string                  `json:"vpc_id"`
	Region      string                  `json:"region"`
	Rules       []SecurityGroupRuleInfo `json:"rules,omitempty"`
	Tags        map[string]string       `json:"tags,omitempty"`
}

// SecurityGroupRuleInfo represents security group rule information
type SecurityGroupRuleInfo struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"` // ingress or egress
	Protocol     string   `json:"protocol"`
	FromPort     int32    `json:"from_port,omitempty"`
	ToPort       int32    `json:"to_port,omitempty"`
	CIDRBlocks   []string `json:"cidr_blocks,omitempty"`
	SourceGroups []string `json:"source_groups,omitempty"`
	Description  string   `json:"description,omitempty"`
}

// Network List DTOs

// ListVPCsRequest represents a request to list VPCs
type ListVPCsRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	Region       string `json:"region,omitempty"`
	VPCID        string `json:"vpc_id,omitempty"`
}

// ListVPCsResponse represents the response after listing VPCs
type ListVPCsResponse struct {
	VPCs []VPCInfo `json:"vpcs"`
}

// ListSubnetsRequest represents a request to list subnets
type ListSubnetsRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id,omitempty"`
	Region       string `json:"region,omitempty"`
	SubnetID     string `json:"subnet_id,omitempty"`
}

// ListSubnetsResponse represents the response after listing subnets
type ListSubnetsResponse struct {
	Subnets []SubnetInfo `json:"subnets"`
}

// ListSecurityGroupsRequest represents a request to list security groups
type ListSecurityGroupsRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	VPCID           string `json:"vpc_id,omitempty"`
	Region          string `json:"region,omitempty"`
	SecurityGroupID string `json:"security_group_id,omitempty"`
}

// ListSecurityGroupsResponse represents the response after listing security groups
type ListSecurityGroupsResponse struct {
	SecurityGroups []SecurityGroupInfo `json:"security_groups"`
}

// Network CRUD DTOs

// CreateVPCRequest represents a request to create a VPC
type CreateVPCRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=255"`
	CIDRBlock    string            `json:"cidr_block" validate:"required,cidr"`
	Region       string            `json:"region" validate:"required"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// UpdateVPCRequest represents a request to update a VPC
type UpdateVPCRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateSubnetRequest represents a request to create a subnet
type CreateSubnetRequest struct {
	CredentialID     string            `json:"credential_id" validate:"required,uuid"`
	Name             string            `json:"name" validate:"required,min=1,max=255"`
	VPCID            string            `json:"vpc_id" validate:"required"`
	CIDRBlock        string            `json:"cidr_block" validate:"required,cidr"`
	AvailabilityZone string            `json:"availability_zone" validate:"required"`
	Region           string            `json:"region" validate:"required"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// UpdateSubnetRequest represents a request to update a subnet
type UpdateSubnetRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateSecurityGroupRequest represents a request to create a security group
type CreateSecurityGroupRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=255"`
	Description  string            `json:"description" validate:"required,min=1,max=255"`
	VPCID        string            `json:"vpc_id" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// UpdateSecurityGroupRequest represents a request to update a security group
type UpdateSecurityGroupRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description string            `json:"description,omitempty" validate:"omitempty,min=1,max=255"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// GetVPCRequest represents a request to get a specific VPC
type GetVPCRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// GetSubnetRequest represents a request to get a specific subnet
type GetSubnetRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	SubnetID     string `json:"subnet_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// GetSecurityGroupRequest represents a request to get a specific security group
type GetSecurityGroupRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	Region          string `json:"region" validate:"required"`
}

// DeleteVPCRequest represents a request to delete a VPC
type DeleteVPCRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	VPCID        string `json:"vpc_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// DeleteSubnetRequest represents a request to delete a subnet
type DeleteSubnetRequest struct {
	CredentialID string `json:"credential_id" validate:"required,uuid"`
	SubnetID     string `json:"subnet_id" validate:"required"`
	Region       string `json:"region" validate:"required"`
}

// DeleteSecurityGroupRequest represents a request to delete a security group
type DeleteSecurityGroupRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	Region          string `json:"region" validate:"required"`
}

// Security Group Rule Management DTOs

// AddSecurityGroupRuleRequest represents a request to add a security group rule
type AddSecurityGroupRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
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

// RemoveSecurityGroupRuleRequest represents a request to remove a security group rule
type RemoveSecurityGroupRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Type            string   `json:"type" validate:"required,oneof=ingress egress"`
	Protocol        string   `json:"protocol" validate:"required"`
	FromPort        int32    `json:"from_port"`
	ToPort          int32    `json:"to_port"`
	CIDRBlocks      []string `json:"cidr_blocks,omitempty"`
	SourceGroups    []string `json:"source_groups,omitempty"`
}

// UpdateSecurityGroupRulesRequest represents a request to update security group rules
type UpdateSecurityGroupRulesRequest struct {
	CredentialID    string                  `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string                  `json:"security_group_id" validate:"required"`
	Region          string                  `json:"region" validate:"required"`
	IngressRules    []SecurityGroupRuleInfo `json:"ingress_rules,omitempty"`
	EgressRules     []SecurityGroupRuleInfo `json:"egress_rules,omitempty"`
}

// SecurityGroupRuleResponse represents a response for security group rule operations
type SecurityGroupRuleResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    *SecurityGroupInfo `json:"data,omitempty"`
}
