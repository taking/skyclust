package network

// Network Resource DTOs

// VPCInfo represents VPC information
type VPCInfo struct {
	ID                string            `json:"id"`   // Clean format: projects/{project}/global/networks/{name}
	Name              string            `json:"name"` // Network name
	State             string            `json:"state"`
	IsDefault         bool              `json:"is_default"`
	Region            string            `json:"-"`                      // Internal field, not exposed in JSON
	NetworkMode       string            `json:"network_mode,omitempty"` // subnet or legacy
	RoutingMode       string            `json:"routing_mode,omitempty"` // REGIONAL or GLOBAL
	MTU               int64             `json:"mtu,omitempty"`
	AutoSubnets       bool              `json:"auto_subnets,omitempty"`
	Description       string            `json:"description,omitempty"`
	FirewallRuleCount int               `json:"firewall_rule_count,omitempty"`
	Gateway           *GatewayInfo      `json:"gateway,omitempty"`
	CreationTimestamp string            `json:"creation_timestamp,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}

// GatewayInfo represents gateway information
type GatewayInfo struct {
	Type      string `json:"type,omitempty"` // NAT, Internet Gateway, etc.
	IPAddress string `json:"ip_address,omitempty"`
	Name      string `json:"name,omitempty"`
}

// SubnetInfo represents subnet information
type SubnetInfo struct {
	ID                    string            `json:"id"`     // Clean format: projects/{project}/regions/{region}/subnetworks/{name}
	Name                  string            `json:"name"`   // Subnet name
	VPCID                 string            `json:"vpc_id"` // Clean format: projects/{project}/global/networks/{name}
	CIDRBlock             string            `json:"cidr_block"`
	AvailabilityZone      string            `json:"availability_zone"` // Clean format: projects/{project}/regions/{region}
	State                 string            `json:"state"`
	IsPublic              bool              `json:"is_public"`
	Region                string            `json:"region"` // Region name only (e.g., asia-northeast3)
	Description           string            `json:"description,omitempty"`
	GatewayAddress        string            `json:"gateway_address,omitempty"`
	PrivateIPGoogleAccess bool              `json:"private_ip_google_access,omitempty"`
	FlowLogs              bool              `json:"flow_logs,omitempty"`
	CreationTimestamp     string            `json:"creation_timestamp,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
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
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	Region        string `json:"region,omitempty"`
	VPCID         string `json:"vpc_id,omitempty"`
	ResourceGroup string `json:"resource_group,omitempty" form:"resource_group"` // Azure-specific: Resource Group filter
	// Pagination parameters
	Page  int `json:"page,omitempty" form:"page" validate:"omitempty,min=1"`
	Limit int `json:"limit,omitempty" form:"limit" validate:"omitempty,min=1,max=100"`
	// Sorting parameters
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`                          // name, state, created_at
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
	// Filtering parameters
	Search string `json:"search,omitempty" form:"search"` // Search in name, description
}

// ListVPCsResponse represents the response after listing VPCs
type ListVPCsResponse struct {
	VPCs []VPCInfo `json:"vpcs"`
	Total int64    `json:"total,omitempty"` // Total count after filtering (before pagination)
}

// ListSubnetsRequest represents a request to list subnets
type ListSubnetsRequest struct {
	CredentialID  string `json:"credential_id" validate:"required,uuid"`
	VPCID         string `json:"vpc_id,omitempty"`
	Region        string `json:"region,omitempty"`
	SubnetID      string `json:"subnet_id,omitempty"`
	ResourceGroup string `json:"resource_group,omitempty" form:"resource_group"` // Azure-specific: Resource Group filter
	// Pagination parameters
	Page  int `json:"page,omitempty" form:"page" validate:"omitempty,min=1"`
	Limit int `json:"limit,omitempty" form:"limit" validate:"omitempty,min=1,max=100"`
	// Sorting parameters
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`                          // name, state, cidr_block, created_at
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
	// Filtering parameters
	Search string `json:"search,omitempty" form:"search"` // Search in name, description, cidr_block
}

// ListSubnetsResponse represents the response after listing subnets
type ListSubnetsResponse struct {
	Subnets []SubnetInfo `json:"subnets"`
	Total   int64        `json:"total,omitempty"` // Total count after filtering (before pagination)
}

// ListSecurityGroupsRequest represents a request to list security groups
type ListSecurityGroupsRequest struct {
	CredentialID    string `json:"credential_id" validate:"required,uuid"`
	VPCID           string `json:"vpc_id,omitempty"`
	Region          string `json:"region,omitempty"`
	SecurityGroupID string `json:"security_group_id,omitempty"`
	ResourceGroup   string `json:"resource_group,omitempty" form:"resource_group"` // Azure-specific: Resource Group filter
	// Pagination parameters
	Page  int `json:"page,omitempty" form:"page" validate:"omitempty,min=1"`
	Limit int `json:"limit,omitempty" form:"limit" validate:"omitempty,min=1,max=100"`
	// Sorting parameters
	SortBy    string `json:"sort_by,omitempty" form:"sort_by"`                          // name, created_at
	SortOrder string `json:"sort_order,omitempty" form:"sort_order" validate:"omitempty,oneof=asc desc"`
	// Filtering parameters
	Search string `json:"search,omitempty" form:"search"` // Search in name, description
}

// ListSecurityGroupsResponse represents the response after listing security groups
type ListSecurityGroupsResponse struct {
	SecurityGroups []SecurityGroupInfo `json:"security_groups"`
	Total          int64               `json:"total,omitempty"` // Total count after filtering (before pagination)
}

// Network CRUD DTOs

// CreateVPCRequest represents a request to create a VPC
type CreateVPCRequest struct {
	CredentialID      string            `json:"credential_id" validate:"omitempty,uuid"` // Optional: can be from body or query
	Name              string            `json:"name" validate:"required,min=1,max=255"`
	Description       string            `json:"description,omitempty"`
	CIDRBlock         string            `json:"cidr_block,omitempty"`          // Optional for GCP subnet mode
	Region            string            `json:"region,omitempty"`              // Optional for GCP (Global resource)
	ProjectID         string            `json:"project_id,omitempty"`          // GCP specific
	AutoCreateSubnets *bool             `json:"auto_create_subnets,omitempty"` // GCP specific
	RoutingMode       string            `json:"routing_mode,omitempty"`        // GCP specific
	MTU               int64             `json:"mtu,omitempty"`                 // GCP specific
	Tags              map[string]string `json:"tags,omitempty"`
}

// UpdateVPCRequest represents a request to update a VPC
type UpdateVPCRequest struct {
	Name string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Tags map[string]string `json:"tags,omitempty"`
}

// CreateSubnetRequest represents a request to create a subnet
type CreateSubnetRequest struct {
	CredentialID          string            `json:"credential_id" validate:"omitempty,uuid"` // Optional: can be from body or query
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

// UpdateSubnetRequest represents a request to update a subnet
type UpdateSubnetRequest struct {
	Name                  string            `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description           string            `json:"description,omitempty"`
	PrivateIPGoogleAccess *bool             `json:"private_ip_google_access,omitempty"`
	FlowLogs              *bool             `json:"flow_logs,omitempty"`
	Tags                  map[string]string `json:"tags,omitempty"`
}

// CreateSecurityGroupRequest represents a request to create a security group
type CreateSecurityGroupRequest struct {
	CredentialID string            `json:"credential_id" validate:"required,uuid"`
	Name         string            `json:"name" validate:"required,min=1,max=255"`
	Description  string            `json:"description" validate:"required,min=1,max=255"`
	VPCID        string            `json:"vpc_id" validate:"required"`
	Region       string            `json:"region" validate:"required"`
	ProjectID    string            `json:"project_id,omitempty"` // GCP specific
	Direction    string            `json:"direction,omitempty"`  // INGRESS/EGRESS
	Priority     int64             `json:"priority,omitempty"`   // GCP specific
	Action       string            `json:"action,omitempty"`     // ALLOW/DENY
	Protocol     string            `json:"protocol,omitempty"`   // tcp/udp/icmp
	Ports        []string          `json:"ports,omitempty"`      // Port numbers
	SourceRanges []string          `json:"source_ranges,omitempty"`
	TargetTags   []string          `json:"target_tags,omitempty"`
	Allowed      []FirewallAllowed `json:"allowed,omitempty"` // GCP specific allowed rules
	Denied       []FirewallDenied  `json:"denied,omitempty"`  // GCP specific denied rules
	Tags         map[string]string `json:"tags,omitempty"`
}

// FirewallAllowed represents GCP firewall allowed rule
type FirewallAllowed struct {
	Protocol string   `json:"protocol"` // tcp/udp/icmp
	Ports    []string `json:"ports,omitempty"`
}

// FirewallDenied represents GCP firewall denied rule
type FirewallDenied struct {
	Protocol string   `json:"protocol"` // tcp/udp/icmp
	Ports    []string `json:"ports,omitempty"`
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

// AddFirewallRuleRequest represents a request to add a specific firewall rule (GCP specific)
type AddFirewallRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Protocol        string   `json:"protocol" validate:"required"`
	Ports           []string `json:"ports,omitempty"`
	SourceRanges    []string `json:"source_ranges,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
	Action          string   `json:"action,omitempty"` // ALLOW/DENY
}

// RemoveFirewallRuleRequest represents a request to remove a specific firewall rule (GCP specific)
type RemoveFirewallRuleRequest struct {
	CredentialID    string   `json:"credential_id" validate:"required,uuid"`
	SecurityGroupID string   `json:"security_group_id" validate:"required"`
	Region          string   `json:"region" validate:"required"`
	Protocol        string   `json:"protocol" validate:"required"`
	Ports           []string `json:"ports,omitempty"`
	SourceRanges    []string `json:"source_ranges,omitempty"`
	TargetTags      []string `json:"target_tags,omitempty"`
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

// GCP-specific VPC DTOs

// CreateGCPVPCRequest represents a request to create a GCP VPC
type CreateGCPVPCRequest struct {
	CredentialID      string            `json:"credential_id" validate:"required,uuid"`
	Name              string            `json:"name" validate:"required,min=1,max=255"`
	Description       string            `json:"description,omitempty"`
	CIDRBlock         string            `json:"cidr_block" validate:"required,cidr"`
	Region            string            `json:"region,omitempty"` // Optional for VPC (Global resource)
	ProjectID         string            `json:"project_id" validate:"required"`
	AutoCreateSubnets bool              `json:"auto_create_subnets,omitempty"`
	RoutingMode       string            `json:"routing_mode,omitempty"` // "REGIONAL" or "GLOBAL"
	MTU               int64             `json:"mtu,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
}
