package repositories

import (
	"context"
	"skyclust/internal/domain"
)

// VMRepository defines the interface for VM data operations
type VMRepository interface {
	// Basic CRUD operations
	Create(vm *domain.VM) error
	GetByID(ctx context.Context, id string) (*domain.VM, error)
	Update(vm *domain.VM) error
	Delete(ctx context.Context, id string) error

	// List operations
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*domain.VM, error)
	List(limit, offset int) ([]*domain.VM, error)
	Count() (int64, error)

	// Search operations
	Search(query string, limit, offset int) ([]*domain.VM, error)
	GetByStatus(status domain.VMStatus, limit, offset int) ([]*domain.VM, error)
	GetByProvider(provider string, limit, offset int) ([]*domain.VM, error)

	// Status operations
	UpdateStatus(id string, status domain.VMStatus) error
	GetStatus(id string) (domain.VMStatus, error)
}
