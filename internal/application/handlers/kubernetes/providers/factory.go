package providers

import (
	"fmt"

	kubernetesservice "skyclust/internal/application/services/kubernetes"
	"skyclust/internal/domain"

	"github.com/gin-gonic/gin"
)

// ProviderHandler defines the interface that all provider-specific handlers must implement
type ProviderHandler interface {
	// Cluster management
	CreateCluster(c *gin.Context)
	ListClusters(c *gin.Context)
	GetCluster(c *gin.Context)
	DeleteCluster(c *gin.Context)
	GetKubeconfig(c *gin.Context)

	// Node pool management
	CreateNodePool(c *gin.Context)
	ListNodePools(c *gin.Context)
	GetNodePool(c *gin.Context)
	DeleteNodePool(c *gin.Context)
	ScaleNodePool(c *gin.Context)

	// Node group management (provider-specific, e.g., EKS)
	CreateNodeGroup(c *gin.Context)
	ListNodeGroups(c *gin.Context)
	GetNodeGroup(c *gin.Context)
	UpdateNodeGroup(c *gin.Context)
	DeleteNodeGroup(c *gin.Context)

	// Cluster operations
	UpgradeCluster(c *gin.Context)
	GetUpgradeStatus(c *gin.Context)

	// Node management
	ListNodes(c *gin.Context)
	GetNode(c *gin.Context)
	DrainNode(c *gin.Context)
	CordonNode(c *gin.Context)
	UncordonNode(c *gin.Context)

	// SSH access
	GetNodeSSHConfig(c *gin.Context)
	ExecuteNodeCommand(c *gin.Context)

	// Metadata endpoints (AWS only, other providers return NotImplemented)
	GetEKSVersions(c *gin.Context)
	GetAWSRegions(c *gin.Context)
	GetAvailabilityZones(c *gin.Context)
	GetInstanceTypes(c *gin.Context)
	GetEKSAmitTypes(c *gin.Context)
	CheckGPUQuota(c *gin.Context)
}

// Factory creates and manages provider-specific Kubernetes handlers
type Factory struct {
	handlers map[string]ProviderHandler
}

// NewFactory creates a new handler factory and registers all available providers
func NewFactory(
	k8sService *kubernetesservice.Service,
	credentialService domain.CredentialService,
) *Factory {
	factory := &Factory{
		handlers: make(map[string]ProviderHandler),
	}

	// Register all providers
	factory.Register(domain.ProviderAWS, NewAWSHandler(k8sService, credentialService))
	factory.Register(domain.ProviderGCP, NewGCPHandler(k8sService, credentialService))
	factory.Register(domain.ProviderAzure, NewAzureHandler(k8sService, credentialService))
	factory.Register(domain.ProviderNCP, NewNCPHandler(k8sService, credentialService))

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
