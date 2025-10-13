package domain

import (
	"context"
	"time"
)

// Workspace represents a workspace in the system
type Workspace struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

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
}

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

// CreateWorkspaceRequest represents the request to create a new workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	OwnerID     string `json:"owner_id" validate:"required"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// Validate performs validation on the CreateWorkspaceRequest
func (r *CreateWorkspaceRequest) Validate() error {
	if len(r.Name) < 3 || len(r.Name) > 100 {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if len(r.Description) > 500 {
		return NewDomainError(ErrCodeValidationFailed, "description must be less than 500 characters", 400)
	}
	if len(r.OwnerID) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "owner_id is required", 400)
	}
	return nil
}

// Validate performs validation on the UpdateWorkspaceRequest
func (r *UpdateWorkspaceRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 3 || len(*r.Name) > 100) {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if r.Description != nil && len(*r.Description) > 500 {
		return NewDomainError(ErrCodeValidationFailed, "description must be less than 500 characters", 400)
	}
	return nil
}
