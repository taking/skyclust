package services

// CloudProviderService defines the interface for cloud provider operations
type CloudProviderService interface {
	// GetProviders retrieves available cloud providers
	GetProviders() ([]interface{}, error)

	// GetProvider retrieves a specific cloud provider
	GetProvider(providerID string) (interface{}, error)

	// RegisterProvider registers a new cloud provider
	RegisterProvider(provider interface{}) error

	// UpdateProvider updates an existing cloud provider
	UpdateProvider(provider interface{}) error

	// DeleteProvider deletes a cloud provider
	DeleteProvider(providerID string) error

	// TestProviderConnection tests connection to a cloud provider
	TestProviderConnection(providerID string) error

	// GetProviderRegions retrieves available regions for a provider
	GetProviderRegions(providerID string) ([]interface{}, error)

	// GetProviderResources retrieves resources from a cloud provider
	GetProviderResources(providerID string, region string) ([]interface{}, error)

	// GetProviderPricing retrieves pricing information for a provider
	GetProviderPricing(providerID string, region string) (interface{}, error)

	// ValidateProviderCredentials validates provider credentials
	ValidateProviderCredentials(providerID string, credentials map[string]string) error
}
