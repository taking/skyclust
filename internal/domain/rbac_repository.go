package domain

import (
	"github.com/google/uuid"
)

// RBACRepository defines the interface for RBAC data operations
type RBACRepository interface {
	// UserRole operations
	GetUserRole(userID uuid.UUID, role Role) (*UserRole, error)
	CreateUserRole(userRole *UserRole) error
	DeleteUserRole(userID uuid.UUID, role Role) (int64, error) // Returns rows affected
	GetUserRolesByUserID(userID uuid.UUID) ([]UserRole, error)
	GetUserRolesByUserIDs(userIDs []uuid.UUID) (map[uuid.UUID][]UserRole, error) // Batch fetch user roles
	CountUserRoles(userID uuid.UUID, role Role) (int64, error)
	GetRoleDistribution() (map[Role]int, error)

	// RolePermission operations
	GetRolePermission(role Role, permission Permission) (*RolePermission, error)
	CreateRolePermission(rolePermission *RolePermission) error
	DeleteRolePermission(role Role, permission Permission) (int64, error) // Returns rows affected
	GetRolePermissionsByRole(role Role) ([]RolePermission, error)
	CountRolePermissions(role Role, permission Permission) (int64, error)
}
