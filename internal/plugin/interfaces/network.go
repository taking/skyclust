package interfaces

import "context"

// NetworkProvider defines the interface for network management
type NetworkProvider interface {
	// GetName returns the name of the network provider
	GetName() string

	// GetVersion returns the version of the network provider
	GetVersion() string

	// Initialize initializes the network provider with configuration
	Initialize(config map[string]interface{}) error

	// VPC/Network Management
	CreateVPC(ctx context.Context, req CreateVPCRequest) (*VPC, error)
	GetVPC(ctx context.Context, vpcID string) (*VPC, error)
	ListVPCs(ctx context.Context) ([]VPC, error)
	UpdateVPC(ctx context.Context, vpcID string, req UpdateVPCRequest) (*VPC, error)
	DeleteVPC(ctx context.Context, vpcID string) error

	// Subnet Management
	CreateSubnet(ctx context.Context, req CreateSubnetRequest) (*Subnet, error)
	GetSubnet(ctx context.Context, subnetID string) (*Subnet, error)
	ListSubnets(ctx context.Context, vpcID string) ([]Subnet, error)
	UpdateSubnet(ctx context.Context, subnetID string, req UpdateSubnetRequest) (*Subnet, error)
	DeleteSubnet(ctx context.Context, subnetID string) error

	// Security Group Management
	CreateSecurityGroup(ctx context.Context, req CreateSecurityGroupRequest) (*SecurityGroup, error)
	GetSecurityGroup(ctx context.Context, sgID string) (*SecurityGroup, error)
	ListSecurityGroups(ctx context.Context, vpcID string) ([]SecurityGroup, error)
	UpdateSecurityGroup(ctx context.Context, sgID string, req UpdateSecurityGroupRequest) (*SecurityGroup, error)
	DeleteSecurityGroup(ctx context.Context, sgID string) error

	// Security Group Rules
	CreateSecurityGroupRule(ctx context.Context, req CreateSecurityGroupRuleRequest) (*SecurityGroupRule, error)
	GetSecurityGroupRule(ctx context.Context, ruleID string) (*SecurityGroupRule, error)
	ListSecurityGroupRules(ctx context.Context, sgID string) ([]SecurityGroupRule, error)
	DeleteSecurityGroupRule(ctx context.Context, ruleID string) error

	// Key Pair Management
	CreateKeyPair(ctx context.Context, req CreateKeyPairRequest) (*KeyPair, error)
	GetKeyPair(ctx context.Context, keyPairID string) (*KeyPair, error)
	ListKeyPairs(ctx context.Context) ([]KeyPair, error)
	DeleteKeyPair(ctx context.Context, keyPairID string) error

	// Load Balancer Management
	CreateLoadBalancer(ctx context.Context, req CreateLoadBalancerRequest) (*LoadBalancer, error)
	GetLoadBalancer(ctx context.Context, lbID string) (*LoadBalancer, error)
	ListLoadBalancers(ctx context.Context) ([]LoadBalancer, error)
	UpdateLoadBalancer(ctx context.Context, lbID string, req UpdateLoadBalancerRequest) (*LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, lbID string) error
}

// VPC represents a Virtual Private Cloud
type VPC struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CIDR      string            `json:"cidr"`
	State     string            `json:"state"`
	Region    string            `json:"region"`
	CreatedAt string            `json:"created_at"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Subnet represents a subnet within a VPC
type Subnet struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	VPCID            string            `json:"vpc_id"`
	CIDR             string            `json:"cidr"`
	AvailabilityZone string            `json:"availability_zone"`
	State            string            `json:"state"`
	CreatedAt        string            `json:"created_at"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// SecurityGroup represents a security group
type SecurityGroup struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	VPCID       string            `json:"vpc_id"`
	CreatedAt   string            `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// SecurityGroupRule represents a security group rule
type SecurityGroupRule struct {
	ID              string `json:"id"`
	SecurityGroupID string `json:"security_group_id"`
	Type            string `json:"type"`     // ingress, egress
	Protocol        string `json:"protocol"` // tcp, udp, icmp, all
	PortFrom        int    `json:"port_from"`
	PortTo          int    `json:"port_to"`
	Source          string `json:"source"` // CIDR or security group ID
	Description     string `json:"description"`
	CreatedAt       string `json:"created_at"`
}

// KeyPair represents a key pair for SSH access
type KeyPair struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Fingerprint string            `json:"fingerprint"`
	PublicKey   string            `json:"public_key"`
	PrivateKey  string            `json:"private_key,omitempty"` // Only returned on creation
	CreatedAt   string            `json:"created_at"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// LoadBalancer represents a load balancer
type LoadBalancer struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"` // application, network, classic
	State     string            `json:"state"`
	VPCID     string            `json:"vpc_id"`
	SubnetIDs []string          `json:"subnet_ids"`
	DNSName   string            `json:"dns_name"`
	Port      int               `json:"port"`
	Protocol  string            `json:"protocol"`
	CreatedAt string            `json:"created_at"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Request types
type CreateVPCRequest struct {
	Name   string            `json:"name"`
	CIDR   string            `json:"cidr"`
	Region string            `json:"region"`
	Tags   map[string]string `json:"tags,omitempty"`
}

type UpdateVPCRequest struct {
	Name string            `json:"name,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

type CreateSubnetRequest struct {
	Name             string            `json:"name"`
	VPCID            string            `json:"vpc_id"`
	CIDR             string            `json:"cidr"`
	AvailabilityZone string            `json:"availability_zone"`
	Tags             map[string]string `json:"tags,omitempty"`
}

type UpdateSubnetRequest struct {
	Name string            `json:"name,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

type CreateSecurityGroupRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	VPCID       string            `json:"vpc_id"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type UpdateSecurityGroupRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

type CreateSecurityGroupRuleRequest struct {
	SecurityGroupID string `json:"security_group_id"`
	Type            string `json:"type"` // ingress, egress
	Protocol        string `json:"protocol"`
	PortFrom        int    `json:"port_from"`
	PortTo          int    `json:"port_to"`
	Source          string `json:"source"`
	Description     string `json:"description,omitempty"`
}

type CreateKeyPairRequest struct {
	Name      string            `json:"name"`
	PublicKey string            `json:"public_key,omitempty"` // If not provided, will be generated
	Tags      map[string]string `json:"tags,omitempty"`
}

type CreateLoadBalancerRequest struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	VPCID     string            `json:"vpc_id"`
	SubnetIDs []string          `json:"subnet_ids"`
	Port      int               `json:"port"`
	Protocol  string            `json:"protocol"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type UpdateLoadBalancerRequest struct {
	Name string            `json:"name,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}
