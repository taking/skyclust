package service

import (
	"context"
	"fmt"

	"skyclust/pkg/logger"
)

// cloudProviderService implements the CloudProviderService interface
// NOTE: This service is deprecated and will be replaced by gRPC-based provider services
type cloudProviderService struct {
	// TODO: Replace with gRPC client manager
}

// NewCloudProviderService creates a new cloud provider service
// DEPRECATED: Use gRPC-based provider services instead
func NewCloudProviderService() CloudProviderService {
	return &cloudProviderService{}
}

// CreateInstance creates a new cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.CreateInstance is deprecated, use gRPC-based provider services")
	return nil, fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}

// GetInstance retrieves a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) GetInstance(ctx context.Context, provider, instanceID string) (*CloudInstance, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.GetInstance is deprecated, use gRPC-based provider services")
	return nil, fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}

// DeleteInstance deletes a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) DeleteInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.DeleteInstance is deprecated, use gRPC-based provider services")
	return fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}

// StartInstance starts a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) StartInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.StartInstance is deprecated, use gRPC-based provider services")
	return fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}

// StopInstance stops a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) StopInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.StopInstance is deprecated, use gRPC-based provider services")
	return fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}

// GetInstanceStatus gets the status of a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.GetInstanceStatus is deprecated, use gRPC-based provider services")
	return "", fmt.Errorf("cloud_provider_service is deprecated, use gRPC-based provider services")
}
