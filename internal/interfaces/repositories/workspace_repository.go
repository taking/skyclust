package repositories

import (
	"context"
	"skyclust/internal/domain"
)

// WorkspaceRepository defines the interface for workspace data operations
type WorkspaceRepository interface {
	// Basic CRUD operations
	Create(workspace *domain.Workspace) error
	GetByID(ctx context.Context, id string) (*domain.Workspace, error)
	Update(workspace *domain.Workspace) error
	Delete(ctx context.Context, id string) error

	// List operations
	GetByOwnerID(ctx context.Context, ownerID string) ([]*domain.Workspace, error)
	List(limit, offset int) ([]*domain.Workspace, error)
	Count() (int64, error)

	// Search operations
	Search(query string, limit, offset int) ([]*domain.Workspace, error)
	GetByOwner(ownerID string, limit, offset int) ([]*domain.Workspace, error)

	// Status operations
	Activate(id string) error
	Deactivate(id string) error
	IsActive(id string) (bool, error)
}
