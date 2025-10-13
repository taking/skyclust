package domain

import (
	"context"
	"time"
)

// VMStatus represents the status of a virtual machine
type VMStatus string

const (
	VMStatusPending    VMStatus = "pending"
	VMStatusRunning    VMStatus = "running"
	VMStatusStopped    VMStatus = "stopped"
	VMStatusStopping   VMStatus = "stopping"
	VMStatusStarting   VMStatus = "starting"
	VMStatusTerminated VMStatus = "terminated"
	VMStatusError      VMStatus = "error"
)

// VM represents a virtual machine in the system
type VM struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	WorkspaceID string                 `json:"workspace_id" db:"workspace_id"`
	Provider    string                 `json:"provider" db:"provider"`
	InstanceID  string                 `json:"instance_id" db:"instance_id"`
	Status      VMStatus               `json:"status" db:"status"`
	Type        string                 `json:"type" db:"type"`
	Region      string                 `json:"region" db:"region"`
	ImageID     string                 `json:"image_id" db:"image_id"`
	CPUs        int                    `json:"cpus" db:"cpus"`
	Memory      int                    `json:"memory" db:"memory"`   // in MB
	Storage     int                    `json:"storage" db:"storage"` // in GB
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// VMRepository defines the interface for VM data operations
type VMRepository interface {
	Create(ctx context.Context, vm *VM) error
	GetByID(ctx context.Context, id string) (*VM, error)
	GetByWorkspaceID(ctx context.Context, workspaceID string) ([]*VM, error)
	GetVMsByWorkspace(ctx context.Context, workspaceID string) ([]*VM, error)
	GetByProvider(ctx context.Context, provider string) ([]*VM, error)
	Update(ctx context.Context, vm *VM) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, workspaceID string, limit, offset int) ([]*VM, error)
	UpdateStatus(ctx context.Context, id string, status VMStatus) error
}

// VMService defines the business logic interface for VMs
type VMService interface {
	CreateVM(ctx context.Context, req CreateVMRequest) (*VM, error)
	GetVM(ctx context.Context, id string) (*VM, error)
	UpdateVM(ctx context.Context, id string, req UpdateVMRequest) (*VM, error)
	DeleteVM(ctx context.Context, id string) error
	GetVMs(ctx context.Context, workspaceID string) ([]*VM, error)
	StartVM(ctx context.Context, id string) error
	StopVM(ctx context.Context, id string) error
	RestartVM(ctx context.Context, id string) error
	GetVMStatus(ctx context.Context, id string) (VMStatus, error)
}

// CreateVMRequest represents the request to create a new VM
type CreateVMRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=100"`
	WorkspaceID string                 `json:"workspace_id" validate:"required"`
	Provider    string                 `json:"provider" validate:"required"`
	Type        string                 `json:"type" validate:"required"`
	Region      string                 `json:"region" validate:"required"`
	ImageID     string                 `json:"image_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateVMRequest represents the request to update a VM
type UpdateVMRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Type     *string                `json:"type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validate performs validation on the CreateVMRequest
func (r *CreateVMRequest) Validate() error {
	if len(r.Name) < 3 || len(r.Name) > 100 {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if len(r.WorkspaceID) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "workspace_id is required", 400)
	}
	if len(r.Provider) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "provider is required", 400)
	}
	if len(r.Type) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "type is required", 400)
	}
	if len(r.Region) == 0 {
		return NewDomainError(ErrCodeValidationFailed, "region is required", 400)
	}
	return nil
}

// Validate performs validation on the UpdateVMRequest
func (r *UpdateVMRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 3 || len(*r.Name) > 100) {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	return nil
}
