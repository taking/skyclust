package service

import (
	"context"
	"fmt"

	"skyclust/internal/plugin"
	plugininterfaces "skyclust/internal/plugin/interfaces"
	"skyclust/pkg/logger"
)

// cloudProviderService implements the CloudProviderService interface
type cloudProviderService struct {
	pluginManager *plugin.Manager
}

// NewCloudProviderService creates a new cloud provider service
func NewCloudProviderService(pluginManager *plugin.Manager) CloudProviderService {
	return &cloudProviderService{
		pluginManager: pluginManager,
	}
}

// CreateInstance creates a new cloud instance
func (s *cloudProviderService) CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error) {
	// Get the cloud provider plugin
	cloudProvider, err := s.pluginManager.GetProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", provider, err)
	}

	// Convert request to plugin format
	pluginReq := plugininterfaces.CreateInstanceRequest{
		Name:   req.Name,
		Type:   req.Type,
		Region: req.Region,
		Tags:   convertToStringMap(req.Metadata),
	}

	// Create instance using plugin
	instance, err := cloudProvider.CreateInstance(ctx, pluginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	// Convert plugin response to service format
	cloudInstance := &CloudInstance{
		ID:       instance.ID,
		Status:   instance.Status,
		Type:     instance.Type,
		Region:   instance.Region,
		Metadata: convertToInterfaceMap(instance.Tags),
	}

	logger.Info(fmt.Sprintf("Created instance %s on provider %s", instance.ID, provider))
	return cloudInstance, nil
}

// GetInstance retrieves a cloud instance
func (s *cloudProviderService) GetInstance(ctx context.Context, provider, instanceID string) (*CloudInstance, error) {
	// Get the cloud provider plugin
	cloudProvider, err := s.pluginManager.GetProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider %s: %w", provider, err)
	}

	// List instances and find the one with matching ID
	instances, err := cloudProvider.ListInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	for _, instance := range instances {
		if instance.ID == instanceID {
			return &CloudInstance{
				ID:       instance.ID,
				Status:   instance.Status,
				Type:     instance.Type,
				Region:   instance.Region,
				Metadata: convertToInterfaceMap(instance.Tags),
			}, nil
		}
	}

	return nil, fmt.Errorf("instance %s not found", instanceID)
}

// DeleteInstance deletes a cloud instance
func (s *cloudProviderService) DeleteInstance(ctx context.Context, provider, instanceID string) error {
	// Get the cloud provider plugin
	cloudProvider, err := s.pluginManager.GetProvider(provider)
	if err != nil {
		return fmt.Errorf("failed to get provider %s: %w", provider, err)
	}

	// Delete instance using plugin
	err = cloudProvider.DeleteInstance(ctx, instanceID)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	logger.Info(fmt.Sprintf("Deleted instance %s on provider %s", instanceID, provider))
	return nil
}

// StartInstance starts a cloud instance
func (s *cloudProviderService) StartInstance(ctx context.Context, provider, instanceID string) error {
	// Note: Start/Stop operations are typically handled by the cloud provider's API
	// This is a placeholder implementation
	logger.Info(fmt.Sprintf("Start instance operation requested for %s on provider %s", instanceID, provider))
	return fmt.Errorf("start instance operation not implemented for provider %s", provider)
}

// StopInstance stops a cloud instance
func (s *cloudProviderService) StopInstance(ctx context.Context, provider, instanceID string) error {
	// Note: Start/Stop operations are typically handled by the cloud provider's API
	// This is a placeholder implementation
	logger.Info(fmt.Sprintf("Stop instance operation requested for %s on provider %s", instanceID, provider))
	return fmt.Errorf("stop instance operation not implemented for provider %s", provider)
}

// GetInstanceStatus gets the status of a cloud instance
func (s *cloudProviderService) GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error) {
	// Get the cloud provider plugin
	cloudProvider, err := s.pluginManager.GetProvider(provider)
	if err != nil {
		return "", fmt.Errorf("failed to get provider %s: %w", provider, err)
	}

	// Get instance status using plugin
	status, err := cloudProvider.GetInstanceStatus(ctx, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get instance status: %w", err)
	}

	return status, nil
}

// Helper functions for type conversion
func convertToStringMap(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			result[k] = str
		}
	}
	return result
}

func convertToInterfaceMap(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}
