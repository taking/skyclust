package usecase

import (
	"context"
	"fmt"
	"time"

	"cmp/internal/domain"
	"cmp/pkg/shared/logger"
	"cmp/pkg/shared/security"
)

// UserService implements the UserService interface
type UserService struct {
	userRepo domain.UserRepository
	hasher   security.PasswordHasher
}

// NewUserService creates a new UserService
func NewUserService(userRepo domain.UserRepository, hasher security.PasswordHasher) *UserService {
	return &UserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Check username availability
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
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
		ID:        generateID(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info(fmt.Sprintf("User created successfully: %s (%s)", user.ID, user.Username))
	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Update fields
	if req.Username != nil {
		// Check if username is available
		existingUser, err := s.userRepo.GetByUsername(ctx, *req.Username)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("username already exists")
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if email is available
		existingUser, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, fmt.Errorf("email already exists")
		}
		user.Email = *req.Email
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info(fmt.Sprintf("User updated successfully: %s", user.ID))
	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info(fmt.Sprintf("User deleted successfully: %s", id))
	return nil
}

// Authenticate authenticates a user
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
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
	if !s.hasher.VerifyPassword(password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	logger.Info(fmt.Sprintf("User authenticated successfully: %s (%s)", user.ID, user.Email))
	return user, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Verify old password
	if !s.hasher.VerifyPassword(oldPassword, user.Password) {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info(fmt.Sprintf("Password changed successfully: %s", userID))
	return nil
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
