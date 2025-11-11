package vm

import (
	"context"
	"fmt"
	"skyclust/internal/application/services/common"
	computeservice "skyclust/internal/application/services/compute"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
)

// Service: domain.VMService 인터페이스 구현체
type Service struct {
	vmRepo          domain.VMRepository
	vmDomainService *domain.VMDomainService
	workspaceRepo   domain.WorkspaceRepository
	computeService  computeservice.ComputeService
	eventService    domain.EventService
	auditLogRepo    domain.AuditLogRepository
	cacheService    domain.CacheService
	logger          domain.LoggerService
}

// NewService: 새로운 VMService를 생성합니다
func NewService(
	vmRepo domain.VMRepository,
	vmDomainService *domain.VMDomainService,
	workspaceRepo domain.WorkspaceRepository,
	computeService computeservice.ComputeService,
	eventService domain.EventService,
	auditLogRepo domain.AuditLogRepository,
	cacheService domain.CacheService,
	logger domain.LoggerService,
) *Service {
	return &Service{
		vmRepo:          vmRepo,
		vmDomainService: vmDomainService,
		workspaceRepo:   workspaceRepo,
		computeService:  computeService,
		eventService:    eventService,
		auditLogRepo:    auditLogRepo,
		cacheService:    cacheService,
		logger:          logger,
	}
}

// CreateVM: 새로운 가상 머신을 생성합니다
func (s *Service) CreateVM(ctx context.Context, req domain.CreateVMRequest) (*domain.VM, error) {
	// Extract user ID from context (application-level concern)
	userIDValue := ctx.Value("user_id")
	if userIDValue == nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "user ID not found in context", 401)
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		if userIDStr, ok := userIDValue.(string); ok {
			var err error
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid user ID format", 401)
			}
		} else {
			return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid user ID type", 401)
		}
	}

	// Use domain service to validate business rules and create VM entity
	vm, err := s.vmDomainService.CreateVM(ctx, req, userID)
	if err != nil {
		return nil, err
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

	// Update VM entity with compute instance details
	vm.InstanceID = computeInstance.ID
	vm.Status = domain.VMStatus(computeInstance.Status)

	if err := s.vmRepo.Create(ctx, vm); err != nil {
		// Rollback compute instance creation
		if rollbackErr := s.computeService.DeleteInstance(ctx, req.Provider, computeInstance.ID); rollbackErr != nil {
			s.logger.Error(ctx, "Failed to rollback compute instance creation", rollbackErr)
		}
		s.logger.Error(ctx, "Failed to create VM record", err)
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create VM record: %v", err), 500)
	}

	// 캐시 무효화: VM 목록 캐시 삭제
	if s.cacheService != nil {
		cacheKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, cacheKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(vm.Status),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		if err := s.eventService.Publish(ctx, domain.EventVMCreated, vmData); err != nil {
			s.logger.Error(ctx, "Failed to publish VM created event", err)
		}
	}

	// 감사로그 기록
	common.LogAction(ctx, s.auditLogRepo, nil, domain.ActionVMCreate,
		"POST /api/v1/vms",
		map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"region":       vm.Region,
			"type":         vm.Type,
		},
	)

	s.logger.Info(ctx, "VM created successfully",
		domain.NewLogField("vm_id", vm.ID),
		domain.NewLogField("name", vm.Name),
		domain.NewLogField("provider", vm.Provider))
	return vm, nil
}

// GetVM retrieves a VM by ID
func (s *Service) GetVM(ctx context.Context, id string) (*domain.VM, error) {
	// 캐시 키 생성
	cacheKey := buildVMItemKey(id)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVM, ok := cachedValue.(*domain.VM); ok {
				s.logger.Debug(ctx, "VM retrieved from cache",
					domain.NewLogField("vm_id", id))
				return cachedVM, nil
			} else if cachedVM, ok := cachedValue.(domain.VM); ok {
				s.logger.Debug(ctx, "VM retrieved from cache",
					domain.NewLogField("vm_id", id))
				return &cachedVM, nil
			}
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
	if s.cacheService != nil {
		if err := s.cacheService.Set(ctx, cacheKey, vm, defaultVMTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache VM, continuing without cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
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
		s.logger.Error(ctx, "Failed to update VM", err,
			domain.NewLogField("vm_id", id))
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update VM: %v", err), 500)
	}

	// 캐시 무효화
	if s.cacheService != nil {
		listKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
		itemKey := buildVMItemKey(id)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM item cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(vm.Status),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventService.Publish(ctx, domain.EventVMUpdated, vmData)
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

	s.logger.Info(ctx, "VM updated successfully",
		domain.NewLogField("vm_id", vm.ID))
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
		s.logger.Error(ctx, "Failed to delete compute instance", err,
			domain.NewLogField("provider", vm.Provider),
			domain.NewLogField("instance_id", vm.InstanceID))
		// Continue with VM record deletion even if compute deletion fails
	}

	// Delete VM record
	if err := s.vmRepo.Delete(ctx, id); err != nil {
		s.logger.Error(ctx, "Failed to delete VM record", err,
			domain.NewLogField("vm_id", id))
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete VM record: %v", err), 500)
	}

	// 캐시 무효화
	if s.cacheService != nil {
		listKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
		itemKey := buildVMItemKey(id)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM item cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(vm.Status),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		if err := s.eventService.Publish(ctx, domain.EventVMDeleted, vmData); err != nil {
			s.logger.Error(ctx, "Failed to publish VM deleted event", err)
		}
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

	s.logger.Info(ctx, "VM deleted successfully",
		domain.NewLogField("vm_id", vm.ID))
	return nil
}

// GetVMs: 워크스페이스의 VM 목록을 조회합니다
func (s *Service) GetVMs(ctx context.Context, workspaceID string) ([]*domain.VM, error) {
	// 캐시 키 생성
	cacheKey := buildVMListKey(workspaceID)

	// 캐시에서 조회 시도
	if s.cacheService != nil {
		cachedValue, err := s.cacheService.Get(ctx, cacheKey)
		if err == nil && cachedValue != nil {
			if cachedVMs, ok := cachedValue.([]*domain.VM); ok {
				s.logger.Debug(ctx, "VMs retrieved from cache",
					domain.NewLogField("workspace_id", workspaceID))
				return cachedVMs, nil
			}
		}
	}

	// 캐시 미스 시 실제 DB 조회
	vms, err := s.vmRepo.GetByWorkspaceID(ctx, workspaceID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to list VMs: %v", err), 500)
	}

	// 응답을 캐시에 저장 (캐시 실패해도 계속 진행)
	if s.cacheService != nil && vms != nil {
		if err := s.cacheService.Set(ctx, cacheKey, vms, defaultVMTTL); err != nil {
			s.logger.Warn(ctx, "Failed to cache VMs, continuing without cache",
				domain.NewLogField("workspace_id", workspaceID),
				domain.NewLogField("error", err))
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
		s.logger.Error(ctx, "Failed to update VM status", err,
			domain.NewLogField("vm_id", id))
	}

	// 캐시 무효화
	if s.cacheService != nil {
		listKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
		itemKey := buildVMItemKey(id)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM item cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(domain.VMStatusStarting),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventService.Publish(ctx, domain.EventVMStarted, vmData)
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

	s.logger.Info(ctx, "VM start initiated",
		domain.NewLogField("vm_id", id))
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
		s.logger.Error(ctx, "Failed to update VM status", err,
			domain.NewLogField("vm_id", id))
	}

	// 캐시 무효화
	if s.cacheService != nil {
		listKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
		itemKey := buildVMItemKey(id)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM item cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(domain.VMStatusStopping),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventService.Publish(ctx, domain.EventVMStopped, vmData)
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

	s.logger.Info(ctx, "VM stop initiated",
		domain.NewLogField("vm_id", id))
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
		s.logger.Error(ctx, "Failed to update VM status", err,
			domain.NewLogField("vm_id", id))
	}

	// 캐시 무효화
	if s.cacheService != nil {
		listKey := buildVMListKey(vm.WorkspaceID)
		if err := s.cacheService.Delete(ctx, listKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM list cache",
				domain.NewLogField("workspace_id", vm.WorkspaceID),
				domain.NewLogField("error", err))
		}
		itemKey := buildVMItemKey(id)
		if err := s.cacheService.Delete(ctx, itemKey); err != nil {
			s.logger.Warn(ctx, "Failed to invalidate VM item cache",
				domain.NewLogField("vm_id", id),
				domain.NewLogField("error", err))
		}
	}

	// 이벤트 발행
	if s.eventService != nil {
		vmData := map[string]interface{}{
			"vm_id":        vm.ID,
			"workspace_id": vm.WorkspaceID,
			"provider":     vm.Provider,
			"name":         vm.Name,
			"status":       string(domain.VMStatusStarting),
			"region":       vm.Region,
			"type":         vm.Type,
		}
		_ = s.eventService.Publish(ctx, domain.EventVMStarted, vmData)
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

	s.logger.Info(ctx, "VM restart initiated",
		domain.NewLogField("vm_id", id))
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
		s.logger.Error(ctx, "Failed to get compute instance status", err,
			domain.NewLogField("provider", vm.Provider),
			domain.NewLogField("instance_id", vm.InstanceID))
		return vm.Status, nil // Return cached status if compute call fails
	}

	// Update status if it changed
	newStatus := domain.VMStatus(status)
	if newStatus != vm.Status {
		if err := s.vmRepo.UpdateStatus(ctx, id, newStatus); err != nil {
			s.logger.Error(ctx, "Failed to update VM status", err,
				domain.NewLogField("vm_id", id))
		}
	}

	return newStatus, nil
}
