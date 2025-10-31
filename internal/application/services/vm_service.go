package service

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"skyclust/pkg/logger"
)

// VMService implements the VMService interface
type VMService struct {
	vmRepo        domain.VMRepository
	workspaceRepo domain.WorkspaceRepository
	cloudProvider CloudProviderService
	eventService  domain.EventService
	auditLogRepo  domain.AuditLogRepository
}

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

// NewVMService creates a new VMService
func NewVMService(
	vmRepo domain.VMRepository,
	workspaceRepo domain.WorkspaceRepository,
	cloudProvider CloudProviderService,
	eventService domain.EventService,
	auditLogRepo domain.AuditLogRepository,
) *VMService {
	return &VMService{
		vmRepo:        vmRepo,
		workspaceRepo: workspaceRepo,
		cloudProvider: cloudProvider,
		eventService:  eventService,
		auditLogRepo:  auditLogRepo,
	}
}

// CreateVM creates a new virtual machine
func (s *VMService) CreateVM(ctx context.Context, req domain.CreateVMRequest) (*domain.VM, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// Check if VM name already exists in workspace
	existingVMs, err := s.vmRepo.GetByWorkspaceID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing VMs: %w", err)
	}

	for _, vm := range existingVMs {
		if vm.Name == req.Name {
			return nil, domain.ErrVMAlreadyExists
		}
	}

	// Create cloud instance
	cloudReq := CreateInstanceRequest{
		Name:     req.Name,
		Type:     req.Type,
		Region:   req.Region,
		ImageID:  req.ImageID,
		Metadata: req.Metadata,
	}

	cloudInstance, err := s.cloudProvider.CreateInstance(ctx, req.Provider, cloudReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud instance: %w", err)
	}

	// Create VM record
	vm := &domain.VM{
		ID:          uuid.New().String(),
		Name:        req.Name,
		WorkspaceID: req.WorkspaceID,
		Provider:    req.Provider,
		InstanceID:  cloudInstance.ID,
		Status:      domain.VMStatus(cloudInstance.Status),
		Type:        req.Type,
		Region:      req.Region,
		ImageID:     req.ImageID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	if err := s.vmRepo.Create(ctx, vm); err != nil {
		// Rollback cloud instance creation
		if rollbackErr := s.cloudProvider.DeleteInstance(ctx, req.Provider, cloudInstance.ID); rollbackErr != nil {
			logger.Error(fmt.Sprintf("Failed to rollback cloud instance creation: %v", rollbackErr))
		}
		return nil, fmt.Errorf("failed to create VM record: %w", err)
	}

	// Publish event
	if s.eventService != nil {
		if err := s.eventService.Publish(ctx, domain.EventVMCreated, map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		}); err != nil {
			logger.Error(fmt.Sprintf("Failed to publish VM created event: %v", err))
		}
	}

	logger.Info(fmt.Sprintf("VM created successfully: %s (%s) - %s", vm.ID, vm.Name, vm.Provider))
	return vm, nil
}

// GetVM retrieves a VM by ID
func (s *VMService) GetVM(ctx context.Context, id string) (*domain.VM, error) {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return nil, domain.ErrVMNotFound
	}
	return vm, nil
}

// UpdateVM updates a VM
func (s *VMService) UpdateVM(ctx context.Context, id string, req domain.UpdateVMRequest) (*domain.VM, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing VM
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return nil, domain.ErrVMNotFound
	}

	// Update fields
	if req.Name != nil {
		vm.Name = *req.Name
	}
	if req.Type != nil {
		vm.Type = *req.Type
	}
	if req.Metadata != nil {
		vm.Metadata = req.Metadata
	}

	vm.UpdatedAt = time.Now()

	if err := s.vmRepo.Update(ctx, vm); err != nil {
		return nil, fmt.Errorf("failed to update VM: %w", err)
	}

	logger.Info(fmt.Sprintf("VM updated successfully: %s", vm.ID))
	return vm, nil
}

// DeleteVM deletes a VM
func (s *VMService) DeleteVM(ctx context.Context, id string) error {
	// Get VM
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	// Delete cloud instance
	if err := s.cloudProvider.DeleteInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete cloud instance: %v (provider: %s, instance: %s)", err, vm.Provider, vm.InstanceID))
		// Continue with VM record deletion even if cloud deletion fails
	}

	// Delete VM record
	if err := s.vmRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete VM record: %w", err)
	}

	// Publish event
	if s.eventService != nil {
		if err := s.eventService.Publish(ctx, domain.EventVMDeleted, map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		}); err != nil {
			logger.Error(fmt.Sprintf("Failed to publish VM deleted event: %v", err))
		}
	}

	logger.Info(fmt.Sprintf("VM deleted successfully: %s", vm.ID))
	return nil
}

// GetVMs lists VMs for a workspace
func (s *VMService) GetVMs(ctx context.Context, workspaceID string) ([]*domain.VM, error) {
	vms, err := s.vmRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", err)
	}
	return vms, nil
}

// StartVM starts a VM
func (s *VMService) StartVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	if vm.Status == domain.VMStatusRunning {
		return fmt.Errorf("VM is already running")
	}

	if err := s.cloudProvider.StartInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return fmt.Errorf("failed to start cloud instance: %w", err)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStarting); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	logger.Info(fmt.Sprintf("VM start initiated: %s", id))
	return nil
}

// StopVM stops a VM
func (s *VMService) StopVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	if vm.Status == domain.VMStatusStopped {
		return fmt.Errorf("VM is already stopped")
	}

	if err := s.cloudProvider.StopInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return fmt.Errorf("failed to stop cloud instance: %w", err)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStopping); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	logger.Info(fmt.Sprintf("VM stop initiated: %s", id))
	return nil
}

// RestartVM restarts a VM
func (s *VMService) RestartVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	// Stop first
	if err := s.cloudProvider.StopInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return fmt.Errorf("failed to stop cloud instance: %w", err)
	}

	// Wait a bit before starting
	time.Sleep(2 * time.Second)

	// Start
	if err := s.cloudProvider.StartInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return fmt.Errorf("failed to start cloud instance: %w", err)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStarting); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	logger.Info(fmt.Sprintf("VM restart initiated: %s", id))
	return nil
}

// GetVMStatus gets the current status of a VM
func (s *VMService) GetVMStatus(ctx context.Context, id string) (domain.VMStatus, error) {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get VM: %w", err)
	}
	if vm == nil {
		return "", domain.ErrVMNotFound
	}

	// Get current status from cloud provider
	status, err := s.cloudProvider.GetInstanceStatus(ctx, vm.Provider, vm.InstanceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get cloud instance status: %v (provider: %s, instance: %s)", err, vm.Provider, vm.InstanceID))
		return vm.Status, nil // Return cached status if cloud call fails
	}

	// Update status if it changed
	newStatus := domain.VMStatus(status)
	if newStatus != vm.Status {
		if err := s.vmRepo.UpdateStatus(ctx, id, newStatus); err != nil {
			logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
		}
	}

	return newStatus, nil
}
