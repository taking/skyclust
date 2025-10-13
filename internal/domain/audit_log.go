package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	Action    string                 `json:"action" gorm:"not null;size:50;index"` // login, logout, create_credential, etc.
	Resource  string                 `json:"resource" gorm:"size:100"`             // api endpoint
	IPAddress string                 `json:"ip_address" gorm:"type:inet"`
	UserAgent string                 `json:"user_agent" gorm:"type:text"`
	Details   map[string]interface{} `json:"details" gorm:"type:jsonb"`
	CreatedAt time.Time              `json:"created_at" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	Create(log *AuditLog) error
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	CountByAction(action string) (int64, error)
	DeleteOldLogs(before time.Time) error
}

// AuditLogService defines the interface for audit log business logic
type AuditLogService interface {
	LogAction(userID uuid.UUID, action, resource string, details map[string]interface{}) error
	GetUserLogs(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetLogsByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetLogsByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CleanupOldLogs(retentionDays int) error
}

// AuditAction constants for different actions
const (
	ActionUserRegister     = "user_register"
	ActionUserLogin        = "user_login"
	ActionUserLogout       = "user_logout"
	ActionOIDCLogin        = "oidc_login"
	ActionOIDCLogout       = "oidc_logout"
	ActionUserUpdate       = "user_update"
	ActionUserDelete       = "user_delete"
	ActionCredentialCreate = "credential_create"
	ActionCredentialUpdate = "credential_update"
	ActionCredentialDelete = "credential_delete"
	ActionProviderList     = "provider_list"
	ActionInstanceList     = "instance_list"
	ActionInstanceCreate   = "instance_create"
	ActionInstanceDelete   = "instance_delete"
	ActionPasswordChange   = "password_change"
)

// AuditLogRequest represents the request to create an audit log
type AuditLogRequest struct {
	UserID   uuid.UUID              `json:"user_id"`
	Action   string                 `json:"action"`
	Resource string                 `json:"resource"`
	Details  map[string]interface{} `json:"details,omitempty"`
}
