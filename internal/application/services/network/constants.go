package network

import "time"

// Network Service Constants
// These constants are specific to network resource operations

// Resource prefixes
const (
	ResourcePrefixVPC = "vpc-%s-%s"
)

// VPC states
const (
	StateCreating = "creating"
	StateActive   = "active"
	StateDeleting = "deleting"
	StateError    = "error"
)

// Network modes
const (
	NetworkModeSubnet = "subnet"
	NetworkModeGlobal = "global"
)

// Security group actions
const (
	ActionAllow = "allow"
	ActionDeny  = "deny"
)

// Protocol constants
const (
	ProtocolICMP = "icmp"
	ProtocolTCP  = "tcp"
	ProtocolUDP  = "udp"
)

// Operation constants for async operations
const (
	OperationPollInterval  = 5 * time.Second
	OperationTimeout       = 30 * time.Minute
	OperationStatusDone    = "done"
	OperationStatusPending = "pending"
	OperationStatusRunning = "running"
)

// Error message constants
const (
	// Provider errors
	ErrMsgUnsupportedProvider    = "Unsupported provider: %s"
	ErrMsgProviderNotImplemented = "Provider %s is not implemented"
	ErrMsgProjectIDNotFound      = "Project ID not found in credential data"

	// AWS errors
	ErrMsgFailedToCreateEC2Client = "Failed to create AWS EC2 client: %v"
	ErrMsgFailedToDescribeVPCs    = "Failed to describe AWS VPCs: %v"
	ErrMsgFailedToDescribeVPC     = "Failed to describe AWS VPC: %v"
	ErrMsgVPCNotFound             = "VPC not found: %s"
	ErrMsgInvalidVPCIDFormat      = "Invalid VPC ID format: %s"
	ErrMsgRegionRequired           = "Region is required for AWS EC2 client"
	ErrMsgInvalidRegionFormat      = "Invalid region: '%s' appears to be a VPC ID, not a region"

	// GCP errors
	ErrMsgFailedToDecryptCredential   = "Failed to decrypt credential: %v"
	ErrMsgFailedToMarshalCredential  = "Failed to marshal credential data: %v"
	ErrMsgFailedToCreateGCPCompute   = "Failed to create GCP compute service: %v"
	ErrMsgFailedToCreateGCPNetwork   = "Failed to create GCP network: %v"
	ErrMsgFailedToListGCPNetworks    = "Failed to list GCP networks: %v"
	ErrMsgFailedToGetGCPNetwork      = "Failed to get GCP network: %v"
	ErrMsgFailedToDeleteGCPNetwork   = "Failed to delete GCP network: %v"
	ErrMsgFailedToGetGCPVPC          = "Failed to get GCP VPC: %v"
	ErrMsgInvalidVPCIDFormatGCP      = "Invalid VPC ID format: %s"
	ErrMsgInvalidVPCIDOrResourceGroup = "Invalid VPC ID format or resource group not found"

	// Azure errors
	ErrMsgResourceGroupRequired = "Resource group is required for Azure Virtual Network. Please provide it in the credential or request."
	ErrMsgRegionRequiredAzure   = "Region is required for Azure Virtual Network"
	ErrMsgCIDRBlockRequired     = "CIDR block is required for Azure Virtual Network"

	// Subnet errors
	ErrMsgFailedToCreateSubnet = "Failed to create subnet: %v"
	ErrMsgFailedToGetSubnet    = "Failed to get subnet: %v"
	ErrMsgFailedToUpdateSubnet = "Failed to update subnet: %v"
	ErrMsgFailedToDeleteSubnet = "Failed to delete subnet: %v"
	ErrMsgSubnetNotFound       = "Subnet not found: %s"

	// Security Group errors
	ErrMsgFailedToCreateSecurityGroup = "Failed to create security group: %v"
	ErrMsgFailedToGetSecurityGroup     = "Failed to get security group: %v"
	ErrMsgFailedToUpdateSecurityGroup  = "Failed to update security group: %v"
	ErrMsgFailedToDeleteSecurityGroup  = "Failed to delete security group: %v"
	ErrMsgSecurityGroupNotFound        = "Security group not found: %s"
	ErrMsgFailedToDescribeSecurityGroups = "Failed to describe security groups: %v"
	ErrMsgFailedToCreateGCPFirewallRule  = "Failed to create GCP firewall rule: %v"
	ErrMsgFailedToGetGCPFirewallRule     = "Failed to get GCP firewall rule: %v"
	ErrMsgFailedToListGCPFirewallRules   = "Failed to list GCP firewall rules: %v"
	ErrMsgFailedToUpdateSecurityGroupTags = "Failed to update security group tags: %v"
	ErrMsgFailedToGetCurrentFirewallRule = "Failed to get current firewall rule: %v"

	// AWS errors (additional)
	ErrMsgFailedToLoadAWSConfig = "Failed to load AWS config: %v"
	ErrMsgFailedToCreateVPC     = "Failed to create VPC: %v"
	ErrMsgFailedToUpdateVPCTags = "Failed to update VPC tags: %v"

	// GCP errors (additional)
	ErrMsgFailedToListFirewallRules   = "Failed to list firewall rules: %v"
	ErrMsgFailedToListSubnets         = "Failed to list subnets: %v"
	ErrMsgFailedToListInstances       = "Failed to list instances: %v"
	ErrMsgFailedToUpdateGCPFirewallRule = "Failed to update GCP firewall rule: %v"
	ErrMsgFailedToDeleteGCPFirewallRule = "Failed to delete GCP firewall rule: %v"
	ErrMsgFailedToMarshalServiceAccountKey = "Failed to marshal service account key: %v"
	ErrMsgFailedToCreateCredentials = "Failed to create credentials: %v"
	ErrMsgFailedToCreateComputeService = "Failed to create compute service: %v"

	// Security Group Rule errors
	ErrMsgFailedToAddSecurityGroupRule    = "Failed to add security group rule: %v"
	ErrMsgFailedToRemoveSecurityGroupRule = "Failed to remove security group rule: %v"

	// Credential errors
	ErrMsgAccessKeyNotFound = "access_key not found in credential"
	ErrMsgSecretKeyNotFound = "secret_key not found in credential"

	// Additional GCP errors
	ErrMsgFailedToGetCurrentSecurityGroup = "Failed to get current security group: %v"
	ErrMsgFailedToAddIngressRule          = "Failed to add ingress rule: %v"
	ErrMsgFailedToAddEgressRule          = "Failed to add egress rule: %v"
	ErrMsgFailedToCreateGCPComputeClient = "Failed to create GCP compute client: %v"
	ErrMsgFailedToListGCPSubnets          = "Failed to list GCP subnets: %v"
	ErrMsgFailedToGetGCPSubnet            = "Failed to get GCP subnet: %v"
	ErrMsgFailedToCreateGCPSubnet         = "Failed to create GCP subnet: %v"
	ErrMsgFailedToGetCurrentSubnet         = "Failed to get current subnet: %v"
	ErrMsgFailedToUpdateGCPSubnet         = "Failed to update GCP subnet: %v"
	ErrMsgFailedToDeleteGCPSubnet         = "Failed to delete GCP subnet: %v"
	ErrMsgFailedToListRouters             = "Failed to list routers: %v"

	// Additional AWS errors
	ErrMsgFailedToDescribeSubnets = "Failed to describe subnets: %v"
	ErrMsgFailedToDescribeSubnet  = "Failed to describe subnet: %v"
	ErrMsgFailedToUpdateSubnetTags = "Failed to update subnet tags: %v"
)
