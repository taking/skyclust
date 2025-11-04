package domain

import (
	"context"
)

// WorkspaceService defines the business logic interface for workspaces
type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, req CreateWorkspaceRequest) (*Workspace, error)
	GetWorkspace(ctx context.Context, id string) (*Workspace, error)
	UpdateWorkspace(ctx context.Context, id string, req UpdateWorkspaceRequest) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	GetUserWorkspaces(ctx context.Context, userID string) ([]*Workspace, error)
	AddUserToWorkspace(ctx context.Context, workspaceID, userID string) error
	RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error
	GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*User, error)
}

