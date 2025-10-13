package usecase

import (
	"context"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/security"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// authService implements the authentication business logic
type authService struct {
	userRepo     domain.UserRepository
	auditLogRepo domain.AuditLogRepository
	hasher       security.PasswordHasher
	blacklist    *cache.TokenBlacklist
	jwtSecret    string
	jwtExpiry    time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
	hasher security.PasswordHasher,
	blacklist *cache.TokenBlacklist,
	jwtSecret string,
	jwtExpiry time.Duration,
) domain.AuthService {
	return &authService{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
		hasher:       hasher,
		blacklist:    blacklist,
		jwtSecret:    jwtSecret,
		jwtExpiry:    jwtExpiry,
	}
}

// Register creates a new user account
func (s *authService) Register(req domain.CreateUserRequest) (*domain.User, string, error) {
	// Check if username already exists
	if existing, _ := s.userRepo.GetByUsername(req.Username); existing != nil {
		return nil, "", domain.ErrUserAlreadyExists
	}

	// Check if email already exists
	if existing, _ := s.userRepo.GetByEmail(req.Email); existing != nil {
		return nil, "", domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to hash password", 500)
	}

	// Create user
	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to create user", 500)
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate token", 500)
	}

	// Log registration
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   user.ID,
		Action:   domain.ActionUserRegister,
		Resource: "POST /api/v1/auth/register",
		Details: map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		},
	})

	return user, token, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(username, password string) (*domain.User, string, error) {
	// Get user by username or email
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		// Try email
		user, err = s.userRepo.GetByEmail(username)
		if err != nil {
			return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
		}
		if user == nil {
			return nil, "", domain.ErrInvalidCredentials
		}
	}

	// Check if user is active
	if !user.IsActive {
		return nil, "", domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	// Validate password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID, user.Username)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate token", 500)
	}

	// Log login
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   user.ID,
		Action:   domain.ActionUserLogin,
		Resource: "POST /api/v1/auth/login",
		Details: map[string]interface{}{
			"username": user.Username,
		},
	})

	return user, token, nil
}

// ValidateToken validates a JWT token and returns user information
func (s *authService) ValidateToken(tokenString string) (*domain.User, error) {
	// Check if token is blacklisted
	if s.blacklist != nil && s.blacklist.IsBlacklisted(context.Background(), tokenString) {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "token has been revoked", 401)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
	}

	if !token.Valid {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token", 401)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token claims", 401)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid token claims", 401)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "invalid user ID", 401)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "user not found", 401)
	}

	if !user.IsActive {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	return user, nil
}

// Logout logs out a user and invalidates their token
func (s *authService) Logout(userID uuid.UUID, token string) error {
	// Add token to blacklist
	expiry := 24 * time.Hour // Set expiry based on your JWT expiry
	if err := s.blacklist.AddToBlacklist(context.Background(), token, expiry); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to invalidate token", 500)
	}

	// Log logout
	_ = s.auditLogRepo.Create(&domain.AuditLog{
		UserID:   userID,
		Action:   domain.ActionUserLogout,
		Resource: "POST /api/v1/auth/logout",
		Details: map[string]interface{}{
			"token_invalidated": true,
		},
	})

	return nil
}

// generateJWT generates a JWT token for a user
func (s *authService) generateJWT(userID uuid.UUID, username string) (string, error) {
	now := time.Now()
	// Use 24 hours expiry if jwtExpiry is 0
	expiry := s.jwtExpiry
	if expiry == 0 {
		expiry = 24 * time.Hour
	}
	exp := now.Add(expiry)

	claims := jwt.MapClaims{
		"user_id":  userID.String(),
		"username": username,
		"exp":      exp.Unix(),
		"iat":      now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
