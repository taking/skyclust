package services

import (
	"context"
	"skyclust/internal/domain"
)

// VMService defines the interface for VM business operations
type VMService interface {
	// VM management
	CreateVM(ctx context.Context, req domain.CreateVMRequest) (*domain.VM, error)
	GetVM(ctx context.Context, id string) (*domain.VM, error)
	UpdateVM(ctx context.Context, id string, req domain.UpdateVMRequest) (*domain.VM, error)
	DeleteVM(ctx context.Context, id string) error
	ListVMs(ctx context.Context, limit, offset int) ([]*domain.VM, error)

	// Workspace VM operations
	GetWorkspaceVMs(ctx context.Context, workspaceID string) ([]*domain.VM, error)
	GetVMsByStatus(ctx context.Context, status domain.VMStatus, limit, offset int) ([]*domain.VM, error)

	// VM operations
	StartVM(ctx context.Context, id string) error
	StopVM(ctx context.Context, id string) error
	RestartVM(ctx context.Context, id string) error
	TerminateVM(ctx context.Context, id string) error

	// Status management
	UpdateVMStatus(ctx context.Context, id string, status domain.VMStatus) error
	GetVMStatus(ctx context.Context, id string) (domain.VMStatus, error)

	// Access control
	CheckVMAccess(ctx context.Context, userID string, vmID string) (bool, error)
}
