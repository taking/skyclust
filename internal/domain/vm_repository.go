package domain

import (
	"context"
)

// VMRepository defines the interface for VM data operations
type VMRepository interface {
	Create(ctx context.Context, vm *VM) error
	GetByID(ctx context.Context, id string) (*VM, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*VM, error)
	GetVMsByWorkspace(ctx context.Context, workspaceID string) ([]*VM, error)
	GetByProvider(ctx context.Context, provider string) ([]*VM, error)
	Update(ctx context.Context, vm *VM) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, workspaceID string, limit, offset int) ([]*VM, error)
	UpdateStatus(ctx context.Context, id string, status VMStatus) error
}

