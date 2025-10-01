package interfaces

import "context"

// CloudProvider defines the interface that all cloud provider plugins must implement
type CloudProvider interface {
	// GetName returns the name of the cloud provider
	GetName() string

	// GetVersion returns the version of the plugin
	GetVersion() string

	// Initialize initializes the cloud provider with configuration
	Initialize(config map[string]interface{}) error

	// Instance Management
	ListInstances(ctx context.Context) ([]Instance, error)
	CreateInstance(ctx context.Context, req CreateInstanceRequest) (*Instance, error)
	DeleteInstance(ctx context.Context, instanceID string) error
	GetInstanceStatus(ctx context.Context, instanceID string) (string, error)

	// Region Management
	ListRegions(ctx context.Context) ([]Region, error)

	// Cost Estimation
	GetCostEstimate(ctx context.Context, req CostEstimateRequest) (*CostEstimate, error)

	// Network Management (optional - can return nil if not supported)
	GetNetworkProvider() NetworkProvider

	// IAM Management (optional - can return nil if not supported)
	GetIAMProvider() IAMProvider
}

// Instance represents a cloud instance
type Instance struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Region    string            `json:"region"`
	CreatedAt string            `json:"created_at"`
	Tags      map[string]string `json:"tags,omitempty"`
	PublicIP  string            `json:"public_ip,omitempty"`
	PrivateIP string            `json:"private_ip,omitempty"`
}

// CreateInstanceRequest represents a request to create an instance
type CreateInstanceRequest struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Region   string            `json:"region"`
	ImageID  string            `json:"image_id"`
	Tags     map[string]string `json:"tags,omitempty"`
	UserData string            `json:"user_data,omitempty"`

	// Network Configuration
	VPCID          string   `json:"vpc_id,omitempty"`
	SubnetID       string   `json:"subnet_id,omitempty"`
	SecurityGroups []string `json:"security_groups,omitempty"`
	PublicIP       bool     `json:"public_ip,omitempty"`

	// Key Pair Configuration
	KeyPairID   string `json:"key_pair_id,omitempty"`
	KeyPairName string `json:"key_pair_name,omitempty"`

	// Storage Configuration
	RootVolumeSize int    `json:"root_volume_size,omitempty"` // GB
	RootVolumeType string `json:"root_volume_type,omitempty"`
}

// Region represents a cloud region
type Region struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
}

// CostEstimateRequest represents a request for cost estimation
type CostEstimateRequest struct {
	InstanceType string `json:"instance_type"`
	Region       string `json:"region"`
	Duration     string `json:"duration"` // e.g., "1h", "1d", "1m"
}

// CostEstimate represents cost estimation result
type CostEstimate struct {
	InstanceType string  `json:"instance_type"`
	Region       string  `json:"region"`
	Duration     string  `json:"duration"`
	Cost         float64 `json:"cost"`
	Currency     string  `json:"currency"`
}

