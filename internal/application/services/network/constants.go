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
	ErrMsgUnsupportedProvider    = "Unsupported provider: %s"
	ErrMsgProviderNotImplemented = "Provider %s is not implemented"
	ErrMsgProjectIDNotFound      = "Project ID not found in credential data"
)
