package domain

import (
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
	ID          string                 `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"not null;size:100"`
	WorkspaceID string                 `json:"workspace_id" gorm:"not null;index"`
	Provider    string                 `json:"provider" gorm:"not null;size:50;index"`
	InstanceID  string                 `json:"instance_id" gorm:"size:255;index"`
	Status      VMStatus               `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Type        string                 `json:"type" gorm:"size:50;not null"`
	Region      string                 `json:"region" gorm:"size:50;not null"`
	ImageID     string                 `json:"image_id" gorm:"size:255"`
	CPUs        int                    `json:"cpus" gorm:"not null"`
	Memory      int                    `json:"memory" gorm:"not null"`   // in MB
	Storage     int                    `json:"storage" gorm:"not null"` // in GB
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName specifies the table name for VM
func (VM) TableName() string {
	return "vms"
}

// Business methods for VM entity

// IsRunning checks if the VM is running
func (v *VM) IsRunning() bool {
	return v.Status == VMStatusRunning
}

// IsStopped checks if the VM is stopped
func (v *VM) IsStopped() bool {
	return v.Status == VMStatusStopped
}

// IsPending checks if the VM is pending
func (v *VM) IsPending() bool {
	return v.Status == VMStatusPending
}

// IsError checks if the VM is in error state
func (v *VM) IsError() bool {
	return v.Status == VMStatusError
}

// CanStart checks if the VM can be started
func (v *VM) CanStart() bool {
	return v.Status == VMStatusStopped || v.Status == VMStatusError
}

// CanStop checks if the VM can be stopped
func (v *VM) CanStop() bool {
	return v.Status == VMStatusRunning
}

// CanRestart checks if the VM can be restarted
func (v *VM) CanRestart() bool {
	return v.Status == VMStatusRunning || v.Status == VMStatusStopped
}

// CanTerminate checks if the VM can be terminated
func (v *VM) CanTerminate() bool {
	return v.Status != VMStatusTerminated
}

// Start changes VM status to starting
func (v *VM) Start() error {
	if !v.CanStart() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be started in current status", 400)
	}
	v.Status = VMStatusStarting
	v.UpdatedAt = time.Now()
	return nil
}

// Stop changes VM status to stopping
func (v *VM) Stop() error {
	if !v.CanStop() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be stopped in current status", 400)
	}
	v.Status = VMStatusStopping
	v.UpdatedAt = time.Now()
	return nil
}

// Restart changes VM status to starting
func (v *VM) Restart() error {
	if !v.CanRestart() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be restarted in current status", 400)
	}
	v.Status = VMStatusStarting
	v.UpdatedAt = time.Now()
	return nil
}

// Terminate changes VM status to terminated
func (v *VM) Terminate() error {
	if !v.CanTerminate() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be terminated in current status", 400)
	}
	v.Status = VMStatusTerminated
	v.UpdatedAt = time.Now()
	return nil
}

// SetStatus sets the VM status
func (v *VM) SetStatus(status VMStatus) {
	v.Status = status
	v.UpdatedAt = time.Now()
}

// SetInstanceID sets the cloud instance ID
func (v *VM) SetInstanceID(instanceID string) {
	v.InstanceID = instanceID
	v.UpdatedAt = time.Now()
}

// UpdateSpecs updates VM specifications
func (v *VM) UpdateSpecs(cpus, memory, storage int) error {
	if cpus <= 0 {
		return NewDomainError(ErrCodeValidationFailed, "CPU count must be positive", 400)
	}
	if memory <= 0 {
		return NewDomainError(ErrCodeValidationFailed, "Memory must be positive", 400)
	}
	if storage <= 0 {
		return NewDomainError(ErrCodeValidationFailed, "Storage must be positive", 400)
	}

	v.CPUs = cpus
	v.Memory = memory
	v.Storage = storage
	v.UpdatedAt = time.Now()
	return nil
}

// SetMetadata sets VM metadata
func (v *VM) SetMetadata(key string, value interface{}) {
	if v.Metadata == nil {
		v.Metadata = make(map[string]interface{})
	}
	v.Metadata[key] = value
	v.UpdatedAt = time.Now()
}

// GetMetadata gets VM metadata
func (v *VM) GetMetadata(key string) (interface{}, bool) {
	if v.Metadata == nil {
		return nil, false
	}
	value, exists := v.Metadata[key]
	return value, exists
}

// RemoveMetadata removes VM metadata
func (v *VM) RemoveMetadata(key string) {
	if v.Metadata != nil {
		delete(v.Metadata, key)
		v.UpdatedAt = time.Now()
	}
}

// GetDisplayName returns the display name for the VM
func (v *VM) GetDisplayName() string {
	return v.Name
}

// GetStatusDisplayName returns a human-readable status
func (v *VM) GetStatusDisplayName() string {
	switch v.Status {
	case VMStatusPending:
		return "Pending"
	case VMStatusRunning:
		return "Running"
	case VMStatusStopped:
		return "Stopped"
	case VMStatusStopping:
		return "Stopping"
	case VMStatusStarting:
		return "Starting"
	case VMStatusTerminated:
		return "Terminated"
	case VMStatusError:
		return "Error"
	default:
		return "Unknown"
	}
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
