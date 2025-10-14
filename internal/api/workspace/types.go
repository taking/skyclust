package workspace

import "time"

// CreateWorkspaceRequest represents a workspace creation request
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=500"`
}

// UpdateWorkspaceRequest represents a workspace update request
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// WorkspaceResponse represents a workspace in API responses
type WorkspaceResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkspaceListResponse represents a list of workspaces
type WorkspaceListResponse struct {
	Workspaces []*WorkspaceResponse `json:"workspaces"`
	Total      int64                `json:"total"`
}
