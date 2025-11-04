package domain

import (
	"github.com/google/uuid"
)

// RBACService defines the interface for role-based access control
type RBACService interface {
	// User role management
	AssignRole(userID uuid.UUID, role Role) error
	RemoveRole(userID uuid.UUID, role Role) error
	GetUserRoles(userID uuid.UUID) ([]Role, error)
	HasRole(userID uuid.UUID, role Role) (bool, error)

	// Permission management
	GrantPermission(role Role, permission Permission) error
	RevokePermission(role Role, permission Permission) error
	GetRolePermissions(role Role) ([]Permission, error)

	// Access control
	CheckPermission(userID uuid.UUID, permission Permission) (bool, error)
	CheckAnyPermission(userID uuid.UUID, permissions []Permission) (bool, error)
	CheckAllPermissions(userID uuid.UUID, permissions []Permission) (bool, error)

	// Role hierarchy
	GetUserEffectivePermissions(userID uuid.UUID) ([]Permission, error)
	GetInheritedRoles(role Role) ([]Role, error)
	HasInheritedRole(userID uuid.UUID, role Role) (bool, error)

	// Statistics
	GetRoleDistribution() (map[Role]int, error)
}

