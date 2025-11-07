package domain

import (
	"time"
)

// Workspace: 시스템의 워크스페이스를 나타내는 도메인 엔티티
type Workspace struct {
	ID          string                 `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name        string                 `json:"name" gorm:"uniqueIndex;not null"`
	Description string                 `json:"description" gorm:"type:text"`
	OwnerID     string                 `json:"owner_id" gorm:"not null;type:uuid"`
	Settings    map[string]interface{} `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time              `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"autoUpdateTime"`
	Active      bool                   `json:"is_active" gorm:"default:true"`
}

// TableName: Workspace의 테이블 이름을 반환합니다
func (Workspace) TableName() string {
	return "workspaces"
}

// IsActive: 워크스페이스가 활성화되어 있는지 확인합니다
func (w *Workspace) IsActive() bool {
	return w.Active
}

// Activate: 워크스페이스를 활성화합니다
func (w *Workspace) Activate() {
	w.Active = true
	w.UpdatedAt = time.Now()
}

// Deactivate: 워크스페이스를 비활성화합니다
func (w *Workspace) Deactivate() {
	w.Active = false
	w.UpdatedAt = time.Now()
}

// UpdateInfo: 워크스페이스 정보를 업데이트합니다
func (w *Workspace) UpdateInfo(name, description string) error {
	if name == "" {
		return NewDomainError(ErrCodeValidationFailed, "workspace name cannot be empty", 400)
	}

	w.Name = name
	w.Description = description
	w.UpdatedAt = time.Now()
	return nil
}

// SetSetting: 워크스페이스 설정을 저장합니다
func (w *Workspace) SetSetting(key string, value interface{}) {
	if w.Settings == nil {
		w.Settings = make(map[string]interface{})
	}
	w.Settings[key] = value
	w.UpdatedAt = time.Now()
}

// GetSetting: 워크스페이스 설정을 조회합니다
func (w *Workspace) GetSetting(key string) (interface{}, bool) {
	if w.Settings == nil {
		return nil, false
	}
	value, exists := w.Settings[key]
	return value, exists
}

// RemoveSetting: 워크스페이스 설정을 제거합니다
func (w *Workspace) RemoveSetting(key string) {
	if w.Settings != nil {
		delete(w.Settings, key)
		w.UpdatedAt = time.Now()
	}
}

// CanUserAccess: 사용자가 이 워크스페이스에 접근할 수 있는지 확인합니다
func (w *Workspace) CanUserAccess(userID string, userRole Role) bool {
	// 관리자는 모든 워크스페이스 접근 가능
	if userRole == AdminRoleType {
		return true
	}
	// 소유자는 자신의 워크스페이스 접근 가능
	return w.OwnerID == userID
}

// IsOwner: 사용자가 이 워크스페이스의 소유자인지 확인합니다
func (w *Workspace) IsOwner(userID string) bool {
	return w.OwnerID == userID
}

// GetDisplayName: 워크스페이스의 표시 이름을 반환합니다
func (w *Workspace) GetDisplayName() string {
	return w.Name
}

// WorkspaceUser: 워크스페이스 내 사용자를 나타내는 엔티티
type WorkspaceUser struct {
	UserID      string    `json:"user_id" gorm:"primaryKey;type:uuid"`
	WorkspaceID string    `json:"workspace_id" gorm:"primaryKey;type:uuid"`
	Role        string    `json:"role" gorm:"not null;default:member"`
	JoinedAt    time.Time `json:"joined_at" gorm:"autoCreateTime"`
}

// TableName: WorkspaceUser의 테이블 이름을 반환합니다
func (WorkspaceUser) TableName() string {
	return "workspace_users"
}

// CreateWorkspaceRequest: 워크스페이스 생성 요청
type CreateWorkspaceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=500"`
	OwnerID     string `json:"owner_id,omitempty"` // 핸들러에서 토큰으로 설정, 클라이언트에서 전달하지 않음
}

// UpdateWorkspaceRequest: 워크스페이스 업데이트 요청
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// Validate: CreateWorkspaceRequest의 유효성을 검사합니다
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

// Validate: UpdateWorkspaceRequest의 유효성을 검사합니다
func (r *UpdateWorkspaceRequest) Validate() error {
	if r.Name != nil && (len(*r.Name) < 3 || len(*r.Name) > 100) {
		return NewDomainError(ErrCodeValidationFailed, "name must be between 3 and 100 characters", 400)
	}
	if r.Description != nil && len(*r.Description) > 500 {
		return NewDomainError(ErrCodeValidationFailed, "description must be less than 500 characters", 400)
	}
	return nil
}
