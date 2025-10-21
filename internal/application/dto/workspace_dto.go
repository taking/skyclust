package dto

import (
	"time"
)

// WorkspaceDTO represents a workspace data transfer object
type WorkspaceDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	OwnerID     string            `json:"owner_id"`
	Active      bool              `json:"is_active"`
	Settings    map[string]string `json:"settings,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateWorkspaceRequest represents a request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string            `json:"name" validate:"required,min=3,max=100"`
	Description string            `json:"description,omitempty" validate:"omitempty,max=500"`
	Settings    map[string]string `json:"settings,omitempty"`
}

// UpdateWorkspaceRequest represents a request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        string            `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description string            `json:"description,omitempty" validate:"omitempty,max=500"`
	Settings    map[string]string `json:"settings,omitempty"`
	Active      *bool             `json:"is_active,omitempty"`
}
