package domain

import (
	"github.com/google/uuid"
	"time"
)

// Role: 시스템의 사용자 역할을 나타내는 타입
type Role string

const (
	AdminRoleType  Role = "admin"  // 관리자 역할
	UserRoleType   Role = "user"   // 일반 사용자 역할
	ViewerRoleType Role = "viewer" // 조회자 역할
)

// Permission: 시스템 권한을 나타내는 타입
type Permission string

const (
	// 사용자 관리 권한
	UserCreate Permission = "user:create"
	UserRead   Permission = "user:read"
	UserUpdate Permission = "user:update"
	UserDelete Permission = "user:delete"
	UserManage Permission = "user:manage"

	// 시스템 관리 권한
	SystemRead   Permission = "system:read"
	SystemUpdate Permission = "system:update"
	SystemManage Permission = "system:manage"

	// 감사 로그 권한
	AuditRead   Permission = "audit:read"
	AuditExport Permission = "audit:export"
	AuditManage Permission = "audit:manage"

	// 워크스페이스 권한
	WorkspaceCreate Permission = "workspace:create"
	WorkspaceRead   Permission = "workspace:read"
	WorkspaceUpdate Permission = "workspace:update"
	WorkspaceDelete Permission = "workspace:delete"
	WorkspaceManage Permission = "workspace:manage"

	// 제공자 권한
	ProviderRead   Permission = "provider:read"
	ProviderManage Permission = "provider:manage"
)

// UserRole: 사용자와 역할 간의 관계를 나타내는 엔티티
type UserRole struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      Role      `json:"role" gorm:"not null;size:20;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName: UserRole의 테이블 이름을 반환합니다
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission: 역할과 권한 간의 관계를 나타내는 엔티티
type RolePermission struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Role       Role       `json:"role" gorm:"not null;size:20;index"`
	Permission Permission `json:"permission" gorm:"not null;size:50;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName: RolePermission의 테이블 이름을 반환합니다
func (RolePermission) TableName() string {
	return "role_permissions"
}

// DefaultRolePermissions: 각 역할에 대한 기본 권한을 정의합니다
var DefaultRolePermissions = map[Role][]Permission{
	AdminRoleType: {
		UserCreate, UserRead, UserUpdate, UserDelete, UserManage,
		SystemRead, SystemUpdate, SystemManage,
		AuditRead, AuditExport, AuditManage,
		WorkspaceCreate, WorkspaceRead, WorkspaceUpdate, WorkspaceDelete, WorkspaceManage,
		ProviderRead, ProviderManage,
	},
	UserRoleType: {
		WorkspaceCreate, WorkspaceRead, WorkspaceUpdate,
		ProviderRead,
	},
	ViewerRoleType: {
		WorkspaceRead,
		ProviderRead,
	},
}

// RoleHierarchy: 역할 상속 계층 구조를 정의합니다
var RoleHierarchy = map[Role][]Role{
	AdminRoleType:  {UserRoleType, ViewerRoleType},
	UserRoleType:   {ViewerRoleType},
	ViewerRoleType: {},
}

