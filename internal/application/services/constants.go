package service

import "time"

// Cloud Provider constants
const (
	ProviderAWS       = "aws"
	ProviderGCP       = "gcp"
	ProviderAzure     = "azure"
	ProviderNCP       = "ncp"
	ProviderOpenStack = "openstack"
)

// Pagination constants
const (
	DefaultPageLimit = 50
	MaxPageLimit     = 1000
	MinPageLimit     = 1
)

// GCP Operation constants
const (
	OperationPollInterval = 1 * time.Second
	OperationTimeout      = 5 * time.Minute
	OperationStatusDone   = "DONE"
)

// Network resource states
const (
	StateCreating  = "creating"
	StateAvailable = "available"
	StateReady     = "READY"
	StatePending   = "pending"
	StateDeleting  = "deleting"
)

// GCP Network modes
const (
	NetworkModeSubnet = "subnet"
	NetworkModeLegacy = "legacy"
)

// Default values
const (
	DefaultMTU        = 1460
	MaxVPCNameLength  = 63
	MinVPCNameLength  = 1
	DefaultPriority   = 1000
	MaxRetryAttempts  = 3
	RetryDelaySeconds = 1
)

// GCP Routing modes
const (
	RoutingModeRegional = "REGIONAL"
	RoutingModeGlobal   = "GLOBAL"
)

// Firewall directions
const (
	DirectionIngress = "INGRESS"
	DirectionEgress  = "EGRESS"
)

// Firewall actions
const (
	ActionAllow = "ALLOW"
	ActionDeny  = "DENY"
)

// Common protocols
const (
	ProtocolTCP  = "tcp"
	ProtocolUDP  = "udp"
	ProtocolICMP = "icmp"
	ProtocolAll  = "all"
)

// Error messages
const (
	ErrMsgUnsupportedProvider    = "provider %s is not supported"
	ErrMsgProviderNotImplemented = "provider %s is not implemented yet"
	ErrMsgInvalidCredential      = "failed to decrypt credential"
	ErrMsgInvalidVPCID           = "invalid VPC ID format"
	ErrMsgInvalidSubnetID        = "invalid Subnet ID format"
	ErrMsgProjectIDNotFound      = "project_id not found in credential"
	ErrMsgAccessKeyNotFound      = "access_key not found in credential"
	ErrMsgSecretKeyNotFound      = "secret_key not found in credential"
)

// Resource name prefixes
const (
	ResourcePrefixVPC      = "projects/%s/global/networks/%s"
	ResourcePrefixSubnet   = "projects/%s/regions/%s/subnetworks/%s"
	ResourcePrefixFirewall = "projects/%s/global/firewalls/%s"
	ResourcePrefixRegion   = "projects/%s/regions/%s"
)
