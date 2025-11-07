package compute

import (
	"context"

	"skyclust/internal/domain"
	"skyclust/pkg/logger"
)

// ComputeService defines the interface for cloud compute operations
// Supports AWS EC2, GCP Compute Engine, Azure Compute Service, etc.
type ComputeService interface {
	CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*ComputeInstance, error)
	GetInstance(ctx context.Context, provider, instanceID string) (*ComputeInstance, error)
	DeleteInstance(ctx context.Context, provider, instanceID string) error
	StartInstance(ctx context.Context, provider, instanceID string) error
	StopInstance(ctx context.Context, provider, instanceID string) error
	GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error)
}

// computeService: ComputeService 인터페이스 구현체
// TODO: AWS EC2, GCP Compute Engine, Azure Compute Service 지원 구현 필요
type computeService struct {
	// TODO: 프로바이더별 클라이언트 추가 필요 (AWS SDK, GCP Client, Azure SDK)
}

// NewService: 새로운 컴퓨트 서비스를 생성합니다
// TODO: 프로바이더별 구현 추가 필요 (AWS EC2, GCP Compute, Azure Compute)
func NewService() ComputeService {
	return &computeService{}
}

// CreateInstance: 새로운 컴퓨트 인스턴스를 생성합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) CreateInstance(ctx context.Context, provider string, req CreateInstanceRequest) (*ComputeInstance, error) {
	logger.DefaultLogger.Warn("compute_service.CreateInstance is not yet implemented, provider-specific implementations pending")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}

// GetInstance: 컴퓨트 인스턴스를 조회합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) GetInstance(ctx context.Context, provider, instanceID string) (*ComputeInstance, error) {
	logger.DefaultLogger.Warn("compute_service.GetInstance is not yet implemented, provider-specific implementations pending")
	return nil, domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}

// DeleteInstance: 컴퓨트 인스턴스를 삭제합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) DeleteInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("compute_service.DeleteInstance is not yet implemented, provider-specific implementations pending")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}

// StartInstance: 컴퓨트 인스턴스를 시작합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) StartInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("compute_service.StartInstance is not yet implemented, provider-specific implementations pending")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}

// StopInstance: 컴퓨트 인스턴스를 중지합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) StopInstance(ctx context.Context, provider, instanceID string) error {
	logger.DefaultLogger.Warn("compute_service.StopInstance is not yet implemented, provider-specific implementations pending")
	return domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}

// GetInstanceStatus: 컴퓨트 인스턴스의 상태를 조회합니다
// TODO: 프로바이더별 로직 구현 필요 (AWS EC2, GCP Compute Engine, Azure Compute Service)
func (s *computeService) GetInstanceStatus(ctx context.Context, provider, instanceID string) (string, error) {
	logger.DefaultLogger.Warn("compute_service.GetInstanceStatus is not yet implemented, provider-specific implementations pending")
	return "", domain.NewDomainError(domain.ErrCodeNotImplemented, "compute service is not yet implemented, AWS EC2, GCP Compute, Azure Compute support pending", 501)
}
