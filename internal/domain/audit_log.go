package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONBMap is a custom type for handling JSONB fields in PostgreSQL
type JSONBMap map[string]interface{}

// Value implements the driver.Valuer interface for JSONBMap
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONBMap
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return json.Unmarshal([]byte{}, j)
	}

	if len(bytes) == 0 {
		*j = make(JSONBMap)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID              `json:"user_id" gorm:"type:uuid;not null;index"`
	Action    string                 `json:"action" gorm:"not null;size:50;index"` // login, logout, create_credential, etc.
	Resource  string                 `json:"resource" gorm:"size:100"`             // api endpoint
	IPAddress string                 `json:"ip_address" gorm:"type:inet"`
	UserAgent string                 `json:"user_agent" gorm:"type:text"`
	Details   JSONBMap               `json:"details" gorm:"type:jsonb"`
	CreatedAt time.Time              `json:"created_at" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	Create(log *AuditLog) error
	GetByID(id uuid.UUID) (*AuditLog, error)
	GetByUserID(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	CountByAction(action string) (int64, error)
	DeleteOldLogs(before time.Time) (int64, error)
	// Statistics methods
	GetTotalCount(filters AuditStatsFilters) (int64, error)
	GetUniqueUsersCount(filters AuditStatsFilters) (int64, error)
	GetTopActions(filters AuditStatsFilters, limit int) ([]map[string]interface{}, error)
	GetTopResources(filters AuditStatsFilters, limit int) ([]map[string]interface{}, error)
	GetEventsByDay(filters AuditStatsFilters) ([]map[string]interface{}, error)
	// Summary methods
	GetMostActiveUser(startTime, endTime time.Time) (uuid.UUID, int64, error)
	GetSecurityEventsCount(startTime, endTime time.Time) (int64, error)
	GetErrorEventsCount(startTime, endTime time.Time) (int64, error)
}

// AuditLogFilters represents filters for audit log queries
type AuditLogFilters struct {
	UserID    *uuid.UUID
	Action    string
	Resource  string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Limit     int
}

// AuditStatsFilters represents filters for audit statistics
type AuditStatsFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
}

// AuditStats represents audit log statistics
type AuditStats struct {
	TotalEvents  int64                    `json:"total_events"`
	UniqueUsers  int64                    `json:"unique_users"`
	TopActions   []map[string]interface{} `json:"top_actions"`
	TopResources []map[string]interface{} `json:"top_resources"`
	EventsByDay  []map[string]interface{} `json:"events_by_day"`
}

// AuditLogSummary represents a summary of audit log activities
type AuditLogSummary struct {
	TotalEvents    int64                    `json:"total_events"`
	UniqueUsers    int64                    `json:"unique_users"`
	MostActiveUser string                   `json:"most_active_user"`
	TopActions     []map[string]interface{} `json:"top_actions"`
	SecurityEvents int64                    `json:"security_events"`
	ErrorEvents    int64                    `json:"error_events"`
}

// AuditLogService defines the interface for audit log business logic
type AuditLogService interface {
	LogAction(userID uuid.UUID, action, resource string, details map[string]interface{}) error
	GetUserLogs(userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	GetLogsByAction(action string, limit, offset int) ([]*AuditLog, error)
	GetLogsByDateRange(start, end time.Time, limit, offset int) ([]*AuditLog, error)
	CleanupOldLogs(retentionDays int) error

	// Admin-specific methods
	GetAuditLogs(filters AuditLogFilters) ([]*AuditLog, int64, error)
	GetAuditLogByID(id uuid.UUID) (*AuditLog, error)
	GetAuditStats(filters AuditStatsFilters) (*AuditStats, error)
	ExportAuditLogs(filters AuditLogFilters, format string) ([]byte, error)
	CleanupAuditLogs(retentionDays int) (int64, error)
	GetAuditLogSummary(startTime, endTime time.Time) (*AuditLogSummary, error)
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
