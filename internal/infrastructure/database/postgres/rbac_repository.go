package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// rbacRepository: domain.RBACRepository 인터페이스 구현체
type rbacRepository struct {
	db *gorm.DB
}

// NewRBACRepository: 새로운 RBACRepository를 생성합니다
func NewRBACRepository(db *gorm.DB) domain.RBACRepository {
	return &rbacRepository{db: db}
}

// GetUserRole: 사용자 ID와 역할로 사용자 역할을 조회합니다
func (r *rbacRepository) GetUserRole(userID uuid.UUID, role domain.Role) (*domain.UserRole, error) {
	var userRole domain.UserRole
	result := r.db.Where("user_id = ? AND role = ?", userID, role).First(&userRole)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user role: %v", result.Error)
		return nil, result.Error
	}
	return &userRole, nil
}

// CreateUserRole: 새로운 사용자 역할을 생성합니다
func (r *rbacRepository) CreateUserRole(userRole *domain.UserRole) error {
	if err := r.db.Create(userRole).Error; err != nil {
		logger.Errorf("Failed to create user role: %v", err)
		return err
	}
	return nil
}

// DeleteUserRole: 사용자 역할을 삭제합니다
func (r *rbacRepository) DeleteUserRole(userID uuid.UUID, role domain.Role) (int64, error) {
	result := r.db.Unscoped().Where("user_id = ? AND role = ?", userID, role).Delete(&domain.UserRole{})
	if result.Error != nil {
		logger.Errorf("Failed to delete user role: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetUserRolesByUserID: 사용자의 모든 역할을 조회합니다
func (r *rbacRepository) GetUserRolesByUserID(userID uuid.UUID) ([]domain.UserRole, error) {
	var userRoles []domain.UserRole
	if err := r.db.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		logger.Errorf("Failed to get user roles: %v", err)
		return nil, err
	}
	return userRoles, nil
}

// GetUserRolesByUserIDs: 여러 사용자 ID로 사용자 역할 목록을 배치 조회합니다 (N+1 쿼리 방지)
func (r *rbacRepository) GetUserRolesByUserIDs(userIDs []uuid.UUID) (map[uuid.UUID][]domain.UserRole, error) {
	if len(userIDs) == 0 {
		return make(map[uuid.UUID][]domain.UserRole), nil
	}

	var userRoles []domain.UserRole
	if err := r.db.Where("user_id IN ?", userIDs).Find(&userRoles).Error; err != nil {
		logger.Errorf("Failed to get user roles by user IDs: %v", err)
		return nil, err
	}

	// Group by user ID
	result := make(map[uuid.UUID][]domain.UserRole)
	for _, userRole := range userRoles {
		result[userRole.UserID] = append(result[userRole.UserID], userRole)
	}

	// Ensure all user IDs are in the map (even if they have no roles)
	for _, userID := range userIDs {
		if _, exists := result[userID]; !exists {
			result[userID] = []domain.UserRole{}
		}
	}

	return result, nil
}

// CountUserRoles: 조건에 맞는 사용자 역할 수를 반환합니다
func (r *rbacRepository) CountUserRoles(userID uuid.UUID, role domain.Role) (int64, error) {
	var count int64
	result := r.db.Model(&domain.UserRole{}).Where("user_id = ? AND role = ?", userID, role).Count(&count)
	if result.Error != nil {
		logger.Errorf("Failed to count user roles: %v", result.Error)
		return 0, result.Error
	}
	return count, nil
}

// GetRoleDistribution: 사용자 간 역할 분포를 반환합니다
func (r *rbacRepository) GetRoleDistribution() (map[domain.Role]int, error) {
	var results []struct {
		Role  domain.Role `json:"role"`
		Count int         `json:"count"`
	}

	if err := r.db.Model(&domain.UserRole{}).
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

// GetRolePermission: 역할과 권한으로 역할 권한을 조회합니다
func (r *rbacRepository) GetRolePermission(role domain.Role, permission domain.Permission) (*domain.RolePermission, error) {
	var rolePermission domain.RolePermission
	result := r.db.Where("role = ? AND permission = ?", role, permission).First(&rolePermission)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get role permission: %v", result.Error)
		return nil, result.Error
	}
	return &rolePermission, nil
}

// CreateRolePermission: 새로운 역할 권한을 생성합니다
func (r *rbacRepository) CreateRolePermission(rolePermission *domain.RolePermission) error {
	if err := r.db.Create(rolePermission).Error; err != nil {
		logger.Errorf("Failed to create role permission: %v", err)
		return err
	}
	return nil
}

// DeleteRolePermission: 역할 권한을 삭제합니다
func (r *rbacRepository) DeleteRolePermission(role domain.Role, permission domain.Permission) (int64, error) {
	result := r.db.Unscoped().Where("role = ? AND permission = ?", role, permission).Delete(&domain.RolePermission{})
	if result.Error != nil {
		logger.Errorf("Failed to delete role permission: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// GetRolePermissionsByRole: 역할의 모든 권한을 조회합니다
func (r *rbacRepository) GetRolePermissionsByRole(role domain.Role) ([]domain.RolePermission, error) {
	var rolePermissions []domain.RolePermission
	if err := r.db.Where("role = ?", role).Find(&rolePermissions).Error; err != nil {
		logger.Errorf("Failed to get role permissions: %v", err)
		return nil, err
	}
	return rolePermissions, nil
}

// CountRolePermissions: 조건에 맞는 역할 권한 수를 반환합니다
func (r *rbacRepository) CountRolePermissions(role domain.Role, permission domain.Permission) (int64, error) {
	var count int64
	result := r.db.Model(&domain.RolePermission{}).Where("role = ? AND permission = ?", role, permission).Count(&count)
	if result.Error != nil {
		logger.Errorf("Failed to count role permissions: %v", result.Error)
		return 0, result.Error
	}
	return count, nil
}
