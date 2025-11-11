package user

import (
	"context"
	"fmt"
	"skyclust/internal/domain"
	"skyclust/internal/shared/logging"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"skyclust/pkg/logger"
	"skyclust/pkg/security"
)

// Service: 사용자 서비스 인터페이스 구현체
type Service struct {
	userRepo          domain.UserRepository
	userDomainService *domain.UserDomainService
	hasher            security.PasswordHasher
	auditLogRepo      domain.AuditLogRepository
	logger            *zap.Logger
}

// NewService: 새로운 사용자 서비스를 생성합니다
func NewService(userRepo domain.UserRepository, userDomainService *domain.UserDomainService, hasher security.PasswordHasher, auditLogRepo domain.AuditLogRepository) *Service {
	return &Service{
		userRepo:          userRepo,
		userDomainService: userDomainService,
		hasher:            hasher,
		auditLogRepo:      auditLogRepo,
		logger:            logger.DefaultLogger.GetLogger(),
	}
}

// CreateUser: 새로운 사용자를 생성합니다
func (s *Service) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	// Use domain service to validate business rules and create user entity
	user, err := s.userDomainService.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	// Hash password (application-level concern)
	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to hash password: %v", err), 500)
	}
	user.PasswordHash = hashedPassword

	// Persist to repository (application-level concern)
	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user",
			logging.WithOperation("create_user"),
			logging.WithError(err),
			logging.WithUserID(user.ID.String()),
		)
		// Check for unique constraint violation
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
			if strings.Contains(errStr, "email") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "email already exists", 409)
			}
			if strings.Contains(errStr, "username") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "username already exists", 409)
			}
		}
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to create user: %v", err), 500)
	}

	s.logger.Info("User created successfully",
		logging.WithOperation("create_user"),
		logging.WithUserID(user.ID.String()),
		logging.WithUsername(user.Username),
	)
	return user, nil
}

// GetUser: ID로 사용자를 조회합니다
func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid user ID format", 400)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	return user, nil
}

// GetUsers: 페이지네이션과 필터링을 포함한 사용자 목록을 조회합니다
func (s *Service) GetUsers(ctx context.Context, limit, offset int, filters map[string]interface{}) ([]*domain.User, int64, error) {
	users, total, err := s.userRepo.List(limit, offset, filters)
	if err != nil {
		return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get users: %v", err), 500)
	}

	s.logger.Debug("Users retrieved",
		logging.WithOperation("list_users"),
		logging.WithCount(len(users)),
		logging.WithLimit(limit),
		logging.WithOffset(offset),
	)
	return users, total, nil
}

// UpdateUser: 사용자 정보를 업데이트합니다
func (s *Service) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.User, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err // req.Validate() already returns domain.NewDomainError
	}

	// Get existing user
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeBadRequest, "invalid user ID format", 400)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	// Update fields
	if req.Username != nil {
		// Check if username is available
		existingUser, err := s.userRepo.GetByUsername(*req.Username)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "username already exists", 409)
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		// Check if email is available
		existingUser, err := s.userRepo.GetByEmail(*req.Email)
		if err == nil && existingUser != nil && existingUser.ID != userID {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "email already exists", 409)
		}
		user.Email = *req.Email
	}

	if req.IsActive != nil {
		user.Active = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user",
			logging.WithOperation("update_user"),
			logging.WithError(err),
			logging.WithUserID(user.ID.String()),
		)
		// Check for unique constraint violation
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") {
			if strings.Contains(errStr, "email") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "email already exists", 409)
			}
			if strings.Contains(errStr, "username") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "username already exists", 409)
			}
		}
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update user: %v", err), 500)
	}

	s.logger.Info("User updated successfully",
		logging.WithOperation("update_user"),
		logging.WithUserID(user.ID.String()),
	)
	return user, nil
}

// DeleteUser: 사용자를 삭제합니다
func (s *Service) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists
	userID, err := uuid.Parse(id)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid user ID format", 400)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	if err := s.userRepo.Delete(userID); err != nil {
		s.logger.Error("Failed to delete user",
			logging.WithOperation("delete_user"),
			logging.WithError(err),
			logging.WithUserID(id),
		)
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete user: %v", err), 500)
	}

	s.logger.Info("User deleted successfully",
		logging.WithOperation("delete_user"),
		logging.WithUserID(id),
	)
	return nil
}

// Authenticate: 사용자를 인증합니다
func (s *Service) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if !user.IsActive() {
		return nil, domain.NewDomainError(domain.ErrCodeForbidden, "user account is disabled", 403)
	}

	// Verify password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	s.logger.Info("User authenticated successfully",
		logging.WithOperation("authenticate_user"),
		logging.WithUserID(user.ID.String()),
		logging.WithEmail(user.Email),
	)
	return user, nil
}

// ChangePassword: 사용자 비밀번호를 변경합니다
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	// Get user
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeBadRequest, "invalid user ID format", 400)
	}

	user, err := s.userRepo.GetByID(userUUID)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	// Verify old password
	if !s.hasher.VerifyPassword(oldPassword, user.PasswordHash) {
		return domain.NewDomainError(domain.ErrCodeInvalidCredentials, "invalid old password", 401)
	}

	// Hash new password
	hashedPassword, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to hash new password: %v", err), 500)
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update password",
			logging.WithOperation("change_password"),
			logging.WithError(err),
			logging.WithUserID(userID),
		)
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update password: %v", err), 500)
	}

	s.logger.Info("Password changed successfully",
		logging.WithOperation("change_password"),
		logging.WithUserID(userID),
	)
	return nil
}

// GetUserByID: ID로 사용자를 조회합니다 (관리자 메서드)
func (s *Service) GetUserByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}
	return user, nil
}

// GetUsersWithFilters: 필터를 사용하여 사용자 목록을 조회합니다 (관리자 메서드)
func (s *Service) GetUsersWithFilters(filters domain.UserFilters) ([]*domain.User, int64, error) {
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
		return nil, 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get users: %v", err), 500)
	}

	return users, total, nil
}

// UpdateUserDirect: 사용자 정보를 직접 업데이트합니다 (관리자 메서드)
// 주의: 비밀번호는 이 메서드를 호출하기 전에 이미 해시되어 있어야 합니다
func (s *Service) UpdateUserDirect(user *domain.User) (*domain.User, error) {
	// Check if username is being changed and if it conflicts with existing user
	if user.Username != "" {
		existingUser, err := s.userRepo.GetByUsername(user.Username)
		if err == nil && existingUser != nil && existingUser.ID != user.ID {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "username already exists", 409)
		}
	}

	// Check if email is being changed and if it conflicts with existing user
	if user.Email != "" {
		existingUser, err := s.userRepo.GetByEmail(user.Email)
		if err == nil && existingUser != nil && existingUser.ID != user.ID {
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "email already exists", 409)
		}
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user", zap.Error(err), zap.String("user_id", user.ID.String()))

		// Check for unique constraint violation
		errStr := err.Error()
		if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint") ||
			strings.Contains(errStr, "username") || strings.Contains(errStr, "email") {
			if strings.Contains(errStr, "username") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "username already exists", 409)
			}
			if strings.Contains(errStr, "email") {
				return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "email already exists", 409)
			}
			return nil, domain.NewDomainError(domain.ErrCodeAlreadyExists, "user information already exists", 409)
		}

		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to update user: %v", err), 500)
	}

	s.logger.Info("User updated successfully", zap.String("user_id", user.ID.String()))
	return user, nil
}

// HashPassword: 서비스의 해셔를 사용하여 비밀번호를 해시합니다
func (s *Service) HashPassword(password string) (string, error) {
	return s.hasher.HashPassword(password)
}

// DeleteUserByID: ID로 사용자를 삭제합니다 (관리자 메서드)
func (s *Service) DeleteUserByID(id uuid.UUID) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.ErrUserNotFound
		}
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user: %v", err), 500)
	}

	if err := s.userRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete user",
			logging.WithOperation("delete_user_by_id"),
			logging.WithError(err),
			logging.WithUserID(id.String()),
		)
		return domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to delete user: %v", err), 500)
	}

	s.logger.Info("User deleted successfully",
		logging.WithOperation("delete_user_by_id"),
		logging.WithUserID(id.String()),
	)
	return nil
}

// GetUserStats: 사용자 통계를 조회합니다 (관리자 메서드)
func (s *Service) GetUserStats() (*domain.UserStats, error) {
	// Get all users to calculate stats
	users, total, err := s.userRepo.List(1000, 0, map[string]interface{}{})
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get users: %v", err), 500)
	}

	// Calculate statistics
	activeUsers := int64(0)
	inactiveUsers := int64(0)
	newUsersToday := int64(0)

	today := time.Now().Truncate(24 * time.Hour)

	for _, user := range users {
		if user.IsActive() {
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

// GetUserCount: 전체 사용자 수를 반환합니다 (시스템 초기화 확인용)
func (s *Service) GetUserCount() (int64, error) {
	count, err := s.userRepo.Count()
	if err != nil {
		return 0, domain.NewDomainError(domain.ErrCodeInternalError, fmt.Sprintf("failed to get user count: %v", err), 500)
	}
	return count, nil
}
