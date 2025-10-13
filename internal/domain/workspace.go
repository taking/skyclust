package domain

import (
	"context"
	"time"
)

// Workspace represents a workspace in the system
type Workspace struct {
	ID          string                 `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string                 `json:"name" gorm:"uniqueIndex;not null"`
	Description string                 `json:"description" gorm:"type:text"`
	OwnerID     string                 `json:"owner_id" gorm:"not null;type:uuid"`
	Settings    map[string]interface{} `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time             `json:"-" gorm:"index"`
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
}

// TableName specifies the table name for Workspace
func (Workspace) TableName() string {
	return "workspaces"
}

// WorkspaceUser represents a user in a workspace
type WorkspaceUser struct {
	UserID      string     `json:"user_id" gorm:"primaryKey;type:uuid"`
	WorkspaceID string     `json:"workspace_id" gorm:"primaryKey;type:uuid"`
	Role        string     `json:"role" gorm:"not null;default:member"`
	JoinedAt    time.Time  `json:"joined_at" gorm:"autoCreateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

// TableName specifies the table name for WorkspaceUser
func (WorkspaceUser) TableName() string {
	return "workspace_users"
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
	// Optimized methods for N+1 query prevention
	GetWorkspacesWithUsers(ctx context.Context, userID string) ([]*Workspace, error)
	GetWorkspaceWithMembers(ctx context.Context, workspaceID string) (*Workspace, []*User, error)
	GetUserWorkspacesOptimized(ctx context.Context, userID string) ([]*Workspace, error)
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
