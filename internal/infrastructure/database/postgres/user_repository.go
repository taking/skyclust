package postgres

import (
	"skyclust/internal/domain"
	"skyclust/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		logger.Errorf("Failed to create user: %v", err)
		return err
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by ID: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by username: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by email: %v", err)
		return nil, err
	}
	return &user, nil
}

// GetByOIDC retrieves a user by OIDC provider and subject
func (r *userRepository) GetByOIDC(provider, subject string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("oidc_provider = ? AND oidc_subject = ? AND deleted_at IS NULL", provider, subject).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		logger.Errorf("Failed to get user by OIDC: %v", err)
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(user *domain.User) error {
	if err := r.db.Save(user).Error; err != nil {
		logger.Errorf("Failed to update user: %v", err)
		return err
	}
	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(id uuid.UUID) error {
	if err := r.db.Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
		logger.Errorf("Failed to delete user: %v", err)
		return err
	}
	return nil
}

// List retrieves a list of users with pagination and filtering
func (r *userRepository) List(limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	// Build base query
	query := r.db.Model(&domain.User{}).Where("deleted_at IS NULL")

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

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute query
	if err := query.Find(&users).Error; err != nil {
		logger.Errorf("Failed to list users: %v", err)
		return nil, 0, err
	}

	return users, total, nil
}

// Count returns the total number of users
func (r *userRepository) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		logger.Errorf("Failed to count users: %v", err)
		return 0, err
	}
	return count, nil
}
