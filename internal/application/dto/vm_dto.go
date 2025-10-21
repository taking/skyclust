package dto

import (
	"time"
)

// VMDTO represents a VM data transfer object
type VMDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	WorkspaceID string            `json:"workspace_id"`
	Provider    string            `json:"provider"`
	Type        string            `json:"type"`
	Region      string            `json:"region"`
	ImageID     string            `json:"image_id"`
	Status      string            `json:"status"`
	InstanceID  string            `json:"instance_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateVMRequest represents a request to create a VM
type CreateVMRequest struct {
	Name        string            `json:"name" validate:"required,min=3,max=100"`
	WorkspaceID string            `json:"workspace_id" validate:"required"`
	Provider    string            `json:"provider" validate:"required"`
	Type        string            `json:"type" validate:"required"`
	Region      string            `json:"region" validate:"required"`
	ImageID     string            `json:"image_id" validate:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateVMRequest represents a request to update a VM
type UpdateVMRequest struct {
	Name     string            `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Type     string            `json:"type,omitempty"`
	Region   string            `json:"region,omitempty"`
	ImageID  string            `json:"image_id,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// VMOperationRequest represents a request to perform a VM operation
type VMOperationRequest struct {
	Operation string `json:"operation" validate:"required,oneof=start stop restart terminate"`
}
