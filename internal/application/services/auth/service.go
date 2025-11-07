package auth

import (
	"context"
	"skyclust/internal/application/services/common"
	"skyclust/internal/domain"
	"skyclust/pkg/cache"
	"skyclust/pkg/security"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// contextKey는 context.WithValue에서 사용할 커스텀 키 타입입니다
type contextKey string

const (
	contextKeyClientIP  contextKey = "client_ip"
	contextKeyUserAgent contextKey = "user_agent"
)

// Service: 인증 비즈니스 로직 구현체
type Service struct {
	userRepo     domain.UserRepository
	auditLogRepo domain.AuditLogRepository
	rbacService  domain.RBACService
	hasher       security.PasswordHasher
	blacklist    *cache.TokenBlacklist
	jwtSecret    string
	jwtExpiry    time.Duration
}

// NewService: 새로운 인증 서비스를 생성합니다
func NewService(
	userRepo domain.UserRepository,
	auditLogRepo domain.AuditLogRepository,
	rbacService domain.RBACService,
	hasher security.PasswordHasher,
	blacklist *cache.TokenBlacklist,
	jwtSecret string,
	jwtExpiry time.Duration,
) domain.AuthService {
	return &Service{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
		rbacService:  rbacService,
		hasher:       hasher,
		blacklist:    blacklist,
		jwtSecret:    jwtSecret,
		jwtExpiry:    jwtExpiry,
	}
}

// Register: 새로운 사용자 계정을 생성합니다
func (s *Service) Register(req domain.CreateUserRequest) (*domain.User, string, error) {
	// Check if email already exists (email is unique)
	if existing, _ := s.userRepo.GetByEmail(req.Email); existing != nil {
		return nil, "", domain.ErrUserAlreadyExists
	}

	// Check if this is the first user (make them admin) - MUST check BEFORE creating user
	userCount, err := s.userRepo.Count()
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to check user count", 500)
	}

	// Assign role based on user count
	var defaultRole domain.Role
	if userCount == 0 {
		// First user becomes admin
		defaultRole = domain.AdminRoleType
	} else {
		// Subsequent users get user role
		defaultRole = domain.UserRoleType
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
		Active:       true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to create user", 500)
	}

	if err := s.rbacService.AssignRole(user.ID, defaultRole); err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to assign role", 500)
	}

	// Get user roles for JWT token
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate JWT token with primary role (first role)
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	token, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate token", 500)
	}

	// Log registration
	ctx := context.Background()
	common.LogAction(ctx, s.auditLogRepo, &user.ID, domain.ActionUserRegister,
		"POST /api/v1/auth/register",
		map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		},
	)

	return user, token, nil
}

// Login: 사용자를 인증하고 JWT 토큰을 반환합니다
func (s *Service) Login(email, password string) (*domain.User, string, error) {
	// Get user by email (email is unique)
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user", 500)
	}
	if user == nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, "", domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	// Validate password
	if !s.hasher.VerifyPassword(password, user.PasswordHash) {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Get user roles for JWT token
	userRoles, err := s.rbacService.GetUserRoles(user.ID)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to get user roles", 500)
	}

	// Generate JWT token with primary role (first role)
	var primaryRole domain.Role
	if len(userRoles) > 0 {
		primaryRole = userRoles[0]
	} else {
		primaryRole = domain.UserRoleType
	}

	token, err := s.generateJWT(user.ID, user.Username, primaryRole)
	if err != nil {
		return nil, "", domain.NewDomainError(domain.ErrCodeInternalError, "failed to generate token", 500)
	}

	// Note: Audit log is now created in LoginWithContext method

	return user, token, nil
}

// LoginWithContext: 클라이언트 컨텍스트 정보를 포함하여 로그인을 수행합니다
func (s *Service) LoginWithContext(email, password, clientIP, userAgent string) (*domain.User, string, error) {
	// Use the existing Login method for authentication
	user, token, err := s.Login(email, password)
	if err != nil {
		return nil, "", err
	}

	// Create audit log with client context using common helper
	ctx := context.Background()
	// Set IP and UserAgent in context for extraction
	if clientIP == "" {
		clientIP = "127.0.0.1" // Default to localhost if no IP
	}
	ctx = context.WithValue(ctx, contextKeyClientIP, clientIP)
	ctx = context.WithValue(ctx, contextKeyUserAgent, userAgent)

	common.LogActionWithContext(ctx, s.auditLogRepo, &user.ID, domain.ActionUserLogin,
		"POST /api/v1/auth/login",
		map[string]interface{}{
			"email": user.Email,
		},
		clientIP,
		userAgent,
	)

	return user, token, nil
}

// ValidateToken: JWT 토큰을 검증하고 사용자 정보를 반환합니다
func (s *Service) ValidateToken(tokenString string) (*domain.User, error) {
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

	if !user.IsActive() {
		return nil, domain.NewDomainError(domain.ErrCodeUnauthorized, "account is deactivated", 401)
	}

	return user, nil
}

// Logout: 사용자를 로그아웃하고 토큰을 무효화합니다
func (s *Service) Logout(userID uuid.UUID, token string) error {
	// Add token to blacklist
	ctx := context.Background()
	expiry := BlacklistTokenExpiry
	if err := s.blacklist.AddToBlacklist(ctx, token, expiry); err != nil {
		return domain.NewDomainError(domain.ErrCodeInternalError, "failed to invalidate token", 500)
	}

	// Log logout using common helper
	common.LogAction(ctx, s.auditLogRepo, &userID, domain.ActionUserLogout,
		"POST /api/v1/auth/logout",
		map[string]interface{}{
			"token_invalidated": true,
		},
	)

	return nil
}

// generateJWT generates a JWT token for a user
func (s *Service) generateJWT(userID uuid.UUID, username string, role domain.Role) (string, error) {
	now := time.Now()
	// Use default expiry if jwtExpiry is 0
	expiry := s.jwtExpiry
	if expiry == 0 {
		expiry = DefaultTokenExpiry
	}
	exp := now.Add(expiry)

	claims := jwt.MapClaims{
		"user_id":  userID.String(),
		"username": username,
		"role":     string(role),
		"exp":      exp.Unix(),
		"iat":      now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
