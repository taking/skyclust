package rbac

import (
	"fmt"
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
)

// rbacService implements the RBACService interface
type rbacService struct {
	rbacRepo domain.RBACRepository
}

// NewService creates a new RBAC service
func NewService(rbacRepo domain.RBACRepository) domain.RBACService {
	// Initialize default role permissions
	service := &rbacService{rbacRepo: rbacRepo}
	service.initializeDefaultPermissions()
	return service
}

// initializeDefaultPermissions sets up default role permissions
func (r *rbacService) initializeDefaultPermissions() {
	for role, permissions := range domain.DefaultRolePermissions {
		for _, permission := range permissions {
			// Check if permission already exists
			existing, err := r.rbacRepo.GetRolePermission(role, permission)
			if err != nil {
				logger.Warnf("Failed to check role permission %s for role %s: %v", permission, role, err)
				continue
			}
			if existing == nil {
				// Create new permission
				rolePermission := &domain.RolePermission{
					Role:       role,
					Permission: permission,
				}
				if err := r.rbacRepo.CreateRolePermission(rolePermission); err != nil {
					logger.Errorf("Failed to create role permission %s for role %s: %v", permission, role, err)
				}
			}
		}
	}
}

// AssignRole assigns a role to a user
func (r *rbacService) AssignRole(userID uuid.UUID, role domain.Role) error {
	// Check if user already has this role
	existing, err := r.rbacRepo.GetUserRole(userID, role)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check existing role: %v", err), 500)
	}
	if existing != nil {
		return domain.NewDomainError(domain.ErrCodeAlreadyExists, fmt.Sprintf("user already has role %s", role), 409)
	}

	userRole := &domain.UserRole{
		UserID: userID,
		Role:   role,
	}

	if err := r.rbacRepo.CreateUserRole(userRole); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to assign role: %v", err), 500)
	}

	logger.Infof("Assigned role %s to user %s", role, userID)
	return nil
}

// RemoveRole removes a role from a user
func (r *rbacService) RemoveRole(userID uuid.UUID, role domain.Role) error {
	rowsAffected, err := r.rbacRepo.DeleteUserRole(userID, role)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to remove role: %v", err), 500)
	}

	if rowsAffected == 0 {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("user does not have role %s", role), 404)
	}

	logger.Infof("Removed role %s from user %s", role, userID)
	return nil
}

// GetUserRoles returns all roles assigned to a user
func (r *rbacService) GetUserRoles(userID uuid.UUID) ([]domain.Role, error) {
	userRoles, err := r.rbacRepo.GetUserRolesByUserID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user roles: %v", err), 500)
	}

	roles := make([]domain.Role, len(userRoles))
	for i, userRole := range userRoles {
		roles[i] = userRole.Role
	}

	return roles, nil
}

// HasRole checks if a user has a specific role
func (r *rbacService) HasRole(userID uuid.UUID, role domain.Role) (bool, error) {
	count, err := r.rbacRepo.CountUserRoles(userID, role)
	if err != nil {
		return false, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check user role: %v", err), 500)
	}

	return count > 0, nil
}

// GrantPermission grants a permission to a role
func (r *rbacService) GrantPermission(role domain.Role, permission domain.Permission) error {
	// Check if permission already exists
	existing, err := r.rbacRepo.GetRolePermission(role, permission)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check existing permission: %v", err), 500)
	}
	if existing != nil {
		return domain.NewDomainError(domain.ErrCodeAlreadyExists, fmt.Sprintf("role %s already has permission %s", role, permission), 409)
	}

	rolePermission := &domain.RolePermission{
		Role:       role,
		Permission: permission,
	}

	if err := r.rbacRepo.CreateRolePermission(rolePermission); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to grant permission: %v", err), 500)
	}

	logger.Infof("Granted permission %s to role %s", permission, role)
	return nil
}

// RevokePermission revokes a permission from a role
func (r *rbacService) RevokePermission(role domain.Role, permission domain.Permission) error {
	rowsAffected, err := r.rbacRepo.DeleteRolePermission(role, permission)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to revoke permission: %v", err), 500)
	}

	if rowsAffected == 0 {
		return domain.NewDomainError(domain.ErrCodeNotFound, fmt.Sprintf("role %s does not have permission %s", role, permission), 404)
	}

	logger.Infof("Revoked permission %s from role %s", permission, role)
	return nil
}

// GetRolePermissions returns all permissions for a role
func (r *rbacService) GetRolePermissions(role domain.Role) ([]domain.Permission, error) {
	rolePermissions, err := r.rbacRepo.GetRolePermissionsByRole(role)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get role permissions: %v", err), 500)
	}

	permissions := make([]domain.Permission, len(rolePermissions))
	for i, rp := range rolePermissions {
		permissions[i] = rp.Permission
	}

	return permissions, nil
}

// CheckPermission checks if a user has a specific permission
func (r *rbacService) CheckPermission(userID uuid.UUID, permission domain.Permission) (bool, error) {
	// Get user roles
	roles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	// Check if any role has the permission
	for _, role := range roles {
		hasPermission, err := r.roleHasPermission(role, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// CheckAnyPermission checks if a user has any of the specified permissions
func (r *rbacService) CheckAnyPermission(userID uuid.UUID, permissions []domain.Permission) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := r.CheckPermission(userID, permission)
		if err != nil {
			return false, err
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// CheckAllPermissions checks if a user has all of the specified permissions
func (r *rbacService) CheckAllPermissions(userID uuid.UUID, permissions []domain.Permission) (bool, error) {
	for _, permission := range permissions {
		hasPermission, err := r.CheckPermission(userID, permission)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}

	return true, nil
}

// GetUserEffectivePermissions returns all effective permissions for a user
func (r *rbacService) GetUserEffectivePermissions(userID uuid.UUID) ([]domain.Permission, error) {
	// Get user roles
	roles, err := r.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	// Collect all permissions from all roles
	permissionSet := make(map[domain.Permission]bool)
	for _, role := range roles {
		permissions, err := r.GetRolePermissions(role)
		if err != nil {
			return nil, err
		}

		for _, permission := range permissions {
			permissionSet[permission] = true
		}
	}

	// Convert map to slice
	permissions := make([]domain.Permission, 0, len(permissionSet))
	for permission := range permissionSet {
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// roleHasPermission checks if a role has a specific permission
func (r *rbacService) roleHasPermission(role domain.Role, permission domain.Permission) (bool, error) {
	count, err := r.rbacRepo.CountRolePermissions(role, permission)
	if err != nil {
		return false, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to check role permission: %v", err), 500)
	}

	return count > 0, nil
}

// GetRoleDistribution returns the distribution of roles across users
func (r *rbacService) GetRoleDistribution() (map[domain.Role]int, error) {
	distribution, err := r.rbacRepo.GetRoleDistribution()
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get role distribution: %v", err), 500)
	}

	return distribution, nil
}

// GetInheritedRoles returns all roles that inherit from the given role
func (r *rbacService) GetInheritedRoles(role domain.Role) ([]domain.Role, error) {
	inheritedRoles, exists := domain.RoleHierarchy[role]
	if !exists {
		return []domain.Role{}, nil
	}

	// Recursively get all inherited roles
	var allInheritedRoles []domain.Role
	for _, inheritedRole := range inheritedRoles {
		allInheritedRoles = append(allInheritedRoles, inheritedRole)

		// Get roles inherited by this role
		subInherited, err := r.GetInheritedRoles(inheritedRole)
		if err != nil {
			return nil, err
		}
		allInheritedRoles = append(allInheritedRoles, subInherited...)
	}

	return allInheritedRoles, nil
}

// HasInheritedRole checks if a user has a role through inheritance
func (r *rbacService) HasInheritedRole(userID uuid.UUID, role domain.Role) (bool, error) {
	// Get user's direct roles
	userRoles, err := r.GetUserRoles(userID)
	if err != nil {
		return false, err
	}

	// Check if user has the role directly
	for _, userRole := range userRoles {
		if userRole == role {
			return true, nil
		}

		// Check if user's role inherits the target role
		inheritedRoles, err := r.GetInheritedRoles(userRole)
		if err != nil {
			return false, err
		}

		for _, inheritedRole := range inheritedRoles {
			if inheritedRole == role {
				return true, nil
			}
		}
	}

	return false, nil
}
