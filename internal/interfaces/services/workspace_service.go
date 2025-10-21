package services

import (
	"context"
	"skyclust/internal/domain"

	"github.com/google/uuid"
)

// WorkspaceService defines the interface for workspace business operations
type WorkspaceService interface {
	// Workspace management
	CreateWorkspace(ctx context.Context, req domain.CreateWorkspaceRequest) (*domain.Workspace, error)
	GetWorkspace(ctx context.Context, id string) (*domain.Workspace, error)
	UpdateWorkspace(ctx context.Context, id string, req domain.UpdateWorkspaceRequest) (*domain.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	ListWorkspaces(ctx context.Context, limit, offset int) ([]*domain.Workspace, error)

	// User workspace operations
	GetUserWorkspaces(ctx context.Context, userID string) ([]*domain.Workspace, error)
	GetWorkspaceByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.Workspace, error)

	// Workspace status
	ActivateWorkspace(ctx context.Context, id string) error
	DeactivateWorkspace(ctx context.Context, id string) error

	// Access control
	CheckWorkspaceAccess(ctx context.Context, userID uuid.UUID, workspaceID string) (bool, error)
	CheckWorkspaceOwnership(ctx context.Context, userID uuid.UUID, workspaceID string) (bool, error)
}
