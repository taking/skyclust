package domain

import (
	"context"
)

// WorkspaceRepository defines the interface for workspace data operations
type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id string) (*Workspace, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Workspace, error)
	GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error)
	AddUserToWorkspace(ctx context.Context, userID, workspaceID string, role string) error
	RemoveUserFromWorkspace(ctx context.Context, userID, workspaceID string) error
	// Optimized methods for N+1 query prevention
	GetWorkspacesWithUsers(ctx context.Context, userID string) ([]*Workspace, error)
	GetWorkspaceWithMembers(ctx context.Context, workspaceID string) (*Workspace, []*User, error)
	GetUserWorkspacesOptimized(ctx context.Context, userID string) ([]*Workspace, error)
}

