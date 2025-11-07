package domain

import (
	"time"
)

// VMStatus: 가상 머신의 상태를 나타내는 타입
type VMStatus string

const (
	VMStatusPending    VMStatus = "pending"    // 대기 중
	VMStatusRunning    VMStatus = "running"    // 실행 중
	VMStatusStopped    VMStatus = "stopped"    // 중지됨
	VMStatusStopping   VMStatus = "stopping"   // 중지 중
	VMStatusStarting   VMStatus = "starting"   // 시작 중
	VMStatusTerminated VMStatus = "terminated" // 종료됨
	VMStatusError      VMStatus = "error"      // 오류
)

// VM: 시스템의 가상 머신을 나타내는 도메인 엔티티
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
	Memory      int                    `json:"memory" gorm:"not null"`  // in MB
	Storage     int                    `json:"storage" gorm:"not null"` // in GB
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
}

// TableName: VM의 테이블 이름을 반환합니다
func (VM) TableName() string {
	return "vms"
}

// IsRunning: VM이 실행 중인지 확인합니다
func (v *VM) IsRunning() bool {
	return v.Status == VMStatusRunning
}

// IsStopped: VM이 중지되었는지 확인합니다
func (v *VM) IsStopped() bool {
	return v.Status == VMStatusStopped
}

// IsPending: VM이 대기 중인지 확인합니다
func (v *VM) IsPending() bool {
	return v.Status == VMStatusPending
}

// IsError: VM이 오류 상태인지 확인합니다
func (v *VM) IsError() bool {
	return v.Status == VMStatusError
}

// CanStart: VM을 시작할 수 있는지 확인합니다
func (v *VM) CanStart() bool {
	return v.Status == VMStatusStopped || v.Status == VMStatusError
}

// CanStop: VM을 중지할 수 있는지 확인합니다
func (v *VM) CanStop() bool {
	return v.Status == VMStatusRunning
}

// CanRestart: VM을 재시작할 수 있는지 확인합니다
func (v *VM) CanRestart() bool {
	return v.Status == VMStatusRunning || v.Status == VMStatusStopped
}

// CanTerminate: VM을 종료할 수 있는지 확인합니다
func (v *VM) CanTerminate() bool {
	return v.Status != VMStatusTerminated
}

// Start: VM 상태를 시작 중으로 변경합니다
func (v *VM) Start() error {
	if !v.CanStart() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be started in current status", 400)
	}
	v.Status = VMStatusStarting
	v.UpdatedAt = time.Now()
	return nil
}

// Stop: VM 상태를 중지 중으로 변경합니다
func (v *VM) Stop() error {
	if !v.CanStop() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be stopped in current status", 400)
	}
	v.Status = VMStatusStopping
	v.UpdatedAt = time.Now()
	return nil
}

// Restart: VM 상태를 시작 중으로 변경합니다 (재시작)
func (v *VM) Restart() error {
	if !v.CanRestart() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be restarted in current status", 400)
	}
	v.Status = VMStatusStarting
	v.UpdatedAt = time.Now()
	return nil
}

// Terminate: VM 상태를 종료됨으로 변경합니다
func (v *VM) Terminate() error {
	if !v.CanTerminate() {
		return NewDomainError(ErrCodeValidationFailed, "VM cannot be terminated in current status", 400)
	}
	v.Status = VMStatusTerminated
	v.UpdatedAt = time.Now()
	return nil
}

// SetStatus: VM 상태를 설정합니다
func (v *VM) SetStatus(status VMStatus) {
	v.Status = status
	v.UpdatedAt = time.Now()
}

// SetInstanceID: 클라우드 인스턴스 ID를 설정합니다
func (v *VM) SetInstanceID(instanceID string) {
	v.InstanceID = instanceID
	v.UpdatedAt = time.Now()
}

// UpdateSpecs: VM 사양을 업데이트합니다
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

// SetMetadata: VM 메타데이터를 설정합니다
func (v *VM) SetMetadata(key string, value interface{}) {
	if v.Metadata == nil {
		v.Metadata = make(map[string]interface{})
	}
	v.Metadata[key] = value
	v.UpdatedAt = time.Now()
}

// GetMetadata: VM 메타데이터를 조회합니다
func (v *VM) GetMetadata(key string) (interface{}, bool) {
	if v.Metadata == nil {
		return nil, false
	}
	value, exists := v.Metadata[key]
	return value, exists
}

// RemoveMetadata: VM 메타데이터를 제거합니다
func (v *VM) RemoveMetadata(key string) {
	if v.Metadata != nil {
		delete(v.Metadata, key)
		v.UpdatedAt = time.Now()
	}
}

// GetDisplayName: VM의 표시 이름을 반환합니다
func (v *VM) GetDisplayName() string {
	return v.Name
}

// GetStatusDisplayName: 사람이 읽을 수 있는 상태 이름을 반환합니다
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

// CreateVMRequest: VM 생성 요청 DTO
type CreateVMRequest struct {
	Name        string                 `json:"name" validate:"required,min=3,max=100"`
	WorkspaceID string                 `json:"workspace_id" validate:"required"`
	Provider    string                 `json:"provider" validate:"required"`
	Type        string                 `json:"type" validate:"required"`
	Region      string                 `json:"region" validate:"required"`
	ImageID     string                 `json:"image_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateVMRequest: VM 업데이트 요청 DTO
type UpdateVMRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Type     *string                `json:"type,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validate: CreateVMRequest의 유효성을 검사합니다
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

// Validate: UpdateVMRequest의 유효성을 검사합니다
func (r *UpdateVMRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 3 || len(*r.Name) > 100) {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	return nil
}
