package services

import (
	"context"
	"fmt"

	"cmp/pkg/credentials"
	"cmp/pkg/events"
	"cmp/pkg/interfaces"
)

// cloudService implements the CloudService interface
type cloudService struct {
	credentialsService credentials.Service
	eventBus           events.Bus
	providers          map[string]interfaces.CloudProvider
}

// NewCloudService creates a new cloud service
func NewCloudService(credentialsService credentials.Service, eventBus events.Bus) CloudService {
	return &cloudService{
		credentialsService: credentialsService,
		eventBus:           eventBus,
		providers:          make(map[string]interfaces.CloudProvider),
	}
}

// ListProviders lists all available cloud providers
func (cs *cloudService) ListProviders() ([]interfaces.CloudProvider, error) {
	var providers []interfaces.CloudProvider
	for _, provider := range cs.providers {
		providers = append(providers, provider)
	}
	return providers, nil
}

// GetProvider gets a specific provider
func (cs *cloudService) GetProvider(name string) (interfaces.CloudProvider, error) {
	provider, exists := cs.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// InitializeProvider initializes a cloud provider
func (cs *cloudService) InitializeProvider(name string, config map[string]interface{}) error {
	// In a real implementation, you would load the provider plugin
	// For now, we'll create a mock provider
	provider := &mockProvider{name: name}

	if err := provider.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize provider: %w", err)
	}

	cs.providers[name] = provider
	return nil
}

// ListVMs lists all VMs in a workspace
func (cs *cloudService) ListVMs(workspaceID, provider string) ([]interfaces.Instance, error) {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// Get credentials for the workspace
	creds, err := cs.credentialsService.ListCredentials(context.Background(), workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Find credentials for the provider
	var providerCreds *credentials.Credentials
	for _, cred := range creds {
		if cred.Provider == provider {
			providerCreds = cred
			break
		}
	}

	if providerCreds == nil {
		return nil, fmt.Errorf("no credentials found for provider: %s", provider)
	}

	// Decrypt credentials
	config, err := cs.credentialsService.DecryptCredentials(context.Background(), providerCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Convert config to interface{}
	configInterface := make(map[string]interface{})
	for k, v := range config {
		configInterface[k] = v
	}

	// Initialize provider with credentials
	if err := prov.Initialize(configInterface); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	// List instances
	instances, err := prov.ListInstances(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	return instances, nil
}

// CreateVM creates a new VM
func (cs *cloudService) CreateVM(workspaceID, provider string, req CreateInstanceRequest) (*interfaces.Instance, error) {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// Get credentials for the workspace
	creds, err := cs.credentialsService.ListCredentials(context.Background(), workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Find credentials for the provider
	var providerCreds *credentials.Credentials
	for _, cred := range creds {
		if cred.Provider == provider {
			providerCreds = cred
			break
		}
	}

	if providerCreds == nil {
		return nil, fmt.Errorf("no credentials found for provider: %s", provider)
	}

	// Decrypt credentials
	config, err := cs.credentialsService.DecryptCredentials(context.Background(), providerCreds)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	// Convert config to interface{}
	configInterface := make(map[string]interface{})
	for k, v := range config {
		configInterface[k] = v
	}

	// Initialize provider with credentials
	if err := prov.Initialize(configInterface); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	// Create instance
	instance, err := prov.CreateInstance(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Publish event
	cs.eventBus.PublishToWorkspace(context.Background(), workspaceID, &events.Event{
		Type:        "vm.created",
		WorkspaceID: workspaceID,
		Provider:    provider,
		Data: map[string]interface{}{
			"vm_id": instance.ID,
			"name":  instance.Name,
		},
	})

	return instance, nil
}

// GetVM gets a specific VM
func (cs *cloudService) GetVM(workspaceID, provider, vmID string) (*interfaces.Instance, error) {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// Get credentials and initialize provider (similar to ListVMs)
	// ... (omitted for brevity)

	// Get instance status
	status, err := prov.GetInstanceStatus(context.Background(), vmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance status: %w", err)
	}

	// Return mock instance
	return &interfaces.Instance{
		ID:     vmID,
		Name:   "mock-instance",
		Status: status,
		Type:   "t3.micro",
		Region: "us-east-1",
	}, nil
}

// DeleteVM deletes a VM
func (cs *cloudService) DeleteVM(workspaceID, provider, vmID string) error {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return err
	}

	// Get credentials and initialize provider (similar to ListVMs)
	// ... (omitted for brevity)

	// Delete instance
	if err := prov.DeleteInstance(context.Background(), vmID); err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	// Publish event
	cs.eventBus.PublishToWorkspace(context.Background(), workspaceID, &events.Event{
		Type:        "vm.deleted",
		WorkspaceID: workspaceID,
		Provider:    provider,
		Data: map[string]interface{}{
			"vm_id": vmID,
		},
	})

	return nil
}

// StartVM starts a VM
func (cs *cloudService) StartVM(workspaceID, provider, vmID string) error {
	// Mock implementation
	return nil
}

// StopVM stops a VM
func (cs *cloudService) StopVM(workspaceID, provider, vmID string) error {
	// Mock implementation
	return nil
}

// ListRegions lists all regions for a provider
func (cs *cloudService) ListRegions(provider string) ([]interfaces.Region, error) {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	regions, err := prov.ListRegions(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list regions: %w", err)
	}

	return regions, nil
}

// GetCostEstimate gets cost estimate for a provider
func (cs *cloudService) GetCostEstimate(provider string, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
	prov, err := cs.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	estimate, err := prov.GetCostEstimate(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost estimate: %w", err)
	}

	return estimate, nil
}

// CreateInstanceRequest represents a request to create an instance
type CreateInstanceRequest = interfaces.CreateInstanceRequest

// mockProvider is a mock implementation of CloudProvider
type mockProvider struct {
	name   string
	config map[string]interface{}
}

func (mp *mockProvider) GetName() string {
	return mp.name
}

func (mp *mockProvider) GetVersion() string {
	return "1.0.0"
}

func (mp *mockProvider) Initialize(config map[string]interface{}) error {
	mp.config = config
	return nil
}

func (mp *mockProvider) ListInstances(ctx context.Context) ([]interfaces.Instance, error) {
	return []interfaces.Instance{
		{
			ID:        "mock-instance-1",
			Name:      "mock-vm-1",
			Status:    "running",
			Type:      "t3.micro",
			Region:    "us-east-1",
			CreatedAt: "2024-01-01T00:00:00Z",
			Tags: map[string]string{
				"Environment": "development",
			},
		},
	}, nil
}

func (mp *mockProvider) CreateInstance(ctx context.Context, req interfaces.CreateInstanceRequest) (*interfaces.Instance, error) {
	return &interfaces.Instance{
		ID:        "mock-instance-new",
		Name:      req.Name,
		Status:    "creating",
		Type:      req.Type,
		Region:    req.Region,
		CreatedAt: "2024-01-01T00:00:00Z",
		Tags:      req.Tags,
	}, nil
}

func (mp *mockProvider) DeleteInstance(ctx context.Context, instanceID string) error {
	return nil
}

func (mp *mockProvider) GetInstanceStatus(ctx context.Context, instanceID string) (string, error) {
	return "running", nil
}

func (mp *mockProvider) ListRegions(ctx context.Context) ([]interfaces.Region, error) {
	return []interfaces.Region{
		{ID: "us-east-1", Name: "us-east-1", DisplayName: "US East (N. Virginia)", Status: "available"},
		{ID: "us-west-2", Name: "us-west-2", DisplayName: "US West (Oregon)", Status: "available"},
	}, nil
}

func (mp *mockProvider) GetCostEstimate(ctx context.Context, req interfaces.CostEstimateRequest) (*interfaces.CostEstimate, error) {
	return &interfaces.CostEstimate{
		InstanceType: req.InstanceType,
		Region:       req.Region,
		Duration:     req.Duration,
		Cost:         0.01,
		Currency:     "USD",
	}, nil
}

func (mp *mockProvider) GetNetworkProvider() interfaces.NetworkProvider {
	return nil
}

func (mp *mockProvider) GetIAMProvider() interfaces.IAMProvider {
	return nil
}
