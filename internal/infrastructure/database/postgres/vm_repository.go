package postgres

import (
	"context"
	"fmt"
	"skyclust/internal/domain"

	"gorm.io/gorm"
	"skyclust/pkg/logger"
)

// VMRepository: domain.VMRepository 인터페이스 구현체
type VMRepository struct {
	db *gorm.DB
}

// NewVMRepository: 새로운 VMRepository를 생성합니다
func NewVMRepository(db *gorm.DB) *VMRepository {
	return &VMRepository{db: db}
}

// Create: 새로운 VM을 생성합니다
func (r *VMRepository) Create(ctx context.Context, vm *domain.VM) error {
	result := r.db.WithContext(ctx).Create(vm)
	if result.Error != nil {
		return fmt.Errorf("failed to create VM: %w", result.Error)
	}

	logger.Info(fmt.Sprintf("VM created in database: %s (%s) - %s", vm.ID, vm.Name, vm.Provider))
	return nil
}

// GetByID: ID로 VM을 조회합니다
func (r *VMRepository) GetByID(ctx context.Context, id string) (*domain.VM, error) {
	var vm domain.VM
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&vm)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get VM by ID: %w", result.Error)
	}

	return &vm, nil
}

// GetByWorkspaceID: 워크스페이스 ID로 VM 목록을 조회합니다
func (r *VMRepository) GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.VM, error) {
	var vms []*domain.VM
	result := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Find(&vms)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get VMs by workspace ID: %w", result.Error)
	}

	return vms, nil
}

// GetVMsByWorkspace: 워크스페이스 ID로 VM 목록을 조회합니다 (GetByWorkspaceID의 별칭)
func (r *VMRepository) GetVMsByWorkspace(ctx context.Context, workspaceID string) ([]*domain.VM, error) {
	return r.GetByWorkspaceID(ctx, workspaceID)
}

// GetByProvider: 프로바이더로 VM 목록을 조회합니다
func (r *VMRepository) GetByProvider(ctx context.Context, provider string) ([]*domain.VM, error) {
	var vms []*domain.VM
	result := r.db.WithContext(ctx).
		Where("provider = ?", provider).
		Order("created_at DESC").
		Find(&vms)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get VMs by provider: %w", result.Error)
	}

	return vms, nil
}

// Update: VM 정보를 업데이트합니다
func (r *VMRepository) Update(ctx context.Context, vm *domain.VM) error {
	result := r.db.WithContext(ctx).Save(vm)
	if result.Error != nil {
		return fmt.Errorf("failed to update VM: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("VM not found")
	}

	logger.Info(fmt.Sprintf("VM updated in database: %s", vm.ID))
	return nil
}

// Delete: VM을 삭제합니다
func (r *VMRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.VM{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete VM: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("VM not found")
	}

	logger.Info(fmt.Sprintf("VM deleted from database: %s", id))
	return nil
}

// List: 페이지네이션을 포함한 VM 목록을 조회합니다
func (r *VMRepository) List(ctx context.Context, workspaceID string, limit, offset int) ([]*domain.VM, error) {
	var vms []*domain.VM
	result := r.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&vms)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to list VMs: %w", result.Error)
	}

	return vms, nil
}

// UpdateStatus: VM 상태를 업데이트합니다
func (r *VMRepository) UpdateStatus(ctx context.Context, id string, status domain.VMStatus) error {
	result := r.db.WithContext(ctx).
		Model(&domain.VM{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update VM status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("VM not found")
	}

	logger.Info(fmt.Sprintf("VM status updated in database: %s -> %s", id, status))
	return nil
}
