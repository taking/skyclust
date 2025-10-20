package service

import (
	"testing"

	"skyclust/internal/domain"
	"skyclust/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockRBACService is a mock implementation of RBACService
type MockRBACService struct {
	mock.Mock
}

func (m *MockRBACService) AssignRole(userID uuid.UUID, role domain.Role) error {
	args := m.Called(userID, role)
	return args.Error(0)
}

func (m *MockRBACService) RemoveRole(userID uuid.UUID, role domain.Role) error {
	args := m.Called(userID, role)
	return args.Error(0)
}

func (m *MockRBACService) GetUserRoles(userID uuid.UUID) ([]domain.Role, error) {
	args := m.Called(userID)
	return args.Get(0).([]domain.Role), args.Error(1)
}

func (m *MockRBACService) HasRole(userID uuid.UUID, role domain.Role) (bool, error) {
	args := m.Called(userID, role)
	return args.Bool(0), args.Error(1)
}

func (m *MockRBACService) GrantPermission(role domain.Role, permission domain.Permission) error {
	args := m.Called(role, permission)
	return args.Error(0)
}

func (m *MockRBACService) RevokePermission(role domain.Role, permission domain.Permission) error {
	args := m.Called(role, permission)
	return args.Error(0)
}

func (m *MockRBACService) GetRolePermissions(role domain.Role) ([]domain.Permission, error) {
	args := m.Called(role)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockRBACService) CheckPermission(userID uuid.UUID, permission domain.Permission) (bool, error) {
	args := m.Called(userID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockRBACService) CheckAnyPermission(userID uuid.UUID, permissions []domain.Permission) (bool, error) {
	args := m.Called(userID, permissions)
	return args.Bool(0), args.Error(1)
}

func (m *MockRBACService) CheckAllPermissions(userID uuid.UUID, permissions []domain.Permission) (bool, error) {
	args := m.Called(userID, permissions)
	return args.Bool(0), args.Error(1)
}

func (m *MockRBACService) GetUserEffectivePermissions(userID uuid.UUID) ([]domain.Permission, error) {
	args := m.Called(userID)
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockRBACService) GetRoleDistribution() (map[domain.Role]int, error) {
	args := m.Called()
	return args.Get(0).(map[domain.Role]int), args.Error(1)
}

func TestRBACService_AssignRole(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	userID := uuid.New()
	role := domain.AdminRoleType

	// Test assigning role
	err = rbacService.AssignRole(userID, role)
	assert.NoError(t, err)

	// Verify role was assigned
	hasRole, err := rbacService.HasRole(userID, role)
	assert.NoError(t, err)
	assert.True(t, hasRole)

	// Test getting user roles
	roles, err := rbacService.GetUserRoles(userID)
	assert.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, role, roles[0])
}

func TestRBACService_RemoveRole(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	userID := uuid.New()
	role := domain.AdminRoleType

	// First assign a role
	err = rbacService.AssignRole(userID, role)
	assert.NoError(t, err)

	// Verify role was assigned
	hasRole, err := rbacService.HasRole(userID, role)
	assert.NoError(t, err)
	assert.True(t, hasRole)

	// Now remove the role
	err = rbacService.RemoveRole(userID, role)
	assert.NoError(t, err)

	// Verify role was removed
	hasRole, err = rbacService.HasRole(userID, role)
	assert.NoError(t, err)
	assert.False(t, hasRole)
}

func TestRBACService_GrantPermission(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	role := domain.AdminRoleType
	permission := domain.UserCreate

	// Test granting permission
	err = rbacService.GrantPermission(role, permission)
	assert.NoError(t, err)

	// Verify permission was granted
	permissions, err := rbacService.GetRolePermissions(role)
	assert.NoError(t, err)
	assert.Contains(t, permissions, permission)
}

func TestRBACService_CheckPermission(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	userID := uuid.New()
	role := domain.AdminRoleType
	permission := domain.UserCreate

	// Assign role to user
	err = rbacService.AssignRole(userID, role)
	assert.NoError(t, err)

	// Grant permission to role
	err = rbacService.GrantPermission(role, permission)
	assert.NoError(t, err)

	// Test checking permission
	hasPermission, err := rbacService.CheckPermission(userID, permission)
	assert.NoError(t, err)
	assert.True(t, hasPermission)

	// Test checking non-existent permission
	hasPermission, err = rbacService.CheckPermission(userID, domain.AuditRead)
	assert.NoError(t, err)
	assert.False(t, hasPermission)
}

func TestRBACService_GetRoleDistribution(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()

	// Assign roles
	err = rbacService.AssignRole(user1ID, domain.AdminRoleType)
	assert.NoError(t, err)

	err = rbacService.AssignRole(user2ID, domain.UserRoleType)
	assert.NoError(t, err)

	err = rbacService.AssignRole(user3ID, domain.UserRoleType)
	assert.NoError(t, err)

	// Test getting role distribution
	distribution, err := rbacService.GetRoleDistribution()
	assert.NoError(t, err)
	assert.Equal(t, 1, distribution[domain.AdminRoleType])
	assert.Equal(t, 2, distribution[domain.UserRoleType])
}

func TestRBACService_CheckAnyPermission(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	userID := uuid.New()
	role := domain.AdminRoleType
	permission1 := domain.UserCreate
	permission2 := domain.AuditRead

	// Assign role to user
	err = rbacService.AssignRole(userID, role)
	assert.NoError(t, err)

	// Grant only one permission
	err = rbacService.GrantPermission(role, permission1)
	assert.NoError(t, err)

	// Test checking any permission (should return true for permission1)
	hasAnyPermission, err := rbacService.CheckAnyPermission(userID, []domain.Permission{permission1, permission2})
	assert.NoError(t, err)
	assert.True(t, hasAnyPermission)

	// Test checking any permission with only permission2 (should return false)
	hasAnyPermission, err = rbacService.CheckAnyPermission(userID, []domain.Permission{permission2})
	assert.NoError(t, err)
	assert.False(t, hasAnyPermission)
}

func TestRBACService_CheckAllPermissions(t *testing.T) {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate tables
	err = db.AutoMigrate(&domain.UserRole{}, &domain.RolePermission{})
	assert.NoError(t, err)

	// Create RBAC service
	rbacService := service.NewRBACService(db)

	// Test data
	userID := uuid.New()
	role := domain.AdminRoleType
	permission1 := domain.UserCreate
	permission2 := domain.AuditRead

	// Assign role to user
	err = rbacService.AssignRole(userID, role)
	assert.NoError(t, err)

	// Grant both permissions
	err = rbacService.GrantPermission(role, permission1)
	assert.NoError(t, err)

	err = rbacService.GrantPermission(role, permission2)
	assert.NoError(t, err)

	// Test checking all permissions (should return true)
	hasAllPermissions, err := rbacService.CheckAllPermissions(userID, []domain.Permission{permission1, permission2})
	assert.NoError(t, err)
	assert.True(t, hasAllPermissions)

	// Test checking all permissions with one missing (should return false)
	hasAllPermissions, err = rbacService.CheckAllPermissions(userID, []domain.Permission{permission1, domain.SystemManage})
	assert.NoError(t, err)
	assert.False(t, hasAllPermissions)
}
