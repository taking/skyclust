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

// AddMemberRequest represents a request to add a member to a workspace
type AddMemberRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member"`
}

// UpdateMemberRoleRequest represents a request to update a member's role
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

// WorkspaceMemberResponse represents a workspace member in API responses
type WorkspaceMemberResponse struct {
	UserID      string    `json:"user_id"`
	WorkspaceID string    `json:"workspace_id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	User        struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	} `json:"user"`
}
