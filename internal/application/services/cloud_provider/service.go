package cloud_provider

import (
	"context"

	"skyclust/internal/domain"
	"skyclust/pkg/logger"
)

// CloudProviderService defines the interface for cloud provider operations
type CloudProviderService interface {
	CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error)
	GetInstance(ctx context.Context, provider, instanceID string) (*CloudInstance, error)
	DeleteInstance(ctx context.Context, provider, instanceID string) error
	StartInstance(ctx context.Context, provider, instanceID string) error
	StopInstance(ctx context.Context, provider, instanceID string) error
	GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error)
}

// CloudInstance represents a cloud instance
type CloudInstance struct {
	ID       string                 `json:"id"`
	Status   string                 `json:"status"`
	Type     string                 `json:"type"`
	Region   string                 `json:"region"`
	ImageID  string                 `json:"image_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateInstanceRequest represents a request to create a cloud instance
type CreateInstanceRequest struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Region   string                 `json:"region"`
	ImageID  string                 `json:"image_id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// cloudProviderService implements the CloudProviderService interface
// NOTE: This service is deprecated and will be replaced by gRPC-based provider services
type cloudProviderService struct {
	// TODO: Replace with gRPC client manager
}

// NewService creates a new cloud provider service
// DEPRECATED: Use gRPC-based provider services instead
func NewService() CloudProviderService {
	return &cloudProviderService{}
}

// CreateInstance creates a new cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*CloudInstance, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.CreateInstance is deprecated, use gRPC-based provider services")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}

// GetInstance retrieves a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) GetInstance(ctx context.Context, provider, instanceID string) (*CloudInstance, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.GetInstance is deprecated, use gRPC-based provider services")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}

// DeleteInstance deletes a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) DeleteInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.DeleteInstance is deprecated, use gRPC-based provider services")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}

// StartInstance starts a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) StartInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.StartInstance is deprecated, use gRPC-based provider services")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}

// StopInstance stops a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) StopInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("cloud_provider_service.StopInstance is deprecated, use gRPC-based provider services")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}

// GetInstanceStatus gets the status of a cloud instance
// DEPRECATED: Use gRPC-based provider services instead
func (s *cloudProviderService) GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error) {
	logger.DefaultLogger.Warn("cloud_provider_service.GetInstanceStatus is deprecated, use gRPC-based provider services")
		return "", domain.NewDomainError(domain.ErrCodeNotImplemented, "cloud_provider_service is deprecated, use gRPC-based provider services", 501)
}
