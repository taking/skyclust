package domain

import (
	"github.com/google/uuid"
	"time"
)

// Role represents user roles in the system
type Role string

const (
	AdminRoleType  Role = "admin"
	UserRoleType   Role = "user"
	ViewerRoleType Role = "viewer"
)

// Permission represents system permissions
type Permission string

const (
	// User management permissions
	UserCreate Permission = "user:create"
	UserRead   Permission = "user:read"
	UserUpdate Permission = "user:update"
	UserDelete Permission = "user:delete"
	UserManage Permission = "user:manage"

	// System management permissions
	SystemRead   Permission = "system:read"
	SystemUpdate Permission = "system:update"
	SystemManage Permission = "system:manage"

	// Audit permissions
	AuditRead   Permission = "audit:read"
	AuditExport Permission = "audit:export"
	AuditManage Permission = "audit:manage"

	// Workspace permissions
	WorkspaceCreate Permission = "workspace:create"
	WorkspaceRead   Permission = "workspace:read"
	WorkspaceUpdate Permission = "workspace:update"
	WorkspaceDelete Permission = "workspace:delete"
	WorkspaceManage Permission = "workspace:manage"

	// Provider permissions
	ProviderRead   Permission = "provider:read"
	ProviderManage Permission = "provider:manage"
)

// UserRole represents the relationship between users and roles
type UserRole struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Role      Role       `json:"role" gorm:"not null;size:20;index"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission represents the relationship between roles and permissions
type RolePermission struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Role       Role       `json:"role" gorm:"not null;size:20;index"`
	Permission Permission `json:"permission" gorm:"not null;size:50;index"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `json:"-" gorm:"index"`
}

// TableName specifies the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}

// DefaultRolePermissions defines the default permissions for each role
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

// RoleHierarchy defines the role inheritance hierarchy
var RoleHierarchy = map[Role][]Role{
	AdminRoleType:  {UserRoleType, ViewerRoleType},
	UserRoleType:   {ViewerRoleType},
	ViewerRoleType: {},
}

