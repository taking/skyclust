package workspace

import (
	"context"
	"fmt"
	"time"

	"cmp/pkg/database"
)

// Workspace represents a workspace in the system
type Workspace = database.Workspace

// WorkspaceUser represents a user in a workspace
type WorkspaceUser = database.WorkspaceUser

// Service defines the workspace service interface
type Service interface {
	// Workspace management
	CreateWorkspace(ctx context.Context, name, ownerID string) (*Workspace, error)
	GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error)
	ListWorkspaces(ctx context.Context, userID string) ([]*Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *Workspace) error
	DeleteWorkspace(ctx context.Context, workspaceID string) error

	// User management
	AddUser(ctx context.Context, workspaceID, userID, role string) error
	RemoveUser(ctx context.Context, workspaceID, userID string) error
	ListUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error)
	GetUserRole(ctx context.Context, workspaceID, userID string) (string, error)

	// Permission checking
	HasPermission(ctx context.Context, workspaceID, userID, permission string) (bool, error)
}

type service struct {
	db database.Service
}

// NewService creates a new workspace service
func NewService(db database.Service) Service {
	return &service{db: db}
}

// CreateWorkspace creates a new workspace
func (s *service) CreateWorkspace(ctx context.Context, name, ownerID string) (*Workspace, error) {
	workspace := &database.Workspace{
		ID:        generateID(),
		Name:      name,
		OwnerID:   ownerID,
		Settings:  make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.CreateWorkspace(ctx, workspace); err != nil {
		return nil, err
	}

	// Add owner as admin
	if err := s.AddUser(ctx, workspace.ID, ownerID, "admin"); err != nil {
		return nil, err
	}

	return (*Workspace)(workspace), nil
}

// GetWorkspace retrieves a workspace by ID
func (s *service) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	return s.db.GetWorkspace(ctx, workspaceID)
}

// ListWorkspaces lists all workspaces for a user
func (s *service) ListWorkspaces(ctx context.Context, userID string) ([]*Workspace, error) {
	return s.db.ListWorkspacesByUser(ctx, userID)
}

// UpdateWorkspace updates a workspace
func (s *service) UpdateWorkspace(ctx context.Context, workspace *Workspace) error {
	workspace.UpdatedAt = time.Now()
	return s.db.UpdateWorkspace(ctx, workspace)
}

// DeleteWorkspace deletes a workspace
func (s *service) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return s.db.DeleteWorkspace(ctx, workspaceID)
}

// AddUser adds a user to a workspace
func (s *service) AddUser(ctx context.Context, workspaceID, userID, role string) error {
	return s.db.AddUserToWorkspace(ctx, userID, workspaceID, role)
}

// RemoveUser removes a user from a workspace
func (s *service) RemoveUser(ctx context.Context, workspaceID, userID string) error {
	return s.db.RemoveUserFromWorkspace(ctx, userID, workspaceID)
}

// ListUsers lists all users in a workspace
func (s *service) ListUsers(ctx context.Context, workspaceID string) ([]*WorkspaceUser, error) {
	return s.db.GetWorkspaceUsers(ctx, workspaceID)
}

// GetUserRole gets the role of a user in a workspace
func (s *service) GetUserRole(ctx context.Context, workspaceID, userID string) (string, error) {
	users, err := s.ListUsers(ctx, workspaceID)
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.UserID == userID {
			return user.Role, nil
		}
	}

	return "", fmt.Errorf("user not found in workspace")
}

// HasPermission checks if a user has a specific permission in a workspace
func (s *service) HasPermission(ctx context.Context, workspaceID, userID, permission string) (bool, error) {
	role, err := s.GetUserRole(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	// Simple role-based permission system
	permissions := map[string][]string{
		"admin":  {"read", "write", "delete", "manage_users"},
		"member": {"read", "write"},
		"viewer": {"read"},
	}

	userPermissions, exists := permissions[role]
	if !exists {
		return false, nil
	}

	for _, perm := range userPermissions {
		if perm == permission {
			return true, nil
		}
	}

	return false, nil
}

// generateID generates a random ID
func generateID() string {
	// This is a simplified implementation
	// In production, use a proper UUID generator
	return fmt.Sprintf("ws_%d", time.Now().UnixNano())
}
