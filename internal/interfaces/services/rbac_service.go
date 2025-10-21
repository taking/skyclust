package services

import (
	"github.com/google/uuid"
	"skyclust/internal/domain"
)

// RBACService defines the interface for role-based access control operations
type RBACService interface {
	// AssignRole assigns a role to a user
	AssignRole(userID uuid.UUID, role domain.Role) error

	// RemoveRole removes a role from a user
	RemoveRole(userID uuid.UUID, role domain.Role) error

	// GetUserRoles retrieves all roles for a user
	GetUserRoles(userID uuid.UUID) ([]domain.Role, error)

	// HasRole checks if a user has a specific role
	HasRole(userID uuid.UUID, role domain.Role) (bool, error)

	// GrantPermission grants a permission to a role
	GrantPermission(role domain.Role, permission domain.Permission) error

	// RevokePermission revokes a permission from a role
	RevokePermission(role domain.Role, permission domain.Permission) error

	// GetRolePermissions retrieves all permissions for a role
	GetRolePermissions(role domain.Role) ([]domain.Permission, error)

	// CheckPermission checks if a user has a specific permission
	CheckPermission(userID uuid.UUID, permission domain.Permission) (bool, error)

	// CheckAnyPermission checks if a user has any of the specified permissions
	CheckAnyPermission(userID uuid.UUID, permissions []domain.Permission) (bool, error)

	// CheckAllPermissions checks if a user has all of the specified permissions
	CheckAllPermissions(userID uuid.UUID, permissions []domain.Permission) (bool, error)

	// GetUserEffectivePermissions retrieves all effective permissions for a user
	GetUserEffectivePermissions(userID uuid.UUID) ([]domain.Permission, error)

	// GetRoleDistribution returns the distribution of roles across users
	GetRoleDistribution() (map[domain.Role]int, error)
}
