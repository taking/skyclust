package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository: domain.UserRepository 인터페이스 구현체
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository: 새로운 UserRepository를 생성합니다
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create: 새로운 사용자를 생성합니다
func (r *userRepository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return err
	}
	return nil
}

// GetByID: ID로 사용자를 조회합니다
func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByUsername: 사용자명으로 사용자를 조회합니다
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by username: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByEmail: 이메일로 사용자를 조회합니다
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByOIDC: OIDC 프로바이더와 subject로 사용자를 조회합니다
func (r *userRepository) GetByOIDC(provider, subject string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("oidc_provider = ? AND oidc_subject = ?", provider, subject).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by OIDC: %v", err)
		return nil, err
	}
	return &user, nil
}

// Update: 사용자 정보를 업데이트합니다
func (r *userRepository) Update(user *domain.User) error {
	if err := r.db.Save(user).Error; err != nil {
		logger.Errorf("Failed to update user: %v", err)
		return err
	}
	return nil
}

// Delete: 사용자를 영구 삭제합니다
func (r *userRepository) Delete(id uuid.UUID) error {
	if err := r.db.Unscoped().Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
		logger.Errorf("Failed to delete user: %v", err)
		return err
	}
	return nil
}

// List: 페이지네이션과 필터링을 포함한 사용자 목록을 조회합니다
func (r *userRepository) List(limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	// Build base query
	query := r.db.Model(&domain.User{})

	// Apply filters
	if search, ok := filters["search"]; ok && search != "" {
		searchStr := "%" + search.(string) + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ?", searchStr, searchStr)
	}

	if isActive, ok := filters["is_active"]; ok {
		query = query.Where("is_active = ?", isActive)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		logger.Errorf("Failed to count users: %v", err)
		return nil, 0, err
	}

	// Apply pagination and ordering
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Order by created_at DESC (newest first)
	query = query.Order("created_at DESC")

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		logger.Errorf("Failed to list users: %v", err)
		return nil, 0, err
	}

	return users, total, nil
}

// Count: 전체 사용자 수를 반환합니다
func (r *userRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Count(&count).Error; err != nil {
		logger.Errorf("Failed to count users: %v", err)
		return 0, err
	}
	return count, nil
}
