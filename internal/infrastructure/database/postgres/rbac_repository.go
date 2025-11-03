package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// rbacRepository implements the RBACRepository interface
type rbacRepository struct {
	db *gorm.DB
}

// NewRBACRepository creates a new RBAC repository
func NewRBACRepository(db *gorm.DB) domain.RBACRepository {
	return &rbacRepository{db: db}
}

// GetUserRole retrieves a user role by user ID and role
func (r *rbacRepository) GetUserRole(userID uuid.UUID, role domain.Role) (*domain.UserRole, error) {
	var userRole domain.UserRole
	result := r.db.Where("user_id = ? AND role = ? AND deleted_at IS NULL", userID, role).First(&userRole)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user role: %v", result.Error)
		return nil, result.Error
	}
	return &userRole, nil
}

// CreateUserRole creates a new user role
func (r *rbacRepository) CreateUserRole(userRole *domain.UserRole) error {
	if err := r.db.Create(userRole).Error; err != nil {
		logger.Errorf("Failed to create user role: %v", err)
		return err
	}
	return nil
}

// DeleteUserRole deletes a user role
func (r *rbacRepository) DeleteUserRole(userID uuid.UUID, role domain.Role) (int64, error) {
	result := r.db.Where("user_id = ? AND role = ? AND deleted_at IS NULL", userID, role).Delete(&domain.UserRole{})
	if result.Error != nil {
		logger.Errorf("Failed to delete user role: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetUserRolesByUserID retrieves all roles for a user
func (r *rbacRepository) GetUserRolesByUserID(userID uuid.UUID) ([]domain.UserRole, error) {
	var userRoles []domain.UserRole
	if err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Find(&userRoles).Error; err != nil {
		logger.Errorf("Failed to get user roles: %v", err)
		return nil, err
	}
	return userRoles, nil
}

// CountUserRoles counts user roles matching the criteria
func (r *rbacRepository) CountUserRoles(userID uuid.UUID, role domain.Role) (int64, error) {
	var count int64
	result := r.db.Model(&domain.UserRole{}).Where("user_id = ? AND role = ? AND deleted_at IS NULL", userID, role).Count(&count)
	if result.Error != nil {
		logger.Errorf("Failed to count user roles: %v", result.Error)
		return 0, result.Error
	}
	return count, nil
}

// GetRoleDistribution returns the distribution of roles across users
func (r *rbacRepository) GetRoleDistribution() (map[domain.Role]int, error) {
	var results []struct {
		Role  domain.Role `json:"role"`
		Count int         `json:"count"`
	}

	if err := r.db.Model(&domain.UserRole{}).
		Where("deleted_at IS NULL").
		Select("role, COUNT(*) as count").
		Group("role").
		Scan(&results).Error; err != nil {
		logger.Errorf("Failed to get role distribution: %v", err)
		return nil, err
	}

	distribution := make(map[domain.Role]int)
	for _, result := range results {
		distribution[result.Role] = result.Count
	}

	return distribution, nil
}

// GetRolePermission retrieves a role permission by role and permission
func (r *rbacRepository) GetRolePermission(role domain.Role, permission domain.Permission) (*domain.RolePermission, error) {
	var rolePermission domain.RolePermission
	result := r.db.Where("role = ? AND permission = ? AND deleted_at IS NULL", role, permission).First(&rolePermission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get role permission: %v", result.Error)
		return nil, result.Error
	}
	return &rolePermission, nil
}

// CreateRolePermission creates a new role permission
func (r *rbacRepository) CreateRolePermission(rolePermission *domain.RolePermission) error {
	if err := r.db.Create(rolePermission).Error; err != nil {
		logger.Errorf("Failed to create role permission: %v", err)
		return err
	}
	return nil
}

// DeleteRolePermission deletes a role permission
func (r *rbacRepository) DeleteRolePermission(role domain.Role, permission domain.Permission) (int64, error) {
	result := r.db.Where("role = ? AND permission = ? AND deleted_at IS NULL", role, permission).Delete(&domain.RolePermission{})
	if result.Error != nil {
		logger.Errorf("Failed to delete role permission: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetRolePermissionsByRole retrieves all permissions for a role
func (r *rbacRepository) GetRolePermissionsByRole(role domain.Role) ([]domain.RolePermission, error) {
	var rolePermissions []domain.RolePermission
	if err := r.db.Where("role = ? AND deleted_at IS NULL", role).Find(&rolePermissions).Error; err != nil {
		logger.Errorf("Failed to get role permissions: %v", err)
		return nil, err
	}
	return rolePermissions, nil
}

// CountRolePermissions counts role permissions matching the criteria
func (r *rbacRepository) CountRolePermissions(role domain.Role, permission domain.Permission) (int64, error) {
	var count int64
	result := r.db.Model(&domain.RolePermission{}).Where("role = ? AND permission = ? AND deleted_at IS NULL", role, permission).Count(&count)
	if result.Error != nil {
		logger.Errorf("Failed to count role permissions: %v", result.Error)
		return 0, result.Error
	}
	return count, nil
}

