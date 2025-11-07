package vm

import (
	"context"
	"fmt"
	computeservice "skyclust/internal/application/services/compute"
	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/internal/infrastructure/messaging"
	"skyclust/pkg/cache"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"skyclust/pkg/logger"
)

// Service: domain.VMService 인터페이스 구현체
type Service struct {
	vmRepo        domain.VMRepository
	workspaceRepo domain.WorkspaceRepository
	computeService computeservice.ComputeService
	eventService  domain.EventService
	auditLogRepo  domain.AuditLogRepository
	cache         cache.Cache
	keyBuilder    *cache.CacheKeyBuilder
	invalidator   *cache.Invalidator
	eventPublisher *messaging.Publisher
	logger        *zap.Logger
}

// NewService: 새로운 VMService를 생성합니다
func NewService(
	vmRepo domain.VMRepository,
	workspaceRepo domain.WorkspaceRepository,
	computeService computeservice.ComputeService,
	eventService domain.EventService,
	auditLogRepo domain.AuditLogRepository,
	cache cache.Cache,
	keyBuilder *cache.CacheKeyBuilder,
	invalidator *cache.Invalidator,
	eventPublisher *messaging.Publisher,
	logger *zap.Logger,
) *Service {
	return &Service{
		vmRepo:         vmRepo,
		workspaceRepo:  workspaceRepo,
		computeService: computeService,
		eventService:   eventService,
		auditLogRepo:   auditLogRepo,
		cache:          cache,
		keyBuilder:     keyBuilder,
		invalidator:    invalidator,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// CreateVM: 새로운 가상 머신을 생성합니다
func (s *Service) CreateVM(ctx context.Context, req domain.CreateVMRequest) (*domain.VM, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err // req.Validate() already returns domain.NewDomainError
	}

	// Check if workspace exists
	workspace, err := s.workspaceRepo.GetByID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get workspace: %v", err), 500)
	}
	if workspace == nil {
		return nil, domain.ErrWorkspaceNotFound
	}

	// Check if VM name already exists in workspace
	existingVMs, err := s.vmRepo.GetByWorkspaceID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check existing VMs: %v", err), 500)
	}

	for _, vm := range existingVMs {
		if vm.Name == req.Name {
			return nil, domain.ErrVMAlreadyExists
		}
	}

	// Create compute instance
	computeReq := computeservice.CreateInstanceRequest{
		Name:     req.Name,
		Type:     req.Type,
		Region:   req.Region,
		ImageID:  req.ImageID,
		Metadata: req.Metadata,
	}

	computeInstance, err := s.computeService.CreateInstance(ctx, req.Provider, computeReq)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to create compute instance: %v", err), 502)
	}

	// Create VM record
	vm := &domain.VM{
		ID:          uuid.New().String(),
		Name:        req.Name,
		WorkspaceID: req.WorkspaceID,
		Provider:    req.Provider,
		InstanceID:  computeInstance.ID,
		Status:      domain.VMStatus(computeInstance.Status),
		Type:        req.Type,
		Region:      req.Region,
		ImageID:     req.ImageID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	if err := s.vmRepo.Create(ctx, vm); err != nil {
		// Rollback compute instance creation
		if rollbackErr := s.computeService.DeleteInstance(ctx, req.Provider, computeInstance.ID); rollbackErr != nil {
			logger.Error(fmt.Sprintf("Failed to rollback compute instance creation: %v", rollbackErr))
		}
		logger.Error(fmt.Sprintf("Failed to create VM record: %v", err))
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create VM record: %v", err), 500)
	}

	// 캐시 무효화: VM 목록 캐시 삭제
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			} else {
				logger.Warn(fmt.Sprintf("Failed to invalidate VM list cache: %v", err))
			}
		}
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

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       vm.Status,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "created", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMCreate,
		fmt.Sprintf("POST /api/v1/vms"),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"region":       vm.Region,
			"type":         vm.Type,
		},
	)

	logger.Info(fmt.Sprintf("VM created successfully: %s (%s) - %s", vm.ID, vm.Name, vm.Provider))
	return vm, nil
}

// GetVM retrieves a VM by ID
func (s *Service) GetVM(ctx context.Context, id string) (*domain.VM, error) {
	// 캐시 키 생성
	cacheKey := s.keyBuilder.BuildVMItemKey(id)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedVM domain.VM
		if err := s.cache.Get(ctx, cacheKey, &cachedVM); err == nil {
			if s.logger != nil {
				s.logger.Debug("VM retrieved from cache",
					zap.String("vm_id", id))
			}
			return &cachedVM, nil
		}
	}

	// 캐시 미스 시 실제 DB 조회
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return nil, domain.ErrVMNotFound
	}

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cache != nil && vm != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceVM)
		if err := s.cache.Set(ctx, cacheKey, vm, ttl); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to cache VM, continuing without cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
	}

	return vm, nil
}

// UpdateVM: VM 정보를 업데이트합니다
func (s *Service) UpdateVM(ctx context.Context, id string, req domain.UpdateVMRequest) (*domain.VM, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err // req.Validate() already returns domain.NewDomainError
	}

	// Get existing VM
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
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
		logger.Error(fmt.Sprintf("Failed to update VM: %v", err))
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update VM: %v", err), 500)
	}

	// 캐시 무효화
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			}
		}
		if err := s.invalidator.InvalidateByKey(ctx, s.keyBuilder.BuildVMItemKey(id)); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM item cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
	}

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       vm.Status,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "updated", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMUpdate,
		fmt.Sprintf("PUT /api/v1/vms/%s", id),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		},
	)

	logger.Info(fmt.Sprintf("VM updated successfully: %s", vm.ID))
	return vm, nil
}

// DeleteVM: VM을 삭제합니다
func (s *Service) DeleteVM(ctx context.Context, id string) error {
	// Get VM
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	// Delete compute instance
	if err := s.computeService.DeleteInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete compute instance: %v (provider: %s, instance: %s)", err, vm.Provider, vm.InstanceID))
		// Continue with VM record deletion even if compute deletion fails
	}

	// Delete VM record
	if err := s.vmRepo.Delete(ctx, id); err != nil {
		logger.Error(fmt.Sprintf("Failed to delete VM record: %v", err))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete VM record: %v", err), 500)
	}

	// 캐시 무효화
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			}
		}
		if err := s.invalidator.InvalidateByKey(ctx, s.keyBuilder.BuildVMItemKey(id)); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM item cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
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

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       vm.Status,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "deleted", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMDelete,
		fmt.Sprintf("DELETE /api/v1/vms/%s", id),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		},
	)

	logger.Info(fmt.Sprintf("VM deleted successfully: %s", vm.ID))
	return nil
}

// GetVMs: 워크스페이스의 VM 목록을 조회합니다
func (s *Service) GetVMs(ctx context.Context, workspaceID string) ([]*domain.VM, error) {
	// 캐시 키 생성
	cacheKey := s.keyBuilder.BuildVMListKey(workspaceID)

	// 캐시에서 조회 시도
	if s.cache != nil {
		var cachedVMs []*domain.VM
		if err := s.cache.Get(ctx, cacheKey, &cachedVMs); err == nil {
			if s.logger != nil {
				s.logger.Debug("VMs retrieved from cache",
					zap.String("workspace_id", workspaceID))
			}
			return cachedVMs, nil
		}
	}

	// 캐시 미스 시 실제 DB 조회
	vms, err := s.vmRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to list VMs: %v", err), 500)
	}

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cache != nil && vms != nil {
		ttl := cache.GetDefaultTTL(cache.ResourceVM)
		if err := s.cache.Set(ctx, cacheKey, vms, ttl); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to cache VMs, continuing without cache",
					zap.String("workspace_id", workspaceID),
					zap.Error(err))
			}
		}
	}

	return vms, nil
}

// StartVM: VM을 시작합니다
func (s *Service) StartVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	if vm.Status == domain.VMStatusRunning {
		return domain.NewDomainError(domain.ErrCodeConflict, "VM is already running", 409)
	}

	if err := s.computeService.StartInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to start compute instance: %v", err), 502)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStarting); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	// 캐시 무효화
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			}
		}
		if err := s.invalidator.InvalidateByKey(ctx, s.keyBuilder.BuildVMItemKey(id)); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM item cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
	}

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       domain.VMStatusStarting,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "updated", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMStart,
		fmt.Sprintf("POST /api/v1/vms/%s/start", id),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		},
	)

	logger.Info(fmt.Sprintf("VM start initiated: %s", id))
	return nil
}

// StopVM: VM을 중지합니다
func (s *Service) StopVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	if vm.Status == domain.VMStatusStopped {
		return domain.NewDomainError(domain.ErrCodeConflict, "VM is already stopped", 409)
	}

	if err := s.computeService.StopInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to stop compute instance: %v", err), 502)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStopping); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	// 캐시 무효화
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			}
		}
		if err := s.invalidator.InvalidateByKey(ctx, s.keyBuilder.BuildVMItemKey(id)); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM item cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
	}

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       domain.VMStatusStopping,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "updated", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMStop,
		fmt.Sprintf("POST /api/v1/vms/%s/stop", id),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		},
	)

	logger.Info(fmt.Sprintf("VM stop initiated: %s", id))
	return nil
}

// RestartVM: VM을 재시작합니다
func (s *Service) RestartVM(ctx context.Context, id string) error {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return domain.ErrVMNotFound
	}

	// Stop first
	if err := s.computeService.StopInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to stop compute instance: %v", err), 502)
	}

	// Wait a bit before starting
	time.Sleep(2 * time.Second)

	// Start
	if err := s.computeService.StartInstance(ctx, vm.Provider, vm.InstanceID); err != nil {
		return domain.NewDomainError(domain.ErrCodeProviderError, fmt.Sprintf("failed to start compute instance: %v", err), 502)
	}

	// Update status
	if err := s.vmRepo.UpdateStatus(ctx, id, domain.VMStatusStarting); err != nil {
		logger.Error(fmt.Sprintf("Failed to update VM status: %v (vm_id: %s)", err, id))
	}

	// 캐시 무효화
	if s.invalidator != nil {
		if err := s.invalidator.InvalidateVMList(ctx, vm.WorkspaceID); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM list cache",
					zap.String("workspace_id", vm.WorkspaceID),
					zap.Error(err))
			}
		}
		if err := s.invalidator.InvalidateByKey(ctx, s.keyBuilder.BuildVMItemKey(id)); err != nil {
			if s.logger != nil {
				s.logger.Warn("Failed to invalidate VM item cache",
					zap.String("vm_id", id),
					zap.Error(err))
			}
		}
	}

	// NATS 이벤트 발행
	if s.eventPublisher != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       domain.VMStatusStarting,
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventPublisher.PublishVMEvent(ctx, vm.Provider, vm.WorkspaceID, vm.ID, "updated", vmData)
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMRestart,
		fmt.Sprintf("POST /api/v1/vms/%s/restart", id),
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
		},
	)

	logger.Info(fmt.Sprintf("VM restart initiated: %s", id))
	return nil
}

// GetVMStatus: VM의 현재 상태를 조회합니다
func (s *Service) GetVMStatus(ctx context.Context, id string) (domain.VMStatus, error) {
	vm, err := s.vmRepo.GetByID(ctx, id)
	if err != nil {
		return "", domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get VM: %v", err), 500)
	}
	if vm == nil {
		return "", domain.ErrVMNotFound
	}

	// Get current status from compute service
	status, err := s.computeService.GetInstanceStatus(ctx, vm.Provider, vm.InstanceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get compute instance status: %v (provider: %s, instance: %s)", err, vm.Provider, vm.InstanceID))
		return vm.Status, nil // Return cached status if compute call fails
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
