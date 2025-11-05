package providers

import (
	"fmt"

	networkservice "skyclust/internal/application/services/network"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// ProviderHandler defines the interface that all provider-specific network handlers must implement
type ProviderHandler interface {
	// VPC management
	ListVPCs(c *gin.Context)
	CreateVPC(c *gin.Context)
	GetVPC(c *gin.Context)
	UpdateVPC(c *gin.Context)
	DeleteVPC(c *gin.Context)

	// Subnet management
	ListSubnets(c *gin.Context)
	CreateSubnet(c *gin.Context)
	GetSubnet(c *gin.Context)
	UpdateSubnet(c *gin.Context)
	DeleteSubnet(c *gin.Context)

	// Security Group management
	ListSecurityGroups(c *gin.Context)
	CreateSecurityGroup(c *gin.Context)
	GetSecurityGroup(c *gin.Context)
	UpdateSecurityGroup(c *gin.Context)
	DeleteSecurityGroup(c *gin.Context)

	// Security Group Rule management
	AddSecurityGroupRule(c *gin.Context)
	RemoveSecurityGroupRule(c *gin.Context)
	UpdateSecurityGroupRules(c *gin.Context)
}

// Factory creates and manages provider-specific network handlers
type Factory struct {
	handlers map[string]ProviderHandler
}

// NewFactory creates a new handler factory and registers all available providers
func NewFactory(
	networkService *networkservice.Service,
	credentialService domain.CredentialService,
	logger interface{},
) *Factory {
	factory := &Factory{
		handlers: make(map[string]ProviderHandler),
	}

	// Register all providers
	factory.Register(domain.ProviderAWS, NewAWSHandler(networkService, credentialService))
	factory.Register(domain.ProviderGCP, NewGCPHandler(networkService, credentialService, logger))
	factory.Register(domain.ProviderAzure, NewAzureHandler(networkService, credentialService))
	factory.Register(domain.ProviderNCP, NewNCPHandler(networkService, credentialService))

	return factory
}

// Register registers a provider handler
func (f *Factory) Register(provider string, handler ProviderHandler) {
	f.handlers[provider] = handler
}

// GetHandler returns the handler for a specific provider
func (f *Factory) GetHandler(provider string) (ProviderHandler, error) {
	handler, exists := f.handlers[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
	return handler, nil
}

// GetAllProviders returns a list of all registered providers
func (f *Factory) GetAllProviders() []string {
	providers := make([]string, 0, len(f.handlers))
	for provider := range f.handlers {
		providers = append(providers, provider)
	}
	return providers
}

// IsProviderSupported checks if a provider is supported
func (f *Factory) IsProviderSupported(provider string) bool {
	_, exists := f.handlers[provider]
	return exists
}

