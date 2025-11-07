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
	AddMemberByEmail(ctx context.Context, workspaceID, email, role string) error
	RemoveUserFromWorkspace(ctx context.Context, workspaceID, userID string) error
	GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]*User, error)
	GetWorkspaceMembersWithRoles(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error)
	UpdateMemberRole(ctx context.Context, workspaceID, userID, role string) error
}
