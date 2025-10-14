package usecase

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"
)

// UserService implements the UserService interface
type UserService struct {
	userRepo     domain.UserRepository
	hasher       security.PasswordHasher
	auditLogRepo domain.AuditLogRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo domain.UserRepository, hasher security.PasswordHasher, auditLogRepo domain.AuditLogRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		hasher:       hasher,
		auditLogRepo: auditLogRepo,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Check username availability
	existingUser, err = s.userRepo.GetByUsername(req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info(fmt.Sprintf("User created successfully: %s (%s)", user.ID, user.Username))
	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GetUsers retrieves a list of users with pagination and filtering
func (s *UserService) GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	users, total, err := s.userRepo.List(limit, offset, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	logger.Info(fmt.Sprintf("Retrieved %d users (limit: %d, offset: %d)", len(users), limit, offset))
	return users, total, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing user
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Update fields
	if req.Username != nil {
		// Check if username is available
		existingUser, err := s.userRepo.GetByUsername(*req.Username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if email is available
		existingUser, err := s.userRepo.GetByEmail(*req.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = *req.Email
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info(fmt.Sprintf("User updated successfully: %s", user.ID))
	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists
	userID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := s.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info(fmt.Sprintf("User deleted successfully: %s", id))
	return nil
}

// Authenticate authenticates a user
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Verify password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	logger.Info(fmt.Sprintf("User authenticated successfully: %s (%s)", user.ID, user.Email))
	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get user
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Verify old password
	if !s.hasher.VerifyPassword(oldPassword, user.PasswordHash) {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info(fmt.Sprintf("Password changed successfully: %s", userID))
	return nil
}

// GetUserByID retrieves a user by ID (admin method)
func (s *UserService) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUsersWithFilters retrieves users with filters (admin method)
func (s *UserService) GetUsersWithFilters(filters domain.UserFilters) ([]*domain.User, int64, error) {
	// Convert filters to map for repository
	filterMap := make(map[string]interface{})
	if filters.Search != "" {
		filterMap["search"] = filters.Search
	}
	if filters.Role != "" {
		filterMap["role"] = filters.Role
	}
	if filters.Status != "" {
		filterMap["status"] = filters.Status
	}

	offset := (filters.Page - 1) * filters.Limit
	users, total, err := s.userRepo.List(filters.Limit, offset, filterMap)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users: %w", err)
	}

	return users, total, nil
}

// UpdateUserDirect updates a user (admin method)
func (s *UserService) UpdateUserDirect(user *domain.User) (*domain.User, error) {
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info(fmt.Sprintf("User updated successfully: %s", user.ID))
	return user, nil
}

// DeleteUserByID deletes a user (admin method)
func (s *UserService) DeleteUserByID(id uuid.UUID) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info(fmt.Sprintf("User deleted successfully: %s", id))
	return nil
}

// GetUserStats retrieves user statistics (admin method)
func (s *UserService) GetUserStats() (*domain.UserStats, error) {
	// Get all users to calculate stats
	users, total, err := s.userRepo.List(1000, 0, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Calculate statistics
	activeUsers := int64(0)
	inactiveUsers := int64(0)
	newUsersToday := int64(0)

	today := time.Now().Truncate(24 * time.Hour)

	for _, user := range users {
		if user.IsActive {
			activeUsers++
		} else {
			inactiveUsers++
		}

		if user.CreatedAt.After(today) {
			newUsersToday++
		}
	}

	return &domain.UserStats{
		TotalUsers:    total,
		ActiveUsers:   activeUsers,
		InactiveUsers: inactiveUsers,
		NewUsersToday: newUsersToday,
	}, nil
}
