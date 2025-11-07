package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// JSONBMap: PostgreSQL의 JSONB 필드를 처리하기 위한 커스텀 타입
type JSONBMap map[string]interface{}

// Value: driver.Valuer 인터페이스를 구현합니다
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan: sql.Scanner 인터페이스를 구현합니다
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

// AuditLog: 감사 로그 항목을 나타내는 도메인 엔티티
type AuditLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Action    string    `json:"action" gorm:"not null;size:50;index"` // login, logout, create_credential, etc.
	Resource  string    `json:"resource" gorm:"size:100"`             // api endpoint
	IPAddress string    `json:"ip_address" gorm:"type:inet"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Details   JSONBMap  `json:"details" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// AuditLogFilters: 감사 로그 쿼리를 위한 필터
type AuditLogFilters struct {
	UserID    *uuid.UUID
	Action    string
	Resource  string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	Limit     int
}

// AuditStatsFilters: 감사 통계를 위한 필터
type AuditStatsFilters struct {
	StartTime *time.Time
	EndTime   *time.Time
}

// AuditStats: 감사 로그 통계를 나타내는 타입
type AuditStats struct {
	TotalEvents  int64                    `json:"total_events"`
	UniqueUsers  int64                    `json:"unique_users"`
	TopActions   []map[string]interface{} `json:"top_actions"`
	TopResources []map[string]interface{} `json:"top_resources"`
	EventsByDay  []map[string]interface{} `json:"events_by_day"`
}

// AuditLogSummary: 감사 로그 활동 요약을 나타내는 타입
type AuditLogSummary struct {
	TotalEvents    int64                    `json:"total_events"`
	UniqueUsers    int64                    `json:"unique_users"`
	MostActiveUser string                   `json:"most_active_user"`
	TopActions     []map[string]interface{} `json:"top_actions"`
	SecurityEvents int64                    `json:"security_events"`
	ErrorEvents    int64                    `json:"error_events"`
}

// AuditAction: 다양한 액션을 나타내는 상수
const (
	// 사용자 관련 액션
	ActionUserRegister   = "user_register"
	ActionUserLogin      = "user_login"
	ActionUserLogout     = "user_logout"
	ActionOIDCLogin      = "oidc_login"
	ActionOIDCLogout     = "oidc_logout"
	ActionUserUpdate     = "user_update"
	ActionUserDelete     = "user_delete"
	ActionPasswordChange = "password_change"

	// 자격증명 관련 액션
	ActionCredentialCreate = "credential_create"
	ActionCredentialUpdate = "credential_update"
	ActionCredentialDelete = "credential_delete"

	// 워크스페이스 관련 액션
	ActionWorkspaceCreate          = "workspace_create"
	ActionWorkspaceUpdate          = "workspace_update"
	ActionWorkspaceDelete          = "workspace_delete"
	ActionWorkspaceUserAdded       = "workspace_user_added"
	ActionWorkspaceUserRemoved     = "workspace_user_removed"
	ActionWorkspaceUserRoleUpdated = "workspace_user_role_updated"

	// VM 관련 액션
	ActionVMCreate  = "vm_create"
	ActionVMUpdate  = "vm_update"
	ActionVMDelete  = "vm_delete"
	ActionVMStart   = "vm_start"
	ActionVMStop    = "vm_stop"
	ActionVMRestart = "vm_restart"

	// Kubernetes 관련 액션
	ActionKubernetesClusterCreate   = "kubernetes_cluster_create"
	ActionKubernetesClusterUpdate   = "kubernetes_cluster_update"
	ActionKubernetesClusterDelete   = "kubernetes_cluster_delete"
	ActionKubernetesClusterUpgrade  = "kubernetes_cluster_upgrade"
	ActionKubernetesNodePoolCreate  = "kubernetes_node_pool_create"
	ActionKubernetesNodePoolDelete  = "kubernetes_node_pool_delete"
	ActionKubernetesNodeGroupCreate = "kubernetes_node_group_create"
	ActionKubernetesNodeGroupDelete = "kubernetes_node_group_delete"

	// 네트워크 관련 액션
	ActionVPCCreate               = "vpc_create"
	ActionVPCUpdate               = "vpc_update"
	ActionVPCDelete               = "vpc_delete"
	ActionSubnetCreate            = "subnet_create"
	ActionSubnetUpdate            = "subnet_update"
	ActionSubnetDelete            = "subnet_delete"
	ActionSecurityGroupCreate     = "security_group_create"
	ActionSecurityGroupUpdate     = "security_group_update"
	ActionSecurityGroupDelete     = "security_group_delete"
	ActionSecurityGroupRuleAdd    = "security_group_rule_add"
	ActionSecurityGroupRuleRemove = "security_group_rule_remove"

	// 제공자 관련 액션 (레거시)
	ActionProviderList   = "provider_list"
	ActionInstanceList   = "instance_list"
	ActionInstanceCreate = "instance_create"
	ActionInstanceDelete = "instance_delete"
)

// AuditLogRequest: 감사 로그 생성을 위한 요청 DTO
type AuditLogRequest struct {
	UserID   uuid.UUID              `json:"user_id"`
	Action   string                 `json:"action"`
	Resource string                 `json:"resource"`
	Details  map[string]interface{} `json:"details,omitempty"`
}
